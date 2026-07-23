# Vuna — Viability: the honest assessment

> This is the "should we build this" document, not the "how" (`02-architecture.md`) or the "what"
> (`01-design.md`). It exists because the research that produced this design
> (`00-discussion-history.md`) found real, cited structural problems in the space Vuna operates
> in, and a design doc that only argues for itself is not honest accounting. Read this before
> trusting the architecture doc's confidence.

**Bottom line up front:** the index-not-pages design genuinely solves the problem it targets
(storage-driven recentralization) and genuinely defuses one real future-proofing risk
(embedding-model lock-in). It does **not** solve, and does not claim to solve, the two problems
that have sunk every prior attempt at fully decentralized search: Sybil resistance in a
permissionless-ish network, and who pays the recurring cost of keeping the corpus current. Those
are open, and this document says so plainly rather than around it.

---

## 1. The volunteer-cost flip — and the incentive gap it creates

The pitch that motivated this whole design: open-web search's "reliable storage tier"
recentralizes at real scale (300–800 PB committed = datacenter operators, not volunteers —
`00-discussion-history.md` §1). Storing only the derived index instead of the pages collapses that
to single-digit GB per node even at 10-billion-page scale (`02-architecture.md` §6.2). That part
of the flip is real and the arithmetic backs it.

**But it's a flip, not a deletion.** Storage was never the only recurring cost of a search engine —
crawl bandwidth and embedding compute were always there too, just dwarfed by storage at Google's
scale. Removing the storage bottleneck doesn't remove those; it **promotes compute to the binding
constraint** (`02-architecture.md` §6.3):

- Every page needs a forward pass through an embedding model per served space, at ingest and again
  on every content-changing re-crawl.
- Freshness — the highest-value query dimension, and the piece the original open-web brief was
  silent on (`00-discussion-history.md` §1) — is a recurring compute tax with no ceiling, not a
  one-time ingest cost.

**Here is the part the design doesn't resolve: GPU/CPU-cycle volunteering has no comparably strong
precedent to disk volunteering.** Idle disk is a near-universal, near-zero-marginal-cost donation —
people already have spare gigabytes and don't notice giving them up. Idle GPU-seconds are not the
same kind of asset: they have a live market price (people rent them out, or mine with them), the
donor notices the electricity bill, and the strongest historical examples of large-scale volunteer
*compute* (BOINC-style @home projects) succeeded on donated idle cycles for scientific prestige or
curiosity, not as an ongoing infrastructure dependency competing with paid alternatives, and never
at a cadence where staleness accumulates if donors quietly stop showing up. Vuna explicitly rules
out a token (`01-design.md`, "no token, no page archive, no new crypto") — which is the honest,
mission-aligned choice, but it also removes the one mechanism (payment) that has reliably
motivated donated compute at scale elsewhere. **This is the flip's real cost, and it is unresolved
by anything in the current design.** If GPU-volunteer supply doesn't materialize, the failure mode
isn't a crash — it's the frontier quietly growing a stale tail nobody is re-embedding, which is a
much harder failure to notice or fix than an outage.

---

## 2. Sybil-in-a-small-network — still the crux, and it's visible in the code today

The research verdict on token-free Sybil resistance was unambiguous: reputation/social-graph
methods "work only under an honest-majority assumption, which collapses in a small network — exactly
the launch condition" (`00-discussion-history.md` §2), and this is what sank SwarmSearch's specific
numeric claims. Nothing about Vuna's design changes that math; it only decides where the crux
lives.

**It lives, concretely, in `crates/vuna-core/src/quorum.rs` today.** `reconcile()`'s Sybil floor is
"one distinct `IdentityKey` votes once" — which stops one key from ballot-stuffing to quorum, but
says nothing about whether `k` distinct keys are `k` distinct *people*. The module's own doc
comment is explicit about the punt: anchoring those keys is *"the caller's job via KOTVA ATTEST."*
That's the honest design — core logic shouldn't hardcode a personhood oracle — but it means the
actual security of every retail-vertical reconciliation reduces to: **how expensive is it, at
launch, for one adversary to acquire `k` (default 3) ATTEST-anchored identities?**

