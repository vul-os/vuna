//! vuna-extract — pluggable extractors implementing [`vuna_core::extract::Extractor`].
//! First two verticals: `web` (chunks+links+snippet) and `retail` (JSON-LD/OpenGraph → observations).
//! [`ExtractorRegistry`] is how `vuna-node` wires up "which extractors run on which pages" — the
//! same opt-in plurality as embedding spaces (see `vuna-core` crate docs).

use vuna_core::extract::FetchedPage;

pub mod registry;
pub mod retail;
pub mod web;

pub use registry::ExtractorRegistry;
pub use retail::RetailExtractor;
pub use web::WebExtractor;

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
