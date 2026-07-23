//! The node-local, persistent [`Frontier`] implementation: a `redb` table of [`UrlEntry`], keyed
//! by [`UrlId`], plus staleness-first [`Frontier::due`] selection filtered through an
//! [`Assignment`].
//!
//! **Why redb**: pure Rust (no C/libsqlite3 toolchain dependency — matches vuna-core's "compiles
//! offline" ethos), embedded (no server process, one file per node), and transactional (a crash
//! mid-write cannot corrupt the frontier — a plain in-memory-map-plus-JSON-snapshot v0 would risk
//! losing everything written since the last snapshot). Entries are stored as JSON (via
//! `serde_json`, already a workspace dep) rather than a binary format: the frontier is small
//! (URLs + a handful of timestamps per entry), so trading a few bytes for output you can eyeball
//! with `redb-dump` during development is the right side of that trade for now.

use crate::assignment::{Assignment, OwnAll};
use crate::canon::canonical_id;
use redb::{Database, ReadableTable, TableDefinition};
use std::path::Path;
use vuna_core::frontier::{Frontier, UrlEntry, UrlId, UrlList};
use vuna_core::{Error, Result, UnixSecs};

const TABLE: TableDefinition<&[u8], &[u8]> = TableDefinition::new("url_entries");

/// The node-local URL frontier: persistent storage plus a pluggable [`Assignment`] deciding which
/// stored entries `due()` will hand back to *this* node.
pub struct FrontierStore<A: Assignment = OwnAll> {
    db: Database,
    assignment: A,
}

impl FrontierStore<OwnAll> {
    /// Open (or create) a frontier store at `path`, owning everything — the dev/single-node
    /// default.
    pub fn open(path: impl AsRef<Path>) -> Result<Self> {
        Self::open_with(path, OwnAll)
    }

    /// An ephemeral, in-memory store — no file, gone on drop. Handy for tests and short-lived
    /// tools; not for a real node (nothing survives a restart).
    pub fn in_memory() -> Result<Self> {
        Self::in_memory_with(OwnAll)
    }
}

impl<A: Assignment> FrontierStore<A> {
    /// Open (or create) a frontier store at `path` under a specific [`Assignment`] (e.g. a
    /// [`crate::assignment::ModuloAssignment`] for a sharded deployment).
    pub fn open_with(path: impl AsRef<Path>, assignment: A) -> Result<Self> {
        let db = Database::create(path)
            .map_err(|e| Error::Frontier(format!("open frontier store: {e}")))?;
        Self::init_table(&db)?;
        Ok(Self { db, assignment })
    }

    /// In-memory store under a specific [`Assignment`] — used by this crate's own tests to
    /// exercise `due()`'s ownership filter without touching disk.
    pub fn in_memory_with(assignment: A) -> Result<Self> {
        let db = Database::builder()
            .create_with_backend(redb::backends::InMemoryBackend::new())
            .map_err(|e| Error::Frontier(format!("open in-memory frontier store: {e}")))?;
        Self::init_table(&db)?;
        Ok(Self { db, assignment })
    }

    fn init_table(db: &Database) -> Result<()> {
        // Opening a table that doesn't exist yet inside a write txn creates it; committing an
        // empty write is how redb wants a table declared before the first read.
        let txn = db.begin_write().map_err(frontier_err)?;
        txn.open_table(TABLE).map_err(frontier_err)?;
        txn.commit().map_err(frontier_err)?;
        Ok(())
    }

    fn get(&self, id: &UrlId) -> Result<Option<UrlEntry>> {
        let txn = self.db.begin_read().map_err(frontier_err)?;
        let table = txn.open_table(TABLE).map_err(frontier_err)?;
        match table.get(id.0.as_slice()).map_err(frontier_err)? {
            Some(v) => Ok(Some(serde_json::from_slice(v.value())?)),
            None => Ok(None),
        }
    }

    fn put(&self, entry: &UrlEntry) -> Result<()> {
        let bytes = serde_json::to_vec(entry)?;
        let txn = self.db.begin_write().map_err(frontier_err)?;
        {
            let mut table = txn.open_table(TABLE).map_err(frontier_err)?;
            table.insert(entry.id.0.as_slice(), bytes.as_slice()).map_err(frontier_err)?;
        }
        txn.commit().map_err(frontier_err)?;
        Ok(())
    }

    /// Every stored entry, in no particular order. Used by `due()`; exposed publicly since it's
    /// generally useful (tooling, tests, debugging) and costs nothing extra to expose.
    pub fn iter_all(&self) -> Result<Vec<UrlEntry>> {
        let txn = self.db.begin_read().map_err(frontier_err)?;
        let table = txn.open_table(TABLE).map_err(frontier_err)?;
        let mut out = Vec::new();
        for row in table.iter().map_err(frontier_err)? {
            let (_key, value) = row.map_err(frontier_err)?;
            out.push(serde_json::from_slice(value.value())?);
        }
        Ok(out)
    }
}

fn frontier_err(e: impl std::fmt::Display) -> Error {
    Error::Frontier(e.to_string())
}

impl<A: Assignment> Frontier for FrontierStore<A> {
    fn subscribe(&mut self, list: &UrlList) -> Result<()> {
        for raw in &list.urls {
            // `discover` already inserts-if-absent (dedup) and canonicalizes. A malformed URL in
            // someone else's third-party list shouldn't sink the whole subscribe, so per-entry
            // failures are swallowed here rather than propagated.
            let _ = self.discover(raw, list.updated_at);
        }
        Ok(())
    }

