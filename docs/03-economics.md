# 03 — Economics (proposal)

**Status: proposal, not spec, not built.** Nothing here is implemented. `vuna-query`
and `vuna-node` are Wave 2 stubs and the engine does not crawl-to-query end to end
yet, so every mechanism below is a design to argue with rather than a commitment.
Where this document and the code disagree, the code is the truth and this file is
wrong. Recorded 2026-07-30.

Cross-repo context: this is one instance of a family-wide payment design recorded in
`patala/docs/shared-economics.md`. The primitives it depends on are being built for
`magnetite` first.

---

## 1. The architecture already answered the hard question

The hard question for a search engine's economics is **"can placement be bought?"**
Google's answer is yes, and the conflict of interest between relevance and revenue is
the structural criticism of the whole industry.

Vuna's answer is **structurally no**, and not because of a policy — because of
`vuna-query`'s design. Ranking is **local-first, computed by the querying node**
(Min-PPR over its own graph, with peer fan-out as a supplement). There is no global
score, so there is no global ranking to sell. This is the same property TRACT states
for commerce: *"ranking is derived data any node computes… There is no global score,
because computing one requires a party that aggregates and ranks."*

**Every mechanism below is constrained by that.** A business can buy things that are
compatible with locally-computed ranking. It cannot buy position, because no party
holds position to sell.

Two other architectural facts shape the design:

- **Vuna stores the index, not the pages.** So publisher compensation cannot be for
  hosting content — it can only be for *citation*, when an answer uses a source.
- **The searcher runs a node** (Tauri desktop). So the searcher is a peer, not a
  customer. Local queries cost nothing. What costs something is **consuming other
  peers' fan-out capacity**, and that is where money belongs.

## 2. The core primitive: reciprocity first, settle the imbalance

Because every participant both serves and consumes, the honest model is not
subscription-to-a-service. It is **net settlement of an asymmetric exchange** —
closer to BitTorrent than to SaaS.

1. Each node publishes a **signed tariff** — price per query served, per result
   returned, per extraction performed. (This is Ephor's `TariffSchedule` shape; do not
   reinvent it.)
2. A query that fans out to peers accrues **signed usage receipts**.
3. Receipts net off against each other. **Only the imbalance settles**, periodically,
   through patala.

A node that serves roughly as much as it consumes pays nothing and receives nothing.
That should be the overwhelmingly common case, and it means the system has an economy
without most participants ever transacting.

**One correction to Ephor's receipt model, and it matters here.** Ephor's own docs are
honest that usage receipts are *"a one-directional audit — proves a claimed operation
happened, can't disconfirm a fabricated one."* For fan-out that is the wrong
direction: the serving node would be attesting to its own volume.

**Invert it — the consumer signs.** The consumer knows what it asked for and what it
received, so it signs the receipt and hands it back. A serving node cannot inflate
volume it was never asked for. Two colluding nodes can still fabricate exchange, but
with no inflation anywhere in the system they would only be paying each other, which
is not an attack.

## 3. Retail radar — the shopping surface

`vuna-extract`'s `retail` extractor (JSON-LD / OG) and `vuna-core`'s **retail quorum
reconciler** already exist. The reconciler is the trust mechanism: price and
availability are agreed by quorum across nodes, not asserted by one authority.

Where product data should come from, in preference order:

