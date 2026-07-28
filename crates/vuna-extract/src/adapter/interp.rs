//! The interpreter: an [`AdapterManifest`] wearing the [`Extractor`] seam, so a `.toml` file in
//! `adapters/` drives real extraction at runtime with no Rust change.
//!
//! Scope, stated plainly: an [`Extractor`] is handed a page `vuna-crawl` already fetched, so this
//! interpreter reads manifests, it does not *fetch* from them. `[fetch].endpoint` is validated to
//! be the page itself (see [`ManifestError::UnfetchableEndpoint`]); the frontier is what decides
//! which URLs get fetched, exactly as it does for the built-in extractors. Secondary lookups and
//! multi-offer aggregation remain unimplemented — see `adapters/README.md`.
//!
//! Hostile-input posture, since a manifest is applied to whatever a third-party server returned:
//! the declared `content_type` is used only to *rule the adapter out*, never to decide how the
//! body is parsed; array iteration is capped; and a body that does not parse is an error for this
//! adapter alone, never for the page (`ExtractorRegistry::extract_all` runs the others regardless).

use std::collections::BTreeMap;
use std::path::Path;

use scraper::{Html, Selector};
use serde_json::Value;
use url::Url;
use vuna_core::extract::{
    Availability, Extraction, Extractor, ExtractorKind, FetchedPage, RetailObservation,
};
use vuna_core::{Error, Result};

use super::manifest::{AdapterManifest, BodyFormat, Expr, IterateSpec, ManifestError};
use super::pattern::UrlPattern;
use super::price;

/// How far into the body `applies()` looks for `[match].requires` markers. JSON-LD and product
/// meta tags live in `<head>`, so a bounded prefix is enough to pre-filter and keeps the sniff
/// cheap on a large page.
const SNIFF_WINDOW: usize = 65_536;

/// Ceiling on observations from one page. A hostile (or merely broken) endpoint returning a
/// million-element array must not turn one fetch into a million index writes.
pub const MAX_OBSERVATIONS: usize = 1_024;

/// A manifest compiled into a runnable extractor.
#[derive(Debug)]
pub struct AdapterExtractor {
    manifest: AdapterManifest,
    pattern: Option<UrlPattern>,
}

impl AdapterExtractor {
    /// Compiles a validated manifest.
    pub fn new(manifest: AdapterManifest) -> std::result::Result<Self, ManifestError> {
        let pattern = manifest.compiled_pattern()?;
        Ok(Self { manifest, pattern })
    }

    /// Parses, validates, and compiles a manifest from TOML source.
    pub fn from_toml(src: &str) -> std::result::Result<Self, ManifestError> {
        Self::new(AdapterManifest::parse(src)?)
    }

    /// Loads every `*.toml` in `dir`, in filename order.
    ///
    /// Deliberately **not** called by [`crate::ExtractorRegistry::with_defaults`]: adapters are
    /// opt-in data a node points at, and the crate compiles, tests, and runs with the directory
    /// absent. Nothing in the build depends on this directory existing.
    pub fn load_dir(dir: &Path) -> Result<Vec<Self>> {
        let mut paths: Vec<_> = std::fs::read_dir(dir)
            .map_err(|e| Error::Other(format!("adapters dir {}: {e}", dir.display())))?
            .filter_map(std::result::Result::ok)
            .map(|e| e.path())
            .filter(|p| p.extension().is_some_and(|x| x == "toml"))
            .collect();
        paths.sort();

        paths
            .into_iter()
            .map(|path| {
                let src = std::fs::read_to_string(&path)
                    .map_err(|e| Error::Other(format!("adapter {}: {e}", path.display())))?;
                let extractor = Self::from_toml(&src)
                    .map_err(|e| Error::Other(format!("adapter {}: {e}", path.display())))?;
                Ok(extractor)
            })
            .collect()
    }

    /// The manifest's `[adapter].name` — the identity a reviewer and an operator both use.
    pub fn name(&self) -> &str {
        &self.manifest.adapter.name
    }

    pub fn manifest(&self) -> &AdapterManifest {
        &self.manifest
    }
}

impl Extractor for AdapterExtractor {
    fn kind(&self) -> ExtractorKind {
        self.manifest.adapter.extractor_kind.clone()
    }

