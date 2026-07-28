//! The shipped `adapters/` corpus, run against the interpreter.
//!
//! Two jobs, and the second is the one that matters over time:
//!
//! 1. Every manifest in the repo loads, validates, and produces the observations a realistic
//!    payload should produce.
//! 2. **Coverage is asserted, not assumed.** [`FIXTURES`] must name every manifest on disk, set
//!    for set. Dropping a new `.toml` into `adapters/` without a fixture fails this file — the
//!    corpus cannot grow while the tests quietly keep exercising three files. A harness that
//!    passes by iterating over an empty directory is worse than no harness, so the directory
//!    itself is asserted non-empty too.

use std::collections::BTreeSet;
use std::path::{Path, PathBuf};

use vuna_extract::adapter::AdapterExtractor;
use vuna_core::extract::{Availability, Extraction, Extractor, FetchedPage, RetailMethod, RetailObservation};

/// Every adapter in `adapters/`, with the page each is exercised against below. Adding a manifest
/// means adding a line here — that is the point.
const FIXTURES: [&str; 3] = ["generic-jsonld", "shopify-products-json", "woocommerce-store-api"];

const OBSERVED_AT: u64 = 1_700_000_000;

fn adapters_dir() -> PathBuf {
    // CARGO_MANIFEST_DIR is crates/vuna-extract.
    Path::new(env!("CARGO_MANIFEST_DIR")).join("../../adapters")
}

fn load_all() -> Vec<AdapterExtractor> {
    AdapterExtractor::load_dir(&adapters_dir())
        .unwrap_or_else(|e| panic!("adapters/ must load: {e}"))
}

fn adapter(name: &str) -> AdapterExtractor {
    let path = adapters_dir().join(format!("{name}.toml"));
    let src = std::fs::read_to_string(&path).unwrap_or_else(|e| panic!("{}: {e}", path.display()));
    AdapterExtractor::from_toml(&src).unwrap_or_else(|e| panic!("{name}: {e}"))
}

fn page(url: &str, content_type: &str, body: &str) -> FetchedPage {
    FetchedPage {
        url: url.to_string(),
        status: 200,
        content_type: Some(content_type.to_string()),
        body: body.as_bytes().to_vec(),
        fetched_at: OBSERVED_AT,
    }
}

fn run(name: &str, page: &FetchedPage) -> Vec<RetailObservation> {
    let extractor = adapter(name);
    assert!(extractor.applies(page), "{name} should apply to {}", page.url);
    match extractor.extract(page).unwrap_or_else(|e| panic!("{name}: {e}")) {
        Extraction::Retail(obs) => obs,
        other => panic!("{name}: expected Extraction::Retail, got {other:?}"),
    }
}

/// The guard: the fixture list and the directory must describe the same set of adapters.
#[test]
fn every_manifest_on_disk_is_exercised_by_a_fixture() {
    let on_disk: BTreeSet<String> = load_all().iter().map(|a| a.name().to_string()).collect();
    let fixtured: BTreeSet<String> = FIXTURES.iter().map(|s| s.to_string()).collect();

    assert!(!on_disk.is_empty(), "adapters/ is empty — this harness would otherwise pass by doing nothing");
    assert_eq!(
        on_disk, fixtured,
        "adapters/ and the FIXTURES list have drifted.\n  on disk:   {on_disk:?}\n  fixtured:  {fixtured:?}\n\
         Every adapter must be exercised by a fixture in this file."
    );
}

/// `[adapter].name` is the identity used in logs, errors, and reviews; a manifest whose name does
/// not match its filename is a rename waiting to confuse someone.
#[test]
fn manifest_names_match_their_filenames() {
    let mut checked: usize = 0;
    for entry in std::fs::read_dir(adapters_dir()).expect("adapters/ readable") {
        let path = entry.expect("dir entry").path();
        if path.extension().map_or(true, |e| e != "toml") {
            continue;
        }
        let stem = path.file_stem().unwrap().to_string_lossy().to_string();
        let src = std::fs::read_to_string(&path).expect("readable");
        let manifest = AdapterExtractor::from_toml(&src).unwrap_or_else(|e| panic!("{stem}: {e}"));
        assert_eq!(manifest.name(), stem, "{}: [adapter].name must match the filename", path.display());
        assert_eq!(manifest.kind(), "retail");
        checked += 1;
    }
    assert_eq!(checked, FIXTURES.len(), "checked {checked} manifests, expected {}", FIXTURES.len());
}

