//! The `adapters/*.toml` manifest: its typed shape, and the validation that runs when one is
//! loaded.
//!
//! Validation is deliberately strict and up-front. A manifest is contributor-supplied data whose
//! whole promise is "a new storefront is a reviewed data change, not a Rust release" — that
//! promise only holds if a manifest that *cannot* work fails at load with a reason, instead of
//! quietly matching nothing (or worse, matching and emitting wrong observations) at runtime.
//! Unknown keys are rejected for the same reason: a misspelled `price_represention` that serde
//! silently ignored would be indistinguishable from a correct manifest until the prices came out
//! wrong.

use std::collections::BTreeMap;

use serde::Deserialize;
use vuna_core::extract::{Availability, RetailMethod};

use super::pattern::UrlPattern;
use super::price::PriceRepresentation;

/// The only manifest format version this interpreter implements.
pub const FORMAT_VERSION: u32 = 1;

/// The only `extractor_kind` this interpreter implements. Kept as a manifest field (rather than
/// assumed) so a future non-retail vertical can add its own interpreter without the format
/// silently changing meaning underneath these files.
pub const SUPPORTED_KIND: &str = "retail";

/// Captures the interpreter supplies itself, so a `url:{…}` expression may name them even though
/// no `[match].url_pattern` declares them.
pub const IMPLICIT_CAPTURES: [&str; 2] = ["host", "page_url"];

/// A parsed, validated adapter manifest.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AdapterManifest {
    pub adapter: AdapterMeta,
    #[serde(rename = "match", default)]
    pub match_rules: MatchRules,
    pub fetch: FetchSpec,
    #[serde(default)]
    pub iterate: Option<IterateSpec>,
    pub fields: FieldMap,
    /// Raw source value (as a string key) -> [`Availability`]. A value absent from this table
    /// resolves to [`Availability::Unknown`]; it is never guessed toward in-stock.
    #[serde(default)]
    pub availability_map: BTreeMap<String, Availability>,
    /// Currency code -> minor-unit exponent, overriding the assumed 2.
    #[serde(default)]
    pub currency_exponents: BTreeMap<String, u32>,
    /// Free text for reviewers. Never machine-read — parsed only so `deny_unknown_fields` can stay
    /// on everywhere else.
    #[serde(default)]
    pub notes: BTreeMap<String, String>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct AdapterMeta {
    pub name: String,
    pub version: u32,
    pub extractor_kind: String,
    pub method: RetailMethod,
}

#[derive(Clone, Debug, Default, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct MatchRules {
    pub url_pattern: Option<String>,
    pub content_type: Option<String>,
    /// Literal substrings that must appear in the fetched body before the adapter is applied — a
    /// cheap pre-filter, not a parse. The real check is whether the fields resolve.
    #[serde(default)]
    pub requires: Vec<String>,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct FetchSpec {
    pub endpoint: String,
    #[serde(default = "default_http_method")]
    pub http_method: String,
    pub format: BodyFormat,
    /// For `format = "html"`: where inside the page the JSON document lives. Only
    /// `"jsonld:<@type>"` is implemented — select the first embedded
    /// `<script type="application/ld+json">` node whose `@type` is (or includes) that type.
    pub embedded: Option<String>,
}

fn default_http_method() -> String {
    "GET".to_string()
}

#[derive(Clone, Copy, Debug, PartialEq, Eq, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum BodyFormat {
    Json,
    Html,
}

#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct IterateSpec {
    /// `"variants[]"` for a named array field, `"$[]"` for the document root itself.
    pub path: String,
}

/// `RetailObservation` field -> source expression. Field names match the struct exactly.
#[derive(Clone, Debug, Deserialize)]
#[serde(deny_unknown_fields)]
pub struct FieldMap {
    pub store: Expr,
    pub sku: Expr,
    /// Used only when `sku` resolves to nothing or an empty string.
    pub sku_fallback: Option<Expr>,
    pub availability: Expr,
    pub quantity: Option<Expr>,
    pub price_minor: Option<Expr>,
    pub currency: Option<Expr>,
    pub price_representation: Option<PriceRepresentation>,
    /// Only meaningful for `price_representation = "minor_int_with_exponent"`.
    pub price_exponent: Option<Expr>,
}

