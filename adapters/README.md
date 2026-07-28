# Vuna retail adapters — the long tail of sites is data, not Rust

The retail vertical (`retail` in `vuna_core::extract::ExtractorKind`) reads price/availability
signal off structured data that stores already publish — JSON-LD, Open Graph, or a store
platform's own JSON endpoints (`extract::RetailMethod`). Every store does this slightly
differently: different endpoint shapes, different field names, different price representations
(integer cents vs. decimal major-unit strings vs. an explicit minor-unit exponent), different
availability vocabularies.

Writing a new Rust `Extractor` impl for every storefront platform would make `vuna-extract`'s Rust
surface grow without bound and turn every new site into a reviewed, compiled, released change.
Instead, one interpreter in `vuna-extract` (`src/adapter/`) reads a **declarative manifest per
site/platform**, and that manifest is the only thing that grows as coverage grows. A manifest is
data: reviewable in a PR diff, safe to accept from a contributor who has never touched Rust, and —
matching the frontier's `UrlList` model — distributable as an ordinary signed object rather than a
code release.

This directory holds the manifest **format** and the three worked examples the interpreter is
tested against.

## What is and is not wired up

The interpreter is real: `vuna_extract::AdapterExtractor` parses a manifest, validates it, and
implements `vuna_core::extract::Extractor`, so a `.toml` file here drives extraction from a fetched
page with no Rust change. Every file in this directory is exercised against realistic payloads by
`crates/vuna-extract/tests/adapter_corpus.rs`, and that test fails if a manifest is added without a
fixture.

