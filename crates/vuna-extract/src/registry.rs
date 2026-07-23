//! The extractor registry: holds boxed [`Extractor`]s and selects the ones applicable to a page.
//! A single page commonly matches more than one — a shop's product page is both a web doc (for
//! RAG) and a source of retail observations (for the price/stock radar) — so selection returns
//! every match, not just the first, mirroring the opt-in-plurality model of embedding spaces.

use vuna_core::extract::{Extraction, Extractor, ExtractorKind, FetchedPage};
use vuna_core::Result;

#[derive(Default)]
pub struct ExtractorRegistry {
    extractors: Vec<Box<dyn Extractor>>,
}

impl ExtractorRegistry {
    pub fn new() -> Self {
        Self::default()
    }

    /// The default registry: `web` + `retail`, the two verticals this crate ships.
    pub fn with_defaults() -> Self {
        let mut registry = Self::new();
        registry.register(Box::new(crate::WebExtractor));
        registry.register(Box::new(crate::RetailExtractor));
        registry
    }

    pub fn register(&mut self, extractor: Box<dyn Extractor>) {
        self.extractors.push(extractor);
    }

    /// The registered extractors that declare themselves applicable to `page`, in registration
    /// order.
    pub fn applicable(&self, page: &FetchedPage) -> Vec<&dyn Extractor> {
        self.extractors.iter().filter(|e| e.applies(page)).map(Box::as_ref).collect()
    }

    /// Runs every applicable extractor and collects each result. One extractor's parse failure
    /// doesn't prevent the others from running — a page a retail extractor chokes on can still be
    /// indexed as a web doc.
    pub fn extract_all(&self, page: &FetchedPage) -> Vec<Result<Extraction>> {
        self.applicable(page).into_iter().map(|e| e.extract(page)).collect()
    }

    /// Kinds of every registered extractor, in registration order.
    pub fn kinds(&self) -> impl Iterator<Item = ExtractorKind> + '_ {
        self.extractors.iter().map(|e| e.kind())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use vuna_core::extract::Extraction;

    fn html_page(url: &str, body: &str) -> FetchedPage {
        FetchedPage {
            url: url.to_string(),
            status: 200,
            content_type: Some("text/html".to_string()),
            body: body.as_bytes().to_vec(),
            fetched_at: 0,
        }
    }

    #[test]
    fn default_registry_registers_web_and_retail() {
        let registry = ExtractorRegistry::with_defaults();
        let kinds: Vec<_> = registry.kinds().collect();
        assert_eq!(kinds, vec!["web".to_string(), "retail".to_string()]);
    }

    /// A page with a JSON-LD product is applicable to BOTH extractors — one page, two verticals.
    #[test]
    fn product_page_matches_both_extractors() {
        let registry = ExtractorRegistry::with_defaults();
        let html = r#"<html><head><title>Widget</title>
            <script type="application/ld+json">
            {"@type":"Product","sku":"W1","offers":{"@type":"Offer","price":"9.99","priceCurrency":"USD","availability":"InStock"}}
            </script>
        </head><body><p>A fine widget.</p></body></html>"#;
        let page = html_page("https://shop.example.com/w1", html);

        let applicable = registry.applicable(&page);
        assert_eq!(applicable.len(), 2);

        let results = registry.extract_all(&page);
        assert_eq!(results.len(), 2);
        let mut saw_web = false;
        let mut saw_retail = false;
        for r in results {
            match r.unwrap() {
                Extraction::Web(_) => saw_web = true,
                Extraction::Retail(obs) => {
                    saw_retail = true;
                    assert_eq!(obs.len(), 1);
                }
            }
        }
        assert!(saw_web && saw_retail);
    }

    /// A plain article page (no structured retail markup) only matches `web`.
    #[test]
    fn plain_article_matches_only_web() {
        let registry = ExtractorRegistry::with_defaults();
        let page = html_page("https://example.com/post", "<html><body><p>Just an article.</p></body></html>");
        let kinds: Vec<_> = registry.applicable(&page).iter().map(|e| e.kind()).collect();
        assert_eq!(kinds, vec!["web".to_string()]);
    }
}
