# Grand Synthesis: Gas Town Multi-Agent Orchestration

The definitive analysis, synthesized from seven independent evaluations of Gas Town's architecture, operational logs, design principles, and competitive position. This document drives decisions.

---

## Section 1: Convergence Map

Findings ranked by the number of independent analyses that arrived at the same conclusion.

### High-Confidence Findings (4+ analyses agree)

**1. The supervision hierarchy is the system's primary pathology, not its strength.**
*Identified by: Wrong-Problem Detector, Innovation Engine, 11-Perspective Analysis (8 of 11 perspectives), Blind Spot Finder, Ford Audit, First Principles*

Every analysis independently concluded that the three-tier watchdog chain (Daemon -> Boot -> Deacon -> Witness) costs more than it prevents. The empirical evidence is unambiguous: the Deacon was restarted every 6 minutes for over two hours, producing zero useful work. The Witness hands off every 8 minutes due to context exhaustion from health-check message traffic. 111 nudges were generated for 5 completed tasks -- a 22:1 coordination-to-work ratio. The supervision infrastructure is the dominant consumer of tokens and the dominant source of failures. This is the single highest-confidence finding across all analyses.

**2. AI agents are being used for mechanical tasks that deterministic code handles better.**
*Identified by: Wrong-Problem Detector, Innovation Engine, Perspectives (SRE, Economist, Lean Engineer), Blind Spot Finder, Ford Audit, First Principles*

Health checking, heartbeat monitoring, merge queue processing, message relay, nudge routing, worktree cleanup, and process restart are all mechanical operations. They do not benefit from LLM reasoning. They are degraded by LLM unreliability (context exhaustion, API failures, stale sessions). The Ford Audit quantified this: of 36 total responsibilities across 8 agent roles, only 7 (19%) require AI reasoning. The other 29 (81%) are mechanizable. Gas Town uses its most expensive, least reliable component (LLM API calls) for its most routine, most reliability-critical functions.

**3. The concept count and naming system create an unnecessarily steep barrier.**
*Identified by: Perspectives (7 of 11: End User, Child, Competitor, Cognitive Scientist, Systems Architect, Lean Engineer, Historian), Blind Spot Finder (metaphor breakdown analysis), First Principles*

13 agent roles, 10 work unit types, 8 message types, 4 gate types, 4 escalation levels, 20+ named concepts with Mad Max-themed names that require translation on every interaction. The Cognitive Scientist perspective identified that this exceeds Miller's Law (7 +/- 2 chunks in working memory). The Blind Spot Finder demonstrated that 6 of 7 core metaphors break down under examination (Town has no economy, Polecat has no continuity, Refinery does not refine, Witness intervenes rather than observes, Handoff has no receiver present, Convoy members do not move together). The Competitor perspective identified this as the primary attack surface: a rival system with 3 concepts and 5-minute onboarding would win the 90% use case.

**4. The core data model and work-as-structured-data insight is genuinely valuable.**
*Identified by: Perspectives (9 of 11), Ford Audit, First Principles, Innovation Engine*

Persistent identity, structured work tracking, attribution, automated merge coordination, git worktree isolation, and hook-based crash-resilient assignment are real differentiators that solve real problems. The Competitor perspective identified attribution and merge coordination as moats. The SRE perspective valued crash-resilient work state. The Security perspective valued non-repudiation. This core must be preserved. The problem is not the core insight -- it is the 80% of machinery built around it.

**5. Polling-based supervision should be replaced with event-driven reactions.**
*Identified by: Innovation Engine, Perspectives (SRE, Economist, Lean Engineer, Evolutionary Biologist, Systems Architect), Ford Audit, First Principles*

Continuous patrol loops fill agent context windows, trigger premature handoffs, and burn tokens during idle periods. The Ford value stream analysis showed that steps 4, 15, 18, and 22 in the current pipeline are pure wait states (0-8 minutes each) caused by polling gaps. An event-driven architecture -- where state changes trigger immediate processing -- eliminates these gaps entirely and reduces idle-period cost to near zero.

**6. The system is designed for a scale it has not reached, and the overhead may prevent reaching it.**
*Identified by: Wrong-Problem Detector, Architecture Analysis, Perspectives (Economist, Competitor), First Principles*

One rig (hermes). A handful of polecats. The architecture supports 5-20 rigs with federation, cross-project references, capability-based routing, and A/B testing. These features provide value at scale but impose overhead at small scale. The Wrong-Problem Detector framed this as a bootstrapping paradox: the overhead of the scale-system consumes the resources that would produce the work that would justify the system.

### Medium-Confidence Findings (2-3 analyses agree)

**7. Session handoffs are harder than Gas Town acknowledges but also more complex than necessary.**
*Identified by: Wrong-Problem Detector, Innovation Engine (Pre-Mortem), Blind Spot Finder*

The Wrong-Problem Detector argued handoffs create an illusion of continuity where none exists. The Blind Spot Finder showed the "handoff" metaphor breaks down (no receiver is present; it is a message in a bottle). Yet the Innovation Engine's Pre-Mortem (Failure Mode 5) warned that oversimplifying handoffs causes 30% of completions to be wrong. The truth is in the middle: the current protocol is over-engineered, but a "just read the filesystem" approach is insufficient. Continuous checkpointing with structured mandatory fields is the right level.

**8. Cross-task dependencies are a real problem that cannot be fully eliminated.**
*Identified by: Innovation Engine (Pre-Mortem Failure Mode 6), Blind Spot Finder, Perspectives (Systems Architect)*