/// Shopify: a path-template JSON endpoint, one observation per variant, integer cents.
#[test]
fn shopify_products_json_reads_every_variant() {
    let body = r#"{
      "id": 4001, "title": "Blue Widget",
      "variants": [
        {"id": 111, "sku": "BW-S",  "available": true,  "inventory_quantity": 12, "price": 1999},
        {"id": 112, "sku": "BW-M",  "available": true,                            "price": 1999},
        {"id": 113, "sku": "",      "available": false, "inventory_quantity": 0,  "price": 2499}
      ]
    }"#;
    let p = page("https://shop.example.com/products/blue-widget.js", "application/json", body);
    let obs = run("shopify-products-json", &p);

    assert_eq!(obs.len(), 3, "one observation per variant");
    for o in &obs {
        assert_eq!(o.store, "shop.example.com");
        assert_eq!(o.method, RetailMethod::JsonEndpoint);
        assert_eq!(o.observed_at, OBSERVED_AT);
        assert_eq!(o.currency, None, "this endpoint carries no currency — it must stay unresolved");
    }

    assert_eq!(obs[0].sku, "BW-S");
    assert_eq!(obs[0].availability, Availability::InStock);
    assert_eq!(obs[0].quantity, Some(12));
    assert_eq!(obs[0].price_minor, Some(1999), "Shopify prices are already integer cents");

    // The merchant hid the count: honest None, never inferred from `available`.
    assert_eq!(obs[1].quantity, None);
    assert_eq!(obs[1].availability, Availability::InStock);

    // Blank sku falls through to the numeric variant id.
    assert_eq!(obs[2].sku, "113");
    assert_eq!(obs[2].availability, Availability::OutOfStock);
    assert_eq!(obs[2].quantity, Some(0), "a reported zero is a fact, not a missing value");
}

#[test]
fn shopify_products_json_ignores_pages_that_are_not_the_json_twin() {
    let extractor = adapter("shopify-products-json");
    let body = r#"{"variants":[]}"#;
    // The human-readable product page, not the `.js` twin.
    assert!(!extractor.applies(&page("https://shop.example.com/products/blue-widget", "application/json", body)));
    // A collection endpoint on the same host.
    assert!(!extractor.applies(&page("https://shop.example.com/collections/all.js", "application/json", body)));
}

/// Generic JSON-LD: a whole HTML page, schema.org vocabulary, decimal major units.
#[test]
fn generic_jsonld_reads_an_embedded_product_offer() {
    let html = r#"<html><head>
      <script type="application/ld+json">{"@type":"BreadcrumbList","itemListElement":[]}</script>
      <script type="application/ld+json">
      {
        "@context": "https://schema.org/",
        "@type": "Product",
        "name": "Kettle",
        "sku": "KTL-1",
        "offers": {
          "@type": "Offer",
          "price": "19.99",
          "priceCurrency": "USD",
          "availability": "https://schema.org/InStock"
        }
      }
      </script>
    </head><body><h1>Kettle</h1></body></html>"#;
    let p = page("https://store.example.org/kettle", "text/html; charset=utf-8", html);
    let obs = run("generic-jsonld", &p);

    assert_eq!(obs.len(), 1, "no [iterate]: one Product, one observation");
    assert_eq!(obs[0].store, "store.example.org", "store comes from the implicit host capture");
    assert_eq!(obs[0].sku, "KTL-1");
    assert_eq!(obs[0].price_minor, Some(1999), "\"19.99\" major units -> 1999 minor units");
    assert_eq!(obs[0].currency.as_deref(), Some("USD"));
    assert_eq!(obs[0].availability, Availability::InStock);
    assert_eq!(obs[0].quantity, None, "schema.org Offer has no exact-count field: const:none");
    assert_eq!(obs[0].method, RetailMethod::StructuredData);
}

/// The 11-value `ItemAvailability` vocabulary collapsed onto Vuna's 4-value enum, exactly as the
/// manifest's `[availability_map]` writes it — including the deliberate `unknown` fallthrough the
/// manifest promises for anything unlisted.
#[test]
fn generic_jsonld_availability_map_is_honoured_including_its_fallthrough() {
    let cases = [
        ("https://schema.org/InStock", Availability::InStock),
        ("https://schema.org/InStoreOnly", Availability::InStock),
        ("https://schema.org/OnlineOnly", Availability::InStock),
        ("https://schema.org/InStockAndOnline", Availability::InStock),
        ("https://schema.org/LimitedAvailability", Availability::LowStock),
        ("https://schema.org/OutOfStock", Availability::OutOfStock),
        ("https://schema.org/SoldOut", Availability::OutOfStock),
        ("https://schema.org/Discontinued", Availability::OutOfStock),
        ("https://schema.org/BackOrder", Availability::OutOfStock),
        ("https://schema.org/PreOrder", Availability::Unknown),
        ("https://schema.org/PreSale", Availability::Unknown),
        // The manifest's stated default: a bare token without the schema.org prefix, a typo, or a
        // custom vocabulary is never guessed toward in_stock.
        ("InStock", Availability::Unknown),
        ("http://schema.org/InStock", Availability::Unknown),
        ("definitely in stock!", Availability::Unknown),
    ];

    for (raw, expected) in cases {
        let html = format!(
            r#"<html><head><script type="application/ld+json">
            {{"@type":"Product","sku":"X","offers":{{"@type":"Offer","price":"1.00","priceCurrency":"USD","availability":"{raw}"}}}}
            </script></head><body></body></html>"#
        );
        let p = page("https://store.example.org/x", "text/html", &html);
        let obs = run("generic-jsonld", &p);
        assert_eq!(obs[0].availability, expected, "availability {raw:?}");
    }
}