/// A source expression. Three kinds, distinguished by prefix.
#[derive(Clone, Debug, PartialEq, Eq, Deserialize)]
#[serde(try_from = "String")]
pub enum Expr {
    /// `url:{store}` — a capture from `[match].url_pattern`, or an implicit one.
    UrlCapture(String),
    /// `json:$.prices.price` — a path into the JSON document (or the `[iterate]` element).
    JsonPath(Vec<String>),
    /// `const:USD` — a fixed value.
    Const(String),
    /// `const:none` — explicitly unresolved. Distinct from `const:` with an empty value, and the
    /// only honest answer for a field the source genuinely does not carry.
    None,
}

impl TryFrom<String> for Expr {
    type Error = ManifestError;

    fn try_from(raw: String) -> Result<Self, Self::Error> {
        let bad = || ManifestError::BadExpression(raw.clone());
        let (kind, rest) = raw.split_once(':').ok_or_else(bad)?;
        match kind {
            "url" => {
                let name = rest.strip_prefix('{').and_then(|r| r.strip_suffix('}')).ok_or_else(bad)?;
                if name.is_empty() {
                    return Err(bad());
                }
                Ok(Expr::UrlCapture(name.to_string()))
            }
            "json" => {
                let path = rest.strip_prefix("$.").ok_or_else(bad)?;
                if path.is_empty() || path.split('.').any(str::is_empty) {
                    return Err(bad());
                }
                Ok(Expr::JsonPath(path.split('.').map(str::to_string).collect()))
            }
            "const" => {
                if rest == "none" {
                    Ok(Expr::None)
                } else if rest.is_empty() {
                    Err(bad())
                } else {
                    Ok(Expr::Const(rest.to_string()))
                }
            }
            _ => Err(bad()),
        }
    }
}

/// Everything that can be wrong with a manifest, as a reason a reviewer can act on.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum ManifestError {
    Toml(String),
    BadExpression(String),
    UnsupportedVersion(u32),
    UnsupportedKind(String),
    BadPattern(String),
    /// A `url:{name}` expression naming a capture no pattern produces.
    UnknownCapture(String),
    /// `[match]` with neither a URL pattern nor a content type would apply to every page fetched.
    MatchesEverything,
    /// `[fetch].endpoint` naming a URL other than the page the extractor was handed.
    UnfetchableEndpoint(String),
    MissingPriceRepresentation,
    MissingPriceExponent,
    UnsupportedEmbedded(String),
    /// `format = "html"` with no `embedded`, or `format = "json"` with one.
    EmbeddedFormatMismatch,
}

impl std::fmt::Display for ManifestError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Toml(e) => write!(f, "not valid manifest TOML: {e}"),
            Self::BadExpression(e) => write!(
                f,
                "field expression {e:?} is not one of url:{{capture}}, json:$.path, const:VALUE, const:none"
            ),
            Self::UnsupportedVersion(v) => {
                write!(f, "manifest version {v} is not supported (this interpreter implements version {FORMAT_VERSION})")
            }
            Self::UnsupportedKind(k) => {
                write!(f, "extractor_kind {k:?} has no interpreter (only {SUPPORTED_KIND:?} does)")
            }
            Self::BadPattern(e) => write!(f, "url_pattern: {e}"),
            Self::UnknownCapture(c) => {
                write!(f, "url:{{{c}}} names a capture no url_pattern produces (implicit ones: {IMPLICIT_CAPTURES:?})")
            }
            Self::MatchesEverything => {
                write!(f, "[match] needs a url_pattern or a content_type, or the adapter claims every page")
            }
            Self::UnfetchableEndpoint(e) => write!(
                f,
                "[fetch].endpoint {e:?} is neither {{page_url}} nor the url_pattern: an Extractor is handed an \
                 already-fetched page and cannot issue a derived request"
            ),
            Self::MissingPriceRepresentation => {
                write!(f, "[fields].price_minor is set but price_representation is not")
            }
            Self::MissingPriceExponent => {
                write!(f, "price_representation = \"minor_int_with_exponent\" requires [fields].price_exponent")
            }
            Self::UnsupportedEmbedded(e) => {
                write!(f, "[fetch].embedded {e:?} is not supported (only \"jsonld:<@type>\")")
            }
            Self::EmbeddedFormatMismatch => {
                write!(f, "[fetch].embedded is required for format = \"html\" and meaningless for format = \"json\"")
            }
        }
    }
}