The Innovation Engine's pre-mortem predicted that decoupling everything leads to invisible cross-repo integration failures. The Blind Spot Finder noted that semantic merge validation is missing -- merges that produce no textual conflicts but break the build are the most common and most damaging failure mode. Dependencies exist in the code itself; the orchestration system must model them at some level. Lightweight dependency tags on tasks (not full Convoy/Molecule abstractions) are the right approach.

**9. Market/stigmergic coordination is theoretically superior but needs contention management.**
*Identified by: Innovation Engine (Architectures A, B, C), First Principles (Architecture C: Claim Board)*

Multiple analyses proposed replacing hierarchical assignment with self-selection from a shared pool. But the Innovation Engine's Pre-Mortem (Failure Mode 1: Thundering Herd) identified the risk: 15 agents all claiming the same hot task creates cascading contention. The fix is jittered polling, capability-based pre-filtering, and lease pre-reservation. Pure self-selection is naive; managed self-selection with contention controls is the right model.

### Notable Solo Findings

**10. The GUPP principle creates a velocity illusion** (Blind Spot Finder only). "If work is on your hook, YOU RUN IT" sounds like high throughput, but starting work before prerequisites are met produces rework cycles. GUPP + "done means gone" creates a context destruction race: the polecat is destroyed before merge success is confirmed.

**11. Persistent identity creates perverse selection oscillation** (Blind Spot Finder only). If an agent gets a good CV from easy tasks, it gets harder tasks, performs worse, gets easy tasks again. The system treats identity as continuous but LLM sessions are stateless.

**12. The system has no backpressure, no token budgeting, and no graceful degradation** (Blind Spot Finder only). Every production queuing system has these. Gas Town cannot slow down when overwhelmed, cannot cap costs, and cannot fall back to a reduced-capability mode when supervision fails.

---

## Section 2: The Verdict on Gas Town

### Is Gas Town solving the right problem?

**Partially.** Gas Town correctly identifies that parallel AI agents working on code need structured coordination. The four irreducible functions -- task distribution, work isolation, merge integration, and failure detection -- are the right problems. But Gas Town wraps these four functions in an organizational metaphor (Mayor, Deacon, Witness, Polecat hierarchy) that imports assumptions from human management that do not apply to LLM-based agents. AI agents do not learn across sessions, do not form working relationships, do not develop institutional knowledge, and are not cheap enough to waste on supervision. The system solves "How do we build an org chart for AI agents?" when the right question is "How does a human direct N parallel coding agents with minimum overhead?"

### What is the right problem?

**"Given N AI agents and M tasks across K repositories, complete the tasks correctly with minimum overhead in tokens, time, and human attention."**

This reframes the problem from orchestration (implying a conductor managing a symphony) to logistics (implying efficient routing of independent workers). The distinction matters because orchestration assumes coordination is the hard part, while logistics assumes the work is the hard part and coordination should be as invisible as possible.

### What is Gas Town's genuine innovation?

Five things Gas Town got right that should be preserved:

1. **Work-as-structured-data.** Treating agent work as queryable, auditable data rather than ephemeral terminal output is a genuine insight. The bead concept (minus its over-engineered schema) is sound.

2. **Git worktree isolation for parallel agents.** Using git's native worktree mechanism to give each agent a conflict-free workspace is the correct primitive. Simple, proven, zero-conflict.

3. **Hook-based crash-resilient assignment.** Durable task assignment that survives agent crashes is essential. If an agent dies, its work must be recoverable. The hook mechanism (now a field on a task record) solves this.

4. **The Propulsion Principle (minus the acronym).** "If you have work, execute immediately" eliminates coordination latency. Agents should not wait for permission. This principle is correct as a default, with the caveat that prerequisite checking should happen before assignment, not after.

5. **Attribution on every action.** Knowing which agent produced which commit is invaluable for debugging, auditing, and quality assessment. This should be preserved, simplified to git author fields rather than a separate attribution system.

### What is Gas Town's fundamental mistake?

**Using AI agents for non-AI work, then building more AI agents to supervise the first set.**

The system does not distinguish between tasks that require AI reasoning (work decomposition, code generation, conflict resolution, architectural judgment) and tasks that are mechanical (health checking, process restart, merge execution, message relay, worktree cleanup, status tracking). It applies expensive, unreliable LLM calls uniformly to both categories, then constructs three layers of AI supervision to compensate for the unreliability that the uniform approach introduces.

The result: a factory where the supervisors outnumber the workers, the supervisors are more expensive and less reliable than the workers, and the primary observable behavior is the supervision system restarting itself.

---

## Section 3: The Seven Essentials

Multi-agent AI orchestration irreducibly requires these seven capabilities. Everything else is implementation choice, optimization, or speculation.

### Essential 1: Task Definition and Storage
**What:** A way to specify what needs doing, with an identifier, description, status, and assignment.
**Why irreducible:** Without defined tasks, agents have nothing to work on. Without durable storage, task state is lost on crash.
**Which analyses:** First Principles (Requirement 1), Innovation Engine (Layer 0), Ford Audit (work tracking core), Architecture Analysis (beads as essential).
**Minimum implementation:** A JSON file per task with 10 fields: id, title, description, status, assigned_to, branch, rig, created_by, timestamps, result.

### Essential 2: Work Isolation
**What:** Agents working in parallel must not corrupt each other's state.
**Why irreducible:** Without isolation, parallel agents overwrite each other's changes. This is the mechanism that makes parallelism safe.
**Which analyses:** All seven analyses agree this is irreducible. The Wrong-Problem Detector called it "the only truly irreducible element."
**Minimum implementation:** Git worktrees. One worktree per active task.

