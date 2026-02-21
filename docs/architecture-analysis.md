# Gas Town: Comprehensive Architectural Analysis

## 1. What Gas Town Is

Gas Town is a **multi-agent orchestration system** for coordinating AI coding agents (primarily Claude Code) across multiple git repositories. It treats agent work as structured data, persists work state in git-backed storage, and provides hierarchical supervision so that 4-30+ agents can work in parallel without losing context, duplicating effort, or going unsupervised.

The `gt` CLI (written in Go, 61 internal packages) is the primary interface. It manages agent lifecycles, work distribution, communication, merge queues, health monitoring, and persistent identity/attribution.

---

## 2. Core Abstractions (Concept Map)

```
┌─────────────────────────────────────────────────────────────────────┐
│                         TOWN  (~/ gt/)                              │
│  The workspace root. Contains all rigs, agents, and coordination.  │
│                                                                     │
│  ┌──────────┐  ┌──────────┐  ┌─────────┐  ┌──────────────────────┐ │
│  │  Mayor   │  │  Deacon  │  │  Boot   │  │  Dogs (infra only)   │ │
│  │ (coord)  │  │ (daemon  │  │ (watch- │  │  - cleanup           │ │
│  │          │  │  beacon) │  │  dog's  │  │  - health checks     │ │
│  │          │  │          │  │  watch-  │  │  - plugin execution  │ │
│  │          │  │          │  │  dog)    │  │                      │ │
│  └──────────┘  └──────────┘  └─────────┘  └──────────────────────┘ │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                     RIG  (per-project)                          ││
│  │  ┌──────────┐ ┌──────────┐ ┌───────────────┐ ┌──────────────┐ ││
│  │  │ Witness  │ │ Refinery │ │   Polecats    │ │    Crew      │ ││
│  │  │ (health  │ │ (merge   │ │ (ephemeral    │ │ (persistent  │ ││
│  │  │  monitor)│ │  queue)  │ │  workers)     │ │  human-dir.) │ ││
│  │  └──────────┘ └──────────┘ └───────────────┘ └──────────────┘ ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

### Abstraction Inventory

| Concept | Metaphor | Purpose | Persistence |
|---------|----------|---------|-------------|
| **Town** | Headquarters | Root workspace, cross-rig coordination | Permanent |
| **Rig** | Project container | Wraps a git repo with agents | Permanent |
| **Mayor** | Chief of staff | Creates convoys, distributes work, escalation target | Singleton, persistent |
| **Deacon** | Daemon beacon | Background patrol, health monitoring, plugin execution | Singleton, persistent |
| **Boot** | Watchdog's watchdog | Ensures Deacon is alive | Ephemeral per tick |
| **Dog** | Infrastructure helper | Deacon's crew for background tasks | Variable lifecycle |
| **Witness** | Per-rig supervisor | Monitors polecats, triggers recovery | One per rig, persistent |
| **Refinery** | Merge processor | Manages merge queue per rig | One per rig, persistent |
| **Polecat** | Worker agent | Executes assigned tasks in isolation | Persistent identity, ephemeral sessions |
| **Crew** | Human workspace | Long-lived, human-directed agent | Persistent |
| **Bead** | Work unit | Git-backed atomic issue/task in JSONL | Permanent |
| **Convoy** | Work order | Batch of beads assigned together | Lifecycle-tracked |
| **Hook** | Assignment pin | Pinned bead = current work assignment | Per-agent |
| **Molecule** | Chained workflow | Multi-step process surviving restarts | Durable |
| **Wisp** | Ephemeral bead | Lightweight transient work record | Destroyed after use |
| **Formula** | Workflow template | TOML-based reusable workflow definition | Permanent |
| **Nudge** | Real-time message | Immediate inter-agent notification | Ephemeral |
| **Mail** | Async message | Structured agent-to-agent communication | Persistent until ack'd |
| **Handoff** | Session transfer | Context preservation across session boundaries | One-time |
| **Seance** | History query | Query predecessor sessions for context | Read-only |
| **Patrol** | Health loop | Continuous monitoring cycle | Ephemeral iterations |

---

## 3. System Architecture Layers

### Layer 1: Infrastructure (Go daemon)
- **Daemon process** runs 3-minute heartbeat ticks
- Spawns/restarts Boot, Deacon, Witness, Refinery sessions via tmux
- Manages Dolt SQL server (port 3307) for beads storage
- Convoy watcher and feed curator run as background goroutines
- Pure mechanical — no AI reasoning at this layer

### Layer 2: Supervision (AI agents)
- **Boot** triages Deacon health (ephemeral, stateless)
- **Deacon** patrols all rigs, dispatches plugins, coordinates recovery
- **Witness** per-rig polecat lifecycle management
- Three-tier watchdog chain: Daemon -> Boot -> Deacon -> Workers

### Layer 3: Work Execution (AI agents)
- **Polecats** execute discrete tasks in isolated git worktrees
- **Crew** provides long-lived human-directed workspaces
- **Refinery** handles automated merge queue processing
- All workers use the Propulsion Principle: hook -> execute immediately

### Layer 4: Data (Beads/Dolt)
- Two-level beads: Town-level (`hq-*`) for coordination, Rig-level for implementation
- Dolt SQL provides versioned, git-like database semantics
- Each polecat gets its own Dolt branch, merged on completion
- Routes file maps ID prefixes to rig locations
- Redirects allow worktrees to share canonical beads databases

### Layer 5: Communication
- **Mail protocol** with typed messages (POLECAT_DONE, MERGE_READY, MERGED, etc.)
- **Nudge** for real-time signaling
- **Handoff** for session continuity
- **Escalation** with severity-based routing (bead -> mayor mail -> email -> SMS)
- **Hooks** (Claude Code hooks) for session lifecycle events (start, compact, submit, stop)

---

## 4. Key Design Principles

| Principle | Description |
|-----------|-------------|
| **GUPP (Propulsion)** | "If work is on your hook, YOU RUN IT." No waiting, no confirming. |
| **NDI (Nondeterministic Idempotence)** | System achieves useful outcomes despite individual agent unreliability |
| **MEOW (Molecular Expression of Work)** | Decompose work into trackable atomic units |
| **Attribution is mandatory** | Every action traces to a specific agent identity via BD_ACTOR |
| **Work is structured data** | Not prose, not tickets — queryable, auditable data |
| **Events are truth, labels are cache** | Immutable event beads + fast-query label cache on role beads |
| **Done means gone** | Polecats self-clean; no idle pool |
| **Three-layer identity** | Identity (permanent) -> Sandbox (per-assignment) -> Session (ephemeral) |

---

## 5. Strategic Rationale (Why These Features)

Gas Town addresses visibility gaps that traditional dev infrastructure doesn't cover: "CI/CD tracks builds, not capability. Git tracks commits, not agent performance."

**Entity Tracking & Attribution**: Every agent gets a distinct identity. Work attribution flows across git commits, event logs, and audit records — enabling precise debugging and compliance rather than generic "AI Assistant" credits. This is the foundation everything else builds on.

**Work History (Agent CVs)**: Agents accumulate performance records. Teams can query success rates by skill area, compare reliability across task types, and A/B test between models with objective metrics. This turns agents from anonymous commodities into trackable resources.

**Capability-Based Routing**: Match work requirements against demonstrated capabilities derived from history data. Eliminates manual assignment bottlenecks; optimizes for task-skill alignment. Only meaningful once you have attribution + work history.

**Recursive Work Decomposition**: Complex initiatives break into hierarchical structures (epics, features, tasks) with automatic rollups. Provides visibility at multiple abstraction levels. Flat issue lists can't capture multi-repo, multi-team reality.

**Cross-Project References**: Explicit dependency tracking between repositories. "Frontend can't ship until backend API lands" — prevents scheduling surprises and clarifies blocking relationships across separate codebases.

**Federation**: Multiple workspaces reference each other for visibility across repos, teams, and organizational boundaries. Designed for distributed environments and contractor coordination.

**Validation & Quality Gates**: Structured verification with attribution metadata. Quality control with audit trails documenting approval chains. "Gates are data, not just policy."

**Real-Time Activity Feed**: Work state streams live, supporting debugging, status awareness, and bottleneck identification across multi-agent operations.

**Design philosophy**: Attribution, data structure, historical tracking, scale, and verification are foundational rather than supplementary. The system is designed for enterprise-scale AI orchestration where you need to answer "who did what, how well, and why" at any point.

---

## 6. Communication Protocol Summary

When a polecat finishes work, it sends POLECAT_DONE to its rig's Witness. The Witness verifies clean git state, then forwards MERGE_READY to the Refinery. The Refinery attempts the merge and responds back to the Witness with one of three outcomes: MERGED (success, Witness nukes the polecat worktree), MERGE_FAILED (non-conflict failure like test errors, Witness notifies polecat for rework), or REWORK_REQUEST (merge conflicts, Witness provides rebase instructions and the polecat retries).

For recovery, the Witness detects zombie or abandoned polecats and sends RECOVERED_BEAD or RECOVERY_NEEDED upstream to the Deacon. The Deacon handles re-dispatch with rate limiting (5-minute cooldown per bead, escalates to Mayor after 3 failures).

Any agent can send HELP to the Mayor for escalation when stuck or blocked.

Agents send HANDOFF messages to themselves (their next session) when context windows fill, preserving work state across session boundaries. The next session reads the injected handoff mail on startup and continues from where the previous session left off.

Health monitoring is passive: multiple Witnesses check the Deacon's agent bead "last_activity" timestamp rather than sending heartbeat mail. They only escalate to the Mayor if the Deacon appears unresponsive (stale for more than 5 minutes).

---

## 7. Local Installation State

The `~/gt/` town has:
- **1 rig**: `hermes` (Swift macOS app — command palette / app launcher)
- **Town agents**: Mayor, Deacon, Boot
- **Rig agents**: Witness, Refinery for hermes
- **Daemon**: ran 97 heartbeat cycles before last shutdown
- **Notable pattern in logs**: Deacon frequently goes stale (restarted every ~6 minutes in the log), suggesting either context exhaustion or the Deacon's AI sessions are hitting limits and not self-recovering well

---

## 8. Complexity Inventory

### Concept Count
- **13 named agent roles** (Mayor, Deacon, Boot, Dog, Witness, Refinery, Polecat, Crew + variants)
- **10 work unit types** (Bead, Convoy, Hook, Molecule, Protomolecule, Wisp, Formula, Digest, Epic, CV chain)
- **8 message types** in mail protocol
- **4 gate types** for plugins
- **4 escalation severity levels**
- **3 polecat lifecycle layers** (Identity, Sandbox, Session)
- **2 beads storage levels** (Town, Rig)

### Infrastructure Requirements
- Go 1.23+, Git 2.25+, SQLite3, Tmux 3.0+
- Dolt SQL server (per-town, port 3307)
- Beads CLI (bd) 0.52.0+
- Claude Code CLI (or alternative AI provider)
- Optional: Nix flake for reproducible environment

### Go Package Count
61 internal packages spanning: agent, beads, boot, checkpoint, claude, cli, cmd, config, connection, constants, convoy, copilot, crew, daemon, deacon, deps, doctor, dog, doltserver, events, feed, formula, gemini, git, hooks, keepalive, krc, lock, mail, mayor, mq, nudge, opencode, plugin, polecat, protocol, quota, refinery, rig, runtime, session, shell, state, style, suggest, swarm, templates, tmux, townlog, tui, ui, util, version, wasteland, web, wisp, witness, workspace, wrappers

---

## 9. Tension Points & Observations

### Complexity vs. Value
The system has very high conceptual density. 13 agent roles and 10+ work unit types create a steep learning curve. Many of these (Protomolecule, Wisp, Formula, Mol Mall, Federation/HOP) appear to be future-facing designs not yet fully realized.

### Supervision Overhead
Three tiers of watchdogs (Daemon -> Boot -> Deacon -> Witness) consume AI tokens just to keep each other alive. The logs show the Deacon being restarted every 6 minutes because it goes "stale" — the supervision is burning resources on self-maintenance.

### Metaphor Load
The naming scheme (Town, Rig, Polecat, Wisp, Molecule, Refinery, Convoy, Seance, MEOW, GUPP, NDI) requires substantial domain translation. Each metaphor adds cognitive overhead for anyone entering the system.

### Single-Rig Reality
The local installation has one rig (hermes). Much of the architecture is designed for 5-20 rigs with cross-rig coordination, federation, and model A/B testing — features that provide value at scale but add overhead at small scale.

### Dolt as a Requirement
A full SQL server (Dolt) for bead storage is heavyweight. For small installations, JSONL files or SQLite would suffice. Dolt's git-like branching for beads data is elegant but adds operational complexity.

### AI Token Burn
Every supervision layer (Boot, Deacon, Witness) is an AI agent consuming tokens on patrol cycles. In a quiet town, these agents spin doing nothing productive. The daemon log shows continuous spawning/checking even when no work exists.

---

## 10. Essence Candidates (What Might Be Core vs. Accidental)

Based on this analysis, here are candidates for what may be **essential** vs. **accidental** complexity:

### Likely Essential
1. **Persistent agent identity + attribution** — knowing who did what
2. **Hook-based work assignment** — durable, crash-surviving task assignment
3. **Git worktree isolation** — agents work in parallel without conflicts
4. **Merge queue** — automated integration of parallel work
5. **Structured work units** (beads) — trackable, queryable work items
6. **Async messaging** — agents coordinate without blocking
7. **Health monitoring** — detect and recover stuck agents
8. **Session handoff** — survive context window limits

### Possibly Accidental / Over-Engineered
1. **Three-tier watchdog chain** — could likely be two tiers or one smart daemon
2. **Molecule/Protomolecule/Wisp/Formula** distinction — four work-unit variants may be reducible
3. **Dolt SQL server** — heavyweight for most installations
4. **13 agent roles** — many could be consolidated
5. **Mol Mall / Federation / HOP** — future features adding present complexity
6. **Plugin gate system with 4 types** — over-specified for current usage
7. **MEOW/GUPP/NDI acronyms** — add naming overhead without functional benefit
8. **Seance** — querying previous sessions could be simplified
