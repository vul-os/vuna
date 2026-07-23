# Scout — Design Discussion Log

*Federated search for the Vulos / DMTAP network. Working notes captured 2026-07-20.*

> Status: **exploration**, not committed. This is the record of a discussion —
> the reasoning behind the current recommendation, so we don't lose it.

---

## TL;DR — where we landed

- **Don't build a Google-style open-web search engine** in any form. Both the
  achievable version (yet another independent index) and the novel version
  (fully decentralized, no-token) are bad bets — see reasoning below.
- **Do build federated search over our OWN network** — index the content that
  lives inside the Vulos/DMTAP federation (member instances/boxes + their
  public/shared content), not the open web.
- **Why this wins:** it dodges every hard problem at once by *not* trying to
  swallow the whole web. No open-web crawler, no SEO adversary, no
  permissionless Sybil wall — because it rides the **vetted-operator trust
  layer we already built in DMTAP.**
- **Name:** **Scout** (product). The underlying federated-query + ranking layer
  should be a DMTAP extension (protocol) rather than separately branded, for now.

---

## 1. The original brief (starting point)

The discussion started from a conceptual brief for a *decentralized, no-crypto
web search engine* — crawl + index the web collaboratively across volunteer
nodes, two-tier node model (reliable storage tier + mass lightweight
query/crawl tier), chunk-level DHT, content-defined chunking for dedup,
reputation-based Sybil resistance without a token, funded by grants/co-op.

It was well-researched but strongest on the **already-solved** layers (storage,
chunking, DHT) and thin on the three layers that actually decide viability.

### The three cruxes (what actually decides feasibility)

1. **Adversarial ranking / SEO-spam resistance** — ranking *is* the search
   engine; the brief punted it. Google's moat is ranking + anti-spam, not the
   crawl.
2. **Sybil resistance without a token/stake** — removing the token (correct for
   mission reasons) also removes the one thing that gives Sybil resistance a
   *cost function*. Makes the trust problem harder, not easier.
3. **Query privacy in a multi-operator system** — routing queries through many
   volunteer nodes means many parties see queries. Potentially *worse* privacy
   than a single incumbent. The brief was silent on it.

Plus two number-breakers: the "reliable" storage tier recentralizes at real
scale (300–800 PB committed = datacenter operators, not volunteers), and
**freshness** (the highest-value queries) is the hardest operational piece and
was absent.

**Verdict on the brief:** not viable as a Google-replacement; the same
components rescoped as a *federated, transparent, censorship-resistant vertical*
search engine are viable. The engineering (storage/chunking/DHT) mostly
survives a rescope; the "as capable as Google" ambition does not.

---

## 2. Research findings (deep-research, 2026-07-20)

23 sources, 25 claims adversarially verified (23 confirmed, 2 refuted).

### Existing projects
- **Marginalia** — real independent index. AGPL. **Centralized, single-operator,
  deliberately monolithic** (designed *against* distribution). Strongest mature
  base, but forking it means fighting its core assumption.
- **Mwmbl** — real independent index. **Crowdsourced/distributed crawl + central
  index** (~500M URLs / 2.5M searchable results, mid-2025). Volunteer Docker
  crawler + Firefox curation extension. The closest live template for
  distributed-crawl / central-index. Still maintained.
