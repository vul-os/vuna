//! URL templates with named captures — the `[match].url_pattern` half of a manifest.
//!
//! A pattern is a literal URL with `{name}` placeholders: `https://{store}/products/{handle}.js`.
//! A placeholder matches one or more characters that are **not** `/`, `?`, `&`, or `#`, so
//! `{store}` can never swallow a path segment and `{handle}` can never swallow a query string.
//! Everything a placeholder matches becomes a capture the `url:{name}` field expressions read.
//!
//! Patterns with a query string (`…/products?sku={sku}`) are split: the path half is matched as a
//! template, the query half as a set of **required** parameters matched in any order. Extra query
//! parameters on the fetched URL are ignored — real URLs carry tracking params, and a positional
//! match would reject a page the adapter genuinely applies to.

use std::collections::BTreeMap;

/// Characters a `{capture}` is never allowed to span.
const CAPTURE_STOPS: [char; 4] = ['/', '?', '&', '#'];

#[derive(Clone, Debug, PartialEq, Eq)]
enum Token {
    Lit(String),
    Cap(String),
}

/// A compiled `[match].url_pattern`.
#[derive(Clone, Debug)]
pub struct UrlPattern {
    path: Vec<Token>,
    /// Required query parameters: name -> literal value or a capture.
    query: Vec<(String, Token)>,
}

/// Why a pattern string could not be compiled. Surfaced at manifest-load time, never at match time.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum PatternError {
    UnclosedPlaceholder,
    EmptyPlaceholder,
    AdjacentPlaceholders,
    /// A `?k=v` pair with no `=`, which would be ambiguous to match.
    MalformedQueryPair(String),
}

impl std::fmt::Display for PatternError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::UnclosedPlaceholder => write!(f, "unclosed '{{' in url_pattern"),
            Self::EmptyPlaceholder => write!(f, "empty '{{}}' placeholder in url_pattern"),
            Self::AdjacentPlaceholders => {
                write!(f, "two placeholders with no literal between them are ambiguous")
            }
            Self::MalformedQueryPair(p) => write!(f, "query parameter {p:?} has no '='"),
        }
    }
}

impl UrlPattern {
    /// Compiles a pattern string. Fails loudly on the shapes that could only ever match by
    /// accident, so a typo in a manifest is a load error rather than an adapter that silently
    /// never applies.
    pub fn compile(pattern: &str) -> Result<Self, PatternError> {
        let (path_part, query_part) = match pattern.find('?') {
            Some(i) => (&pattern[..i], Some(&pattern[i + 1..])),
            None => (pattern, None),
        };

        let path = tokenize(path_part)?;
        if path.windows(2).any(|w| matches!((&w[0], &w[1]), (Token::Cap(_), Token::Cap(_)))) {
            return Err(PatternError::AdjacentPlaceholders);
        }

        let mut query = Vec::new();
        if let Some(q) = query_part {
            for pair in q.split('&').filter(|p| !p.is_empty()) {
                let (k, v) = pair.split_once('=').ok_or_else(|| PatternError::MalformedQueryPair(pair.to_string()))?;
                let mut toks = tokenize(v)?;
                // A query value is either exactly one capture or a plain literal; a mixed value
                // (`v{cap}x`) is not something any real endpoint needs, so it is not supported.
                let tok = match toks.len() {
                    1 => toks.remove(0),
                    _ => Token::Lit(v.to_string()),
                };
                query.push((k.to_string(), tok));
            }
        }

        Ok(Self { path, query })
    }

    /// The capture names this pattern can produce, in declaration order.
    pub fn capture_names(&self) -> Vec<&str> {
        self.path
            .iter()
            .chain(self.query.iter().map(|(_, t)| t))
            .filter_map(|t| match t {
                Token::Cap(name) => Some(name.as_str()),
                Token::Lit(_) => None,
            })
            .collect()
    }

    /// Matches `url`, returning the captures on success. `None` means "this adapter does not apply
    /// to this URL" — never an error, since non-matching is the common case.
    pub fn matches(&self, url: &str) -> Option<BTreeMap<String, String>> {
        let (url_path, url_query) = match url.find('?') {
            Some(i) => (&url[..i], &url[i + 1..]),
            None => (url, ""),
        };
        // A fragment is client-side only and never reaches the server; strip it before matching.
        let url_path = url_path.split('#').next().unwrap_or(url_path);

        let mut captures = match_path(&self.path, url_path)?;

        if !self.query.is_empty() {
            let actual: BTreeMap<&str, &str> = url_query
                .split('&')
                .filter(|p| !p.is_empty())
                .filter_map(|p| p.split_once('='))
                .collect();
            for (key, expected) in &self.query {
                let got = actual.get(key.as_str())?;
                match expected {
                    Token::Lit(lit) => {
                        if got != lit {
                            return None;
                        }
                    }
                    Token::Cap(name) => {
                        if got.is_empty() {
                            return None;
                        }
                        captures.insert(name.clone(), (*got).to_string());
                    }
                }
            }
        }

        Some(captures)
    }
}

