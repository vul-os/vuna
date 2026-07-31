//! vuna-extract — pluggable extractors implementing [`vuna_core::extract::Extractor`].
//! First two verticals: `web` (chunks+links+snippet) and `retail` (JSON-LD/OpenGraph → observations).
//! [`ExtractorRegistry`] is how `vuna-node` wires up "which extractors run on which pages" — the
//! same opt-in plurality as embedding spaces (see `vuna-core` crate docs).
//!
//! The retail vertical has a second, declarative surface: [`adapter`] interprets an
//! `adapters/*.toml` manifest as an [`Extractor`](vuna_core::extract::Extractor), so covering a new
//! storefront platform is a reviewed data change rather than a new Rust impl. Those adapters are
//! opt-in — [`ExtractorRegistry::with_defaults`] is filesystem-free and this crate builds and
//! passes its tests with no `adapters/` directory present.

use vuna_core::extract::FetchedPage;
use vuna_core::ContentId;

pub mod adapter;
pub mod registry;
pub mod retail;
pub mod web;

pub use adapter::{AdapterExtractor, AdapterManifest};
pub use registry::ExtractorRegistry;
pub use retail::RetailExtractor;
pub use web::WebExtractor;

/// Mints the **trust anchor** for an extraction: `BLAKE3-256` over the page's raw body bytes,
/// exactly as fetched. See [`vuna_core::extract::WebDoc::source_hash`] for what the value means and
/// [`vuna_core::trust`] for how it is checked.
///
/// The hash covers the body and *nothing else* — not the URL, not the status, not `fetched_at` —
/// so two nodes that receive the same bytes agree regardless of when or from where they fetched.
///
/// The call form here is deliberately identical to `vuna_frontier::canon::url_id`'s
/// (`ContentId(*blake3::hash(bytes).as_bytes())`): raw BLAKE3-256 of the input, no domain-separation
/// tag and no length framing. That keeps one hashing convention across the workspace. It also means
/// this shares BLAKE3's ordinary preimage domain with the URL hash — a `source_hash` and a `UrlId`
/// are the same 32 bytes if the same input is fed to both. They are never compared to each other,
/// but if a future object ever mixes them, that object must add its own domain separator rather
/// than assume these are distinguishable.
///
/// Honest scope: this is computed by the extractor, over bytes the extractor also reads. It does
/// not stop a dishonest node from publishing a hash of bytes it invented. What it buys is that an
/// honest party holding the same bytes can *recompute and disagree* — which is what turns a
/// symptomless poisoning into a detectable one.
pub fn source_hash(page: &FetchedPage) -> ContentId {
    ContentId(*blake3::hash(&page.body).as_bytes())
}

/// Cheap HTML sniff shared by both extractors' `applies()`. A declared `text/html`-ish
/// content-type is trusted outright; a declared *other* type rules HTML out; a missing type falls
/// back to peeking at the first bytes of the body for a doctype/tag.
pub(crate) fn looks_like_html(page: &FetchedPage) -> bool {
    if let Some(ct) = &page.content_type {
        let ct = ct.to_ascii_lowercase();
        if ct.contains("html") {
            return true;
        }
        if !ct.trim().is_empty() {
            return false;
        }
    }
    let head = &page.body[..page.body.len().min(512)];
    let head = String::from_utf8_lossy(head).to_ascii_lowercase();
    head.contains("<html") || head.contains("<!doctype")
}

/// Collapses all runs of whitespace (including newlines) to a single space and trims the ends.
/// Shared by both extractors so anchor text, titles, and body text are cleaned identically.
pub(crate) fn normalize_whitespace(s: &str) -> String {
    let mut out = String::with_capacity(s.len());
    let mut last_was_space = true; // swallow leading whitespace too
    for ch in s.chars() {
        if ch.is_whitespace() {
            if !last_was_space {
                out.push(' ');
                last_was_space = true;
            }
        } else {
            out.push(ch);
            last_was_space = false;
        }
    }
    if out.ends_with(' ') {
        out.pop();
    }
    out
}

#[cfg(test)]
mod source_hash_tests {
    use super::*;
    use vuna_core::extract::Extractor;
    use vuna_core::trust::TrustAnchored;