- **Stract** — real independent Rust index... **DEAD.** GitHub archived
  (read-only) April 2, 2026. Cautionary tale on solo-index fragility. (This was
  the one candidate that might have flipped the fork recommendation — it's gone.)
- **YaCy** — the *only* genuinely P2P/federated own-index. Architecturally real
  but **weak on scale, freshness, relevance**; aging Java.
- **SwarmSearch (2025)** — a *proposal*, and **not actually crypto-free**
  (settles rewards in Bitcoin/stablecoins). Two of its technical claims
  (MeritRank Sybil figures, Data-Shapley ranking tolerating 50% adversaries)
  were **refuted 0-3** in verification. Do not use as a model.
- **SearXNG / Whoogle / Presearch** — metasearch proxies (widely known), not
  independent indexes.

### The key structural finding
> **No implemented project achieves fully decentralized search over
> decentralized data** (Keizer et al. 2024, Lancaster SSG, ACM survey — verbatim:
> "to the best of our knowledge there are no implemented projects which entirely
> achieve this"). Every real system centralizes at least one of crawl / index /
> ranking.

So: the *achievable* version (central independent index) is a **solved wheel**;
the *fully decentralized* version is an **open research problem**, not a wheel —
building it is a research bet, not an engineering project.

### The three cruxes, per evidence
- **Adversarial ranking → tractable.** Min-PPR (personalized-PageRank variant)
  is *proven* to get low distortion AND high spam resistance simultaneously;
  vanilla PageRank provably cannot. ("Graph Ranking and the Cost of Sybil
  Defense," Farach-Colton, Goldberg et al., ACM EC 2023.) → Use Min-PPR-style
  graph ranking, not classic PageRank.
- **Token-free Sybil → possible but conditional.** Reputation + social-graph
  methods work **only under an honest-majority assumption**, which **collapses
  in a small network** — exactly the launch condition, and exactly what sank
  SwarmSearch's numbers. This is the single biggest risk to any no-token
  decentralized design.
- **Query privacy → achievable, not cheap.** Coeus (SOSP'21), ChalametPIR
  (CCS'24), Wally (2024) give real PIR/DP-grade oblivious search, but with heavy
  cost and non-collusion trust assumptions. **Later layer, not v1.**

### Build stack (if building)
- **Index engine:** **Tantivy** (Rust, Lucene-class library; MIT; ~15.6k stars;
  mature, production use). **Quickwit** = the distributed, object-storage-native
  scale-up path built on Tantivy.
- **Content addressing (optional/later):** **libp2p Kademlia DHT**
  (ADD_PROVIDER / GET_PROVIDERS, keyed by multihash).
- Postings compression (Roaring/Elias-Fano), FastCDC chunking, MinHash/SimHash
  dedup, WARC crawl storage — all standard, reuse don't reinvent.

### Research caveats
- "Reuse-vs-build" framing leans on a single survey (Keizer 2024); one claim
  passed only 2-1 (well-corroborated though).
- Privacy results are self-reported benchmarks under favorable
  hardware/non-collusion assumptions — "achievable" ≠ "cheap."
- Presearch, Brave's independent index, Kagi, Common Crawl indexes, and several
  named tools were **not** confirmed in the verified claim set — unassessed,
  neither endorsed nor dismissed.

---

## 3. Fork vs. build-from-scratch

Key distinction people conflate:
- **Fork a *product*** (Marginalia/YaCy/Mwmbl) = inherit their whole codebase,
  crawler, index, UI, and every baked-in design decision + debt + license.
- **Build on *libraries*** (Tantivy, etc.) = assemble from settled building
  blocks. **This is NOT reinventing the wheel.** "Without a fork" means this.

**When forking pays:** only to inherit the open-web **crawler + ingest
pipeline + spam-tuned ranking** — the brutal, months-long wheel. That's the one
thing worth forking Mwmbl for.

**Do we *have* to fork? No — never mandatory.** And on the recommended path
(federated over our own network) forking is actually a *mistake*, because that
path has **no open-web crawler to inherit** — the main reason to fork evaporates.

Ruled out regardless: build-everything-from-scratch, forking Marginalia (fights
its monolith), YaCy as a base (aging/weak), anything crypto-adjacent.

---

## 4. Open web vs. our own network — THE decision

"Open web" vs "our federation" is a question about **the corpus** — and it
changes everything downstream.

| | Open web | Our federation (Scout) |
|---|---|---|
| Corpus | Entire public internet, billions of pages, strangers' servers | Only content inside our federation (member instances + their public/shared content) |
| Crawler | Yes, and it's huge (robots, JS, traps, freshness, scale) | **None** — content is published *into* the network; just a sync/ingest job |
| SEO adversary | Permanent, well-funded | Effectively **absent** — only vetted members contribute |
| Sybil / trust | Permissionless — the unsolved, fatal crux | **Rides existing DMTAP vetted-operator trust** — dodged |
| Scale | Petabytes, billions of docs | Thousands→millions of docs |
| Write our own crawler? | **No** — reuse Heritrix / Nutch / Common Crawl | **N/A** — write a small protocol-specific ingest client |
| Verdict | Unwinnable Google war | **Shippable feature** |

**Why decentralized open-web search keeps failing:** it's always the
**trust/Sybil layer** — nobody can bootstrap a permissionless network of
anonymous operators you can trust to rank honestly.

**The move:** we already walked through that wall in DMTAP. DMTAP/Vulos *is* a
vetted-operator trust model. So don't build search that needs a *new* trust
layer — build search that **rides the trust layer we already have.**

---

## 5. Recommendation (current)

Build **Scout = federated search native to the Vulos/DMTAP network.**

- **Corpus:** content inside the federation (member instances' public/shared
  content), optionally + a curated vertical. Bounded corpus ⇒ no SEO adversary,
  no open-web crawler.
- **Operators:** the federation members we already vet ⇒ no permissionless
  Sybil problem.
- **Ranking:** Min-PPR-style graph ranking over the (already-trusted)
  federation link graph.
- **Query privacy:** defer. Members query instances they already trust
  (Mastodon-style). PIR is a v3 research layer, if ever.
- **Engine:** Tantivy per node; query fan-out + merge across the federation.
  Boring, standard, correct.
- **Ingest:** small, protocol-specific sync client against known member
  instances (a few hundred lines) — NOT a web crawler.

**No fork. No open-web crawler.** Real effort goes into the index + ranking +
federation-query layer.

**Fallback if we insist on open web:** extend Mwmbl (inherit its distributed
crawler), keep the index central, ship the boring version, and treat
"decentralized/federated" as a research track that may never finish — never as
a v1 promise. Do not build the decentralized version first.

**Should we even start?** If the goal is the sovereign-network search, yes —
it's genuinely buildable and doesn't exist. If the goal is a general web-search
Google-killer, no — the achievable version doesn't matter and the novel version
is gated on unsolved research.

---

## 6. Naming

- **Product:** **Scout** — active, plain, fits the Vulos family
  (Mail/Talk/Meet/Files/Relay), connotes going *out* across the network to find.
- Alternatives considered: **Forage** (most precise to the architecture —
  gathering from known sources), **Dowse** (distinctive, "search for something
  hidden"), **Find** (plainest, in-family, forgettable).
- **Protocol layer:** the federated-query + ranking mechanism should be a
  **DMTAP extension** (e.g. `DMTAP-Query`) rather than separately branded until
  it earns its own identity — same way DMTAP sits under Envoir.

---

## 7. Open questions / next steps

- [ ] Confirm the target: sovereign-network search (Scout) vs. open web. Current
      recommendation assumes **sovereign-network.**
- [ ] Define exactly what "federation content" is in Vulos/DMTAP terms — which
      instance content is indexable (public pages? shared docs? opt-in?).
- [ ] Pressure-test the **small-network Sybil** assumption for *our* federation:
      does vetted-operator membership fully neutralize it, or are there residual
      manipulation vectors (a member gaming ranking within the federation)?
- [ ] Sketch the concrete architecture: ingest-client + Tantivy-per-node +
      Min-PPR-over-federation-graph + query fan-out/merge.
- [ ] Decide product name (Scout confirmed?) and whether protocol layer gets its
      own name now or stays a DMTAP extension.

---

*Sources of record (verified): Marginalia/Mwmbl/Stract/YaCy repos & docs;
Keizer et al. 2024 (Lancaster SSG / ACM); Farach-Colton & Goldberg et al.,
ACM EC 2023 (Min-PPR); Coeus SOSP'21; ChalametPIR CCS'24; Wally 2024; Tantivy /
libp2p docs. Full research artifact in the deep-research task output.*
