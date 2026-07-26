# Viability — the honest assessment

This is the "should we build this" chapter, not the "how"
([Architecture](/products/vuna/docs/architecture)) or the "what"
([Getting started](/products/vuna/docs/getting-started)). It exists because the design has real,
citable structural problems, and a docs set that only argues for itself is not an honest account.
Read this before trusting the architecture chapter's confidence.

**Bottom line up front:** the index-not-pages design genuinely solves the problem it targets
(storage-driven recentralization) and genuinely defuses one real future-proofing risk
(embedding-model lock-in). It does **not** solve, and does not claim to solve, the two problems
that have sunk every prior attempt at fully decentralized search: Sybil resistance in a
permissionless-ish network, and who pays the recurring cost of keeping the corpus current.

## 1. The volunteer-cost flip — and the incentive gap it creates

Storing only the derived index instead of the pages collapses storage from datacenter-operator
scale (hundreds of petabytes, for a "reliable" open-web archive) to single-digit gigabytes per
node even at billion-page scale. That part of the flip is real.

**But it's a flip, not a deletion.** Storage was never the only recurring cost — crawl bandwidth
and embedding compute were always there too, just dwarfed by storage at web scale. Removing the
storage bottleneck **promotes compute to the binding constraint**: every page needs an embedding
forward pass per served space, at ingest and again on every content-changing re-crawl, and
freshness is a compute tax with no ceiling.

**GPU/CPU-cycle volunteering has no comparably strong precedent to disk volunteering.** Idle disk
is a near-universal, near-zero-marginal-cost donation. Idle GPU-seconds have a live market price —
donors notice the electricity bill — and the strongest historical volunteer-compute examples
(BOINC-style @home projects) succeeded on donated cycles for scientific prestige, not as an
ongoing infrastructure dependency competing with paid alternatives. Vuna rules out a token by
design — the honest, mission-aligned choice, but it also removes the one mechanism that has
reliably motivated donated compute elsewhere. **If GPU-volunteer supply doesn't materialize, the
failure mode isn't a crash — it's the frontier quietly growing a stale tail nobody is
re-embedding**, which is a much harder failure to notice or fix than an outage.

## 2. Sybil-in-a-small-network — still the crux

Reputation/social-graph Sybil resistance "works only under an honest-majority assumption, which
collapses in a small network — exactly the launch condition." Nothing about Vuna's design changes
that math; it only decides where the crux lives: `vuna-core`'s quorum module today. Its Sybil
floor is "one distinct identity key votes once," which stops one key from ballot-stuffing to
quorum but says nothing about whether `k` distinct keys are `k` distinct *people* — anchoring that
is explicitly punted to the caller via KOTVA's proof-of-personhood primitive (a real binding, not
vaporware, but single-vendor-fragile and imperfect by its own documentation).

A motivated adversary can plausibly clear a `k = 3` bar of attested identities long before the
network has enough honest participants for plurality-of-3 to mean anything statistically. Raising
`k` helps security and directly hurts liveness at low node counts — reconciliation returns "the
network doesn't know" whenever fewer than `k` distinct identities show up. This is a genuine,
unresolved chicken-and-egg: **the design needs scale to be secure, and needs to look useful before
it has scale to attract that scale.** The retail vertical doesn't dodge this by being a bounded
corpus — it trades the web vertical's Sybil surface (fake pages, link-spam) for a different one (a
non-participant subject with no reason to help you observe it honestly).

## 3. Embedding-model lock-in — mitigated

Unlike the two risks above, this is a place the design earns its future-proofing claim: each
embedding space is an independent, parallel vector index; the keyword index and link graph are
shared and model-agnostic. Chunks retain **text**, not just vectors, so adopting a new space is a
local recompute, never a re-crawl. Adoption is opt-in per node, announced as a signed object — no
fleet-wide flag-day, no coordinated migration window. This genuinely closes the "what if the
embedding model we picked is obsolete in two years" failure mode, at the cost of extra chunk-text
storage — a real, bounded, one-time-per-page price for what would otherwise be an unbounded,
recurring, whole-network re-crawl.

## 4. Bounded-corpus-first — the discipline that makes the rest tractable, and its limit

Decentralized open-web search is an "unwinnable Google war": it inherits a permanent, well-funded
SEO adversary and an unsolved permissionless-trust crux at the same time. The recommendation —
federation content or one curated vertical first, open web only as a research track, never a v1
promise — is exactly what Vuna commits to. A bounded corpus has no SEO adversary and can bootstrap
around the small-network Sybil problem by starting with pre-vetted participants.

**The limit is that this is a discipline, not an architectural guarantee.** Nothing in the
extractor/frontier traits stops a future "let's open the frontier to the whole web" decision — the
whole point of the pluggable-extractor design is that adding scope is *easy*. That's a feature for
adding verticals and a risk for scope creep back into the exact fight this design was built to
avoid. The discipline has to be re-chosen every time growth looks tempting; the code will not
enforce it.

## 5. Comparison to the field

| | Marginalia | Mwmbl | YaCy | Stract | **Vuna** |
|---|---|---|---|---|---|
| Crawl | Own crawler, single operator | Crowdsourced (volunteer Docker crawler + Firefox extension) | P2P | Own crawler | Distributed, DHT-assigned across volunteer nodes |
| Index | Centralized, single operator | Centralized (crawl is distributed, index is not) | Distributed (only project that does this) | Centralized | Distributed, per-node shards + SEARCH-profile fan-out |
| Live? | Yes, mature | Yes, maintained (~500M URLs, mid-2025) | Yes, aging, weak scale/freshness | Archived April 2026 | v0 scaffold, unproven at any scale |

The academic literature's finding, verbatim: *"to the best of our knowledge there are no
implemented projects which entirely achieve"* fully decentralized search over decentralized data.
Every real system above centralizes at least one of crawl, index, or ranking. Mwmbl already proved
distributed *crawl* works; Vuna's actual novelty claim is distributed **index** — the one axis with
no live precedent to lean on.

## 6. What could kill it — stated plainly

- **The compute-volunteer supply never materializes at the needed cadence.** No token means no
  proven mechanism for motivating ongoing GPU-hours; the failure mode is silent staleness, harder
  to detect than an outage.
- **Launch-time Sybil may be structurally unfixable within the "no token" constraint.** If a
  security-adequate `k` also makes the retail vertical mostly unavailable at real launch node
  counts, the vertical doesn't have a soft landing — it has a hole.
- **Single-vendor personhood dependency** imports that vendor's outage risk, policy risk, and
  documented coverage gaps into Vuna's trust model.
- **Scope creep back into the open web** is one product decision away at any time.
- **Retail-vertical collusion is untested, not just theoretical** — nothing in the quorum logic
  distinguishes an honest observer from one the store paid, beyond the same `k`-distinct-identities
  floor.
- **This is a v0 scaffold with zero live nodes and zero users.** Every argument above — including
  the favorable ones — is analysis of a design on paper, not pressure-tested against an actual
  adversary or volunteer population.

None of this means don't build it. The mitigated risk (embedding lock-in) is a real win, the
bounded-corpus discipline is the right call, and an unattempted distributed index is also an
opportunity, not just a danger. It means the two open cruxes — compute-volunteer economics and
small-network Sybil resistance — are the actual bet, and shipping into a bounded, vetted-launch
corpus first is what buys time to find out whether that bet pays off.
