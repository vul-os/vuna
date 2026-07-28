//! The `retail` vertical: price/stock observations of a *non-participant* store. Only STRUCTURED
//! markup is read — JSON-LD (`schema.org/Product` + `Offer`/`AggregateOffer`) and Open Graph
//! product tags — never heuristics over rendered text. That's the whole point of
//! [`RetailMethod::StructuredData`]: near-free to read, and honest about what it actually saw.

use scraper::{Html, Selector};
use serde_json::Value;
use url::Url;
use vuna_core::extract::{Availability, Extraction, Extractor, ExtractorKind, FetchedPage, RetailMethod, RetailObservation};
use vuna_core::{Result, UnixSecs};

/// How far into the body we sniff for structured-data tokens in `applies()`. JSON-LD/OG tags live
/// in `<head>`, so a bounded prefix is enough and keeps the sniff cheap on large pages.
const SNIFF_WINDOW: usize = 65_536;
const SNIFF_TOKENS: [&str; 4] = ["ld+json", "og:price", "product:price", "product:availability"];

/// Reads `schema.org` JSON-LD and Open Graph product tags into [`RetailObservation`]s.
#[derive(Clone, Copy, Debug, Default)]
pub struct RetailExtractor;

impl Extractor for RetailExtractor {
    fn kind(&self) -> ExtractorKind {
        "retail".to_string()
    }

    fn applies(&self, page: &FetchedPage) -> bool {
        if !crate::looks_like_html(page) {
            return false;
        }
        let head = &page.body[..page.body.len().min(SNIFF_WINDOW)];
        let head = String::from_utf8_lossy(head).to_ascii_lowercase();
        SNIFF_TOKENS.iter().any(|t| head.contains(t))
    }

    fn extract(&self, page: &FetchedPage) -> Result<Extraction> {
        let html = String::from_utf8_lossy(&page.body);
        let document = Html::parse_document(&html);
        let store = store_name(&page.url);

        let mut observations = extract_json_ld(&document, &store, &page.url, page.fetched_at);
        // JSON-LD is preferred (richer, per-SKU); Open Graph is a same-page fallback, not additive,
        // to avoid double-reporting the one product a typical page describes.
        if observations.is_empty() {
            observations.extend(extract_open_graph(&document, &store, &page.url, page.fetched_at));
        }

        Ok(Extraction::Retail(observations))
    }
}

/// The observed store's identity: the page's host, or (defensively) the whole URL if it somehow
/// doesn't parse — `store` is never left empty.
fn store_name(url: &str) -> String {
    Url::parse(url).ok().and_then(|u| u.host_str().map(str::to_string)).unwrap_or_else(|| url.to_string())
}

/// Finds every `schema.org/Product` node across all `<script type="application/ld+json">` blocks
/// (a page may have several, or nest products under `@graph`) and turns each into observations.
fn extract_json_ld(document: &Html, store: &str, page_url: &str, observed_at: UnixSecs) -> Vec<RetailObservation> {
    let sel = Selector::parse(r#"script[type="application/ld+json"]"#).expect("static selector");
    let mut out = Vec::new();

    for script in document.select(&sel) {
        let raw: String = script.text().collect();
        // Malformed JSON-LD is common in the wild (trailing commas, HTML entities left in by a
        // buggy template). Skip that one block rather than failing the whole page's extraction.
        let Ok(value) = serde_json::from_str::<Value>(raw.trim()) else { continue };

        let mut products = Vec::new();
        collect_typed(&value, "Product", &mut products);
        for product in products {
            out.extend(observations_from_product(product, store, page_url, observed_at));
        }
    }

    out
}

/// Recursively finds every object whose `@type` is (or includes) `want`, anywhere in the JSON-LD
/// value — handles a bare node, an array of them, and `@graph`-wrapped nodes uniformly since we
/// just walk every nested object/array without special-casing the wrapper key. Document order is
/// preserved, so the first hit is the first the page declared.
///
/// Shared with the declarative [`crate::adapter`] interpreter so a page's JSON-LD is located the
/// same way whether it is read by this Rust extractor or by an `adapters/*.toml` manifest.
pub(crate) fn collect_typed<'a>(value: &'a Value, want: &str, out: &mut Vec<&'a Value>) {
    match value {
        Value::Object(map) => {
            if has_type(map, want) {
                out.push(value);
            }
            for v in map.values() {
                collect_typed(v, want, out);
            }
        }
        Value::Array(items) => {
            for v in items {
                collect_typed(v, want, out);
            }
        }
        _ => {}
    }
}

fn has_type(map: &serde_json::Map<String, Value>, want: &str) -> bool {
    match map.get("@type") {
        Some(Value::String(s)) => s == want,
        Some(Value::Array(types)) => types.iter().any(|v| v.as_str() == Some(want)),
        _ => false,
    }
}