What is **not** wired up: `vuna-node`, the daemon that would subscribe to URL lists and feed pages
through a registry, is still a Wave-2 stub (see the root README's status table). So adapters run
under test and under any caller that builds a registry — they are not yet running against the live
web, because nothing is yet crawling the live web. Loading them is an explicit opt-in:

```rust
let mut registry = ExtractorRegistry::with_defaults();     // never touches the filesystem
for adapter in AdapterExtractor::load_dir(Path::new("adapters"))? {
    registry.register(Box::new(adapter));
}
```

`ExtractorRegistry::with_defaults()` does not read this directory, and `vuna-extract` builds and
passes its unit tests with `adapters/` deleted. A manifest can only ever *add* coverage; it can
never become something the build depends on.

---

## Manifest shape

Each `.toml` file describes one adapter: how to recognize a page it applies to, how to read the
structured data, and how to map that data onto `vuna_core::extract::RetailObservation`.

```toml
[adapter]
name    = "..."          # unique, kebab-case, MUST match the filename
version = 1              # the format version; the interpreter refuses anything else
extractor_kind = "retail" # the only kind with an interpreter today
method  = "..."           # one of: structured_data | availability_flag | json_endpoint | cart_probe | html
                          # — maps 1:1 onto extract::RetailMethod (serde snake_case)

[match]
url_pattern  = "..."      # template with {placeholder} captures, e.g. "https://{store}/products/{handle}.js"
content_type = "..."      # optional; matched against the declared type AND the body's actual shape
requires     = [...]      # optional: literal substrings that must appear in the head of the body
# At least one of url_pattern / content_type is required — an adapter with neither would claim
# every page the node fetches.

[fetch]
endpoint    = "..."       # MUST be "{page_url}" or the same string as url_pattern (see below)
http_method = "GET"       # optional, defaults to GET
format      = "json" | "html"
embedded    = "jsonld:Product"  # required for format = "html", forbidden for format = "json"

[iterate]
path = "..."              # optional: a JSON array path; one RetailObservation emitted per element
                          # "variants[]" for a named field, "$[]" for the document root itself

[fields]
# RetailObservation field -> source expression. Four expression kinds:
#   url:{capture}        — a {placeholder} the url_pattern captured, or an implicit capture
#   json:$.path.to.field — JSON field path, relative to the iterate element if [iterate] is set
#   const:VALUE          — a fixed value
#   const:none           — explicitly unresolved; the honest answer for a field the source lacks
store         = "..."
sku           = "..."
sku_fallback  = "..."     # optional: used when sku is absent or an empty string
availability  = "..."     # resolved value looked up in [availability_map] below
quantity      = "..."     # optional — RetailObservation.quantity is Option<u64>; absent path = None, never a guess
price_minor   = "..."
currency      = "..."
price_representation = "minor_int" | "major_decimal" | "minor_int_with_exponent"
price_exponent = "..."    # required for, and only meaningful for, minor_int_with_exponent
# minor_int              — source value is already integer minor units (e.g. Shopify cents), no conversion
# major_decimal          — source is a decimal string in major units; needs a currency → exponent lookup
# minor_int_with_exponent — source gives both the integer value AND its own exponent explicitly

[availability_map]
# raw source value (as a string key) -> one of: in_stock | out_of_stock | low_stock | unknown
# A JSON boolean is keyed as "true" / "false", so a bool field maps with the same table shape.

[currency_exponents]
# optional overrides for major_decimal conversion; default assumed = 2 (cents) when unlisted.
# Real exceptions exist — JPY has no minor unit (0), some Gulf-state currencies use 3 — so this
# is a table, not a constant, on purpose.

[notes]
# free text: caveats a future adapter author or reviewer needs, not machine-read.
```

Every field maps directly onto a real `vuna-core` type, so the manifest says everything
`RetailObservation` needs:

```rust
pub struct RetailObservation {
    pub store: String,
    pub sku: String,
    pub availability: Availability,          // <- [fields].availability + [availability_map]
    pub quantity: Option<u64>,               // <- [fields].quantity (optional path -> None if absent)
    pub price_minor: Option<i64>,            // <- [fields].price_minor + price_representation
    pub currency: Option<String>,
    pub method: RetailMethod,                // <- [adapter].method
    pub observed_at: UnixSecs,               // <- the page's fetched_at, not the manifest
}
```

`RetailObservation.quantity` and `.price_minor` are already `Option` in the contract — a manifest
that can't find a field for a given store leaves it unresolved and never fabricates a number. This
matters because `quorum::reconcile` treats "no quantity reported" as a legitimate outcome (a median
only forms over observers who *did* report one); a wrong guess is worse than an honest gap.

### Implicit captures

Two captures are always available to `url:{…}` without being declared in `url_pattern`:

| Capture | Value |
|---|---|
| `{host}` | the fetched URL's host — what `store` is normally keyed on |
| `{page_url}` | the fetched URL, verbatim |

### What a manifest error looks like

Validation runs at load, not at extraction, and every failure names the field. A manifest is
rejected — never silently degraded — when it declares an unsupported `version` or
`extractor_kind`, when `[match]` narrows nothing, when a `url:{name}` expression names a capture no
pattern produces, when `price_minor` is set without a `price_representation` (or
`minor_int_with_exponent` without a `price_exponent`), when `format`/`embedded` disagree, or when
the TOML contains **any unknown key**. That last one is deliberate: a misspelled
`price_represention` that serde quietly ignored would be indistinguishable from a correct manifest
until the prices came out wrong.

### Why `[fetch].endpoint` must be the page itself

An `Extractor` is handed a page `vuna-crawl` already fetched; it cannot go and issue a different
request. So `endpoint` is validated to be either `{page_url}` or exactly the `url_pattern`, and a
manifest naming a derived URL is refused at load rather than having that line quietly ignored.
Which URLs get fetched at all is the frontier's job, exactly as it is for the built-in extractors:
an adapter for `https://{store}/products/{handle}.js` only ever sees such a page if a `UrlList`
contains it.

---

## The three examples in this directory

| File | Platform | Fetch shape | Price representation | What it shows |
|---|---|---|---|---|
| `shopify-products-json.toml` | Shopify storefront `/products/{handle}.js` | Path-template JSON endpoint, no auth | `minor_int` — already integer cents | Per-variant `[iterate]`; a field Shopify frequently hides (`inventory_quantity`) handled as an honest `Option`, not a guess; `currency` **not present** in the payload at all — a real gap the manifest documents rather than papers over. |
| `generic-jsonld.toml` | Any site embedding schema.org `Product`/`Offer` JSON-LD | HTML page, `embedded = "jsonld:Product"` | `major_decimal` — schema.org gives `"19.99"` as a string in major units | `[availability_map]` translating schema.org's `ItemAvailability` vocabulary (11 values) down to Vuna's 4-value `Availability` enum; `[currency_exponents]` override table for non-cent currencies. |
| `woocommerce-store-api.toml` | WooCommerce Store API `/wp-json/wc/store/v1/products` | Query-param JSON endpoint, returns an array directly | `minor_int_with_exponent` — the API states its own `currency_minor_unit` | A third, different price shape (explicit exponent in-payload, no guessing needed); query-param matching instead of a path template; boolean-plus-nullable-count availability shape. |

Three different endpoint shapes, three different price representations, one manifest format and
one set of `RetailObservation` fields underneath. That's the point: the Rust surface doesn't grow
when the fourth, fifth, and hundredth storefront platform show up — only this directory does.

## Rules the interpreter enforces, so a manifest author doesn't have to

- **Prices are integer arithmetic on the source's own digits, never `f64`.** Binary floating point
  cannot represent `19.99`; `19.99 * 100.0` truncates to `1998`. Anything the interpreter cannot
  convert exactly (a thousands separator, a currency symbol, an exponent-notation float) resolves
  to `None` rather than to a number that is quietly one cent wrong across a whole corpus.
- **A stated exponent beats an assumed one.** For `minor_int_with_exponent`, the payload's own
  exponent is used as-is unless the manifest explicitly declares a canonical for that currency in
  `[currency_exponents]`. A JPY payload correctly saying `currency_minor_unit = 0` is never
  rescaled up to the assumed 2.
- **An unmapped availability value is `unknown`, never `in_stock`.** `[availability_map]` is an
  exact lookup; a bare `"InStock"` without the schema.org prefix, a typo, or a custom vocabulary
  all fall through to `unknown`.
- **A node with no resolvable SKU is dropped, not keyed on something invented.** `(store, sku)` is
  the join key `quorum::reconcile` groups on; an observation nobody else's can be matched against
  is worse than no observation.
- **The declared `content_type` can only rule an adapter *out*.** It is never trusted to decide how
  a body is parsed — a page claiming `application/json` while serving HTML fails the body-shape
  check and the adapter is not applied.
- **Observations from one page are capped** (`adapter::MAX_OBSERVATIONS`), so a hostile or broken
  endpoint returning a million-element array cannot turn one fetch into a million index writes.
- **A body that doesn't parse fails this adapter alone.** `ExtractorRegistry::extract_all` runs
  every other applicable extractor regardless, so a page one adapter chokes on is still indexed as
  a web doc.

## What this does not (yet) handle

- **No secondary lookups.** `shopify-products-json.toml`'s missing `currency` is a real gap the
  manifest format can't close by itself — it needs either a static per-store override supplied at
  subscription time, or a second fetch (e.g. the theme's locale) that neither the format nor the
  `Extractor` seam expresses.
- **No derived fetches at all**, for the same reason: see "Why `[fetch].endpoint` must be the page
  itself" above.
- **No multi-offer aggregation.** `generic-jsonld.toml` assumes one `Offer` per `Product`; a page
  with an `AggregateOffer` (a price *range* across variants), or a `Product` whose `offers` is an
  array, resolves to no price rather than to a wrong one. Handling it needs a format extension.
- **No templated expressions.** `const:` is a fixed value and `json:` is a bare path, so a
  synthetic-but-stable identifier like `shopify:{store}:{id}` cannot be written — see
  `shopify-products-json.toml`'s `[notes].sku_fallback` for the consequence.
- **No selector beyond `jsonld:<@type>`.** Microdata, RDFa, and CSS-selector extraction from
  rendered HTML are not expressible; a site that publishes only those needs a Rust extractor.
- **`[adapter].method` is recorded, not enforced.** It is written straight through to
  `RetailObservation.method` as the manifest author's declaration of how the reading was obtained;
  the interpreter does not cross-check it against `[fetch].format`.