    fn applies(&self, page: &FetchedPage) -> bool {
        if let Some(pattern) = &self.pattern {
            if pattern.matches(&page.url).is_none() {
                return false;
            }
        }
        if let Some(expected) = &self.manifest.match_rules.content_type {
            if !content_type_ok(page, expected, self.manifest.fetch.format) {
                return false;
            }
        }
        if !self.manifest.match_rules.requires.is_empty() {
            let head = &page.body[..page.body.len().min(SNIFF_WINDOW)];
            let head = String::from_utf8_lossy(head);
            if !self.manifest.match_rules.requires.iter().all(|m| head.contains(m.as_str())) {
                return false;
            }
        }
        true
    }

    fn extract(&self, page: &FetchedPage) -> Result<Extraction> {
        let fail = |reason: String| Error::Extract {
            kind: format!("{}:{}", self.kind(), self.name()),
            url: page.url.clone(),
            reason,
        };

        let mut captures = self.pattern.as_ref().and_then(|p| p.matches(&page.url)).unwrap_or_default();
        captures.insert("page_url".to_string(), page.url.clone());
        if let Some(host) = Url::parse(&page.url).ok().and_then(|u| u.host_str().map(str::to_string)) {
            captures.insert("host".to_string(), host);
        }

        let root = match self.manifest.fetch.format {
            BodyFormat::Json => serde_json::from_slice::<Value>(&page.body)
                .map_err(|e| fail(format!("body is not JSON: {e}")))?,
            BodyFormat::Html => {
                let want = self.manifest.embedded_jsonld_type().ok_or_else(|| fail("no embedded selector".into()))?;
                embedded_jsonld(&page.body, want)
                    .ok_or_else(|| fail(format!("no embedded JSON-LD node with @type {want:?}")))?
            }
        };

        let nodes = select_nodes(&root, self.manifest.iterate.as_ref());
        let observations = nodes
            .into_iter()
            .take(MAX_OBSERVATIONS)
            .filter_map(|node| self.observation(node, &captures, page.fetched_at))
            .collect();

        Ok(Extraction::Retail(observations))
    }
}

impl AdapterExtractor {
    /// Resolves one node into an observation, or `None` when the node carries no usable identity.
    ///
    /// An observation with no SKU cannot be reconciled with anyone else's — `(store, sku)` is the
    /// join key `quorum::reconcile` groups on — so a node whose `sku` and `sku_fallback` both come
    /// up empty is dropped rather than keyed on something invented.
    fn observation(
        &self,
        node: &Value,
        captures: &BTreeMap<String, String>,
        observed_at: u64,
    ) -> Option<RetailObservation> {
        let fields = &self.manifest.fields;

        let store = resolve_string(&fields.store, node, captures)?;
        let sku = resolve_string(&fields.sku, node, captures)
            .or_else(|| fields.sku_fallback.as_ref().and_then(|e| resolve_string(e, node, captures)))?;

        let availability = fields
            .availability
            .as_ref_value(node, captures)
            .and_then(|v| availability_key(&v))
            .and_then(|key| self.manifest.availability_map.get(&key).copied())
            .unwrap_or(Availability::Unknown);

        let quantity = fields
            .quantity
            .as_ref()
            .and_then(|e| e.as_ref_value(node, captures))
            .and_then(|v| v.as_u64());

        let currency = fields.currency.as_ref().and_then(|e| resolve_string(e, node, captures));

        let price_minor = match (&fields.price_minor, fields.price_representation) {
            (Some(expr), Some(repr)) => {
                let stated = fields
                    .price_exponent
                    .as_ref()
                    .and_then(|e| e.as_ref_value(node, captures))
                    .and_then(|v| json_as_u32(&v));
                expr.as_ref_value(node, captures).and_then(|raw| {
                    price::to_minor_units(
                        &raw,
                        repr,
                        currency.as_deref(),
                        &self.manifest.currency_exponents,
                        stated,
                    )
                })
            }
            _ => None,
        };

        Some(RetailObservation {
            store,
            sku,
            availability,
            quantity,
            price_minor,
            currency,
            method: self.manifest.adapter.method,
            observed_at,
        })
    }
}