### Essential 3: Task-Agent Binding
**What:** At most one agent works on each task at a time. The binding survives agent crashes.
**Why irreducible:** Without binding, you get duplicate work or lost work. Without crash resilience, agent death orphans tasks permanently.
**Which analyses:** First Principles (Requirement 2), Innovation Engine (Requirement 2), Ford Audit (hook core as load-bearing).
**Minimum implementation:** A field on the task record (`assigned_to`) plus a lease mechanism (heartbeat timestamp; if stale beyond threshold, task returns to pool).

### Essential 4: Integration Pipeline
**What:** Completed parallel work must be merged back into a shared baseline with conflicts detected and resolved.
**Why irreducible:** Without integration, parallel branches diverge permanently. Merge conflicts are inevitable in parallel development.
**Which analyses:** All seven analyses agree merge queue is load-bearing. The Innovation Engine emphasizes the function is essential but the AI implementation (Refinery) is not.
**Minimum implementation:** A FIFO merge queue with automated CI verification. Mechanical for clean merges; AI agent spawned on-demand for conflict resolution.

### Essential 5: Failure Detection and Recovery
**What:** Stuck or dead agents must be detected and their tasks reclaimed.
**Why irreducible:** Without detection, a single agent crash silently halts all work on its assigned tasks forever.
**Which analyses:** All seven analyses agree detection is essential. Six of seven agree the AI implementation (three-tier watchdog) is wrong. The Innovation Engine's Immune System model and the Ford Audit's mechanical supervision are the converged solution.
**Minimum implementation:** Lease-based heartbeat with two layers: (1) mechanical liveness (daemon checks if agent process is alive), (2) progress assertion (has the worktree had a commit in the last N minutes?). No AI required for detection. AI invoked only for recovery decisions that require reasoning.

### Essential 6: Context Continuity
**What:** Work must survive AI session boundaries (context window exhaustion, API errors, crashes).
**Why irreducible:** LLM context windows are finite. Long-running tasks will exceed them. Without continuity, work restarts from scratch at every session boundary.
**Which analyses:** First Principles (Requirement 9), Innovation Engine (Pre-Mortem Failure Mode 5), Architecture Analysis (handoff as essential). The Wrong-Problem Detector challenged the current implementation but not the need.
**Minimum implementation:** Continuous checkpointing: agent periodically writes a structured state file with mandatory fields (current task, progress state, blockers, decisions made). On crash recovery, next session reads filesystem state + checkpoint file. More than "just read the filesystem" but less than the full Handoff + Seance + Molecule apparatus.

