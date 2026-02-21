# Innovation Engine Applied to Gas Town

Applying the Innovation Engine recipe (Analogical Reasoning, First Principles, Dialectical Synthesis, Pre-Mortem) to Gas Town's multi-agent orchestration system.

---

## Step 1: Analogical Reasoning -- Three Distant Domains

### 1A. Biology: Stigmergic Coordination and Immune Cascades

#### Ant Colonies: Stigmergy over Hierarchy

Ant colonies coordinate 100,000+ workers with zero central command. No ant knows the colony's goals. The mechanism is **stigmergy**: agents modify their shared environment (pheromone trails), and other agents respond to those modifications. The environment *is* the coordination layer.

**Structural mapping to Gas Town:**

| Ant Colony | Gas Town Equivalent | Gap Revealed |
|------------|---------------------|--------------|
| Pheromone trails (environment as message) | Mail protocol (direct agent-to-agent) | Gas Town routes messages *between agents* rather than *through the environment*. Every coordination act requires a named sender and receiver. Ants don't address messages; they deposit state changes. |
| Task switching by threshold response | Mayor assigns convoys | Ants switch tasks when local stimulus exceeds their personal threshold. No dispatcher needed. A hungry larva emits chemical signal; the nearest ant with a low feeding-threshold responds. Gas Town's Mayor is a bottleneck that ants eliminated 100 million years ago. |
| No foreman, no supervisor | Three-tier watchdog chain (Daemon -> Boot -> Deacon -> Witness) | Ant colonies have zero supervision overhead. Dead ants are detected by chemical decay (oleic acid), not by a watchdog polling their heartbeat. The *absence* of expected signals triggers response, not the *presence* of monitoring agents. |
| Nest architecture emerges from simple rules | Architecture is designed top-down | Ant nest chambers (brood, food storage, waste) emerge from local behavioral rules, not a blueprint. Gas Town's Rig/Town/Agent hierarchy is architect-imposed. |
| Colony-level intelligence from simple agents | Complex agents with elaborate prompts | Individual ants follow 3-5 behavioral rules. Colony intelligence is an emergent property of interaction density, not individual sophistication. Gas Town bets on smart agents; ants bet on dumb agents in smart environments. |

**The critical insight:** Gas Town uses agent-to-agent messaging (Mail, Nudge, Handoff) as its coordination primitive. Biology uses **environment-mediated coordination**. The beads database *could* be the pheromone layer -- agents write state changes to beads, other agents poll beads for state changes that match their activation threshold -- but Gas Town overlays a separate messaging system on top of this. The messaging system then requires supervision (who's listening? who's stuck?), which requires watchdogs, which require watchdogs-of-watchdogs. The ant colony says: **delete the messaging layer entirely; make the data layer the only coordination mechanism.**

#### Immune System: Cascading Activation Without Central Planning

The adaptive immune system detects and destroys novel threats it has never encountered. It does this without a "Mayor" cell coordinating the response. The mechanism:

1. **Sentinel cells** (dendritic cells, macrophages) patrol tissues. They don't report to a central authority; they process threats locally and present antigens.
2. **Activation cascades**: A dendritic cell activating a T-cell doesn't "assign work." It presents evidence (antigen) and the T-cell *decides for itself* whether to activate based on receptor match. This is capability-based routing without a router.
3. **Clonal expansion**: When a T-cell activates, it doesn't ask permission to scale. It copies itself thousands of times. Scale-up is a local decision based on signal strength.
4. **Apoptosis**: Cells that are no longer needed self-destruct. They don't wait for a Witness to notice they're idle and clean them up.

**Mapping to Gas Town:**

- Gas Town's Witness monitors polecats and triggers recovery. The immune system equivalent: cells carry their own death timer (apoptosis). If a T-cell doesn't receive survival signals within a window, it kills itself. **Polecats should carry their own liveness timers and self-terminate on stall, rather than relying on external monitoring.**
- Gas Town's Mayor creates convoys and distributes work. The immune system equivalent: antigens are "work" floating in the environment; cells with matching receptors grab them. **Work items (beads) should carry their own activation criteria, and agents should self-select based on capability match.**
- Gas Town's Deacon patrols all rigs. The immune system equivalent: there is no patrol. Sentinel cells are *already in the tissue*. They don't tour the body looking for problems; they're embedded where problems occur. **Monitoring should be embedded in the execution layer, not a separate supervisory agent.**

#### Neural Networks: Hebbian Learning and Backpropagation

The brain doesn't have a "manager neuron" telling other neurons what to do. Coordination emerges from:

- **Connection weights** that strengthen or weaken based on correlated activity (Hebbian: "neurons that fire together wire together")
- **Lateral inhibition**: active neurons suppress nearby competitors, creating winner-take-all dynamics without an arbiter
- **No global error signal in biological nets**: local learning rules produce global coherence

**Gas Town implication:** Agent-to-agent affinity should emerge from successful collaboration history. If Polecat A's work consistently merges cleanly when it handles TypeScript tasks, its "weight" for TypeScript should increase. Gas Town's CV/capability system points in this direction, but the routing is still centralized through the Mayor rather than emerging from weighted self-selection.

---

### 1B. Economics: Price Signals, Transaction Costs, and Spontaneous Order

#### Hayek's Knowledge Problem

Friedrich Hayek's central insight: no single node in an economy can possess all the information needed for optimal allocation. Prices aggregate distributed knowledge into a single signal that enables coordination without central planning.

**Gas Town's knowledge problem:** The Mayor must know (a) what work needs doing, (b) which agents are available, (c) what each agent is good at, (d) current system load, (e) cross-rig dependencies. This is exactly the impossibility Hayek identified -- a central planner cannot aggregate all relevant information fast enough for good decisions. The Mayor's decisions will always be informationally impoverished compared to the distributed knowledge held by the agents themselves.