impl std::error::Error for ManifestError {}

impl From<ManifestError> for vuna_core::Error {
    fn from(e: ManifestError) -> Self {
        vuna_core::Error::Other(format!("adapter manifest: {e}"))
    }
}

impl AdapterManifest {
    /// Parses and validates a manifest. Every failure names the field that is wrong.
    pub fn parse(toml_src: &str) -> Result<Self, ManifestError> {
        let manifest: Self = toml::from_str(toml_src).map_err(|e| ManifestError::Toml(e.to_string()))?;
        manifest.validate()?;
        Ok(manifest)
    }

    /// The compiled URL pattern, if the manifest declares one.
    pub fn compiled_pattern(&self) -> Result<Option<UrlPattern>, ManifestError> {
        self.match_rules
            .url_pattern
            .as_deref()
            .map(|p| UrlPattern::compile(p).map_err(|e| ManifestError::BadPattern(e.to_string())))
            .transpose()
    }

    /// The `@type` an `embedded = "jsonld:<@type>"` selects, if any.
    pub fn embedded_jsonld_type(&self) -> Option<&str> {
        self.fetch.embedded.as_deref().and_then(|e| e.strip_prefix("jsonld:"))
    }

    fn validate(&self) -> Result<(), ManifestError> {
        if self.adapter.version != FORMAT_VERSION {
            return Err(ManifestError::UnsupportedVersion(self.adapter.version));
        }
        if self.adapter.extractor_kind != SUPPORTED_KIND {
            return Err(ManifestError::UnsupportedKind(self.adapter.extractor_kind.clone()));
        }
        if self.match_rules.url_pattern.is_none() && self.match_rules.content_type.is_none() {
            return Err(ManifestError::MatchesEverything);
        }

        let pattern = self.compiled_pattern()?;

        // An Extractor receives a FetchedPage; it cannot go and fetch something else. So the
        // endpoint must BE the page: either the page URL verbatim, or the URL the pattern matched.
        let endpoint = self.fetch.endpoint.as_str();
        let endpoint_is_the_page = endpoint == "{page_url}"
            || self.match_rules.url_pattern.as_deref() == Some(endpoint);
        if !endpoint_is_the_page {
            return Err(ManifestError::UnfetchableEndpoint(endpoint.to_string()));
        }

        match (self.fetch.format, self.fetch.embedded.as_deref()) {
            (BodyFormat::Html, None) | (BodyFormat::Json, Some(_)) => {
                return Err(ManifestError::EmbeddedFormatMismatch)
            }
            (BodyFormat::Html, Some(e)) if !e.starts_with("jsonld:") || e == "jsonld:" => {
                return Err(ManifestError::UnsupportedEmbedded(e.to_string()))
            }
            _ => {}
        }

        // Every url:{capture} must be produced by the pattern or supplied implicitly.
        let mut known: Vec<&str> = IMPLICIT_CAPTURES.to_vec();
        if let Some(p) = &pattern {
            known.extend(p.capture_names());
        }
        for expr in self.fields.all() {
            if let Expr::UrlCapture(name) = expr {
                if !known.contains(&name.as_str()) {
                    return Err(ManifestError::UnknownCapture(name.clone()));
                }
            }
        }

        match (self.fields.price_minor.is_some(), self.fields.price_representation) {
            (true, None) => return Err(ManifestError::MissingPriceRepresentation),
            (true, Some(PriceRepresentation::MinorIntWithExponent)) if self.fields.price_exponent.is_none() => {
                return Err(ManifestError::MissingPriceExponent)
            }
            _ => {}
        }

        Ok(())
    }
}