### Essential 7: Work Decomposition
**What:** Complex goals must be broken into parallelizable tasks.
**Why irreducible:** A human saying "build feature X" cannot be directly executed by a parallel agent pool. Something must translate intent into discrete, assignable units.
**Which analyses:** First Principles (Requirement 1 + Coordinator role), Wrong-Problem Detector (Mayor's decomposition function is "partially load-bearing"), Ford Audit (Mayor becomes "Escalation Judge" but decomposition remains).
**Minimum implementation:** An on-demand AI session (not a persistent agent) invoked when the human submits a goal. It produces task files. Between invocations, it does not run.

---

## Section 4: The Recommended Architecture

Synthesized from the Stigmergic Pool (Innovation Engine Architecture A), Immune System (Architecture B), Market Mesh (Architecture C), Sweet Spot (First Principles Architecture B), and the Pre-Mortem failure analysis. This architecture preserves Gas Town's innovations, eliminates accidental complexity, addresses identified failure modes, and is buildable in 3-4 weeks.

### Overview

```
+------------------------------------------------------------------+
|                    CONSTRAINT LAYER (Config, not agents)           |
|                                                                    |
|  Rules encoded as data:                                           |
|  - Max N concurrent workers per rig                               |
|  - Lease timeout: 30 min without progress commit                  |
|  - Merge requires CI pass                                         |
|  - Token budget: $X/hour ceiling                                  |
|  - Backpressure: pause spawning when merge queue > M              |
+------------------------------------------------------------------+
          |
          v
+------------------------------------------------------------------+
|              SHARED STATE (Filesystem, not database)               |
|                                                                    |
|  ~/gt/.gt/tasks/{id}.json      -- task records                    |
|  ~/gt/.gt/agents/{id}.json     -- agent registry + heartbeats     |
|  ~/gt/.gt/messages/{id}.json   -- inter-agent messages            |
|  ~/gt/.gt/config.json          -- rig paths, constraints          |
|  ~/gt/.gt/events.jsonl         -- append-only event log           |
+------------------------------------------------------------------+
          |
          v
+------------------------------------------------------------------+
|                     GO DAEMON (Mechanical)                         |
|                                                                    |
|  Single background process. No AI. Heartbeat every 60 seconds.   |
|  Responsibilities:                                                |
|  - Check agent heartbeats (restart dead agents)                   |
|  - Check progress markers (flag stalled agents)                   |
|  - Detect new tasks, spawn workers into pre-configured worktrees  |
|  - Run merge queue: git merge + CI for clean merges               |
|  - Token/cost accounting and budget enforcement                   |
|  - Backpressure: stop spawning when system is overloaded          |
|  - Emit structured events to events.jsonl                         |
|  - Manage tmux sessions                                           |
+------------------------------------------------------------------+
          |
          v
+------------------------------------------------------------------+
|                      AI AGENT LAYER                                |
|                                                                    |
|  ROLE 1: Coordinator (on-demand, not persistent)                  |
|    Invoked by: gt work "description" OR daemon escalation         |
|    Does: Decompose goals into tasks, assign to workers            |
|    Handles: Escalation from daemon when automated recovery fails  |
|    Token cost when idle: $0                                       |
|                                                                    |
|  ROLE 2: Worker (ephemeral, N per rig)                            |
|    Invoked by: Daemon when unassigned task exists                 |
|    Does: Write code. That is its only job.                        |
|    Receives: Pre-configured worktree with task description file   |
|    Outputs: Code changes committed to branch                     |
|    Lifecycle: Daemon handles checkout, push, cleanup.             |
|    Worker's context is 100% code, 0% lifecycle management.       |
|    Heartbeat: Claude Code hook touches timestamp file per tool use|
|    Handoff: Writes checkpoint.md on context pressure. Exits.     |
|              Daemon restarts; next session reads checkpoint.       |
|                                                                    |
|  ROLE 3: Resolver (on-demand, rare)                               |
|    Invoked by: Daemon when mechanical merge fails with conflicts  |
|    Does: AI-powered merge conflict resolution                     |
|    Token cost when no conflicts: $0                               |
|                                                                    |
|  ROLE 4: Crew (orthogonal, human-directed)                        |
|    Persistent workspace for interactive human-agent collaboration  |
|    Not part of automated orchestration pipeline                   |
+------------------------------------------------------------------+
```

### Agent Roles: 3 automated + 1 interactive

| Role | Lifecycle | When Active | Token Burn When Idle |
|------|-----------|-------------|---------------------|
| **Coordinator** | On-demand | When human submits goal or escalation occurs | $0 |
| **Worker** | Per-task ephemeral | When unassigned tasks exist | $0 |
| **Resolver** | On-demand | When merge conflicts occur (rare) | $0 |
| **Crew** | Persistent | When human is interacting | Human-controlled |

Total idle supervision cost: **$0**. Compare to current Gas Town: continuous token burn from Mayor + Deacon + Boot + Witness + Refinery.

### Data Model: 3 structures

**Task** (one JSON file per task):
```json
{
  "id": "t-abc123",
  "title": "Add user authentication endpoint",
  "description": "Acceptance criteria...",
  "status": "open | assigned | in_progress | done | failed",
  "assigned_to": "worker-03",
  "branch": "gt/t-abc123",
  "rig": "hermes",
  "created_by": "coordinator",
  "created_at": "2026-02-19T10:00:00Z",
  "updated_at": "2026-02-19T10:05:00Z",
  "result": "",
  "parent_id": "",
  "deps": ["t-xyz789"],
  "tags": ["swift", "auth"],
  "priority": 5
}
```

**Agent** (one JSON file per agent):
```json
{
  "id": "worker-03",
  "role": "worker",
  "rig": "hermes",
  "status": "active | idle | dead",
  "current_task": "t-abc123",
  "worktree": "/path/to/worktree",
  "heartbeat": "2026-02-19T11:29:45Z",
  "last_progress": "2026-02-19T11:28:00Z"
}
```

**Message** (JSON files in a shared directory, 4 types):
- `task_done` -- worker reports completion
- `merge_result` -- daemon reports merge outcome (success, conflict, test failure)
- `help` -- escalation request
- `handoff` -- session continuity checkpoint

### Communication Mechanism

**The filesystem is the message bus.** No custom protocol. No database server. No Dolt.

- Agents write JSON files to `~/gt/.gt/messages/`.
- The daemon polls messages every 30-60 seconds and dispatches actions.
- Workers read merge results from their message files.
- Heartbeats are timestamp fields in agent JSON files, updated by Claude Code hooks on tool use.
- Handoffs are markdown files in the worktree directory.
- Events are appended to `events.jsonl` for observability.

### Infrastructure: 4 components

1. **Go CLI** (`gt`) -- the interface. `gt work`, `gt status`, `gt log`, `gt task list`.
2. **Go daemon** -- mechanical process manager. No AI. Spawns/restarts agents, runs merges, checks heartbeats, enforces constraints.
3. **tmux** -- session management for agent processes.
4. **git** -- worktree isolation and version control. Standard git, no extensions.

No Dolt. No beads CLI. No special database. The filesystem and git provide all needed durability and versioning.

### Estimated Go Packages: 12-15

`cmd/gt`, `daemon`, `coordinator`, `worker`, `resolver`, `task`, `message`, `agent`, `git`, `tmux`, `claude`, `config`, `events`, `constraints`

Down from 61. An 80% reduction.

### How This Addresses Pre-Mortem Failure Modes

| Failure Mode | How Addressed |
|-------------|--------------|
| **Thundering Herd** (15 agents claim same task) | Daemon assigns tasks, not agents self-selecting. Assignment is serialized in the daemon's heartbeat loop. No contention. |
| **Lost Supervisor** (zombie agent renewing lease but not progressing) | Two-layer health: heartbeat (mechanical liveness) + progress markers (has worktree had a commit recently?). Daemon checks both. |
| **Schema Drift** (agents write inconsistent data) | Minimal core schema with 10 required fields on tasks. Enforced by the Go task package, not by agent discipline. |
| **Emergent Monoculture** (capability routing converges on local optimum) | Deferred to later. Start with round-robin assignment or human-directed assignment. Capability routing is an optimization for scale, not a launch requirement. |
| **Broken Handoff Chain** (incomplete context on session restart) | Continuous checkpointing with mandatory fields. Fallback: reconstruct from git history + task description when no checkpoint exists. |
| **Invisible Dependency** (cross-repo integration failure) | `deps` field on tasks. Tasks with unresolved deps are not assignable. Lightweight, not full Convoy/Molecule. |
| **No backpressure** | Daemon enforces: pause spawning when merge queue depth > M or token spend > budget. |
| **No cost accounting** | Daemon tracks token consumption per agent per hour. Circuit breaker: stop all agents if budget exceeded. |

---

## Section 5: What to Kill, What to Keep, What to Transform

### Kill (eliminate entirely)

| Feature | Justification |
|---------|--------------|
| **Boot** | Watchdog for a watchdog. Daemon checks agent liveness directly. Zero token cost vs. AI agent cost. |
| **Deacon** | Middle-management relay layer. Its primary observed behavior is going stale and being restarted. All 6 responsibilities are mechanizable. The Ford Audit found 0 of 6 Deacon responsibilities require AI. |
| **Dog** | Sub-agent of a role being eliminated. Daemon goroutines replace it. |
| **Nudge** | Real-time signaling layer. Polling on 30-60 second intervals is fast enough. Nudges consume context window in receiving agents. Eliminating nudges directly extends agent session life. |
| **Seance** | Convenience wrapper around reading files. The handoff file and filesystem state provide the same information. |
| **Convoy** | A list abstraction that can be replaced by a `parent_id` field on tasks. Convoys do not move together, protect each other, or synchronize -- the metaphor is misleading. |
| **Molecule / Protomolecule** | Workflow engine bolted onto a task runner. Sequential task creation by the Coordinator achieves multi-step workflows without a separate abstraction. |
| **Wisp** | Ephemeral work record. Binary choice: either you need a record (use a task) or you do not. |
| **Formula** | Workflow template. Requires a template ecosystem (Mol Mall) that does not exist. YAGNI. |
| **Digest / Epic / CV chain** | Reporting and analytics layers for data that does not yet exist at volume. Build when needed. |
| **Federation / HOP** | Cross-workspace coordination for a single-workspace system. Speculative. |
| **Dolt SQL server** | Heavyweight infrastructure for what can be JSON files. Dolt's git-like branching for data is elegant but creates merge semantic nightmares (the Blind Spot Finder flagged this) and adds operational complexity disproportionate to value. |
| **Plugin gate system (4 types)** | Over-specified scheduling framework. A single plugin hook type suffices. |
| **MEOW / GUPP / NDI acronyms** | Naming overhead without functional benefit. Keep the principles, drop the acronyms. |
| **Events-as-truth / label-cache pattern** | Event sourcing is powerful but disproportionate to current needs. Direct state mutation on task records is simpler and sufficient. Add event sourcing later if audit requirements demand it. |

### Keep (preserve as-is or with minimal changes)

| Feature | Justification |
|---------|--------------|
| **Git worktree isolation** | The correct primitive for parallel agent work. All seven analyses agree this is irreducible. No changes needed. |
| **Propulsion principle** ("if you have work, execute immediately") | Eliminates coordination latency. Keep the principle, apply it at the daemon level: when a task is assigned and a worker is available, spawn immediately. |
| **"Done means gone"** (workers self-clean) | Prevents resource leaks. Keep, but shift cleanup responsibility to the daemon (worker exits, daemon cleans worktree after merge confirmation). |
| **Attribution on commits** | Agent ID in git commit author field. Invaluable for debugging and audit. Already built into git; just keep using it. |
| **Crew (human-directed workspace)** | The most naturally specialized role. Single responsibility, clear interface. Already at maximum specialization per the Ford Audit. Keep as an orthogonal feature outside automated orchestration. |
| **tmux session management** | Well-tested infrastructure for agent lifecycle management. Keep. |
| **CLI framework** | The `gt` CLI structure, Cobra command framework, config/paths logic. Reuse. |

### Transform (keep concept, fundamentally redesign implementation)

| Feature | Current | Transformed |
|---------|---------|-------------|
| **Mayor** | Persistent AI agent consuming tokens while idle. 5 responsibilities. | **Coordinator**: On-demand AI session invoked by `gt work` command or daemon escalation. Zero idle cost. 2 responsibilities: decompose work, handle escalation. |
| **Witness** | Persistent AI agent per rig. 7 responsibilities. Hands off every 8 minutes from context exhaustion. | **Daemon goroutine**: Mechanical heartbeat + progress checking. No AI session. AI invoked only for novel recovery decisions. |
| **Refinery** | Persistent AI agent for merge queue. Burning tokens to run `git merge`. | **Daemon function** + **Resolver** (on-demand AI): Daemon runs mechanical merges. AI agent spawned only for conflict resolution. Common case (clean merge) costs $0 in AI tokens. |
| **Polecat** | Worker with 6 responsibilities, 80% of which are lifecycle overhead. | **Worker**: Single responsibility -- write code. Daemon handles worktree setup, branch checkout, push, merge submission, and cleanup. Worker's context is 100% code. |
| **Beads** | 10 work unit types, JSONL + Dolt, event/label cache pattern, full schema. | **Tasks**: 1 type, JSON files, direct state mutation, 10 required fields + flexible metadata. |
| **Mail protocol** | 8 message types, typed messages between named agents. | **Messages**: 4 types (`task_done`, `merge_result`, `help`, `handoff`), JSON files in a shared directory. Filesystem is the message bus. |
| **Health monitoring** (Boot + Deacon + Witness chain) | Three AI supervision tiers. AI agents checking AI agents. | **Daemon heartbeat loop**: One mechanical tier. Daemon checks process liveness + progress markers. Escalates to Coordinator (on-demand AI) only for unresolvable failures. |
| **Handoff** | Structured message to next session with elaborate protocol. | **Checkpoint file**: Worker writes `checkpoint.md` with mandatory fields (task state, progress, blockers, decisions). Next session reads it. Continuous checkpointing, not just at session end. |
| **Three-layer identity** (Identity -> Sandbox -> Session) | Three abstractions with separate lifecycle management. | **Two fields**: `agent_id` (persistent) + `session_id` (ephemeral from tmux). Sandbox maps to worktree path, already tracked on the task. |
| **Escalation** (4 severity levels, email/SMS routing) | Over-specified for a system with empty contacts config. | **Binary**: Handle-it (daemon retries/reassigns) or Human-needed (single notification channel). Add graduated severity when multiple escalation channels are actually configured. |
| **Persistent agent identity** | Named agents (Rust, Chrome, Nitro) with CVs, capability routing. | **Agent ID**: Simple identifier (worker-01, worker-02). No CV, no capability routing at launch. Add performance tracking as an optimization layer when data exists to make it meaningful. Round-robin or tag-based assignment initially. |

---

## Section 6: The Compounding Improvement Stack

From the Ford Audit, ordered so each improvement enables the next. The key insight: these improvements multiply rather than add. 2x * 2x * 2x = 8x, not 6x.

### Improvement 1: Mechanize Supervision (2.5x)
**What changes:** Replace AI-based health monitoring (Boot, Deacon, Witness) with daemon process checks. Tmux liveness, file timestamps, PID monitoring. Zero AI tokens for supervision.
**Direct effect:** Eliminates $0.50-2.00 in supervision tokens per completed task. Eliminates the Deacon restart loop. Eliminates 90% of inter-agent communication (health-check chatter).
**Why it enables everything else:** When supervision messages stop flooding agent context windows, the Witness (now eliminated) no longer hands off every 8 minutes. All remaining agents can dedicate 100% of context to their actual work. This is the conveyor belt -- the infrastructure that makes everything else possible.
**Multiplier: 2.5x**

### Improvement 2: Event-Driven Dispatch (2.5x)
**What changes:** Replace poll-based agent coordination with event-driven triggers. Daemon watches for filesystem events (bead file changes, git pushes) and triggers actions immediately.
**Direct effect:** Eliminates 0-8 minute polling gaps at four points in the current pipeline.
**Why it compounds with #1:** Once supervision is mechanical, the daemon can be the single event processor. It watches for pushes, triggers merges, spawns workers -- all within milliseconds. The combination eliminates the entire "waiting for the next patrol cycle" waste category.
**Cumulative multiplier: 2.5 * 2.5 = 6.25x**

### Improvement 3: Worker Context Purity (2x)
**What changes:** Pre-configure worktrees before worker spawn. Worker receives a fully configured environment (correct branch, assignment details in a file, worktree ready). Its ONLY job is writing code. Post-completion, daemon handles commit, push, merge submission, and cleanup.
**Direct effect:** Worker startup overhead drops from ~75 seconds to ~5 seconds. Worker context is 100% code, 0% lifecycle management.
**Why it compounds with #1 and #2:** A worker with 100% code context produces better code. Better code means fewer merge conflicts. Fewer conflicts means less Resolver invocation. Less Resolver invocation means faster throughput for all workers. This is a second-order compounding effect: quality improvement reduces rework, which improves throughput.
**Cumulative multiplier: 6.25 * 2 = 12.5x**

### Improvement 4: Unified Work Types (1.5x)
**What changes:** Collapse 10 work unit types into 1 (Task with status field and tags). Eliminate Convoy, Hook, Molecule, Protomolecule, Wisp, Formula, Digest, Epic, CV chain as separate concepts.
**Direct effect:** Dramatic reduction in system prompt complexity. Agent instructions become simpler. `gt` CLI surface area shrinks. Developer cognitive load drops.
**Why it compounds with #3:** Simpler work types mean simpler worker instructions. Simpler instructions mean more context for code. More context for code means better output quality. Better quality means fewer rework cycles.
**Cumulative multiplier: 12.5 * 1.5 = 18.75x**

### Improvement 5: Mechanical Merge Pipeline (1.5x)
**What changes:** Daemon runs `git merge` and `make test` directly. AI agent (Resolver) spawned only for conflict resolution (rare). Merges happen within seconds of push.
**Direct effect:** Merge latency drops from 5-20 minutes to seconds. Eliminates the Refinery as a persistent AI agent.
**Why it compounds with everything above:** Faster merges mean the main branch stays closer to current work. Closer main branch means fewer conflicts for subsequent workers. Fewer conflicts means even fewer Resolver invocations. This is a positive feedback loop.
**Cumulative multiplier: 18.75 * 1.5 = ~28x**

### Summary

| # | Improvement | Individual | Cumulative |
|---|------------|-----------|------------|
| 1 | Mechanize supervision | 2.5x | 2.5x |
| 2 | Event-driven dispatch | 2.5x | 6.25x |
| 3 | Worker context purity | 2x | 12.5x |
| 4 | Unified work types | 1.5x | 18.75x |
| 5 | Mechanical merge pipeline | 1.5x | **~28x** |

If these were additive: 2.5 + 2.5 + 2 + 1.5 + 1.5 = 10x.
Because they compound: 2.5 * 2.5 * 2 * 1.5 * 1.5 = **~28x**.

The ~28x estimate applies to effective output per token spent. It combines throughput improvement (faster pipeline), quality improvement (better code from purer context), and cost reduction (eliminated supervision tokens).

---

## Section 7: Migration Strategy

### Approach: Strangle Fig Pattern

Do not rewrite Gas Town. Build the new system alongside it, migrate functionality incrementally, let the old system atrophy. This is less risky than a rewrite and allows rollback at any point.

### Phase 0: Stop the Bleeding (2-3 days)
**Incremental. No breaking changes.**

1. **Fix the Deacon restart loop.** Make health checks a daemon function (Go goroutine checking tmux session liveness and heartbeat timestamps). Stop spawning Boot. Stop spawning Deacon for health checking.
2. **Eliminate health-check nudge traffic.** Remove HEALTH_CHECK, WITNESS_PING, DEACON_ALIVE message types from inter-agent communication. This immediately extends Witness and Deacon session life by reducing context fill rate.
3. **Suppress "no wisp config" warning.** 194 occurrences polluting logs.
4. **Fix duplicate convoy watcher events.** Each event logged twice.
5. **Add token cost tracking.** Instrument per-agent token consumption in daemon logs.

**What this accomplishes:** The most visible operational pain (Deacon death spiral, log noise, wasted tokens) is eliminated without architectural changes. Existing agent roles continue to function; they just stop receiving supervision noise.

### Phase 1: Build the New Task System (3-4 days)
**Parallel system. Old system still works.**

1. Create `task` package with JSON file CRUD operations.
2. Create `~/gt/.gt/tasks/` directory structure.
3. Add `gt task create`, `gt task list`, `gt task show`, `gt task assign` CLI commands.
4. Create `message` package with file-based message passing.
5. Create `agent` package with heartbeat tracking.
6. Verify: Tasks can be created, listed, assigned, and updated as JSON files while old beads system still operates.

### Phase 2: Simplify Workers (3-4 days)
**First agents migrated. Old and new coexist.**

1. Create new worker system prompt: reads task JSON file instead of hooks/beads.
2. Daemon pre-configures worktree before spawning worker: correct branch, task description file placed in worktree.
3. Worker writes heartbeat via Claude Code hook (touches timestamp file on tool use).
4. Worker writes `checkpoint.md` on context pressure (replaces handoff protocol).
5. Worker signals completion by updating task file status and exiting. Daemon detects exit.
6. Test with one worker on one task: create task -> daemon spawns worker into pre-configured worktree -> worker writes code -> worker exits -> daemon detects completion.

### Phase 3: Mechanical Merge Pipeline (2-3 days)
**Refinery agent replaced.**

1. Daemon function: on worker exit, attempt `git merge` to main.
2. Daemon function: on merge success, run CI/validation scripts.
3. Daemon function: on merge conflict, spawn Resolver agent with conflict context.
4. Resolver agent: AI session that resolves the specific conflict and exits.
5. Daemon function: on validation failure, set task status to `failed` with error details.
6. Retire the Refinery as a persistent agent. Merge queue is now a daemon loop.

### Phase 4: Supervisor Elimination (2-3 days)
**Witness agent replaced by daemon.**

1. Daemon goroutine: check all agent heartbeat timestamps every 60 seconds.
2. Daemon goroutine: check progress markers (last git commit timestamp in worktree) every 60 seconds.
3. If heartbeat stale > 5 minutes: restart worker, provide checkpoint file.
4. If progress stale > 30 minutes with heartbeat alive: kill worker, mark task failed, reassign.
5. If task fails 3 times: escalate (spawn Coordinator for reassignment or notify human).
6. Retire Witness as an AI agent. Daemon handles all supervision mechanically.

### Phase 5: On-Demand Coordinator (2-3 days)
**Mayor agent replaced.**

1. `gt work "description"` invokes a single Claude Code session that decomposes the description into task files.
2. Coordinator session assigns tasks to available workers (or signals daemon to spawn workers).
3. Coordinator session handles escalation: reads `help` messages, attempts resolution or notifies human.
4. Coordinator exits after work decomposition. Zero idle cost.
5. Daemon triggers Coordinator on-demand when escalation messages accumulate.
6. Retire Mayor as a persistent agent.

### Phase 6: Cleanup (2-3 days)
**Clean break from old system.**

1. Delete unused Go packages (estimated 45+ packages removed).
2. Remove Dolt dependency and beads CLI dependency.
3. Remove Boot, Deacon, Dog, Witness, Refinery agent templates.
4. Remove unused message types, data structures, and protocols.
5. Remove Convoy, Molecule, Wisp, Formula, Seance, Nudge concepts.
6. Update all CLI commands to reflect new architecture.
7. Update system prompts for Coordinator, Worker, Resolver, Crew.
8. Update documentation.

### Sequence Diagram

```
Phase 0 --> Phase 1 --> Phase 2 --> Phase 3 --> Phase 4 --> Phase 5 --> Phase 6
 Stop        Build       Migrate     Replace      Replace     Replace     Remove
 bleeding    new task    workers     Refinery     Witness     Mayor       dead
             system                                                       code

 [------ Old system still operational ------]  [-- Old roles deprecated --] [Old code gone]
 Day 1-3     Day 4-7     Day 8-11    Day 12-14   Day 15-17   Day 18-20    Day 21-23
```

### Total Timeline: 18-23 days

### What Requires Incremental vs. Clean Break

**Incremental (can coexist with old system):**
- Phase 0: Bug fixes and noise elimination
- Phase 1: New task system (parallel to beads)
- Phase 2: New worker template (can run new workers alongside old polecats)

**Clean break (replaces old component entirely):**
- Phase 3: Mechanical merge pipeline replaces Refinery
- Phase 4: Daemon supervision replaces Witness
- Phase 5: On-demand Coordinator replaces Mayor
- Phase 6: Dead code removal

Each clean-break phase replaces exactly one old component. If a phase fails, only that component needs rollback. Git provides the safety net.

---

## Section 8: Open Questions

### Unresolved Disagreements Between Analyses

**1. How much supervision intelligence is actually needed?**

The Wrong-Problem Detector argues the simplest alternative (a bash script with parallel worktrees) might match Gas Town's output. The Innovation Engine's Pre-Mortem warns that pure mechanical supervision misses progress-vs-liveness distinction and leads to zombie agents. The First Principles analysis lands on the Sweet Spot (3 roles with AI supervisor). The recommended architecture above uses a hybrid (mechanical liveness + progress markers, with AI Coordinator on escalation only), but the right threshold for "when to invoke AI" is not empirically validated.

**What to test:** Run the Lobotomy Test (Wrong-Problem Detector Experiment 1). Compare 10 tasks under the recommended architecture vs. a bash script with parallel worktrees and a simple merge script. Measure wall-clock time, tokens consumed, tasks completed, merge conflicts. If the bash script wins, further simplification is warranted.

**2. Is self-selection (market/stigmergic) better than daemon-assigned dispatch?**

The Innovation Engine and First Principles both proposed self-selection (agents pick their own tasks). The recommended architecture uses daemon-assigned dispatch to avoid the Thundering Herd problem. But daemon assignment reintroduces a central coordinator. At larger scale (15+ agents), self-selection with contention management might outperform centralized dispatch.

**What to test:** Implement both dispatch modes behind a flag. At 5 agents, compare throughput and contention rates. The prediction: daemon assignment wins at small scale (fewer wasted claims), self-selection wins at large scale (less central bottleneck).

**3. What is the right handoff/checkpoint fidelity?**

The Wrong-Problem Detector argues filesystem state alone provides 90% of needed context. The Innovation Engine Pre-Mortem argues that without structured handoffs, 30% of completions are wrong. Neither has empirical data.

**What to test:** Run the Context Window Boundary Test (Wrong-Problem Detector Experiment 6). Compare tasks requiring 3+ sessions with structured checkpoints vs. filesystem-state-only recovery. Measure continuation success rate and work duplication.

**4. Do capability routing and agent CVs actually improve outcomes?**

Multiple analyses noted that persistent identity is an illusion (each session is a fresh LLM with no actual memory). The Blind Spot Finder identified perverse selection oscillation. Yet the Perspectives analysis (Competitor, Economist) identified capability routing as a genuine moat.

**What to test:** Run 100 tasks with random assignment vs. 100 tasks with tag-based capability matching. Measure completion rate, quality, and time. If random assignment performs comparably, defer capability routing. If tag-based matching significantly wins, implement it as the first optimization layer.

**5. How should semantic merge validation work?**

The Blind Spot Finder flagged this as the #3 priority gap: merging code without running tests is merging blind. The recommended architecture includes "daemon runs CI/validation scripts after merge," but the details are unspecified. What tests? How do you catch semantic conflicts between two independently-correct changes? Integration testing across the full codebase after each merge is expensive. Targeted test selection based on changed files is cheaper but potentially misses cross-module regressions.

**What to test:** Instrument the merge pipeline to track how often merges that pass CI individually fail when integrated. If the rate is low (<5%), targeted test selection suffices. If high (>15%), full integration testing after each merge is necessary.

### Questions Not Addressed by Any Analysis

**6. What happens when the human is not available?**

All analyses assume a human is reachable for escalation. In practice, agents may run overnight, over weekends, or during meetings. The system needs an autonomous mode where it degrades gracefully without human input -- pausing non-critical work, completing in-progress tasks, and queuing decisions for human review.

**7. How does the system handle multi-model heterogeneity?**

The Blind Spot Finder noted that the system runs exclusively on Claude Code, making it vulnerable to correlated failures. The recommended architecture does not specify how to integrate multiple AI providers. The `gemini` and `opencode` packages in the current codebase suggest this was planned.

**8. What is the right token budget model?**

The Blind Spot Finder flagged missing token accounting. The recommended architecture includes a cost ceiling, but what should the ceiling be? Per-hour? Per-task? Per-day? How should the system behave when the budget is exhausted: pause all work, pause only new work, or notify human and continue?

**9. How does the system handle repository-level access control?**

Multiple workers have full filesystem access within their worktrees and potentially outside them. The Security perspective flagged this but no analysis proposed a concrete sandboxing mechanism compatible with AI agent workflows.

**10. What is the performance profile at true scale?**

All analyses operate from data gathered at 1 rig with a handful of polecats. The recommended architecture claims linear scaling, but this has not been tested. At 20 rigs with 100 concurrent workers, filesystem-based message passing (polling directories) may become a bottleneck. At what point does the system need a real database or message broker?

---

## Closing: The One-Paragraph Directive

Gas Town's core insight -- that parallel AI agents working on code need structured coordination with durable task assignment, work isolation, automated merging, and failure recovery -- is correct and valuable. Its fundamental mistake is implementing mechanical coordination functions with expensive, unreliable AI agents, then building three layers of AI supervision to compensate. The path forward is to move all mechanical work into the Go daemon (supervision, dispatch, merging, cleanup), reduce AI agents to three on-demand roles (Coordinator, Worker, Resolver), replace 10 work types with 1 (Task), replace Dolt with JSON files, and eliminate the 13-role hierarchy in favor of a flat daemon-driven architecture. This yields an estimated 28x improvement in effective output per token through compounding gains in supervision elimination, event-driven dispatch, worker context purity, work type simplification, and mechanical merging. The migration can be executed incrementally over 3 weeks using the Strangle Fig pattern, with each phase independently testable and reversible.