KOTVA's answer is proof-of-personhood via **World ID or Human Passport**
(`primitives/ATTEST.md`, `THREAT-MODEL.md`) — a real binding, not vaporware, but the KOTVA docs
themselves flag it as imperfect: single-vendor fragility (one biometric company's outage or policy
change is a network-wide event), coverage gaps that "exclude the undocumented"
(`primitives/ATTEST.md` §9), and — the part that matters here — **it raises the cost of a Sybil
identity, it does not raise it to "requires being a genuinely different, honest, disinterested
person."** A motivated adversary (a competitor store, a manipulator with a modest budget) can
plausibly clear a `k=3` bar of *attested* identities long before the network has enough honest
participants for plurality-of-3 to mean anything statistically. Raising `k` helps security and
directly hurts liveness at low node counts — `reconcile` returns `None` ("the network doesn't
know") whenever fewer than `k` distinct identities show up in the freshness window, so a
security-motivated higher `k` is also an availability cut at exactly the stage (launch) when
availability matters most for getting anyone to use it. This is a genuine, unresolved
chicken-and-egg: **the design needs scale to be secure, and needs to look useful before it has
scale to attract that scale.**

Worth naming precisely: **the retail vertical does not dodge this by being a bounded corpus** —
it trades the web vertical's Sybil surface (fake pages, link-spam gaming the graph) for a
*different* one (a non-participant subject with no reason to help you observe it honestly, and a
quorum of node-operators as the only ground truth). It is not a smaller problem, just a different
shape of the same one, and it's the one place in the current codebase where the crux is load-bearing
logic rather than a paragraph in a design doc.

---

## 3. Embedding-model lock-in — this one is actually mitigated

Unlike the two risks above, this is a place the design earns its future-proofing claim rather than
asserting it:

- Each embedding space is an independent, parallel vector index; the keyword index and link graph
  are shared and model-agnostic, so they never redo work when a model changes.
- `IndexedDoc` retains chunk **text**, not just vectors — adopting a new space is a local
  `Embedder::embed` recompute over already-fetched text, never a re-crawl.
- Adoption is opt-in per node (`NodeDescriptor.served_spaces`), announced as a signed object —
  there is no fleet-wide flag-day, no coordinated migration window, and an old space keeps working
  for as long as any node still serves it, because it's derived state with no authoritative status.

This genuinely closes the "what if the embedding model we picked is obsolete in two years"
failure mode that would otherwise be a real one-way door. It costs the extra chunk-text storage
noted in `02-architecture.md` §6.1 — a real, bounded, one-time-per-page price for what would
otherwise be an unbounded, recurring, whole-network re-crawl. That's a good trade and the doc
should say so plainly: this risk is handled, not merely discussed.

---

## 4. Bounded-corpus-first — the discipline that makes the rest tractable, and its own limit

The research verdict was equally unambiguous here: decentralized open-web search is an "unwinnable
Google war" because it inherits a permanent, well-funded SEO adversary and an unsolved
permissionless-trust crux at the same time (`00-discussion-history.md` §4). The corresponding
recommendation — federation content or one curated vertical first, open web only as a research
track, never a v1 promise — is exactly what `01-design.md` and the README commit to.

That discipline is what makes §§1–3 above tractable problems instead of unwinnable ones: a bounded
corpus has no SEO adversary (nothing to game because nothing to rank against strangers), and a
curated/vetted-launch corpus can bootstrap around the small-network Sybil problem in §2 by starting
with participants who are pre-vetted rather than permissionless.

**The limit is that this is a discipline, not an architectural guarantee.** Nothing in the
`Extractor`/`Frontier` traits stops a future "let's open the frontier to the whole web" decision —
the whole point of the pluggable-extractor design is that adding scope is *easy*. That ease is a
feature for adding verticals and a risk for scope creep back into the exact unwinnable fight this
design was built to avoid. The discipline has to be re-chosen every time growth looks tempting; the
code will not enforce it.

---

## 5. Comparison to the field

| | Marginalia | Mwmbl | YaCy | Stract | **Vuna** |
|---|---|---|---|---|---|
| Crawl | Own crawler, single operator | Crowdsourced (volunteer Docker crawler + Firefox extension) | P2P | Own crawler | Distributed, DHT-assigned across volunteer nodes |
| Index | **Centralized**, single operator | **Centralized** (crawl is distributed, index is not) | Distributed (only project that does this) | Centralized | Distributed, per-node shards + SEARCH-profile fan-out |
| Ranking | Own | Own | Own, weak | Own | Min-PPR-style, index-local/disclosed, no ranking token |
| Live? | Yes, mature (AGPL) | Yes, maintained (~500M URLs / 2.5M searchable, mid-2025) | Yes, but aging Java, weak scale/freshness/relevance | **No — archived April 2026** | v0 scaffold, unproven at any scale |
| Closest thing Vuna borrows | — | The live template for distributed-crawl (this is the one Vuna is trying to go *further* than) | The only precedent for a genuinely P2P index (cautionary: weak in practice) | Cautionary tale on solo-index fragility | — |

**The structural finding that frames this whole table, verbatim:** *"to the best of our knowledge
there are no implemented projects which entirely achieve"* fully decentralized search over
decentralized data (Keizer et al. 2024, Lancaster SSG / ACM survey, cited in
`00-discussion-history.md` §2). Every real system above centralizes at least one of crawl, index,
or ranking. Vuna is attempting to be the first to *not* centralize the index step — Mwmbl already
proved distributed crawl works; Vuna's actual novelty claim is distributed **index**, which is
precisely the piece nobody has shipped. That makes Vuna's harder claim a research bet on the one
axis where there is no live precedent to lean on, not an engineering assembly of solved parts.

---

## 6. What could kill it — stated plainly

- **The compute-volunteer supply never materializes at the needed cadence.** No token means no
  proven mechanism for motivating ongoing GPU-hours; the failure mode is silent staleness, which is
  worse to detect and recover from than an outage.
- **Launch-time Sybil is not a temporary bootstrapping inconvenience — it may be structurally
  unfixable within the "no token" constraint**, per §2. If a security-adequate `k` also makes the
  retail vertical mostly unavailable at real launch node-counts, the vertical doesn't have a soft
  landing; it has a hole.
- **Single-vendor personhood dependency.** Anchoring Sybil resistance to World ID / Human Passport
  imports that vendor's outage risk, policy risk, and documented-population bias into Vuna's trust
  model — and per KOTVA's own docs, it's a cost-raise, not a solve (§2).
- **Scope creep back into the open web** is one product decision away at any time, and would
  reintroduce the exact "unwinnable Google war" (SEO adversary + permissionless Sybil) this whole
  design was built to dodge (§4).
- **Retail-vertical collusion is untested, not just theoretical.** A store has a direct competitive
  incentive to want its own stock visibility manipulated (hide low stock from competitors doing
  market research, or the reverse); nothing in `quorum::reconcile` distinguishes an honest observer
  from one the store paid, beyond the same `k`-distinct-identities floor discussed in §2.
- **This is a v0 scaffold with zero live nodes and zero users.** Every argument above, including the
  favorable ones in §3 and §4, is analysis of a design on paper. None of it has been pressure-tested
  against an actual adversary, an actual volunteer population, or actual embedding-compute economics
  at any scale beyond the arithmetic in `02-architecture.md`.

None of this means don't build it — the mitigated risks (§3) are real wins, the discipline (§4) is
the right call, and "distributed index" being unattempted is also an opportunity, not just a
danger. It means the two open cruxes (§1, §2) are the actual bet being made, and shipping into a
bounded, vetted-launch corpus first is what buys time to find out whether that bet pays off before
it has to work at open-web scale.