impl Expr {
    /// Resolves the expression to a JSON value. `const:none` and an absent/null JSON path both
    /// yield `None` — "the source did not say", which every caller here propagates rather than
    /// substituting a default.
    fn as_ref_value(&self, node: &Value, captures: &BTreeMap<String, String>) -> Option<Value> {
        match self {
            Expr::None => None,
            Expr::Const(v) => Some(Value::String(v.clone())),
            Expr::UrlCapture(name) => captures.get(name).map(|s| Value::String(s.clone())),
            Expr::JsonPath(path) => {
                let mut cur = node;
                for segment in path {
                    cur = cur.get(segment)?;
                }
                (!cur.is_null()).then(|| cur.clone())
            }
        }
    }
}

/// A resolved value as a non-empty string. An empty string is treated as absent so a merchant's
/// blank `sku` field falls through to `sku_fallback` instead of becoming an empty join key.
fn resolve_string(expr: &Expr, node: &Value, captures: &BTreeMap<String, String>) -> Option<String> {
    let text = match expr.as_ref_value(node, captures)? {
        Value::String(s) => s,
        Value::Number(n) => n.to_string(),
        Value::Bool(b) => b.to_string(),
        _ => return None,
    };
    (!text.trim().is_empty()).then_some(text)
}

/// The `[availability_map]` lookup key for a raw source value. A JSON boolean becomes `"true"` /
/// `"false"` so a manifest can map it with the same table shape as a string vocabulary.
fn availability_key(value: &Value) -> Option<String> {
    match value {
        Value::String(s) => Some(s.clone()),
        Value::Bool(b) => Some(b.to_string()),
        Value::Number(n) => Some(n.to_string()),
        _ => None,
    }
}

fn json_as_u32(value: &Value) -> Option<u32> {
    value
        .as_u64()
        .or_else(|| value.as_str().and_then(|s| s.trim().parse::<u64>().ok()))
        .and_then(|v| u32::try_from(v).ok())
}

/// The nodes `[fields]` expressions are resolved against: the document itself, or each element of
/// the array `[iterate].path` names.
fn select_nodes<'a>(root: &'a Value, iterate: Option<&IterateSpec>) -> Vec<&'a Value> {
    let Some(spec) = iterate else { return vec![root] };

    let path = spec.path.trim_end_matches("[]");
    let target = match path {
        "" | "$" => Some(root),
        p => p.trim_start_matches("$.").split('.').try_fold(root, |cur, seg| cur.get(seg)),
    };

    match target {
        Some(Value::Array(items)) => items.iter().collect(),
        // A path that resolved to a single object is treated as a one-element list: an endpoint
        // that returns `{...}` where it usually returns `[{...}]` is a real shape, not an error.
        Some(other) if !other.is_null() => vec![other],
        _ => Vec::new(),
    }
}

