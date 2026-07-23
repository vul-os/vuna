//! tantivy-backed keyword index — BM25 over `title` + `body`. Schema keeps `url`/`title`/`snippet`
//! stored for display and `body` (the doc's chunk text, concatenated) indexed-only: the raw page
//! is never kept, only what [`vuna_core::extract`] already reduced it to (KOTVA SEARCH SRCH-2:
//! derived, rebuildable, never authoritative).

use tantivy::collector::TopDocs;
use tantivy::query::QueryParser;
use tantivy::schema::{Field, Schema, Value, STORED, STRING, TEXT};
use tantivy::{Index as TantivyIndex, IndexReader, IndexWriter, ReloadPolicy, TantivyDocument, Term};
use vuna_core::index::{Hit, IndexedDoc};

/// In-RAM indexing arena. Small on purpose: this is a node-local shard, not a bulk offline build.
const WRITER_BUDGET: usize = 15_000_000;

struct Fields {
    url: Field,
    title: Field,
    snippet: Field,
    body: Field,
}

/// Owns one tantivy index (schema: url/title/snippet stored, body indexed) plus the writer/reader
/// pair needed to keep it queryable. `upsert` is delete-by-url then re-add, since tantivy has no
/// in-place update (see tantivy's own `deleting_updating_documents` example).
pub struct KeywordIndex {
    index: TantivyIndex,
    writer: IndexWriter,
    reader: IndexReader,
    fields: Fields,
}

impl KeywordIndex {
    pub fn new() -> vuna_core::Result<Self> {
        let mut builder = Schema::builder();
        // STRING (untokenized) + STORED: exact url match for delete-by-term, and returned as-is.
        let url = builder.add_text_field("url", STRING | STORED);
        let title = builder.add_text_field("title", TEXT | STORED);
        // Stored only — the snippet is display text, not a search target (body covers that).
        let snippet = builder.add_text_field("snippet", STORED);
        // Indexed only: the concatenated chunk text. Never stored — see the module doc.
        let body = builder.add_text_field("body", TEXT);
        let schema = builder.build();

        let index = TantivyIndex::create_in_ram(schema);
        let writer: IndexWriter = index.writer(WRITER_BUDGET).map_err(tantivy_err)?;
        // Manual reload: we call `reader.reload()` right after every commit so a write is
        // immediately visible to the next search — no async-reload flakiness in tests.
        let reader = index
            .reader_builder()
            .reload_policy(ReloadPolicy::Manual)
            .try_into()
            .map_err(tantivy_err)?;

        Ok(Self { index, writer, reader, fields: Fields { url, title, snippet, body } })
    }

    /// Replace this url's keyword contribution. Body is the chunks' text joined so BM25 sees the
    /// whole document even though only `snippet` is kept for display.
    pub fn upsert(&mut self, doc: &IndexedDoc) -> vuna_core::Result<()> {
        self.writer.delete_term(Term::from_field_text(self.fields.url, &doc.url));

        let body = doc.chunks.iter().map(|c| c.text.as_str()).collect::<Vec<_>>().join("\n");
        let mut tdoc = TantivyDocument::default();
        tdoc.add_text(self.fields.url, &doc.url);
        if let Some(title) = &doc.title {
            tdoc.add_text(self.fields.title, title);
        }
        tdoc.add_text(self.fields.snippet, &doc.snippet);
        tdoc.add_text(self.fields.body, &body);

        self.writer.add_document(tdoc).map_err(tantivy_err)?;
        self.writer.commit().map_err(tantivy_err)?;
        self.reader.reload().map_err(tantivy_err)?;
        Ok(())
    }

    /// BM25 search over title+body. Never hard-fails on bad query syntax in free text — a
    /// malformed term degrades to zero keyword hits rather than erroring the whole request (the
    /// vector pass in `VunaIndex::search` can still carry the query).
    pub fn search(&self, text: &str, limit: usize) -> vuna_core::Result<Vec<Hit>> {
        if text.trim().is_empty() || limit == 0 {
            return Ok(Vec::new());
        }
        let searcher = self.reader.searcher();
        let parser = QueryParser::for_index(&self.index, vec![self.fields.title, self.fields.body]);
        let Ok(query) = parser.parse_query(text) else {
            return Ok(Vec::new());
        };

        let top = searcher.search(&query, &TopDocs::with_limit(limit)).map_err(tantivy_err)?;
        let mut hits = Vec::with_capacity(top.len());
        for (score, addr) in top {
            let retrieved: TantivyDocument = searcher.doc(addr).map_err(tantivy_err)?;
            let url = retrieved
                .get_first(self.fields.url)
                .and_then(|v| v.as_str())
                .unwrap_or_default()
                .to_string();
            let title = retrieved.get_first(self.fields.title).and_then(|v| v.as_str()).map(str::to_string);
            let snippet = retrieved
                .get_first(self.fields.snippet)
                .and_then(|v| v.as_str())
                .unwrap_or_default()
                .to_string();
            hits.push(Hit { url, title, snippet, score });
        }
        Ok(hits)
    }
}

fn tantivy_err(e: tantivy::TantivyError) -> vuna_core::Error {
    vuna_core::Error::Index(e.to_string())
}
