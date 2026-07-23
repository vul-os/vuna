//! URL canonicalization: a deterministic string form so two nodes independently derive the same
//! [`UrlId`] for "the same" URL. Kept deliberately simple (and documented, not exhaustive) — this
//! is a hygiene pass, not a full dedup oracle. Rules, in order:
//!
//! 1. `raw` must be an absolute URL (have a scheme) — parsed with the `url` crate, which follows
//!    the WHATWG URL Standard. Relative URLs are a caller bug (extractors resolve links against
//!    their page's base before handing them to the frontier) and are rejected as an error.
//! 2. Scheme and host are lowercased. The `url` crate already lowercases ASCII hostnames and the
//!    scheme during parsing, so step 2 falls out of parsing for free — it is not a separate pass.
//! 3. The fragment (`#...`) is dropped. It is resolved client-side and never changes what the
//!    server returns, so `https://x/y#a` and `https://x/y#b` are the same crawl target.
//! 4. An explicit port equal to the scheme's well-known default is dropped, so
//!    `http://example.com:80/p` and `http://example.com/p` collide to the same id.
//! 5. Everything else — path, query string (including parameter order), trailing slash — is left
//!    **as-is**. Query-param sorting and trailing-slash folding are common follow-ups; they are
//!    skipped here because they can introduce false-positive collisions (e.g. `?page=1` vs
//!    `?tab=1` are genuinely different pages) — worth revisiting once a real corpus shows the
//!    win outweighs the risk.
//!
//! The canonical string is then hashed with BLAKE3-256 to produce the [`UrlId`] — the same hash
//! KOTVA uses for content addressing.

use vuna_core::frontier::UrlId;
use vuna_core::{Error, Result};

/// Canonicalize a URL string per the rules above. Errors if `raw` is not an absolute, parseable
/// URL.
pub fn canonicalize(raw: &str) -> Result<String> {
    let mut url = url::Url::parse(raw)
        .map_err(|e| Error::Frontier(format!("invalid URL {raw:?}: {e}")))?;

    url.set_fragment(None);

    if let Some(port) = url.port() {
        let default_port = match url.scheme() {
            "http" => Some(80),
            "https" => Some(443),
            "ftp" => Some(21),
            _ => None,
        };
        if Some(port) == default_port {
            // Only fails for cannot-be-a-base URLs (no host), which can't reach here since
            // `.port()` just returned `Some` — a host was required to parse a port at all.
            let _ = url.set_port(None);
        }
    }

    Ok(url.to_string())
}

/// Hash a canonical URL string into its stable [`UrlId`] (BLAKE3-256 of the UTF-8 bytes).
pub fn url_id(canonical: &str) -> UrlId {
    vuna_core::ContentId(*blake3::hash(canonical.as_bytes()).as_bytes())
}

/// Convenience: canonicalize then hash in one step — what `discover`/`subscribe` use.
pub fn canonical_id(raw: &str) -> Result<(String, UrlId)> {
    let canon = canonicalize(raw)?;
    let id = url_id(&canon);
    Ok((canon, id))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn lowercases_host_and_drops_default_port_and_fragment() {
        let a = canonicalize("https://Example.COM:443/Path?x=1#frag").unwrap();
        let b = canonicalize("https://example.com/Path?x=1").unwrap();
        assert_eq!(a, b);
    }

    #[test]
    fn keeps_non_default_port() {
        let c = canonicalize("http://example.com:8080/p").unwrap();
        assert!(c.contains(":8080"));
    }

    #[test]
    fn rejects_relative_urls() {
        assert!(canonicalize("/just/a/path").is_err());
    }

    #[test]
    fn same_canonical_form_yields_same_id() {
        let (_, id_a) = canonical_id("https://example.com:443/p#a").unwrap();
        let (_, id_b) = canonical_id("https://EXAMPLE.com/p#b").unwrap();
        assert_eq!(id_a, id_b);
    }
}