1. **TRACT offers** — content-addressed product records published by sellers, from
   [soko](https://github.com/vul-os/soko) and any other TRACT implementation. TRACT
   §2's property is that *"two sellers publishing the same record converge on the same
   address by construction"*, so vuna indexes them rather than inventing a product
   schema. **Do not design a product record; index the one that exists.**
2. **JSON-LD / schema.org** from the open web, as today.

### What a business can buy, and what it cannot

| Can buy | Cannot buy |
|---|---|
| **Crawl priority / freshness** — be crawled hourly instead of weekly | Position in anyone's results |
| **Verified-merchant attestation** — a signed, subscribable label | A relevance boost |
| **Structured-feed ingestion** — push a TRACT feed rather than be scraped | Removal of a competitor |

**Crawl priority** is the honest advertising product. The index records a signed
last-crawled timestamp, so freshness is *verifiable and disclosed*, and a node's
tariff states plainly "paying hosts are crawled hourly, others weekly."

**Be honest about the leak:** fresher data often ranks better, so paid freshness is a
soft ranking effect even though it is not sold as one. It must be disclosed as such,
and a querying node must be able to normalise for recency if it chooses. Claiming this
is perfectly clean would be a lie.

**Verified-merchant attestation** is a *label*, in Ephor's `labeler` shape: opt-in,
subscribable, and never on the canonical path. A searcher's node chooses whether to
trust a given attestor. It is a filterable fact — "this merchant has a published
returns policy and a verified legal identity" — not a boost. Selling attestation is
selling *verification work*, which is honest; selling rank is not.

## 4. AI answering — pay for compute, split with the citations

An AI answer costs real inference. It also consumes publishers' work while returning
no traffic to them, which is the live crisis in search and the thing a decentralised
system can actually fix.

**An answer's payment splits atomically between the parties that produced it:**

| Leg | For |
|---|---|
| Answering node | inference compute |
| Indexer(s) | retrieval |
| **Cited publishers** | the content the answer relied on |
| Stewards (optional, voluntary) | the commons |

This is the **atomic N-way split** being built for magnetite, applied to search. It is
the reason that primitive is worth sharing rather than reimplementing: the same
transaction shape pays a game's developer-operator-stewards and a query's
answerer-indexer-publishers.

Two hard constraints:

- **Per-query on-chain settlement is impossible.** A query is worth a fraction of a
  cent, far below any transaction fee. So payment aggregates over signed receipts and
  settles periodically, per §2.
- **Micro-legs must actually be deliverable.** A publisher's share of one answer is
  sub-cent. This is why the *minimum payment granularity* criterion matters: Stellar
  delivers 1 stroop (10⁻⁷ XLM) and has no per-payment floor, whereas Cardano cannot
  create an output below ~0.97 ADA and would make this model impossible. See
  `magnetite/docs/chain-candidates.md`.

### Free tier, deliberately

Unpaid queries get the index — retrieval, links, ranking, retail comparison. **AI
synthesis is the paid surface**, because it is the part with a real marginal cost and
the part that consumes publishers' work. That split is defensible to a user and to a
publisher, and it does not make search itself paywalled.

## 5. Consumer subscriptions reuse the family's recurring primitive

Heavy consumers — an AI agent making thousands of queries — need a funded balance
rather than per-query payment. Use the family recurring primitive: **N pre-signed,
time-bounded transactions on a dedicated source account**, drawn down against signed
receipts. Non-custodial, no smart contract, cancellable by bumping the sequence
number. See `patala/docs/shared-economics.md` §3.

## 6. Honest limits

- **Under-reported citations are the real attack.** A node has an incentive to claim
  it cited *fewer* publishers than it did, to pay less. Partial mitigation: the
  consumer receives the answer *and* its citation list and signs the receipt over
  both, so the list is attested by the party that received it. A node that strips
  citations produces an answer whose claims are unsupported by its declared sources,
  which is detectable but not cheaply. **Not solved.**
- **Sybil resistance for reputation** depends on proof-of-personhood, which KOTVA's
  bindings index flags as *"imperfect, and single-vendor today"*. Any reputation-
  weighted ranking inherits that weakness.
- **Paid freshness leaks into ranking**, per §3.
- **Quorum reconciliation assumes honest majority per product**; a coordinated group
  of retail nodes could agree a wrong price. The reconciler's threshold is the whole
  defence.
- **None of this can be tested yet.** `vuna-query` and `vuna-node` are stubs. The
  first real question is not economic — it is whether crawl-to-query works at all.

## 7. What to build first, if this is pursued

The economics are **not** the next thing. In order:

1. `vuna-query` and `vuna-node` — crawl-to-query end to end. Without this there is
   nothing to charge for.
2. **Signed tariff + consumer-signed usage receipts.** No settlement, no money — just
   the accounting, running and netting to zero. This is testable offline and it is the
   substrate for everything else.
3. Net settlement through patala, once the family has settled one real payment
   anywhere.
4. The citation split, which needs the atomic N-way primitive to exist and be
   validated.
5. Paid crawl priority — trivially implementable once tariffs exist, and the only
   mechanism here that generates revenue without needing a settled chain payment.
