//! The `web` vertical: HTML → chunks (RAG/search), a link graph, and display metadata.
//! Boilerplate (script/style/nav/header/footer/aside/…) is walked around, not deleted from the
//! DOM — `scraper`'s tree is immutable, so we simply skip those subtrees while collecting text.

use scraper::{ElementRef, Html, Node, Selector};
use url::Url;
use vuna_core::extract::{Chunk, Extraction, Extractor, ExtractorKind, FetchedPage, Link, WebDoc};
use vuna_core::{Error, Result};

use crate::normalize_whitespace;

/// Target chunk size. ~256 words is a common sweet spot for retrieval chunking: small enough for
/// precise recall, large enough to keep a sentence's context intact.
const WORDS_PER_CHUNK: usize = 256;

/// Roughly how many characters of visible text/description to keep as the display snippet.
const SNIPPET_CHARS: usize = 200;

/// Elements whose text is layout/navigation noise, not article content, and whose entire subtree
/// is skipped when collecting visible text. Their `<a href>`s are still walked for the link graph
/// — a nav or footer link is a perfectly real edge, it's just not *content*.
const BOILERPLATE_TAGS: &[&str] = &[
    "script", "style", "noscript", "template", "nav", "header", "footer", "aside", "svg",
    "iframe", "object", "embed", "noembed",
];

/// Parses a page into a [`WebDoc`]: title, snippet, ~256-word chunks, and the outbound link graph.
#[derive(Clone, Copy, Debug, Default)]
pub struct WebExtractor;

impl Extractor for WebExtractor {
    fn kind(&self) -> ExtractorKind {
        "web".to_string()
    }

    fn applies(&self, page: &FetchedPage) -> bool {
        crate::looks_like_html(page)
    }

    fn extract(&self, page: &FetchedPage) -> Result<Extraction> {
        let html = String::from_utf8_lossy(&page.body);
        let document = Html::parse_document(&html);

        let base = Url::parse(&page.url).map_err(|e| Error::Extract {
            kind: self.kind(),
            url: page.url.clone(),
            reason: format!("page url is not a valid absolute url: {e}"),
        })?;

        let title = extract_title(&document);
        let cleaned = normalize_whitespace(&collect_visible_text(&document));
        let snippet = build_snippet(&document, &cleaned);
        let chunks = chunk_text(&cleaned);
        let (links, discovered) = extract_links(&document, &base);

        Ok(Extraction::Web(WebDoc {
            url: page.url.clone(),
            title,
            snippet,
            chunks,
            links,
            discovered,
        }))
    }
}

/// `<title>` wins when present and non-empty; `og:title` is the fallback for pages that skip it.
fn extract_title(document: &Html) -> Option<String> {
    let title_sel = Selector::parse("title").expect("static selector");
    if let Some(el) = document.select(&title_sel).next() {
        let text = normalize_whitespace(&el.text().collect::<String>());
        if !text.is_empty() {
            return Some(text);
        }
    }
    meta_content(document, "og:title")
}

/// Prefers the author-curated `meta[name=description]` (that's what it's for); falls back to the
/// head of the cleaned visible text for pages that never set one.
fn build_snippet(document: &Html, cleaned_text: &str) -> String {
    if let Some(desc) = meta_content(document, "description") {
        if !desc.is_empty() {
            return truncate_chars(&desc, SNIPPET_CHARS);
        }
    }
    truncate_chars(cleaned_text, SNIPPET_CHARS)
}

/// Looks up a `<meta>` tag by either `name=` or `property=` (covers both plain meta and
/// Open Graph/Twitter-card conventions) and returns its trimmed `content`.
fn meta_content(document: &Html, key: &str) -> Option<String> {
    let sel = Selector::parse("meta").expect("static selector");
    document.select(&sel).find_map(|el| {
        let matches = el.attr("name") == Some(key) || el.attr("property") == Some(key);
        if !matches {
            return None;
        }
        el.attr("content").map(str::trim).filter(|c| !c.is_empty()).map(str::to_string)
    })
}