/// A `Product` may carry one `Offer`/`AggregateOffer` object or an array of them; each becomes its
/// own observation (same SKU, potentially different sellers/prices).
fn observations_from_product(
    product: &Value,
    store: &str,
    page_url: &str,
    observed_at: UnixSecs,
) -> Vec<RetailObservation> {
    let Value::Object(map) = product else { return Vec::new() };

    let product_sku = map
        .get("sku")
        .and_then(Value::as_str)
        .or_else(|| map.get("productID").and_then(Value::as_str))
        .or_else(|| map.get("mpn").and_then(Value::as_str));

    let mut offers = Vec::new();
    match map.get("offers") {
        Some(Value::Array(items)) => offers.extend(items.iter()),
        Some(single @ Value::Object(_)) => offers.push(single),
        _ => {}
    }

    offers
        .into_iter()
        .filter_map(|offer| {
            let Value::Object(offer_map) = offer else { return None };

            // Falls back to the page URL when no SKU is published anywhere — never an empty
            // string, since `RetailObservation::sku` is the join key consumers key quorum on.
            let sku = product_sku
                .or_else(|| offer_map.get("sku").and_then(Value::as_str))
                .map(str::to_string)
                .unwrap_or_else(|| page_url.to_string());

            let price_minor = offer_map.get("price").and_then(value_as_f64).map(to_minor_units);
            let currency = offer_map.get("priceCurrency").and_then(Value::as_str).map(str::to_string);
            let availability = offer_map
                .get("availability")
                .and_then(Value::as_str)
                .map(parse_availability)
                .unwrap_or(Availability::Unknown);
            let quantity = offer_map
                .get("inventoryLevel")
                .and_then(|v| v.get("value"))
                .and_then(Value::as_u64);

            Some(RetailObservation {
                store: store.to_string(),
                sku,
                availability,
                quantity,
                price_minor,
                currency,
                method: RetailMethod::StructuredData,
                observed_at,
            })
        })
        .collect()
}

/// Fallback for pages with no JSON-LD Product: Open Graph / product-namespace meta tags. Yields
/// at most one observation (OG describes "this page", not a per-SKU list) and only when there is
/// at least a price or an availability signal to report.
fn extract_open_graph(document: &Html, store: &str, page_url: &str, observed_at: UnixSecs) -> Option<RetailObservation> {
    let amount = meta_property(document, "og:price:amount").or_else(|| meta_property(document, "product:price:amount"));
    let currency = meta_property(document, "og:price:currency").or_else(|| meta_property(document, "product:price:currency"));
    let availability_raw = meta_property(document, "product:availability").or_else(|| meta_property(document, "og:availability"));

    if amount.is_none() && availability_raw.is_none() {
        return None;
    }

    let price_minor = amount.as_deref().and_then(|s| s.trim().parse::<f64>().ok()).map(to_minor_units);
    let availability = availability_raw.as_deref().map(parse_availability).unwrap_or(Availability::Unknown);

    Some(RetailObservation {
        store: store.to_string(),
        // Open Graph carries no SKU field; the page URL is the closest stable key we have.
        sku: page_url.to_string(),
        availability,
        quantity: None,
        price_minor,
        currency,
        method: RetailMethod::StructuredData,
        observed_at,
    })
}

/// Looks up a `<meta>` by `property=` (Open Graph convention) or `name=` (some sites use either).
fn meta_property(document: &Html, key: &str) -> Option<String> {
    let sel = Selector::parse("meta").expect("static selector");
    document.select(&sel).find_map(|el| {
        let matches = el.attr("property") == Some(key) || el.attr("name") == Some(key);
        if !matches {
            return None;
        }
        el.attr("content").map(str::trim).filter(|c| !c.is_empty()).map(str::to_string)
    })
}

/// JSON-LD's `price` is sometimes a number, sometimes a numeric string — accept either.
fn value_as_f64(v: &Value) -> Option<f64> {
    v.as_f64().or_else(|| v.as_str().and_then(|s| s.trim().parse::<f64>().ok()))
}

/// Converts a decimal major-unit price (e.g. `19.99`) to integer minor units (`1999`). Assumes a
/// 2-decimal currency, which covers the overwhelming majority of schema.org-tagged storefronts;
/// 0- and 3-decimal currencies (JPY, KWD, …) are a known rounding gap for a later pass.
fn to_minor_units(major: f64) -> i64 {
    (major * 100.0).round() as i64
}

