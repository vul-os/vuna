# Vuna retail adapters — the long tail of sites is data, not Rust

The retail vertical (`retail` in `vuna_core::extract::ExtractorKind`) reads price/availability
signal off structured data that stores already publish — JSON-LD, Open Graph, or a store
platform's own JSON endpoints (`extract::RetailMethod`). Every store does this slightly
differently: different endpoint shapes, different field names, different price representations
(integer cents vs. decimal major-unit strings vs. an explicit minor-unit exponent), different
availability vocabularies.

Writing a new Rust `Extractor` impl for every storefront platform would make `vuna-extract`'s Rust
surface grow without bound and turn every new site into a reviewed, compiled, released change.
Instead, one small interpreter in `vuna-extract` (currently a stub — see `02-architecture.md`)
reads a **declarative manifest per site/platform**, and that manifest is the only thing that grows
as coverage grows. A manifest is data: reviewable in a PR diff, safe to accept from a contributor
who has never touched Rust, and — matching the frontier's `UrlList` model — distributable as an
ordinary signed object rather than a code release.

This directory holds the manifest **format** and worked examples. The interpreter that consumes it
does not exist yet; this is the target shape it will implement against.

---

## Manifest shape

Each `.toml` file describes one adapter: how to recognize a page it applies to, how to fetch the
structured data, and how to map that data onto `vuna_core::extract::RetailObservation`.

```toml
[adapter]
name    = "..."          # unique, kebab-case, matches the filename
version = 1
extractor_kind = "retail"
method  = "..."           # one of: structured_data | availability_flag | json_endpoint | cart_probe | html
                           # — maps 1:1 onto extract::RetailMethod (serde snake_case)

[match]
url_pattern  = "..."      # glob with {placeholder} captures, e.g. "https://{store}/products/{handle}.js"
content_type = "..."      # optional
requires     = [...]      # optional: markers the interpreter must find before applying (e.g. a JSON-LD @type)

[fetch]
endpoint    = "..."       # template; usually == url_pattern, sometimes a derived URL
http_method = "GET"
format      = "json" | "html"

[iterate]
path = "..."              # optional: a JSON array path; one RetailObservation emitted per element

[fields]
# RetailObservation field -> source expression. Three expression kinds:
#   url:{capture}        — pulled from a {placeholder} the url_pattern captured
#   json:$.path.to.field — JSON field path, relative to the iterate element if [iterate] is set
#   const:VALUE          — a fixed value (use when the source doesn't carry the field at all)
store         = "..."
sku           = "..."
availability  = "..."     # resolved value looked up in [availability_map] below
quantity      = "..."     # optional — RetailObservation.quantity is Option<u64>; absent path = None, never a guess
price_minor   = "..."
currency      = "..."
price_representation = "minor_int" | "major_decimal" | "minor_int_with_exponent"
# minor_int              — source value is already integer minor units (e.g. Shopify cents), no conversion
# major_decimal          — source is a decimal string in major units; needs a currency → exponent lookup
# minor_int_with_exponent — source gives both the integer value AND its own exponent explicitly

[availability_map]
# raw source value (as a string key) -> one of: in_stock | out_of_stock | low_stock | unknown

[currency_exponents]
# optional overrides for major_decimal conversion; default assumed = 2 (cents) when unlisted.
# Real exceptions exist — JPY has no minor unit (0), some Gulf-state currencies use 3 — so this
# is a table, not a constant, on purpose.

[notes]
# free text: caveats a future adapter author or reviewer needs, not machine-read.
```

Every field maps directly onto a real `vuna-core` type, so there is nothing to design once the
interpreter lands — the manifest already says everything `RetailObservation` needs:

```rust
pub struct RetailObservation {
    pub store: String,
    pub sku: String,
    pub availability: Availability,          // <- [fields].availability + [availability_map]
    pub quantity: Option<u64>,               // <- [fields].quantity (optional path -> None if absent)
    pub price_minor: Option<i64>,            // <- [fields].price_minor + price_representation
    pub currency: Option<String>,
    pub method: RetailMethod,                // <- [adapter].method
    pub observed_at: UnixSecs,               // <- filled by the interpreter at fetch time, not the manifest
}
```

`RetailObservation.quantity` and `.price_minor` are already `Option` in the contract — a manifest
that can't find a field for a given store should leave it unresolved, never fabricate a number.
This matters because `quorum::reconcile` treats "no quantity reported" as a legitimate outcome (a
median only forms over observers who *did* report one); a wrong guess is worse than an honest gap.

---

## The three examples in this directory

| File | Platform | Fetch shape | Price representation | What it shows |
|---|---|---|---|---|
| `shopify-products-json.toml` | Shopify storefront `/products/{handle}.js` | Path-template JSON endpoint, no auth | `minor_int` — already integer cents | Per-variant `[iterate]`; a field Shopify frequently hides (`inventory_quantity`) handled as an honest `Option`, not a guess; `currency` **not present** in the payload at all — a real gap the manifest documents rather than papers over. |
| `generic-jsonld.toml` | Any site embedding schema.org `Product`/`Offer` JSON-LD | HTML page, extract embedded `<script type="application/ld+json">` | `major_decimal` — schema.org gives `"19.99"` as a string in major units | `[availability_map]` translating schema.org's `ItemAvailability` vocabulary (11 values) down to Vuna's 4-value `Availability` enum; `[currency_exponents]` override table for non-cent currencies. |
| `woocommerce-store-api.toml` | WooCommerce Store API `/wp-json/wc/store/v1/products` | Query-param JSON endpoint, returns an array directly | `minor_int_with_exponent` — the API states its own `currency_minor_unit` | A third, different price shape (explicit exponent in-payload, no guessing needed); query-param matching instead of a path template; boolean-plus-nullable-count availability shape. |

Three different endpoint shapes, three different price representations, one manifest format and
one set of `RetailObservation` fields underneath. That's the point: the Rust surface doesn't grow
when the fourth, fifth, and hundredth storefront platform show up — only this directory does.

## What this does not (yet) handle

- **No secondary lookups.** `shopify-products-json.toml`'s missing `currency` is a real gap the
  manifest format can't close by itself today — it needs either a static per-store override
  supplied at subscription time, or a second fetch (e.g. the theme's locale) the format doesn't
  express yet.
- **No multi-offer aggregation.** `generic-jsonld.toml` assumes one `Offer` per `Product`; a page
  with an `AggregateOffer` (a price *range* across variants) needs a format extension, not covered
  here.
- **No adversarial-input hardening spec.** These manifests describe the happy path; the interpreter
  that reads them (once built) is the thing responsible for bounding array sizes, rejecting
  malformed JSON, and not trusting a hostile page's declared `content_type`.