/// `[currency_exponents]` is a table rather than a constant because "multiply by 100" is wrong for
/// real currencies in both directions.
#[test]
fn generic_jsonld_currency_exponent_table_is_applied() {
    for (currency, price, expected_minor) in
        [("USD", "19.99", 1999), ("JPY", "1200", 1200), ("KRW", "9900", 9900), ("BHD", "19.995", 19995), ("KWD", "1.250", 1250)]
    {
        let html = format!(
            r#"<html><head><script type="application/ld+json">
            {{"@type":"Product","sku":"X","offers":{{"@type":"Offer","price":"{price}","priceCurrency":"{currency}","availability":"https://schema.org/InStock"}}}}
            </script></head><body></body></html>"#
        );
        let p = page("https://store.example.org/x", "text/html", &html);
        let obs = run("generic-jsonld", &p);
        assert_eq!(obs[0].price_minor, Some(expected_minor), "{currency} {price}");
    }
}

/// A documented limitation, pinned so it stays a *limitation* rather than becoming a wrong number:
/// `generic-jsonld.toml` assumes one `Offer` per `Product`. An `AggregateOffer` (a price range) or
/// an `offers` array resolves to no price and `unknown` availability — never to whichever member
/// of the range happened to be first.
#[test]
fn generic_jsonld_declines_to_price_a_multi_offer_product() {
    let html = r#"<html><head><script type="application/ld+json">
      {"@type":"Product","sku":"MULTI-1","offers":[
        {"@type":"Offer","price":"10.00","priceCurrency":"USD","availability":"https://schema.org/InStock"},
        {"@type":"Offer","price":"25.00","priceCurrency":"USD","availability":"https://schema.org/OutOfStock"}
      ]}
    </script></head><body></body></html>"#;
    let p = page("https://store.example.org/multi", "text/html", html);
    let obs = run("generic-jsonld", &p);

    assert_eq!(obs.len(), 1);
    assert_eq!(obs[0].sku, "MULTI-1");
    assert_eq!(obs[0].price_minor, None, "a price range is not a price");
    assert_eq!(obs[0].currency, None);
    assert_eq!(obs[0].availability, Availability::Unknown);
}

#[test]
fn generic_jsonld_skips_pages_with_no_product_node() {
    let extractor = adapter("generic-jsonld");
    // `requires` pre-filter: no JSON-LD at all.
    let plain = page("https://store.example.org/about", "text/html", "<html><body>About us</body></html>");
    assert!(!extractor.applies(&plain));

    // JSON-LD whose prose merely mentions products: the manifest requires the *quoted* token
    // `"Product"`, so the cheap pre-filter still rules this out.
    let article = r#"<html><head><script type="application/ld+json">
        {"@type":"Article","headline":"Our Product philosophy"}
        </script></head><body></body></html>"#;
    assert!(!extractor.applies(&page("https://store.example.org/blog", "text/html", article)));

    // A page that does clear the pre-filter but has no Product node must be a clean per-adapter
    // error, never a bogus observation. The pre-filter is deliberately cheap and over-matches.
    let html = r#"<html><head><script type="application/ld+json">
        {"@type":"Article","keywords":["Product","reviews"]}
        </script></head><body></body></html>"#;
    let p = page("https://store.example.org/reviews", "text/html", html);
    assert!(extractor.applies(&p), "the substring pre-filter is deliberately cheap and over-matches");
    assert!(extractor.extract(&p).is_err(), "no Product node is an error for this adapter alone");
}