impl FieldMap {
    /// Every expression in the map, for validation walks.
    fn all(&self) -> impl Iterator<Item = &Expr> {
        [Some(&self.store), Some(&self.sku), Some(&self.availability)]
            .into_iter()
            .chain([
                self.sku_fallback.as_ref(),
                self.quantity.as_ref(),
                self.price_minor.as_ref(),
                self.currency.as_ref(),
                self.price_exponent.as_ref(),
            ])
            .flatten()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// A minimal manifest every negative test perturbs one field of.
    const MINIMAL: &str = r#"
[adapter]
name = "t"
version = 1
extractor_kind = "retail"
method = "json_endpoint"

[match]
url_pattern = "https://{store}/p/{handle}.json"

[fetch]
endpoint = "https://{store}/p/{handle}.json"
http_method = "GET"
format = "json"

[fields]
store = "url:{store}"
sku = "json:$.sku"
availability = "json:$.available"
"#;

    fn with(replace: &str, replacement: &str) -> String {
        MINIMAL.replace(replace, replacement)
    }

    #[test]
    fn minimal_manifest_parses_and_defaults_http_method() {
        let m = AdapterManifest::parse(MINIMAL).unwrap();
        assert_eq!(m.adapter.name, "t");
        assert_eq!(m.adapter.method, RetailMethod::JsonEndpoint);
        assert_eq!(m.fetch.http_method, "GET");
        assert_eq!(m.fields.store, Expr::UrlCapture("store".into()));
        assert_eq!(m.fields.sku, Expr::JsonPath(vec!["sku".into()]));
    }

    #[test]
    fn expression_kinds_parse() {
        assert_eq!(Expr::try_from("url:{host}".to_string()).unwrap(), Expr::UrlCapture("host".into()));
        assert_eq!(
            Expr::try_from("json:$.prices.currency_code".to_string()).unwrap(),
            Expr::JsonPath(vec!["prices".into(), "currency_code".into()])
        );
        assert_eq!(Expr::try_from("const:USD".to_string()).unwrap(), Expr::Const("USD".into()));
        assert_eq!(Expr::try_from("const:none".to_string()).unwrap(), Expr::None);
    }

    #[test]
    fn malformed_expressions_are_rejected() {
        for raw in ["sku", "url:store", "url:{}", "json:sku", "json:$.", "json:$.a..b", "const:", "nope:x"] {
            assert!(
                Expr::try_from(raw.to_string()).is_err(),
                "{raw:?} should not parse as a field expression"
            );
        }
    }

    /// The failure mode `deny_unknown_fields` exists to catch: a typo that would otherwise be
    /// ignored, leaving prices silently unconverted.
    #[test]
    fn a_misspelled_key_is_a_load_error_not_a_silent_default() {
        let src = with("sku = \"json:$.sku\"", "sku = \"json:$.sku\"\nskew = \"json:$.skew\"");
        let err = AdapterManifest::parse(&src).unwrap_err();
        assert!(matches!(err, ManifestError::Toml(_)), "got {err:?}");
    }

    #[test]
    fn unsupported_version_and_kind_are_refused() {
        assert_eq!(
            AdapterManifest::parse(&with("version = 1", "version = 2")).unwrap_err(),
            ManifestError::UnsupportedVersion(2)
        );
        assert_eq!(
            AdapterManifest::parse(&with(r#"extractor_kind = "retail""#, r#"extractor_kind = "web""#)).unwrap_err(),
            ManifestError::UnsupportedKind("web".into())
        );
    }

    #[test]
    fn a_capture_no_pattern_produces_is_refused() {
        let src = with(r#"store = "url:{store}""#, r#"store = "url:{merchant}""#);
        assert_eq!(AdapterManifest::parse(&src).unwrap_err(), ManifestError::UnknownCapture("merchant".into()));
    }

    #[test]
    fn implicit_captures_need_no_pattern() {
        let src = with(r#"store = "url:{store}""#, r#"store = "url:{host}""#);
        assert!(AdapterManifest::parse(&src).is_ok());
    }

    /// The interpreter is handed an already-fetched page. A manifest asking it to fetch a
    /// *different* URL is refused rather than quietly ignored.
    #[test]
    fn a_derived_endpoint_is_refused() {
        let src = with(
            r#"endpoint = "https://{store}/p/{handle}.json""#,
            r#"endpoint = "https://{store}/meta.json""#,
        );
        assert!(matches!(AdapterManifest::parse(&src).unwrap_err(), ManifestError::UnfetchableEndpoint(_)));
    }

    #[test]
    fn a_match_block_with_no_narrowing_is_refused() {
        let src = MINIMAL
            .replace(r#"url_pattern = "https://{store}/p/{handle}.json""#, "")
            .replace(r#"store = "url:{store}""#, r#"store = "url:{host}""#)
            .replace(r#"endpoint = "https://{store}/p/{handle}.json""#, r#"endpoint = "{page_url}""#);
        assert_eq!(AdapterManifest::parse(&src).unwrap_err(), ManifestError::MatchesEverything);
    }

    #[test]
    fn price_fields_must_be_internally_consistent() {
        let src = with(r#"availability = "json:$.available""#, "availability = \"json:$.available\"\nprice_minor = \"json:$.price\"");
        assert_eq!(AdapterManifest::parse(&src).unwrap_err(), ManifestError::MissingPriceRepresentation);

        let src = format!("{src}\nprice_representation = \"minor_int_with_exponent\"");
        assert_eq!(AdapterManifest::parse(&src).unwrap_err(), ManifestError::MissingPriceExponent);

        let src = format!("{src}\nprice_exponent = \"json:$.exp\"");
        assert!(AdapterManifest::parse(&src).is_ok());
    }

    /// `format = "html"` needs to be told where in the page the JSON document lives; `format =
    /// "json"` must not carry that key at all.
    #[test]
    fn html_format_requires_a_supported_embedded_selector() {
        let page_manifest = |embedded: &str| {
            format!(
                r#"
[adapter]
name = "t"
version = 1
extractor_kind = "retail"
method = "structured_data"

[match]
content_type = "text/html"

[fetch]
endpoint = "{{page_url}}"
format = "html"
{embedded}

[fields]
store = "url:{{host}}"
sku = "json:$.sku"
availability = "json:$.offers.availability"
"#
            )
        };

        assert_eq!(AdapterManifest::parse(&page_manifest("")).unwrap_err(), ManifestError::EmbeddedFormatMismatch);
        assert_eq!(
            AdapterManifest::parse(&page_manifest(r#"embedded = "microdata:Product""#)).unwrap_err(),
            ManifestError::UnsupportedEmbedded("microdata:Product".into())
        );

        let m = AdapterManifest::parse(&page_manifest(r#"embedded = "jsonld:Product""#)).unwrap();
        assert_eq!(m.embedded_jsonld_type(), Some("Product"));

        // The mirror image: a JSON endpoint has nowhere to embed anything.
        let json_with_embedded = with(r#"format = "json""#, "format = \"json\"\nembedded = \"jsonld:Product\"");
        assert_eq!(
            AdapterManifest::parse(&json_with_embedded).unwrap_err(),
            ManifestError::EmbeddedFormatMismatch
        );
    }
}