/// Truncates to at most `max` chars, backing off to the last word boundary so words aren't cut
/// mid-way, and marks truncation with `…`. No-op (verbatim) when already short enough.
fn truncate_chars(s: &str, max: usize) -> String {
    if s.chars().count() <= max {
        return s.to_string();
    }
    let mut out: String = s.chars().take(max).collect();
    if let Some(idx) = out.rfind(' ') {
        out.truncate(idx);
    }
    out.push('…');
    out
}

/// Depth-first walk of the document's `<body>` (or the whole tree if there's somehow no body),
/// concatenating text nodes and skipping any subtree rooted at a [`BOILERPLATE_TAGS`] element.
fn collect_visible_text(document: &Html) -> String {
    let body_sel = Selector::parse("body").expect("static selector");
    let root = document.select(&body_sel).next().unwrap_or_else(|| document.root_element());
    let mut out = String::new();
    walk_text(root, &mut out);
    out
}

fn walk_text(el: ElementRef<'_>, out: &mut String) {
    if BOILERPLATE_TAGS.contains(&el.value().name()) {
        return;
    }
    for child in el.children() {
        match child.value() {
            Node::Text(text) => {
                out.push_str(text);
                out.push(' ');
            }
            Node::Element(_) => {
                if let Some(child_el) = ElementRef::wrap(child) {
                    walk_text(child_el, out);
                }
            }
            _ => {}
        }
    }
}

/// Splits already-cleaned, whitespace-normalized text into ~[`WORDS_PER_CHUNK`]-word chunks with
/// stable 0-based ordinals. Empty text yields zero chunks (nothing to embed).
fn chunk_text(cleaned: &str) -> Vec<Chunk> {
    let words: Vec<&str> = cleaned.split_whitespace().collect();
    words
        .chunks(WORDS_PER_CHUNK)
        .enumerate()
        .map(|(ordinal, words)| Chunk { ordinal: ordinal as u32, text: words.join(" ") })
        .collect()
}

