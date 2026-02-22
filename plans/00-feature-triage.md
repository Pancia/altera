# Altera Feature Triage

Gap analysis between docs/ vision (Gas Town critique + design principles) and current Altera implementation.

## Do Now (implementation order)

11. **Resolver Integration (Bug Fixes)** — Three critical bugs and two medium issues in the resolver loop. System is broken without fixes. See `11-resolver-integration.md`.
10. **Mission-Type Orders (Auftragstaktik)** — Tasks carry intent, success criteria, constraints, and context about *why*, not just instructions. Simple struct extension. See `10-mission-type-orders.md`.
18. **Per-Agent Prompt Injection Config** — Config-driven prompt additions keyed by agent type (worker, liaison, resolver). Supports inline text and file includes. See `18-prompt-injection-config.md`.
3. **Token Accounting / Budget Tracking** — Workers report token usage; budget constraint system has data to work with. See `03-token-accounting.md`.
7. **Checkpoint / Resume** — Rich checkpointing so sessions can pick up from where the last one left off. Foundation for handoff. See `07-checkpoint.md`.
1. **Session Continuity / Handoff** — When a worker session dies or hits context limits, hand off context to a new session instead of restarting from scratch. Depends on checkpoint. See `01-handoff.md`.

## Needs More Thought

2. **Structured Work Units** — Task notes, investigation records, decision logs. Not "beads." Connects to a broader wiki/documentation/ADR system. Too involved to build now; needs more design. Related to #6 and #9.
4. **Backpressure** — Need more context on what this looks like in practice before deciding.
5. **Graceful Degradation** — Daemon should restart, but broader partial-operation question needs thought.
6. **Environment-First Learning ("Workshop That Teaches")** — Invest in workspace so each new agent session learns through immersion. Related to #2 (depends on having good structured knowledge to teach from). Needs more thought.
8. **Agent Social Contract** — Partially exists (checkpoint command sends message to liaison). But "I'm stuck" signals aren't meaningfully parsed — daemon just accepts them. May need AI triage or liaison escalation. Needs design on what happens when an agent signals inability.
9. **Stigmergy / Indirect Coordination** — Workers leave signals for other workers. Related to #2 and #6 (structured knowledge, teaching environment, indirect coordination are all facets of the same thing). Needs more design detail.
13. **Liaison Dynamic State Awareness** — Liaison should be able to *ask* for state. Unsure about push-based updates. Need to expand on what dynamic awareness actually means.
17. **Per-Rig Supervisor / Team Lead** — Docs' "Sweet Spot" had a persistent per-rig AI supervisor. Doesn't seem necessary with liaison + on-demand resolver. Alternative: make workers smarter and more self-coordinating. Revisit with real usage data.

## Later

12. **Multi-Rig Coordination** — Much later.
14. **Task Decomposition** — Not sure it's necessary yet.
15. **Observability / Dashboard** — Want all of it eventually. `alt status --live` is a nice near-term addition. Lower priority.
16. **Post-Mortem / Learning Loop** — Yes, but later.