/// Normalizes a schema.org `availability` URI (`https://schema.org/InStock`) or a bare Open Graph
/// token (`instock`, `out of stock`) down to the shared [`Availability`] enum.
fn parse_availability(raw: &str) -> Availability {
    let token = raw.rsplit('/').next().unwrap_or(raw);
    let token: String = token.chars().filter(|c| c.is_ascii_alphanumeric()).collect();
    match token.to_ascii_lowercase().as_str() {
        "instock" | "available" | "onlineonly" | "instoreonly" => Availability::InStock,
        "outofstock" | "soldout" | "discontinued" => Availability::OutOfStock,
        "lowstock" | "limitedavailability" | "preorder" | "backorder" | "backordered" | "presale" => {
            Availability::LowStock
        }
        _ => Availability::Unknown,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn page(url: &str, html: &str) -> FetchedPage {
        FetchedPage {
            url: url.to_string(),
            status: 200,
            content_type: Some("text/html; charset=utf-8".to_string()),
            body: html.as_bytes().to_vec(),
            fetched_at: 1_700_000_000,
        }
    }

    fn extract_retail(url: &str, html: &str) -> Vec<RetailObservation> {
        match RetailExtractor.extract(&page(url, html)).unwrap() {
            Extraction::Retail(obs) => obs,
            other => panic!("expected Extraction::Retail, got {other:?}"),
        }
    }

    /// A JSON-LD `Product` + `Offer` yields one observation with price/currency/availability
    /// parsed and the SKU taken from the structured markup, not guessed.
    #[test]
    fn extracts_price_currency_availability_from_json_ld() {
        let html = r#"
            <html><head>
            <script type="application/ld+json">
            {
              "@context": "https://schema.org/",
              "@type": "Product",
              "name": "Widget",
              "sku": "WIDGET-123",
              "offers": {
                "@type": "Offer",
                "priceCurrency": "USD",
                "price": "19.99",
                "availability": "https://schema.org/InStock"
              }
            }
            </script>
            </head><body></body></html>
        "#;
        let p = page("https://shop.example.com/widget", html);
        assert!(RetailExtractor.applies(&p));

        let obs = extract_retail("https://shop.example.com/widget", html);
        assert_eq!(obs.len(), 1);
        let o = &obs[0];
        assert_eq!(o.store, "shop.example.com");
        assert_eq!(o.sku, "WIDGET-123");
        assert_eq!(o.price_minor, Some(1999));
        assert_eq!(o.currency.as_deref(), Some("USD"));
        assert_eq!(o.availability, Availability::InStock);
        assert_eq!(o.method, RetailMethod::StructuredData);
        assert_eq!(o.observed_at, 1_700_000_000);
    }

    /// `@graph`-wrapped products and a numeric (not string) price both parse correctly.
    #[test]
    fn handles_graph_wrapper_and_numeric_price() {
        let html = r#"
            <html><head>
            <script type="application/ld+json">
            {
              "@context": "https://schema.org/",
              "@graph": [
                { "@type": "WebPage", "name": "irrelevant" },
                {
                  "@type": "Product",
                  "sku": "GRAPH-9",
                  "offers": { "@type": "Offer", "price": 5, "priceCurrency": "ZAR", "availability": "OutOfStock" }
                }
              ]
            }
            </script>
            </head><body></body></html>
        "#;
        let obs = extract_retail("https://shop.example.com/g", html);
        assert_eq!(obs.len(), 1);
        assert_eq!(obs[0].sku, "GRAPH-9");
        assert_eq!(obs[0].price_minor, Some(500));
        assert_eq!(obs[0].currency.as_deref(), Some("ZAR"));
        assert_eq!(obs[0].availability, Availability::OutOfStock);
    }

    /// No JSON-LD: falls back to Open Graph product tags.
    #[test]
    fn falls_back_to_open_graph_when_no_json_ld() {
        let html = r#"
            <html><head>
                <meta property="og:price:amount" content="42.50">
                <meta property="og:price:currency" content="EUR">
                <meta property="product:availability" content="in stock">
            </head><body></body></html>
        "#;
        let p = page("https://shop.example.com/og-item", html);
        assert!(RetailExtractor.applies(&p));

        let obs = extract_retail("https://shop.example.com/og-item", html);
        assert_eq!(obs.len(), 1);
        assert_eq!(obs[0].price_minor, Some(4250));
        assert_eq!(obs[0].currency.as_deref(), Some("EUR"));
        assert_eq!(obs[0].availability, Availability::InStock);
        assert_eq!(obs[0].method, RetailMethod::StructuredData);
    }

    /// A plain page with neither signal applies=false and yields no observations.
    #[test]
    fn does_not_apply_without_structured_signals() {
        let html = "<html><head><title>Blog</title></head><body><p>Just words.</p></body></html>";
        let p = page("https://example.com/blog", html);
        assert!(!RetailExtractor.applies(&p));
    }

    /// Malformed JSON-LD is skipped, not a hard extraction failure.
    #[test]
    fn tolerates_malformed_json_ld() {
        let html = r#"<html><head>
            <script type="application/ld+json">{ this is not json }</script>
        </head><body></body></html>"#;
        let obs = extract_retail("https://shop.example.com/broken", html);
        assert!(obs.is_empty());
    }
}