**What a market-based Gas Town would look like:**

1. **Work items (beads) carry a "price" reflecting urgency, complexity, and value.** Price starts low, rises with time (stale work gets more expensive to ignore). This is an **ascending auction** for agent attention.
2. **Agents "bid" by claiming work.** Their bid is implicit: they select tasks whose price exceeds their opportunity cost (the value of what they'd give up). No Mayor needed.
3. **Completion quality adjusts future prices.** If an agent does poor work, the rework cost is "charged" against that agent's reputation score (analogous to credit rating). Good work earns preferential access to high-value tasks.
4. **Transaction costs determine optimal agent granularity.** Ronald Coase's theory of the firm: organizations exist because market transaction costs sometimes exceed internal coordination costs. Gas Town should have agents exactly as long as the overhead of maintaining an agent is less than the overhead of ad-hoc task execution. The three-tier watchdog chain is a massive transaction cost; a simpler coordination mechanism would shift the equilibrium toward more, smaller agents.

#### Comparative Advantage (Ricardo)

Even if one agent is better at *everything* than another agent, both agents should still specialize. Agent A might be 10x better at refactoring and 3x better at documentation. Agent A should do refactoring; Agent B should do documentation. Both produce more total output than if Agent A does everything.

**Gas Town implication:** The system should route based on *comparative* advantage (where is the gap between this agent and the next-best agent largest?), not *absolute* advantage (who is best?). This requires tracking not just "Agent X succeeded at Task Type Y" but "Agent X succeeded at Task Type Y *relative to other agents*." The CV system captures absolute performance but not relative advantage.

#### Transaction Cost Economics (Coase/Williamson)

Every coordination act has a cost: the tokens spent on Mail messages, the time spent in handoffs, the overhead of the Witness monitoring polecats. Williamson identified three factors that raise transaction costs:

1. **Asset specificity**: How specialized is the work? Highly specific tasks are expensive to re-assign. Gas Town's hook mechanism (pinned assignment) handles this well.
2. **Uncertainty**: How unpredictable is the task? Gas Town handles this poorly -- the Mayor must estimate task complexity before assignment, but uncertainty means these estimates are often wrong.
3. **Frequency**: How often does this transaction occur? Gas Town's watchdog chain runs *continuously*, even when no work exists. High-frequency null transactions are pure waste.

**Key insight from transaction cost economics:** Gas Town's supervision architecture burns continuous transaction costs (Boot checking Deacon, Deacon patrolling rigs, Witness monitoring polecats) regardless of workload. A market mechanism would have *zero* coordination overhead at zero workload, because coordination only happens when work is claimed.

#### Mechanism Design (Auction Theory)

Markets don't "just work" -- they require careful mechanism design. Different auction formats produce different outcomes:

- **First-price sealed bid**: Agents claim tasks privately, best match wins. Risk: agents underbid (take work they're not equipped for).
- **Vickrey auction**: Agents bid truthfully because the winner pays the second-highest price. Incentive-compatible.
- **Double auction**: Both task-posters and task-claimers state prices. Most efficient but requires market thickness (many agents, many tasks).

For Gas Town, the relevant mechanism is a **continuous double auction** where beads post required capability levels and agents post available capability levels. Matching happens when capability supply meets capability demand. This is how modern financial markets work, processing millions of transactions per second without a Mayor.

---

### 1C. Urban Planning: Desire Paths, Zoning, and Emergent Neighborhoods

#### Desire Paths

When a university lays concrete walkways and students cut across the grass, the trampled paths reveal where the actual traffic wants to go. Smart campuses wait a year, observe the desire paths, then pave *those*.

**Gas Town's desire paths:** The architecture-analysis.md reveals that the Deacon goes stale every 6 minutes and must be restarted. This is a desire path -- the system is telling you it doesn't want a continuously-running AI supervisor. The "desire path" is a system that activates supervision on-demand (event-driven) rather than maintaining it as a persistent patrol.

Other desire paths to look for:
- Which Mail message types are actually sent frequently vs. which exist but are rarely used? The frequent ones reveal essential coordination; the rare ones reveal speculative design.
- How often do agents actually use Seance (historical context queries)? If rarely, the system is saying it doesn't need deep history -- it needs good handoffs.
- The single-rig reality (one rig: hermes) vs. the multi-rig architecture. The desire path says "start simple, grow into complexity" rather than "build for scale, operate at small scale."

#### Zoning vs. Emergence

Cities use two coordination mechanisms simultaneously:
1. **Zoning** (top-down): Industrial here, residential there, commercial along this corridor. Prevents harmful adjacencies (factory next to school).
2. **Emergent neighborhoods** (bottom-up): Chinatown, Little Italy, arts districts, tech corridors. Nobody planned these; they emerged from individual location decisions reinforcing each other.

**Gas Town mapping:**
- Gas Town is almost entirely "zoned": the architect defines Rigs, assigns Witnesses, places Refineries, designs the hierarchy. There is no emergence.
- A city-inspired Gas Town would define minimal zoning (safety constraints: "no two agents edit the same file simultaneously," "all merges go through a queue") and let everything else emerge. Agent specialization, work selection, collaboration patterns -- these would develop organically based on what works.

**The urban planning insight:** Over-zoned cities are sterile (think Brasilia -- architecturally perfect, humanly dead). Under-zoned cities are chaotic (no building codes = structural collapse). The optimum is **minimal viable zoning**: enough structure to prevent catastrophe, enough freedom for emergent intelligence.

Gas Town is over-zoned. Thirteen agent roles, ten work unit types, four gate types -- this is Brasilia. The question is: what is the *minimal viable zoning* for multi-agent orchestration?

#### Infrastructure vs. Services

Cities provide infrastructure (roads, water, electricity, sewage) but not services. The city builds the road; it doesn't drive your car. The city provides water pipes; it doesn't cook your food.

**Gas Town conflation:** Gas Town conflates infrastructure with services. The daemon (infrastructure) also manages agent lifecycle (service). The beads system (infrastructure) is tightly coupled with the convoy assignment logic (service). The merge queue (infrastructure) is bound to the Refinery agent role (service).

A city-inspired architecture would separate these cleanly:
- **Infrastructure layer**: Git worktrees, beads database, message bus, tmux sessions. Dumb pipes. No AI.
- **Service layer**: Agents that use the infrastructure to accomplish work. Agents come and go; infrastructure persists.

Currently, Boot, Deacon, and Witness are infrastructure masquerading as services. They're AI agents doing infrastructure work (health checks, lifecycle management). This is like hiring a human to manually open and close water valves instead of installing pressure regulators.

#### Urban Resilience

Cities survive earthquakes, floods, wars, and pandemics not through centralized crisis management but through **redundancy, modularity, and local adaptation**:
- Multiple routes between any two points (redundancy)
- Neighborhoods can function semi-independently if cut off (modularity)
- Local businesses adapt to local conditions without central directives (adaptation)

**Gas Town's resilience model is centralized:** If the Deacon goes down, Boot must detect it, restart it, and the system limps until recovery completes. If the Mayor goes down, convoy assignment halts. These are single points of failure dressed up in supervision layers.

A city-resilient Gas Town would have:
- No single agent whose failure halts the system
- Every function performable by multiple agents (any agent can merge, any agent can assign work)
- Graceful degradation: losing one agent means slightly slower operation, not functional loss

---

## Step 2: First Principles -- Regressive Abstraction

### The Setup

Start from: "I have N AI agents and M tasks. I need the tasks completed correctly across K git repositories."

Strip away every assumption about *how* Gas Town currently works. Ask only: what does this problem *irreducibly* require?

### Layer 0: What Can't Be Eliminated

**Requirement 1: Task Definition.** Something must specify what needs to be done. This is irreducible. You need a task representation.
- Minimum viable form: A text description with an identifier. Not necessarily beads, convoys, molecules, wisps, or formulas. Just: "here's what to do" + "here's how to reference it."

**Requirement 2: Task-Agent Binding.** Each task must be assigned to (or claimed by) at most one agent at a time. Without this, you get duplicate work or no work.
- Minimum viable form: A lock. Any mutual exclusion mechanism. A file lock, a database row lock, a git branch name convention.

**Requirement 3: Isolation.** Agents working in parallel must not corrupt each other's state. In a git context, this means separate working copies.
- Minimum viable form: Git worktrees or separate clones. This is already well-established in Gas Town and is genuinely irreducible.

**Requirement 4: Integration.** Parallel work must be merged back into a shared baseline. Conflicts must be resolved.
- Minimum viable form: A merge queue with conflict detection. The merge can be attempted by any process; it doesn't require a dedicated "Refinery" agent.

**Requirement 5: Failure Detection.** If an agent dies mid-task, the task must eventually be reclaimed. Without this, tasks are permanently lost to zombie assignments.
- Minimum viable form: A lease with expiry. Agent holds a task for T minutes; if not renewed, the task returns to the pool. No watchdog agent needed -- the *absence* of a lease renewal is the detection mechanism.

**Requirement 6: Result Verification.** Someone or something must check that the work meets requirements. Tests pass, code compiles, review criteria are satisfied.
- Minimum viable form: Automated checks (CI, linting, type checking). Human review for subjective quality. Neither requires a dedicated agent role.

### Layer 1: What's Probably Necessary at Scale

**Requirement 7: Prioritization.** When M tasks > N agents, some tasks should be done before others.
- Minimum viable form: A sort order on the task pool. Priority can be a static number or a dynamic function (urgency increasing with age).

**Requirement 8: Capability Matching.** Not every agent is equally suited to every task. Routing based on demonstrated capability improves throughput.
- Minimum viable form: Tags on tasks (required skills) + tags on agents (demonstrated skills) + a matching function. This can be a database query; it doesn't need a Mayor.

**Requirement 9: Context Continuity.** AI agents hit context window limits. Work state must survive session boundaries.
- Minimum viable form: A state file written at session end, read at session start. The handoff mechanism. This is genuinely irreducible for long-running work.

**Requirement 10: Observability.** Humans need to see what's happening.
- Minimum viable form: A log. Structured events written to a queryable store. Dashboard optional but log is essential.

### What's NOT Irreducible

Everything else in Gas Town is a design choice, not a requirement:

- **Hierarchy (Mayor -> Deacon -> Witness -> Polecat)**: Not required. Flat pool + self-selection achieves the same outcome with less overhead.
- **Dedicated supervisor agents**: Not required. Lease-based failure detection + automated CI handles supervision without AI tokens.
- **Rich messaging protocol (8 message types)**: Not required. State changes in a shared database are sufficient. Agents read state; they don't need to send messages.
- **Multiple work unit types (Bead, Convoy, Molecule, Wisp, Formula)**: Not required. One work unit type with optional metadata fields handles all cases.
- **Persistent agent identity with CVs**: Useful for optimization but not required for basic function. Agents are fungible at the base layer; capability tracking is an optimization layer.
- **Named metaphors for every component**: Purely cosmetic. "Polecat" vs. "worker" changes nothing about function.

### The Bedrock Architecture

From first principles, multi-agent orchestration needs exactly:

```
┌─────────────────────────────────────────────────┐
│                 TASK POOL                        │
│  Ordered list of task definitions with locks     │
│  (lease-based, auto-expiring)                    │
├─────────────────────────────────────────────────┤
│              WORK ISOLATION                      │
│  Git worktree per active task                    │
├─────────────────────────────────────────────────┤
│              MERGE QUEUE                         │
│  FIFO queue + automated conflict resolution      │
│  + CI verification                               │
├─────────────────────────────────────────────────┤
│              EVENT LOG                           │
│  Append-only structured log of all state changes │
├─────────────────────────────────────────────────┤
│              AGENT LOOP                          │
│  while true:                                     │
│    task = claim_next_matching_task()             │
│    if task: execute(task), submit(task)          │
│    else: sleep(interval)                         │
└─────────────────────────────────────────────────┘
```

That's it. Everything above this is optimization. The question is which optimizations justify their complexity cost.

---

## Step 3: Dialectical Synthesis -- Both/And Reframe

### Tension 1: Hierarchy vs. Autonomy

**Thesis (Gas Town's position):** Hierarchical control ensures coordination. Mayor assigns work, Deacon supervises, Witness monitors workers.

**Antithesis:** Autonomous agents self-organize more efficiently. Hierarchy creates bottlenecks, single points of failure, and supervision overhead that consumes the resources it's supposed to protect.

**Synthesis: Hierarchical Constraints, Autonomous Execution.**

The resolution is not "hierarchy OR autonomy" but **constraint-based autonomy**. The system defines constraints (rules, boundaries, invariants) hierarchically, but agents operate autonomously within those constraints.

Concrete design:

- **Constraint layer** (replaces Mayor): A set of rules encoded in configuration, not in a running AI agent. Rules like: "No more than 3 agents per rig," "Tasks tagged 'security' require agents with security clearance," "Merge queue freezes during deploy windows." These rules are *data*, not an agent. They're like zoning laws -- they constrain behavior without actively managing it.
- **Autonomous agents** (replaces the Polecat-under-supervision model): Agents self-select tasks from the pool, constrained by the rules. They self-monitor via lease renewal. They self-terminate when done. No Witness watches them; the lease expiry mechanism handles failure detection.
- **Escalation** (replaces Boot/Deacon/Witness chain): When an agent encounters a situation outside its constraint envelope (e.g., conflicting merge with no automated resolution), it escalates. Escalation is *exception-driven*, not *patrol-driven*. An on-demand supervisor agent is spawned only when needed, handles the exception, and terminates. This is the immune system model: dendritic cells activate T-cells only when a threat is detected, not on a continuous patrol loop.

This gives you hierarchy (rules and escalation paths) AND autonomy (self-selecting, self-monitoring agents). The hierarchy lives in the *constraints*, not in *agents supervising other agents*.

### Tension 2: Structured Data vs. Flexibility

**Thesis (Gas Town's position):** Work is structured data. Beads have schemas, typed fields, event/label distinction. This enables querying, auditing, and automated processing.

**Antithesis:** Rigid schemas constrain what can be expressed. Real work is messy. Over-structuring forces agents to fit their work into predefined boxes, losing information that doesn't match the schema.

**Synthesis: Schema-on-Read, Not Schema-on-Write.**

Borrow from data lake architecture. Agents write work records in whatever form captures the full reality -- semi-structured (JSON with required fields + optional arbitrary fields). The "schema" is applied when *reading* data for specific purposes (dashboards, routing decisions, audits), not when writing it.

Concrete design:

- **Write side**: Every work event is a JSON blob with three required fields: `id`, `timestamp`, `actor`. Everything else is optional and agent-determined. If an agent wants to record "I tried three approaches before this one worked," it can. If an agent just records "done," that's also valid.
- **Read side**: Views, queries, and dashboards define their own schemas over the raw data. The "audit view" extracts attribution fields. The "routing view" extracts capability tags. The "dashboard view" extracts status and timing. If a field is missing, the view handles it gracefully (null/default).
- **Progressive structuring**: Over time, frequently-used optional fields get promoted to "conventionally expected" (like HTTP headers -- some are required, some are conventional, some are custom). But this emerges from usage, not from upfront schema design.

This gives you structure (queryable, auditable data) AND flexibility (agents aren't constrained by schemas that can't anticipate every situation).

### Tension 3: Comprehensive Tracking vs. Simplicity

**Thesis:** Track everything. Attribution, CVs, performance metrics, event chains, molecule states. This data enables optimization, debugging, and accountability.

**Antithesis:** Tracking overhead slows the system. Every tracked metric is a tax on agent throughput. Most tracked data is never queried. The system spends more time recording what it's doing than doing things.

**Synthesis: Tiered Observability with Lazy Materialization.**

Not all tracking has the same value. Apply the Pareto lens: 20% of tracked data provides 80% of debugging/optimization value.

Concrete design:

- **Tier 0 (Always On, Zero Cost)**: Git commits. They're already being created as part of work execution. Commit messages + author attribution = free observability. This is the irreducible minimum.
- **Tier 1 (Cheap, High Value)**: Append-only event log. Task claimed, task completed, task failed, merge succeeded, merge failed. Simple events, no joins, no complex state management. Stored in flat files (JSONL) or SQLite -- not Dolt.
- **Tier 2 (On-Demand, Materialized When Needed)**: Agent performance summaries, capability profiles, cross-project dependency graphs. These are *computed from* Tier 0 and Tier 1 data when someone asks for them, not maintained in real-time. Lazy materialization means zero cost when nobody's looking.
- **Tier 3 (Opt-In, Research Grade)**: Full session transcripts, decision traces, model comparison data. Only enabled when actively investigating or experimenting. Never on by default.

This gives you comprehensive tracking (everything *can* be derived) AND simplicity (most tracking is deferred or computed rather than maintained in real-time).

### Tension 4: Centralized Coordination vs. Distributed Execution

**Thesis:** Central coordination (Mayor) ensures coherent work distribution, prevents conflicts, and maintains global awareness.

**Antithesis:** Distributed execution is more resilient, scalable, and eliminates the central bottleneck.

**Synthesis: Shared State, Distributed Decisions.**

The resolution comes from distributed systems theory. You don't need centralized coordination OR fully distributed consensus. You need **shared state with local decision-making**.

Concrete design:

- **Shared state**: The task pool (beads database) is the single source of truth. It's not owned by any agent; it's infrastructure. All agents can read it; writes are serialized through simple locking (optimistic concurrency, not a distributed consensus protocol).
- **Local decisions**: Each agent independently evaluates the shared state and makes its own decisions: "Is there work I should claim? Am I stuck? Should I escalate?" No agent tells another agent what to do.
- **Convergent state**: Even if agents make suboptimal local decisions (two agents both start similar tasks, an agent claims work it can't handle), the system converges to correct state through simple mechanisms: lease expiry reclaims abandoned tasks, merge queue detects duplicates, CI catches quality issues.

This is how Git itself works. There's no "Git Mayor" coordinating commits across developers. There's shared state (the repository), local decisions (each developer commits independently), and convergent state (merge resolution). Gas Town could apply the same pattern one level up.

---

## Step 4: Pre-Mortem -- Assume Failure

### Scenario: You redesigned Gas Town using the above insights. Six months later, it failed catastrophically.

### Failure Mode 1: The Thundering Herd

**What happened:** You replaced the Mayor with a self-selection task pool. Twenty agents all poll the task pool simultaneously. When a high-priority task appears, 15 agents try to claim it at once. The locking mechanism handles correctness (only one succeeds) but the failed claims waste tokens. Worse, the 14 rejected agents all fall back to their second choice simultaneously, creating cascading contention. The system spends 60% of its cycles on failed claims and retries.

**What you underestimated:** The Mayor wasn't just assigning work; it was providing **contention management**. By directing agents to specific tasks, it prevented the thundering herd problem. A centralized dispatcher is inefficient in theory but prevents coordination storms in practice.

**The fix you'd need:** Jittered backoff on failed claims. Affinity-based pre-filtering (agents only see tasks matching their capability profile, reducing the candidate pool per agent). Lease pre-reservation (agents declare intent before claiming, enabling others to self-sort). These are the mechanisms that real markets use -- market makers, bid-ask spreads, price bands -- to prevent coordination chaos.

### Failure Mode 2: The Lost Supervisor

**What happened:** You replaced the three-tier watchdog chain with lease-based failure detection. An agent's lease expires, the task returns to the pool -- elegant in theory. But the agent didn't just fail; it went into an infinite loop consuming tokens. It's *alive* (renewing its lease) but *not making progress*. Without a Witness actively checking "is this agent's work advancing?", zombies consume resources indefinitely. The lease says "alive"; the reality says "brain-dead."

**What you underestimated:** Liveness and progress are different properties. A heartbeat (lease renewal) proves liveness. Only semantic inspection proves progress. The Witness wasn't just checking "is the agent alive?" -- it was checking "is the agent *producing useful output*?" That semantic check requires AI reasoning, not just a timer.

**The fix you'd need:** Two-layer health checking. Layer 1: mechanical lease renewal (cheap, handles crashes). Layer 2: periodic progress assertion -- the agent must demonstrate *forward motion* (new commits, state transitions, intermediate outputs) to maintain its lease. If the lease is renewed but no progress markers have changed in T minutes, the lease is forcibly expired. This is still cheaper than a dedicated Witness agent, but it requires instrumenting "progress" as a measurable signal.

### Failure Mode 3: The Schema Drift

**What happened:** You implemented schema-on-read for maximum flexibility. Six months later, every agent writes work records in a slightly different format. The "dashboard view" breaks because Agent A records completion time as `completed_at` and Agent B records it as `finish_time` and Agent C records it as `duration_ms` (a relative measure, not absolute). The capability routing view can't match tasks to agents because capability tags are inconsistent. The audit view produces garbage because attribution fields use different conventions.

**What you underestimated:** Schema-on-write wasn't just bureaucratic rigidity -- it was a **shared vocabulary**. When 20+ agents are writing data, some degree of structural agreement is essential for the data to be useful. Full flexibility degrades into Babel.

**The fix you'd need:** A minimal core schema (5-10 required fields with defined semantics) plus flexible extension fields. This is the HTTP model: required status codes and headers, optional custom headers. Or the JSON-LD model: a shared vocabulary (@context) that agents reference, with freedom to add terms. You need *just enough* schema to maintain interoperability, not full schema enforcement.

### Failure Mode 4: The Emergent Monoculture

**What happened:** You let agents self-select tasks based on capability matching. Over time, agents that are good at easy tasks accumulate great performance records. Agents that tackle hard tasks fail more often and accumulate worse records. The capability-based routing increasingly steers easy work toward "proven" agents and starves them of challenging work that would expand their capability profiles. Meanwhile, hard tasks pile up because no agent has demonstrated capability in those areas (because no agent is routed to them). The system converges on a local optimum: excellent at easy work, incapable of hard work.

**What you underestimated:** Markets need **exploration** as well as **exploitation**. Pure capability-based routing is a greedy algorithm that converges to local optima. The Mayor, for all its inefficiency, occasionally made "bad" assignments that turned out to be learning opportunities. Random exploration has value.

**The fix you'd need:** Epsilon-greedy routing. 80% of tasks are routed by capability match (exploitation). 20% are assigned randomly or to agents with the *least* experience in that area (exploration). This is the multi-armed bandit solution. You'd also need to weight performance records by task difficulty -- an agent that fails at a hard task is not worse than an agent that succeeds at an easy task.

### Failure Mode 5: The Broken Handoff Chain

**What happened:** You simplified context continuity to "a state file written at session end, read at session start." But agents don't always know when their session will end. Context windows fill up mid-thought. Crash recovery produces no state file at all. And the state file, being unstructured, often omits critical context that the writing agent didn't know was important. Three months in, 30% of task completions are wrong because they were based on corrupted or incomplete handoff state.

**What you underestimated:** Gas Town's Handoff + Seance + Molecule system wasn't over-engineered -- it was handling a genuinely hard problem. Context continuity across session boundaries is one of the hardest problems in agent orchestration. A "simple state file" is to this problem what a "simple to-do list" is to project management: it works for trivial cases and collapses for real ones.

**The fix you'd need:** Continuous checkpointing (not just at session end), structured handoff with mandatory fields (current task, progress state, blockers, key decisions made and why), and a fallback mechanism for crash recovery (reconstruct state from git history + event log when no handoff file exists). This brings back some of the complexity you discarded, but in a more principled form.

### Failure Mode 6: The Invisible Dependency

**What happened:** You decoupled everything for maximum modularity. Agents are independent. Tasks are independent. Rigs are independent. Six months later, a frontend agent ships a UI change that assumes a backend API field that another agent is in the process of renaming. Both tasks pass their individual CI checks. Both merge cleanly. The integration fails in production. Without cross-rig dependency tracking (the Convoy/Molecule system you simplified away), no mechanism detected the conflict before it manifested.

**What you underestimated:** Some complexity in Gas Town encodes *real dependencies* in the problem domain. Cross-project references and convoy-level grouping aren't bureaucratic overhead; they're modeling real-world coupling between repositories. You can't simplify away the coupling that exists in the code itself.

**The fix you'd need:** Lightweight dependency declarations on tasks. "This task touches API contract X" is a tag, not a heavyweight molecule. Any other task touching the same API contract is flagged for sequencing or co-validation. This is less than Gas Town's full Convoy/Molecule system but more than nothing.

---

## Alternative Architectures

### Architecture A: The Stigmergic Pool

Inspired by ant colonies + market economics + first principles analysis.

**Core idea:** Eliminate all supervisor agents. Replace hierarchy with environment-mediated coordination.

```
┌─────────────────────────────────────────────────────────┐
│                    TASK POOL (The Pheromone Layer)       │
│                                                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐      │
│  │ Task A  │ │ Task B  │ │ Task C  │ │ Task D  │ ...  │
│  │ pri: 7  │ │ pri: 3  │ │ pri: 9  │ │ pri: 5  │      │
│  │ tags:   │ │ tags:   │ │ tags:   │ │ tags:   │      │
│  │ [swift] │ │ [go,api]│ │ [swift] │ │ [docs]  │      │
│  │ lease:  │ │ lease:  │ │ lease:  │ │ lease:  │      │
│  │ none    │ │ agent-2 │ │ none    │ │ agent-5 │      │
│  │ deps:   │ │ deps:   │ │ deps:   │ │ deps:   │      │
│  │ [C]     │ │ []      │ │ []      │ │ []      │      │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘      │
│                                                          │
│  Rules: (not agents -- config data)                      │
│  - Max 3 concurrent agents per repo                      │
│  - Tasks with unresolved deps are not claimable          │
│  - Lease expires after 30 min without progress marker    │
│  - Priority increases by 1 per hour unclaimed            │
└─────────────────────────────────────────────────────────┘
          │                              ▲
          │ claim(task, agent)            │ complete(task, result)
          ▼                              │
┌─────────────────────────────────────────────────────────┐
│                    AGENT POOL                            │
│                                                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐                │
│  │ Agent 1  │ │ Agent 2  │ │ Agent 3  │ ...            │
│  │ caps:    │ │ caps:    │ │ caps:    │                │
│  │ [swift,  │ │ [go,api, │ │ [swift,  │                │
│  │  ui]     │ │  db]     │ │  test]   │                │
│  │          │ │          │ │          │                │
│  │ loop:    │ │ loop:    │ │ loop:    │                │
│  │ poll ->  │ │ poll ->  │ │ poll ->  │                │
│  │ match -> │ │ match -> │ │ match -> │                │
│  │ claim -> │ │ claim -> │ │ claim -> │                │
│  │ execute  │ │ execute  │ │ execute  │                │
│  │ -> submit│ │ -> submit│ │ -> submit│                │
│  └──────────┘ └──────────┘ └──────────┘                │
│                                                          │
│  Each agent runs: while true { poll, match, claim, do } │
│  No supervisor. No patrol. No watchdog.                  │
└─────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────┐
│                    MERGE QUEUE                           │
│  FIFO. Automated. Runs CI. Retries on conflict.         │
│  Not an agent -- a mechanical process.                   │
└─────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────┐
│                    EVENT LOG                             │
│  Append-only. Every state transition recorded.           │
│  Queries and dashboards are views over this log.         │
└─────────────────────────────────────────────────────────┘
```

**Properties:**
- Zero supervision overhead. No Boot, Deacon, Witness, or Mayor agents consuming tokens.
- Zero coordination cost at zero workload. Agents poll, find nothing, sleep.
- Failure detection via lease expiry + progress markers. Mechanical, not AI-driven.
- Scaling is linear: add agents, they join the pool. No hierarchy to restructure.
- Simplest possible system that satisfies all irreducible requirements.

**Risks:** Thundering herd on hot tasks. No semantic progress checking. No cross-task dependency intelligence.

**Mitigations:** Jittered polling, capability-based pre-filtering, dependency tags on tasks, periodic progress-gate checks (mechanical: "has the git worktree had a commit in the last 15 minutes?").

---

### Architecture B: The Immune System

Inspired by adaptive immunity + urban resilience + dialectical synthesis of hierarchy and autonomy.

**Core idea:** Agents are undifferentiated stem cells that specialize in response to environmental signals. Supervision is triggered by anomaly, not by patrol.

```
┌──────────────────────────────────────────────────────────────┐
│                 CONSTRAINT ENVELOPE (Zoning Laws)            │
│                                                               │
│  Static rules, not a running agent:                          │
│  - Isolation: one worktree per task                          │
│  - Concurrency: max N agents per repo                        │
│  - Quality: all merges pass CI                               │
│  - Safety: tasks tagged "critical" require 2-agent review    │
│  - Budget: total token spend per hour capped at T            │
└──────────────────────────────────────────────────────────────┘
          │ constrains
          ▼
┌──────────────────────────────────────────────────────────────┐
│                    SHARED STATE (The Tissue)                  │
│                                                               │
│  Task Pool + Agent Registry + Event Log + Capability Index   │
│  All stored in SQLite. No Dolt. No special database.         │
│  Read by all agents. Written through serialized transactions.│
└──────────────────────────────────────────────────────────────┘
          │ read/write                    ▲ anomaly detected
          ▼                               │
┌──────────────────────────────────────────────────────────────┐
│                    AGENT POOL (The Cells)                     │
│                                                               │
│  Undifferentiated at birth. Specialize based on what they    │
│  encounter in the shared state.                               │
│                                                               │
│  Agent lifecycle:                                             │
│  1. IDLE: poll task pool for matching work                   │
│  2. ACTIVE: claimed a task, executing in isolated worktree   │
│  3. MERGING: submitted work to merge queue                   │
│  4. DONE: work merged, agent returns to IDLE or terminates   │
│                                                               │
│  Any agent can become a "responder" if it detects anomaly:   │
│  - Stale lease on a task? Reclaim it.                        │
│  - Merge conflict? Attempt resolution.                       │
│  - CI failure pattern? Spawn diagnostic sub-task.            │
│  - System constraint violated? Escalate to human.            │
│                                                               │
│  No dedicated supervisor role. Supervision is a behavior     │
│  any agent can exhibit when it encounters a trigger.         │
└──────────────────────────────────────────────────────────────┘
          │ anomalies that exceed local resolution
          ▼
┌──────────────────────────────────────────────────────────────┐
│              ON-DEMAND ESCALATION AGENT                       │
│                                                               │
│  Not persistent. Spawned only when an anomaly exceeds the    │
│  constraint envelope's automated resolution rules.            │
│  Handles the exception. Terminates.                          │
│                                                               │
│  This is the T-cell: activated by antigen presentation,      │
│  clonally expands if needed, apoptoses when threat cleared.  │
└──────────────────────────────────────────────────────────────┘
          │ unresolvable exceptions
          ▼
┌──────────────────────────────────────────────────────────────┐
│              HUMAN NOTIFICATION                               │
│  Email / SMS / dashboard alert                               │
│  Last resort. System should self-heal 95% of the time.       │
└──────────────────────────────────────────────────────────────┘
```

**Properties:**
- Agents are homogeneous in capability but heterogeneous in behavior. Any agent can do any job; specialization emerges from task history and capability matching.
- Supervision is embedded in the execution layer. Every agent is a potential supervisor. No dedicated monitoring agents burning tokens on patrol.
- Escalation is exception-driven, not schedule-driven. The on-demand escalation agent is spawned only when needed, like an immune response.
- Constraint envelope provides hierarchy (rules) without hierarchy (agents managing agents).
- Resilience through redundancy: no single agent's failure halts the system.

**Risks:** Agents may lack the context to make good supervisory decisions. On-demand escalation may be too slow for time-critical failures. Emergent specialization may not converge.

**Mitigations:** Pre-computed escalation playbooks (if X then Y, encoded in constraint envelope). Fast anomaly detection via mechanical checks (lease expiry, CI failure, git activity). Seed initial capability profiles from task metadata.

---

### Architecture C: The Market Mesh

Inspired by mechanism design + comparative advantage + the dialectical synthesis of central coordination and distributed execution.

**Core idea:** Tasks and agents participate in a continuous matching market. A lightweight "auctioneer" process (not an AI agent) runs the matching algorithm. Prices (priorities) are dynamic.

```
┌──────────────────────────────────────────────────────────────┐
│                    TASK MARKET                                │
│                                                               │
│  Every task has:                                             │
│  - Base priority (set by human or decomposition)             │
│  - Age premium (+1/hour unclaimed, creates urgency)          │
│  - Difficulty estimate (from historical data)                │
│  - Required capabilities (tags)                              │
│  - Dependency locks (blocked until dep resolved)             │
│  - Contract scope (what "done" means, verifiably)            │
│                                                               │
│  Every agent has:                                            │
│  - Capability vector (from historical performance)           │
│  - Current load (0 if idle, >0 if multitasking)             │
│  - Efficiency record (tasks/hour by category)                │
│  - Exploration quota (% of time for unfamiliar tasks)        │
│                                                               │
│  MATCHING ENGINE (mechanical, not AI):                       │
│  - Runs every N seconds                                      │
│  - Computes optimal task-agent pairings using:               │
│    - Capability overlap score                                │
│    - Comparative advantage (not just absolute fit)           │
│    - Load balancing across agents                            │
│    - Exploration budget (epsilon-greedy)                     │
│  - Posts matches. Agents accept or reject.                   │
│  - Rejected matches return to pool with updated metadata.    │
└──────────────────────────────────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────────────────────────┐
│                    EXECUTION LAYER                            │
│                                                               │
│  Identical to Architecture A:                                │
│  - Git worktree isolation                                    │
│  - Lease-based liveness + progress-based health              │
│  - Submit to merge queue on completion                       │
│  - Continuous checkpointing for handoff resilience           │
└──────────────────────────────────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────────────────────────┐
│                    FEEDBACK LOOP                              │
│                                                               │
│  Post-merge analysis (mechanical):                           │
│  - Did CI pass on first try?                                 │
│  - How long did the task take vs. estimate?                  │
│  - Were there merge conflicts?                               │
│  - Was rework needed?                                        │
│                                                               │
│  Updates:                                                    │
│  - Agent capability vectors (Bayesian update)                │
│  - Task difficulty estimates for similar future tasks        │
│  - Matching algorithm weights                                │
│                                                               │
│  This is the invisible hand: prices (priorities, capability  │
│  scores, difficulty estimates) adjust based on outcomes,     │
│  steering future allocation without central planning.        │
└──────────────────────────────────────────────────────────────┘
```

**Properties:**
- The matching engine replaces the Mayor but is mechanical (a matching algorithm), not an AI agent. Zero token cost for coordination.
- Comparative advantage routing means the system extracts maximum value from heterogeneous agents. Agent A does what Agent A is *relatively* best at, not what Agent A is *absolutely* best at.
- Epsilon-greedy exploration prevents the monoculture failure mode. Agents are occasionally assigned unfamiliar tasks to expand capability profiles.
- Dynamic pricing (age premium, difficulty adjustment) ensures no task is permanently starved of attention.
- Feedback loop creates continuous improvement without explicit "kaizen" agents or review cycles.

**Risks:** Matching algorithm complexity. Cold-start problem (no historical data for new agents or new task types). Over-reliance on quantitative metrics may miss qualitative factors.

**Mitigations:** Start with simple matching (random with capability filter) and add sophistication as data accumulates. Default capability profiles for new agents based on model type. Human override capability for strategic prioritization.

---

## Comparative Assessment

| Dimension | Current Gas Town | A: Stigmergic Pool | B: Immune System | C: Market Mesh |
|-----------|-----------------|--------------------|--------------------|----------------|
| **Supervision cost** | High (3 AI supervisor tiers) | Zero | Near-zero (embedded) | Zero (mechanical matcher) |
| **Failure detection** | Active patrol (token-expensive) | Lease expiry (mechanical) | Embedded + escalation | Lease + progress gates |
| **Work assignment** | Centralized (Mayor) | Self-selection | Self-selection | Algorithmic matching |
| **Scaling** | Requires hierarchy restructuring | Linear (add agents) | Linear (add agents) | Linear (add agents) |
| **Cold start simplicity** | Complex (many components) | Very simple | Simple | Moderate |
| **Coordination intelligence** | High (AI reasoning) | Low (mechanical) | Medium (distributed) | Medium (algorithmic) |
| **Cross-task awareness** | High (convoy/molecule) | Low (dependency tags only) | Medium (any agent can inspect) | Medium (matcher sees all) |
| **Resilience** | Fragile (SPOF at each tier) | High (no SPOF) | Very high (immune model) | High (no SPOF) |
| **Concept count** | 13 roles, 10+ work types | 3 concepts (task, agent, merge) | 4 concepts (task, agent, constraint, escalation) | 5 concepts (task, agent, match, feedback, constraint) |

---

## Recommended Path

No single alternative architecture is categorically superior. The pre-mortem reveals that Gas Town's complexity encodes real problems (semantic progress checking, cross-repo dependencies, context continuity) that the simpler architectures must still solve.

The recommended path is **Architecture B (Immune System) as the foundation, with selective elements from C (Market Mesh)**:

1. **Replace the three-tier watchdog chain with embedded anomaly detection.** Every agent checks its own health. Escalation agents spawn on-demand. This eliminates the largest source of wasted tokens (continuous AI patrol) while preserving failure recovery.

2. **Replace the Mayor with an algorithmic matcher.** A mechanical process (Go code, not AI) that matches tasks to agents based on capability vectors and comparative advantage. This preserves intelligent routing without burning AI tokens on coordination.

3. **Consolidate work unit types.** One type: Task. With fields for priority, capabilities, dependencies, scope, and flexible metadata. Molecules, wisps, formulas, and convoys become query patterns over tasks, not separate types.

4. **Replace Dolt with SQLite.** For a single-installation system, SQLite handles the shared state layer without the operational overhead of a full SQL server. Dolt's branching model for beads is elegant but costs more in complexity than it saves in capability.

5. **Keep what works.** Git worktree isolation, merge queue, persistent identity, structured event logging, session handoff, and the GUPP principle (if work is on your hook, you run it) all survive the redesign. They emerged from the first principles analysis as genuinely irreducible.

6. **Implement the feedback loop from Architecture C.** Post-completion analysis updates capability vectors and difficulty estimates. This creates continuous improvement without dedicated review agents.

The result would be a system with roughly 5 core concepts instead of 23+, zero AI supervision overhead, mechanical failure detection, algorithmic work routing, and the same functional capabilities. The complexity budget freed up by eliminating accidental complexity can be reinvested in the genuinely hard problems: semantic progress checking, cross-repo dependency management, and robust context continuity.