fn tokenize(s: &str) -> Result<Vec<Token>, PatternError> {
    let mut out = Vec::new();
    let mut rest = s;
    while !rest.is_empty() {
        match rest.find('{') {
            None => {
                out.push(Token::Lit(rest.to_string()));
                break;
            }
            Some(open) => {
                if open > 0 {
                    out.push(Token::Lit(rest[..open].to_string()));
                }
                let close = rest[open..].find('}').ok_or(PatternError::UnclosedPlaceholder)? + open;
                let name = &rest[open + 1..close];
                if name.is_empty() {
                    return Err(PatternError::EmptyPlaceholder);
                }
                out.push(Token::Cap(name.to_string()));
                rest = &rest[close + 1..];
            }
        }
    }
    Ok(out)
}

fn match_path(tokens: &[Token], url: &str) -> Option<BTreeMap<String, String>> {
    let mut captures = BTreeMap::new();
    let mut cursor = 0usize;

    for (i, token) in tokens.iter().enumerate() {
        match token {
            Token::Lit(lit) => {
                if !url[cursor..].starts_with(lit.as_str()) {
                    return None;
                }
                cursor += lit.len();
            }
            Token::Cap(name) => {
                let taken = match tokens.get(i + 1) {
                    // Trailing capture: it takes the rest of the path.
                    None => &url[cursor..],
                    Some(Token::Lit(next)) => {
                        // The FIRST occurrence of the following literal ends the capture. A later
                        // occurrence would mean the capture spans a stop character, which is never
                        // legal, so there is nothing to backtrack to.
                        let idx = url[cursor..].find(next.as_str())?;
                        &url[cursor..cursor + idx]
                    }
                    // Rejected at compile time.
                    Some(Token::Cap(_)) => return None,
                };
                if taken.is_empty() || taken.contains(CAPTURE_STOPS) {
                    return None;
                }
                captures.insert(name.clone(), taken.to_string());
                cursor += taken.len();
            }
        }
    }

    // The pattern must consume the whole path — a prefix match would let
    // `https://{store}/products/{handle}.js` claim `…/widget.js.bak`.
    (cursor == url.len()).then_some(captures)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn caps(pattern: &str, url: &str) -> Option<BTreeMap<String, String>> {
        UrlPattern::compile(pattern).unwrap().matches(url)
    }

    #[test]
    fn path_template_captures_host_and_handle() {
        let c = caps("https://{store}/products/{handle}.js", "https://shop.example.com/products/blue-widget.js").unwrap();
        assert_eq!(c["store"], "shop.example.com");
        assert_eq!(c["handle"], "blue-widget");
    }

    #[test]
    fn capture_never_spans_a_path_separator() {
        // `{handle}` would have to swallow "a/b" to match — it must not.
        assert!(caps("https://{store}/products/{handle}.js", "https://shop.example.com/products/a/b.js").is_none());
    }

    #[test]
    fn trailing_garbage_is_not_a_match() {
        assert!(caps("https://{store}/products/{handle}.js", "https://shop.example.com/products/w.js.bak").is_none());
    }

    #[test]
    fn empty_capture_is_not_a_match() {
        assert!(caps("https://{store}/products/{handle}.js", "https://shop.example.com/products/.js").is_none());
    }

    #[test]
    fn query_params_match_in_any_order_and_ignore_extras() {
        let c = caps(
            "https://{site}/wp-json/wc/store/v1/products?sku={sku}",
            "https://store.example/wp-json/wc/store/v1/products?utm_source=x&sku=ABC-1",
        )
        .unwrap();
        assert_eq!(c["site"], "store.example");
        assert_eq!(c["sku"], "ABC-1");
    }

    #[test]
    fn a_missing_required_query_param_is_not_a_match() {
        assert!(caps(
            "https://{site}/wp-json/wc/store/v1/products?sku={sku}",
            "https://store.example/wp-json/wc/store/v1/products?page=2",
        )
        .is_none());
    }

    #[test]
    fn fragment_is_stripped_before_matching() {
        let c = caps("https://{store}/products/{handle}.js", "https://shop.example.com/products/w.js#reviews").unwrap();
        assert_eq!(c["handle"], "w");
    }

    #[test]
    fn capture_names_are_reported_for_manifest_validation() {
        let p = UrlPattern::compile("https://{site}/wp-json/wc/store/v1/products?sku={sku}").unwrap();
        assert_eq!(p.capture_names(), vec!["site", "sku"]);
    }

    #[test]
    fn ambiguous_and_malformed_patterns_fail_to_compile() {
        assert_eq!(UrlPattern::compile("https://{a/x").unwrap_err(), PatternError::UnclosedPlaceholder);
        assert_eq!(UrlPattern::compile("https://{}/x").unwrap_err(), PatternError::EmptyPlaceholder);
        assert_eq!(UrlPattern::compile("https://{a}{b}/x").unwrap_err(), PatternError::AdjacentPlaceholders);
        assert!(matches!(
            UrlPattern::compile("https://x/y?sku").unwrap_err(),
            PatternError::MalformedQueryPair(_)
        ));
    }
}