/// Walks every `<a href>` in the document (boilerplate included — the link graph doesn't care
/// where on the page a link lives), resolving each to an absolute URL against `base`. Returns the
/// full link list plus a deduped `http(s)`-only projection ready to feed the frontier.
fn extract_links(document: &Html, base: &Url) -> (Vec<Link>, Vec<String>) {
    let sel = Selector::parse("a[href]").expect("static selector");
    let mut links = Vec::new();
    let mut discovered = Vec::new();
    let mut seen = std::collections::HashSet::new();

    for el in document.select(&sel) {
        let Some(href) = el.attr("href").map(str::trim) else { continue };
        if href.is_empty() || href.starts_with('#') || href.starts_with("javascript:") {
            continue;
        }
        let Ok(resolved) = base.join(href) else { continue };
        let anchor = normalize_whitespace(&el.text().collect::<String>());
        let rel = el.attr("rel").map(str::to_string);
        let to_url = resolved.to_string();

        links.push(Link { to_url: to_url.clone(), anchor, rel });

        if matches!(resolved.scheme(), "http" | "https") && seen.insert(to_url.clone()) {
            discovered.push(to_url);
        }
    }

    (links, discovered)
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
            fetched_at: 0,
        }
    }

    fn extract_web(url: &str, html: &str) -> WebDoc {
        match WebExtractor.extract(&page(url, html)).unwrap() {
            Extraction::Web(doc) => doc,
            other => panic!("expected Extraction::Web, got {other:?}"),
        }
    }

    /// Basic article: title/og fallback ordering, meta-description snippet, boilerplate skipped
    /// from the text but its links still recorded, relative + absolute links resolved.
    #[test]
    fn basic_article_page() {
        let html = r#"
            <html><head>
                <title>Example Article</title>
                <meta name="description" content="A short article about testing.">
            </head>
            <body>
                <nav><a href="/nav-link">Nav</a></nav>
                <script>var trackingPixel = true;</script>
                <article>
                    <h1>Example Article</h1>
                    <p>This is the first paragraph of real content.</p>
                    <p>This is a second paragraph with a <a href="/relative/path">relative link</a>
                       and an <a href="https://other.example.com/page">absolute link</a>.</p>
                </article>
                <footer><a href="/footer-link">Footer</a></footer>
            </body></html>
        "#;
        let page = page("https://example.com/articles/1", html);
        assert!(WebExtractor.applies(&page));
        let doc = extract_web("https://example.com/articles/1", html);

        assert_eq!(doc.title.as_deref(), Some("Example Article"));
        assert_eq!(doc.snippet, "A short article about testing.");

        // Boilerplate text is excluded from the chunked body...
        let all_text: String = doc.chunks.iter().map(|c| c.text.as_str()).collect();
        assert!(all_text.contains("first paragraph of real content"));
        assert!(!all_text.contains("Nav"));
        assert!(!all_text.contains("trackingPixel"));
        assert!(!all_text.contains("Footer"));

        // ...but nav/footer links are still part of the link graph, resolved absolute.
        let urls: Vec<&str> = doc.links.iter().map(|l| l.to_url.as_str()).collect();
        assert!(urls.contains(&"https://example.com/nav-link"));
        assert!(urls.contains(&"https://example.com/footer-link"));
        assert!(urls.contains(&"https://example.com/relative/path"));
        assert!(urls.contains(&"https://other.example.com/page"));
        assert!(doc.discovered.contains(&"https://example.com/relative/path".to_string()));

        let anchor = doc
            .links
            .iter()
            .find(|l| l.to_url == "https://other.example.com/page")
            .unwrap();
        assert_eq!(anchor.anchor, "absolute link");
    }

    /// Relative hrefs resolve against the *page* URL per RFC 3986, including `..` traversal.
    #[test]
    fn resolves_relative_links_against_page_url() {
        let html = r#"<html><body>
            <a href="../sibling">Sibling</a>
            <a href="deeper/page">Deeper</a>
        </body></html>"#;
        let doc = extract_web("https://example.com/blog/2024/post/", html);
        let urls: Vec<&str> = doc.links.iter().map(|l| l.to_url.as_str()).collect();
        assert!(urls.contains(&"https://example.com/blog/2024/sibling"));
        assert!(urls.contains(&"https://example.com/blog/2024/post/deeper/page"));
    }

    /// Fragment-only and javascript: hrefs are not real navigable links and are dropped.
    #[test]
    fn drops_fragment_and_javascript_hrefs() {
        let html = r##"<html><body>
            <a href="#top">Top</a>
            <a href="javascript:void(0)">Nope</a>
            <a href="/real">Real</a>
        </body></html>"##;
        let doc = extract_web("https://example.com/page", html);
        assert_eq!(doc.links.len(), 1);
        assert_eq!(doc.links[0].to_url, "https://example.com/real");
    }

    /// Chunk boundary correctness: 300 words split into exactly two chunks, 256 + 44, in order.
    #[test]
    fn splits_chunks_at_word_boundary() {
        let words: Vec<String> = (0..300).map(|i| format!("word{i}")).collect();
        let html = format!("<html><body><p>{}</p></body></html>", words.join(" "));
        let doc = extract_web("https://example.com/long", &html);

        assert_eq!(doc.chunks.len(), 2);
        assert_eq!(doc.chunks[0].ordinal, 0);
        assert_eq!(doc.chunks[1].ordinal, 1);
        assert_eq!(doc.chunks[0].text.split_whitespace().count(), 256);
        assert_eq!(doc.chunks[1].text.split_whitespace().count(), 44);
        assert!(doc.chunks[0].text.starts_with("word0 word1 word2"));
        assert!(doc.chunks[0].text.ends_with("word255"));
        assert!(doc.chunks[1].text.starts_with("word256"));
        assert!(doc.chunks[1].text.ends_with("word299"));
    }

    #[test]
    fn does_not_apply_to_non_html_content_type() {
        let p = FetchedPage {
            url: "https://example.com/data.json".to_string(),
            status: 200,
            content_type: Some("application/json".to_string()),
            body: b"{}".to_vec(),
            fetched_at: 0,
        };
        assert!(!WebExtractor.applies(&p));
    }
}