    fn page(url: &str, body: &[u8]) -> FetchedPage {
        FetchedPage {
            url: url.to_string(),
            status: 200,
            content_type: Some("text/html".to_string()),
            body: body.to_vec(),
            fetched_at: 0,
        }
    }

    /// The convention is pinned to **published BLAKE3 known-answer vectors**, not merely to
    /// whatever this workspace's `blake3` crate happens to return. A corpus that only agrees with
    /// the implementation that produced it proves the two match each other, not that either is
    /// right; these two values are the reference BLAKE3 outputs for the empty input and `abc`, and
    /// were cross-checked against an independent (non-Rust) BLAKE3 implementation.
    #[test]
    fn source_hash_matches_published_blake3_known_answer_vectors() {
        assert_eq!(
            source_hash(&page("https://example.com/", b"")).to_hex(),
            "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262",
            "BLAKE3-256 of the empty input"
        );
        assert_eq!(
            source_hash(&page("https://example.com/", b"abc")).to_hex(),
            "6437b3ac38465133ffb63b75273a8db548c558465d79db03fd359c6cd5bd9d85",
            "BLAKE3-256 of `abc`"
        );
    }

    /// The property the whole trust anchor rests on: two nodes that fetched the same bytes must
    /// derive the same value, even though everything *around* the bytes differs — different URL,
    /// different fetch time, different status. If any of that leaked into the hash, two honest
    /// observers of one page could never corroborate each other.
    #[test]
    fn identical_bytes_hash_identically_regardless_of_url_status_or_fetch_time() {
        let body = b"<html><body>stable document</body></html>";
        let a = FetchedPage {
            url: "https://example.com/a".into(),
            status: 200,
            content_type: Some("text/html".into()),
            body: body.to_vec(),
            fetched_at: 1_000,
        };
        let b = FetchedPage {
            url: "https://mirror.example.org/b".into(),
            status: 203,
            content_type: Some("application/xhtml+xml".into()),
            body: body.to_vec(),
            fetched_at: 9_999_999,
        };
        assert_eq!(source_hash(&a), source_hash(&b));
    }

    /// One flipped bit anywhere in the body must move the anchor. This is what makes a substituted
    /// page detectable rather than symptomless.
    #[test]
    fn a_single_flipped_bit_changes_the_anchor() {
        let base = b"<html><body>the quick brown fox</body></html>".to_vec();
        let original = source_hash(&page("https://example.com/", &base));
        for bit in 0..8 {
            let mut tampered = base.clone();
            let idx = tampered.len() / 2;
            tampered[idx] ^= 1 << bit;
            assert_ne!(
                source_hash(&page("https://example.com/", &tampered)),
                original,
                "flipping bit {bit} left the anchor unchanged"
            );
        }
        // Truncation and extension are movements too, not just substitutions.
        assert_ne!(source_hash(&page("https://example.com/", &base[..base.len() - 1])), original);
        let mut longer = base.clone();
        longer.push(b' ');
        assert_ne!(source_hash(&page("https://example.com/", &longer)), original);
    }

    /// An empty body is a real, hashable observation ("we fetched it and got nothing"), and it must
    /// NOT collide with the `ZERO` sentinel that means "no anchor at all". If it did, every empty
    /// page would arrive looking unanchored — and, worse, an unanchored forgery would be
    /// indistinguishable from an honest fetch of an empty document.
    #[test]
    fn an_empty_body_hashes_to_a_real_value_not_the_zero_sentinel() {
        let h = source_hash(&page("https://example.com/", b""));
        assert_ne!(h, ContentId::ZERO);
    }

    /// The extractor must never emit a document that only *looks* anchored. Every real extraction
    /// carries an anchor that the trust seam accepts.
    #[test]
    fn every_web_extraction_carries_an_anchor_the_trust_seam_accepts() {
        let html = b"<html><head><title>T</title></head><body><p>hello</p></body></html>";
        let p = page("https://example.com/doc", html);
        let doc = match web::WebExtractor.extract(&p).unwrap() {
            vuna_core::extract::Extraction::Web(d) => d,
            other => panic!("expected Web, got {other:?}"),
        };
        assert!(doc.is_anchored(), "extractor emitted an unanchored WebDoc");
        assert_eq!(doc.anchor(), Some(source_hash(&p)));
        // …and the anchor is over the bytes, so it is reproducible from the page alone.
        assert_eq!(doc.source_hash, source_hash(&p));
    }
}