    fn discover(&mut self, url: &str, _now: UnixSecs) -> Result<UrlId> {
        let (canon, id) = canonical_id(url)?;
        if self.get(&id)?.is_none() {
            self.put(&UrlEntry::new(canon, id))?;
        }
        Ok(id)
    }

    fn due(&self, now: UnixSecs, limit: usize) -> Result<Vec<UrlEntry>> {
        let mut owned: Vec<(UnixSecs, UrlEntry)> = self
            .iter_all()?
            .into_iter()
            .filter(|e| self.assignment.owns(&e.id))
            .map(|e| {
                // Staleness = time since last crawl; never-crawled (None) is maximally stale, so
                // it's treated as staler than any real elapsed time (saturating_sub with 0 gives
                // `now`, and a never-crawled entry has been "due" since before recorded time —
                // clamp it above that with u64::MAX so it always sorts first).
                let staleness = match e.last_crawled {
                    Some(t) => now.saturating_sub(t),
                    None => u64::MAX,
                };
                (staleness, e)
            })
            .collect();

        // Most stale (largest staleness) first.
        owned.sort_by(|a, b| b.0.cmp(&a.0));
        owned.truncate(limit);
        Ok(owned.into_iter().map(|(_, e)| e).collect())
    }

    fn record(&mut self, entry: &UrlEntry) -> Result<()> {
        self.put(entry)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::assignment::ModuloAssignment;
    use crate::canon::url_id;
    use vuna_core::ContentId;
    use vuna_core::IdentityKey;

    fn list(urls: Vec<&str>) -> UrlList {
        UrlList {
            id: ContentId::ZERO,
            name: "test-list".into(),
            publisher: IdentityKey([0u8; 32]),
            version: 1,
            urls: urls.into_iter().map(String::from).collect(),
            body: None,
            updated_at: 1_000,
        }
    }

    #[test]
    fn subscribe_dedups_equivalent_urls() {
        let mut store = FrontierStore::in_memory().unwrap();
        let l = list(vec![
            "https://Example.com:443/a",
            "https://example.com/a", // same canonical URL as above
            "https://example.com/b",
        ]);
        store.subscribe(&l).unwrap();
        assert_eq!(store.iter_all().unwrap().len(), 2, "the two /a variants must collapse to one entry");
    }

    #[test]
    fn discover_is_idempotent() {
        let mut store = FrontierStore::in_memory().unwrap();
        let id1 = store.discover("https://example.com/x", 1).unwrap();
        let id2 = store.discover("https://example.com/x", 2).unwrap();
        assert_eq!(id1, id2);
        assert_eq!(store.iter_all().unwrap().len(), 1);
    }

    #[test]
    fn discover_does_not_clobber_recorded_state() {
        let mut store = FrontierStore::in_memory().unwrap();
        let id = store.discover("https://example.com/x", 1).unwrap();
        let mut entry = UrlEntry::new("https://example.com/x", id);
        entry.last_crawled = Some(500);
        store.record(&entry).unwrap();

        // Re-discovering the same URL must not reset the crawl state we just recorded.
        store.discover("https://example.com/x", 999).unwrap();
        let stored = store.get(&id).unwrap().unwrap();
        assert_eq!(stored.last_crawled, Some(500));
    }

    #[test]
    fn due_is_staleness_ordered_and_respects_limit() {
        let mut store = FrontierStore::in_memory().unwrap();
        for (path, crawled) in [("/never", None), ("/old", Some(10u64)), ("/new", Some(90u64))] {
            let id = store.discover(&format!("https://example.com{path}"), 0).unwrap();
            let mut entry = UrlEntry::new(format!("https://example.com{path}"), id);
            entry.last_crawled = crawled;
            store.record(&entry).unwrap();
        }

        let due = store.due(100, 10).unwrap();
        let urls: Vec<&str> = due.iter().map(|e| e.url.as_str()).collect();
        assert_eq!(urls, vec!["https://example.com/never", "https://example.com/old", "https://example.com/new"]);

        let limited = store.due(100, 2).unwrap();
        assert_eq!(limited.len(), 2);
    }

    #[test]
    fn due_only_returns_owned_entries() {
        // Two-node cluster; find one url each node owns by construction, then verify each store
        // only ever hands back its own.
        let node0 = ModuloAssignment::new(0, 2);
        let node1 = ModuloAssignment::new(1, 2);

        let mut url_for_node0 = None;
        let mut url_for_node1 = None;
        for i in 0.. {
            let u = format!("https://example.com/p/{i}");
            let id = url_id(&u);
            if url_for_node0.is_none() && node0.owns(&id) {
                url_for_node0 = Some(u.clone());
            }
            if url_for_node1.is_none() && node1.owns(&id) {
                url_for_node1 = Some(u.clone());
            }
            if url_for_node0.is_some() && url_for_node1.is_some() {
                break;
            }
        }
        let url_for_node0 = url_for_node0.unwrap();
        let url_for_node1 = url_for_node1.unwrap();

        let mut store0 = FrontierStore::in_memory_with(ModuloAssignment::new(0, 2)).unwrap();
        store0.discover(&url_for_node0, 0).unwrap();
        store0.discover(&url_for_node1, 0).unwrap();

        let due0 = store0.due(0, 10).unwrap();
        assert_eq!(due0.len(), 1);
        assert_eq!(due0[0].url, url_for_node0);
    }
}