/// Finds the first embedded JSON-LD node whose `@type` is (or includes) `want`.
///
/// Reuses the same walk the built-in `retail` extractor uses, so `@graph` wrappers, arrays of
/// nodes, and multiple `<script>` blocks behave identically whether a page is read by Rust or by a
/// manifest. Blocks that are not valid JSON are skipped, not fatal — malformed JSON-LD alongside a
/// good block is common in the wild.
fn embedded_jsonld(body: &[u8], want: &str) -> Option<Value> {
    let html = String::from_utf8_lossy(body);
    let document = Html::parse_document(&html);
    let selector = Selector::parse(r#"script[type="application/ld+json"]"#).expect("static selector");

    for script in document.select(&selector) {
        let raw: String = script.text().collect();
        let Ok(value) = serde_json::from_str::<Value>(raw.trim()) else { continue };
        let mut found = Vec::new();
        crate::retail::collect_typed(&value, want, &mut found);
        if let Some(node) = found.first() {
            return Some((*node).clone());
        }
    }
    None
}

/// Whether the page's type is consistent with what the manifest expects.
///
/// Two independent checks, and both must pass when the server declared a type: the declared type
/// must contain the expected token, **and** the body must actually start like that format. A
/// hostile page declaring `application/json` while serving HTML fails the second; a
/// misconfigured-but-honest server that declares nothing is judged on its body alone.
fn content_type_ok(page: &FetchedPage, expected: &str, format: BodyFormat) -> bool {
    let expected = expected.split(';').next().unwrap_or(expected).trim().to_ascii_lowercase();
    if let Some(declared) = &page.content_type {
        let declared = declared.split(';').next().unwrap_or(declared).trim().to_ascii_lowercase();
        if declared != expected {
            return false;
        }
    }
    match format {
        BodyFormat::Json => matches!(first_non_space(&page.body), Some(b'{') | Some(b'[')),
        BodyFormat::Html => crate::looks_like_html(page),
    }
}

fn first_non_space(body: &[u8]) -> Option<u8> {
    body.iter().copied().find(|b| !b.is_ascii_whitespace())
}

#[cfg(test)]
mod tests {
    use super::*;
    use vuna_core::extract::RetailMethod;

    const SHOPIFY_LIKE: &str = r#"
[adapter]
name = "t-shopify"
version = 1
extractor_kind = "retail"
method = "json_endpoint"

[match]
url_pattern = "https://{store}/products/{handle}.js"
content_type = "application/json"

[fetch]
endpoint = "https://{store}/products/{handle}.js"
format = "json"

[iterate]
path = "variants[]"

[fields]
store = "url:{store}"
sku = "json:$.sku"
sku_fallback = "json:$.id"
availability = "json:$.available"
quantity = "json:$.inventory_quantity"
price_minor = "json:$.price"
price_representation = "minor_int"
currency = "const:none"

[availability_map]
"true" = "in_stock"
"false" = "out_of_stock"
"#;

    fn page(url: &str, content_type: Option<&str>, body: &str) -> FetchedPage {
        FetchedPage {
            url: url.to_string(),
            status: 200,
            content_type: content_type.map(str::to_string),
            body: body.as_bytes().to_vec(),
            fetched_at: 1_700_000_000,
        }
    }

    fn observations(extractor: &AdapterExtractor, page: &FetchedPage) -> Vec<RetailObservation> {
        match extractor.extract(page).unwrap() {
            Extraction::Retail(obs) => obs,
            other => panic!("expected Extraction::Retail, got {other:?}"),
        }
    }

    /// The whole point of the module: a TOML file, not a Rust impl, produced these observations.
    #[test]
    fn a_manifest_drives_a_real_extraction() {
        let extractor = AdapterExtractor::from_toml(SHOPIFY_LIKE).unwrap();
        let body = r#"{"id":1,"variants":[
            {"id":11,"sku":"W-S","available":true,"inventory_quantity":7,"price":1999},
            {"id":12,"sku":"","available":false,"price":2499}
        ]}"#;
        let p = page("https://shop.example.com/products/widget.js", Some("application/json"), body);

        assert!(extractor.applies(&p));
        let obs = observations(&extractor, &p);
        assert_eq!(obs.len(), 2, "one observation per variant");

        assert_eq!(obs[0].store, "shop.example.com");
        assert_eq!(obs[0].sku, "W-S");
        assert_eq!(obs[0].availability, Availability::InStock);
        assert_eq!(obs[0].quantity, Some(7));
        assert_eq!(obs[0].price_minor, Some(1999));
        assert_eq!(obs[0].currency, None, "const:none must stay unresolved, not become a string");
        assert_eq!(obs[0].method, RetailMethod::JsonEndpoint);
        assert_eq!(obs[0].observed_at, 1_700_000_000);

        // Blank sku falls through to sku_fallback; a hidden inventory count stays unknown.
        assert_eq!(obs[1].sku, "12");
        assert_eq!(obs[1].availability, Availability::OutOfStock);
        assert_eq!(obs[1].quantity, None);
    }

    #[test]
    fn applies_is_narrowed_by_url_pattern_and_content_type() {
        let extractor = AdapterExtractor::from_toml(SHOPIFY_LIKE).unwrap();
        let body = r#"{"variants":[]}"#;
        assert!(extractor.applies(&page("https://shop.example.com/products/w.js", Some("application/json"), body)));
        // Wrong path.
        assert!(!extractor.applies(&page("https://shop.example.com/products/w", Some("application/json"), body)));
        // Right path, wrong declared type.
        assert!(!extractor.applies(&page("https://shop.example.com/products/w.js", Some("text/html"), body)));
    }

    /// The declared content type can only rule an adapter out. A page claiming `application/json`
    /// while serving HTML is not applied to, and is never parsed as JSON on the header's word.
    #[test]
    fn a_lying_content_type_does_not_get_the_adapter_applied() {
        let extractor = AdapterExtractor::from_toml(SHOPIFY_LIKE).unwrap();
        let p = page(
            "https://shop.example.com/products/w.js",
            Some("application/json"),
            "<html><body>definitely not json</body></html>",
        );
        assert!(!extractor.applies(&p));
    }

    #[test]
    fn a_body_that_is_not_json_is_this_adapters_error_only() {
        let extractor = AdapterExtractor::from_toml(SHOPIFY_LIKE).unwrap();
        let p = page("https://shop.example.com/products/w.js", Some("application/json"), "{not json");
        let err = extractor.extract(&p).unwrap_err();
        assert!(matches!(err, Error::Extract { .. }), "got {err:?}");
    }

    #[test]
    fn unmapped_availability_values_resolve_to_unknown_never_in_stock() {
        let extractor = AdapterExtractor::from_toml(SHOPIFY_LIKE).unwrap();
        let body = r#"{"variants":[{"id":1,"sku":"A","available":"maybe","price":100}]}"#;
        let p = page("https://shop.example.com/products/w.js", Some("application/json"), body);
        assert_eq!(observations(&extractor, &p)[0].availability, Availability::Unknown);
    }

    #[test]
    fn a_node_with_no_sku_at_all_is_dropped_not_keyed_on_something_invented() {
        let extractor = AdapterExtractor::from_toml(SHOPIFY_LIKE).unwrap();
        let body = r#"{"variants":[{"available":true,"price":100},{"id":9,"sku":"B","available":true,"price":100}]}"#;
        let p = page("https://shop.example.com/products/w.js", Some("application/json"), body);
        let obs = observations(&extractor, &p);
        assert_eq!(obs.len(), 1);
        assert_eq!(obs[0].sku, "B");
    }

    /// A hostile endpoint must not turn one fetch into unbounded index writes.
    #[test]
    fn observation_count_is_capped() {
        let extractor = AdapterExtractor::from_toml(SHOPIFY_LIKE).unwrap();
        let variants: Vec<String> = (0..MAX_OBSERVATIONS + 500)
            .map(|i| format!(r#"{{"id":{i},"sku":"S{i}","available":true,"price":1}}"#))
            .collect();
        let body = format!(r#"{{"variants":[{}]}}"#, variants.join(","));
        let p = page("https://shop.example.com/products/w.js", Some("application/json"), &body);
        assert_eq!(observations(&extractor, &p).len(), MAX_OBSERVATIONS);
    }

    #[test]
    fn iterate_over_the_document_root_array() {
        let src = SHOPIFY_LIKE.replace(r#"path = "variants[]""#, r#"path = "$[]""#);
        let extractor = AdapterExtractor::from_toml(&src).unwrap();
        let body = r#"[{"id":1,"sku":"A","available":true,"price":500}]"#;
        let p = page("https://shop.example.com/products/w.js", Some("application/json"), body);
        let obs = observations(&extractor, &p);
        assert_eq!(obs.len(), 1);
        assert_eq!(obs[0].price_minor, Some(500));
    }

    #[test]
    fn missing_iterate_target_yields_no_observations_not_an_error() {
        let extractor = AdapterExtractor::from_toml(SHOPIFY_LIKE).unwrap();
        let p = page("https://shop.example.com/products/w.js", Some("application/json"), r#"{"id":1}"#);
        assert!(observations(&extractor, &p).is_empty());
    }

    #[test]
    fn embedded_jsonld_selection_walks_graph_wrappers() {
        let html = r#"<html><head>
            <script type="application/ld+json">{ broken </script>
            <script type="application/ld+json">
            {"@graph":[{"@type":"BreadcrumbList"},{"@type":"Product","sku":"G-1"}]}
            </script>
        </head><body></body></html>"#;
        let node = embedded_jsonld(html.as_bytes(), "Product").unwrap();
        assert_eq!(node.get("sku").and_then(Value::as_str), Some("G-1"));
        assert!(embedded_jsonld(html.as_bytes(), "Recipe").is_none());
    }

    #[test]
    fn kind_comes_from_the_manifest() {
        let extractor = AdapterExtractor::from_toml(SHOPIFY_LIKE).unwrap();
        assert_eq!(extractor.kind(), "retail");
        assert_eq!(extractor.name(), "t-shopify");
    }
}