/// WooCommerce: query-param matching, a bare top-level array, and the source stating its own
/// minor-unit exponent.
#[test]
fn woocommerce_store_api_reads_a_sku_query_response() {
    let body = r#"[{
      "id": 77, "name": "Lamp", "sku": "LMP-9",
      "is_in_stock": true, "stock_quantity": 4,
      "prices": {"price": "2450", "currency_code": "EUR", "currency_minor_unit": 2}
    }]"#;
    let p = page("https://shop.example.net/wp-json/wc/store/v1/products?sku=LMP-9", "application/json", body);
    let obs = run("woocommerce-store-api", &p);

    assert_eq!(obs.len(), 1);
    assert_eq!(obs[0].store, "shop.example.net");
    assert_eq!(obs[0].sku, "LMP-9");
    assert_eq!(obs[0].availability, Availability::InStock);
    assert_eq!(obs[0].quantity, Some(4));
    assert_eq!(obs[0].price_minor, Some(2450));
    assert_eq!(obs[0].currency.as_deref(), Some("EUR"));
    assert_eq!(obs[0].method, RetailMethod::JsonEndpoint);
}

/// The Store API's `stock_quantity` is nullable even with stock management on. Null must land as
/// `None`, which `quorum::reconcile` treats as "this observer reported no count" — not as zero.
#[test]
fn woocommerce_store_api_null_stock_is_unknown_not_zero() {
    let body = r#"[{"id":78,"sku":"LMP-X","is_in_stock":true,"stock_quantity":null,
                    "prices":{"price":"1000","currency_code":"EUR","currency_minor_unit":2}}]"#;
    let p = page("https://shop.example.net/wp-json/wc/store/v1/products?sku=LMP-X", "application/json", body);
    let obs = run("woocommerce-store-api", &p);
    assert_eq!(obs[0].quantity, None);
    assert_eq!(obs[0].availability, Availability::InStock);
}

/// The distinguishing feature of this adapter: an exponent stated in the payload beats an assumed
/// one. A JPY price at `currency_minor_unit = 0` must not be multiplied by 100.
#[test]
fn woocommerce_store_api_uses_the_payload_stated_exponent() {
    let body = r#"[{"id":79,"sku":"JP-1","is_in_stock":true,"stock_quantity":2,
                    "prices":{"price":"1200","currency_code":"JPY","currency_minor_unit":0}}]"#;
    let p = page("https://shop.example.net/wp-json/wc/store/v1/products?sku=JP-1", "application/json", body);
    let obs = run("woocommerce-store-api", &p);
    assert_eq!(obs[0].price_minor, Some(1200));
    assert_eq!(obs[0].currency.as_deref(), Some("JPY"));
}

/// A SKU query "normally resolves to zero or one element" — the manifest says the interpreter must
/// not assume exactly one, so both ends of that are checked.
#[test]
fn woocommerce_store_api_handles_zero_and_many_results() {
    let p = page("https://shop.example.net/wp-json/wc/store/v1/products?sku=NOPE", "application/json", "[]");
    assert!(run("woocommerce-store-api", &p).is_empty());

    let body = r#"[{"id":1,"sku":"A","is_in_stock":true,"prices":{"price":"100","currency_code":"EUR","currency_minor_unit":2}},
                   {"id":2,"sku":"B","is_in_stock":false,"prices":{"price":"200","currency_code":"EUR","currency_minor_unit":2}}]"#;
    let p = page("https://shop.example.net/wp-json/wc/store/v1/products?sku=A", "application/json", body);
    let obs = run("woocommerce-store-api", &p);
    assert_eq!(obs.len(), 2);
    assert_eq!(obs[1].availability, Availability::OutOfStock);
}

/// Adapters are additive coverage, never a replacement: a Shopify product page still reads as a
/// web doc, and one page may feed several adapters at once.
#[test]
fn adapters_register_alongside_the_built_in_extractors() {
    let mut registry = vuna_extract::ExtractorRegistry::with_defaults();
    for a in load_all() {
        registry.register(Box::new(a));
    }

    let html = r#"<html><head><title>Kettle</title><script type="application/ld+json">
        {"@type":"Product","sku":"KTL-1","offers":{"@type":"Offer","price":"19.99","priceCurrency":"USD","availability":"https://schema.org/InStock"}}
        </script></head><body><p>A kettle.</p></body></html>"#;
    let p = page("https://store.example.org/kettle", "text/html", html);

    let applicable: Vec<_> = registry.applicable(&p).iter().map(|e| e.kind()).collect();
    assert_eq!(
        applicable,
        vec!["web".to_string(), "retail".to_string(), "retail".to_string()],
        "web + the built-in retail extractor + the generic-jsonld adapter"
    );

    // Every result is Ok or a clean per-adapter error; none of them can stop the others.
    let results = registry.extract_all(&p);
    assert_eq!(results.len(), 3);
    assert!(results.iter().filter(|r| r.is_ok()).count() >= 2);
}
