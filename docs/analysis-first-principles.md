# First Principles Essential Extraction: Gas Town

An irreducible analysis of what multi-agent AI orchestration actually needs, derived by stripping Gas Town to bedrock and rebuilding from nothing.

---

## Part 1: Regressive Abstraction

Starting from the full Gas Town system (13 agent roles, 10 work unit types, 8 message types, 61 Go packages, Dolt SQL server) and removing layers until only the load-bearing structure remains.

### Round 1: Remove the Unrealized and Future-Facing

**What gets removed:**

| Removed | Reason |
|---------|--------|
| **Federation / HOP** | Cross-workspace linking. No current multi-workspace deployment exists. Pure forward speculation. |
| **Mol Mall** | Marketplace for molecules/workflows. Zero current usage. Concept only. |
| **Protomolecule** | A "proto" version of an already-speculative concept (Molecule). Two levels of unrealized. |
| **Formula** | TOML-based workflow templates. Requires Mol Mall ecosystem to have value. Templates without a template marketplace are just config files. |
| **CV chains** | Agent performance history and capability routing. Requires significant data accumulation before providing value. Premature optimization of agent selection. |
| **Digest** | Summarized work records. Nice-to-have reporting layer. |
| **Epic** | Hierarchical work decomposition above the bead level. Current usage is single-rig, single-project. Not needed until multi-project coordination is real. |
| **Cross-project references** | No second project exists. |
| **A/B model testing** | Requires CV chains and statistical infrastructure that does not exist. |
| **Plugin gate system (4 types)** | Over-specified. Current plugin usage is minimal. A single plugin hook type suffices. |

**What remains after Round 1:**
- 13 -> ~10 agent concepts (Mayor, Deacon, Boot, Dog, Witness, Refinery, Polecat, Crew, plus infrastructure)
- 10 -> 5 work unit types (Bead, Convoy, Hook, Molecule, Wisp)
- 8 message types (unchanged)
- Dolt, tmux, daemon infrastructure (unchanged)

**Reduction: ~35% of concepts removed. No functional capability lost for current operations.**

---

### Round 2: Remove Convenience Over Necessity

**What gets removed:**

| Removed | Reason |
|---------|--------|
| **Wisp** | "Ephemeral bead" -- a lightweight transient work record destroyed after use. If you need a work record, use a bead. If you do not need a record, do not create one. The distinction between "permanent work record" and "ephemeral work record" is a convenience optimization, not a structural necessity. |
| **Molecule** | Chained multi-step workflows surviving restarts. This is a workflow engine bolted onto a task runner. The core system assigns tasks and merges results. Multi-step workflows can be decomposed into sequential single-step tasks by the coordinator. The coordinator already exists (Mayor). |
| **Convoy** | Batch of beads assigned together. A convoy is a list. A list does not need its own abstraction. The Mayor can assign beads individually or track groups in a plain field on the bead itself. |
| **Seance** | Query predecessor sessions for context. This is a convenience wrapper around "read the handoff file." The handoff mechanism already preserves context. Seance adds a query interface on top, which is nice but not load-bearing. |
| **Nudge** | Real-time signaling. The mail protocol already handles inter-agent communication. Nudge is an optimization for latency -- "notify immediately instead of waiting for the next poll." The system functions correctly (just slightly slower) with polling-only. |
| **Dog** | Infrastructure helper spawned by Deacon. The Deacon can perform its own background tasks. Spawning sub-agents to do cleanup is convenience delegation. |
| **Feed curator** | Real-time activity feed. Useful for dashboards but not for orchestration. Agents do not read the feed to make decisions. |
| **Dolt SQL server** | Heavyweight. The beads system uses JSONL as its underlying format. SQLite or flat files provide the same durability without running a server process. Dolt's git-like branching for data is elegant but the actual branching is per-polecat, which can be achieved with file-level isolation (each polecat writes to its own directory). |
| **4 escalation severity levels** | Two levels suffice: "handle it" and "human needed." The intermediate levels (bead vs. mayor mail vs. email vs. SMS) are delivery channel choices, not architectural necessities. |

**What remains after Round 2:**
- ~10 -> 7 agent concepts (Mayor, Deacon, Boot, Witness, Refinery, Polecat, Crew)
- 5 -> 2 work unit types (Bead, Hook)
- 8 -> ~6 message types
- Infrastructure: daemon, tmux, git, SQLite or JSONL

**Reduction: ~55% of original concepts removed. System still assigns work, executes in parallel, merges results, monitors health, and recovers from failures.**

---

### Round 3: Consolidate Roles

Ask of each remaining role: "Does this role make decisions that no other role can make? Or does it relay decisions that could be made elsewhere?"

| Role | Question | Verdict |
|------|----------|---------|
| **Mayor** | Creates work assignments, handles escalation. Could this be a function of the daemon rather than a separate AI agent? | **Keep but redefine.** The Mayor's core function -- decomposing human intent into agent tasks -- requires AI reasoning. But it does not need to be a persistent, always-running agent. It can be invoked on-demand when work needs to be created or escalation needs handling. |
| **Deacon** | Patrols all rigs, dispatches plugins, coordinates recovery. | **Merge into Witness.** The Deacon is a cross-rig Witness. With one rig, there is no "cross-rig" coordination. Even with multiple rigs, each Witness can report directly to the Mayor. The Deacon is a middle-management layer that relays information. |
| **Boot** | Ensures Deacon is alive. | **Remove.** If the Deacon is merged into Witness, Boot watches nothing. Even before that merge, Boot is a watchdog for a watchdog -- the daemon process itself can check if supervision agents are alive via timestamp checks. No AI reasoning required for "is this process still running?" |
| **Witness** | Per-rig polecat lifecycle management, health monitoring. | **Keep.** This is the essential supervision function: detect stuck agents, trigger recovery, verify completion. |
| **Refinery** | Merge queue management. | **Merge into Witness or make it a function, not an agent.** Merging a git branch is a deterministic operation. It does not require AI reasoning. A Go function called by the Witness (or daemon) can attempt the merge, detect conflicts, and report results. The "Refinery" as a separate AI agent burning tokens to run `git merge` is over-engineering. |
| **Polecat** | Worker agent executing tasks. | **Keep.** This is the irreducible worker. |
| **Crew** | Long-lived, human-directed agent. | **Keep but recognize as separate concern.** Crew is not part of orchestration -- it is a direct human-to-agent interface. It is essentially "Claude Code running in a persistent tmux pane with Gas Town identity." |

**Consolidated roles (3 agent roles + 1 infrastructure process):**

1. **Coordinator** (was Mayor) -- AI agent invoked on-demand to decompose work and handle escalation
2. **Supervisor** (was Witness, absorbing Deacon) -- AI agent running per-rig that monitors workers, triggers recovery, and manages merge operations
3. **Worker** (was Polecat) -- AI agent executing assigned tasks in isolated worktrees
4. **Daemon** (infrastructure, not AI) -- Go process managing heartbeats, spawning agents via tmux, checking liveness via timestamps

Crew remains as an orthogonal concept (human-directed agent, not part of automated orchestration).

**Reduction: 13 original roles -> 3 AI agent roles + 1 infrastructure process. Boot, Dog, Deacon, Refinery eliminated. Crew recognized as a separate interface concern.**

---

### Round 4: Simplify Data Structures

**Beads -> Tasks.** Strip the bead to its minimal fields:

```
Task {
  id:          string     // unique identifier
  title:       string     // what needs to be done
  description: string     // details and acceptance criteria
  status:      enum       // open | assigned | in_progress | done | failed
  assigned_to: string     // worker identity (empty if unassigned)
  branch:      string     // git branch name for this work
  rig:         string     // which project/repo
  created_by:  string     // attribution
  created_at:  timestamp
  updated_at:  timestamp
  result:      string     // outcome summary, error message, or merge commit
  parent_id:   string     // optional, for task decomposition
}
```

**Hook -> Assignment.** The hook is just `task.assigned_to != ""` combined with `task.status == assigned | in_progress`. It does not need a separate data structure.

**Handoff -> File.** A handoff is a markdown file written by a dying session and read by the next session. It does not need a protocol -- it needs a file path convention: `{worker_dir}/.handoff.md`.

**Mail -> Messages.** Simplify to a message queue (directory of JSON files):

```
Message {
  id:        string
  from:      string
  to:        string
  type:      enum    // task_done | merge_result | help | handoff
  payload:   json
  timestamp: timestamp
  ack:       bool
}
```

Four message types instead of eight:
1. **task_done** -- worker reports completion (was POLECAT_DONE + MERGE_READY)
2. **merge_result** -- merge outcome: success, conflict, or test failure (was MERGED + MERGE_FAILED + REWORK_REQUEST)
3. **help** -- escalation request (was HELP)
4. **handoff** -- session continuity (was HANDOFF)

RECOVERED_BEAD and RECOVERY_NEEDED become internal supervisor operations, not messages.

**Storage: Directory of JSON files.** One directory per rig, one file per task, one directory for messages. No database server required. Git tracks history.

---

### Round 5: Reduce Communication to Minimum Viable Set

**The essential communication flow (3 interactions):**

```
Human -> Coordinator: "Build feature X"
  Coordinator -> Tasks: creates task files
  Coordinator -> Workers: assigns via task.assigned_to field

Worker -> Supervisor: "I'm done" (writes task_done message)
  Supervisor -> Git: attempts merge (deterministic operation)
  Supervisor -> Worker: merge_result message (success or rework needed)

Supervisor -> Coordinator: "Worker Y is stuck/dead" (help message)
  Coordinator -> Tasks: reassigns or escalates to human
```

**Removed communication:**
- Nudge (replaced by polling)
- Seance (replaced by reading files)
- Boot-to-Deacon health checks (replaced by daemon timestamp check)
- Deacon-to-Witness relay (Witness reports directly)
- Feed events (not needed for orchestration)

**Health monitoring reduces to one mechanism:** The daemon checks each agent's heartbeat file timestamp. If stale beyond threshold, it restarts the agent (for Supervisor) or notifies the Supervisor (for Workers).

---

### Regressive Abstraction Summary

| Metric | Original Gas Town | After 5 Rounds |
|--------|-------------------|----------------|
| Agent roles | 13 | 3 + daemon |
| Work unit types | 10 | 1 (Task) |
| Message types | 8 | 4 |
| Data storage | Dolt SQL server | JSON files in directories |
| Communication protocols | 5 (mail, nudge, handoff, seance, hooks) | 2 (messages, files) |
| Infrastructure | Go daemon + Dolt + tmux + beads CLI | Go daemon + tmux |
| Estimated Go packages | 61 | ~12-15 |

---

## Part 2: The Minimal Viable Orchestrator

### The Core Value Proposition (Restated)

A system where a human says "do these things" and multiple AI agents work on them in parallel, without stepping on each other, with stuck agents detected and recovered, and completed work merged into the codebase.

### Agent Roles: 3

**1. Coordinator**
- Invoked on-demand (not always-running)
- Receives human intent, decomposes into tasks
- Assigns tasks to available workers
- Handles escalation when supervisor reports failures
- Can reassign, split, or cancel tasks
- One per town

**2. Supervisor**
- Persistent, runs on a patrol loop (check every 60-90 seconds)
- One per rig (project/repo)
- Monitors worker health via heartbeat file timestamps
- Receives "task_done" messages from workers
- Executes merge operations (deterministic git commands, not AI-required)
- Sends merge results back to workers
- Reports unrecoverable failures to coordinator
- Cleans up completed worker worktrees

**3. Worker**
- Ephemeral identity, persistent work assignment
- Receives a task, works in an isolated git worktree on a named branch
- Writes heartbeat file periodically (touch a timestamp file)
- Sends "task_done" when finished
- Reads merge_result and performs rework if needed
- Writes handoff file if context window fills, then exits (daemon restarts it)
- Self-cleans on successful merge

### Data Structures: 3

**Task (JSON file: `{rig}/.gt/tasks/{id}.json`)**
```json
{
  "id": "t-abc123",
  "title": "Add user authentication endpoint",
  "description": "Create POST /auth/login accepting email+password...",
  "status": "assigned",
  "assigned_to": "worker-03",
  "branch": "gt/t-abc123",
  "rig": "hermes",
  "created_by": "coordinator",
  "created_at": "2026-02-19T10:00:00Z",
  "updated_at": "2026-02-19T10:05:00Z",
  "result": "",
  "parent_id": ""
}
```

**Message (JSON file: `{town}/.gt/messages/{id}.json`)**
```json
{
  "id": "m-def456",
  "from": "worker-03",
  "to": "supervisor-hermes",
  "type": "task_done",
  "payload": {"task_id": "t-abc123", "branch": "gt/t-abc123"},
  "timestamp": "2026-02-19T11:30:00Z",
  "ack": false
}
```

**Agent (JSON file: `{town}/.gt/agents/{id}.json`)**
```json
{
  "id": "worker-03",
  "role": "worker",
  "rig": "hermes",
  "status": "active",
  "current_task": "t-abc123",
  "worktree": "/path/to/worktree",
  "heartbeat": "2026-02-19T11:29:45Z",
  "session_count": 3
}
```

### Communication: File-Based Message Passing

No database, no server, no custom protocol. Agents communicate by writing and reading JSON files in shared directories. The filesystem is the message bus.

- **Outbox pattern:** Agent writes a message file to `{town}/.gt/messages/`.
- **Polling:** Supervisor polls the messages directory every 60 seconds for messages addressed to it.
- **Acknowledgment:** Reader sets `ack: true` and moves file to `{town}/.gt/messages/archive/`.
- **Heartbeat:** Each agent touches `{town}/.gt/agents/{id}.json` (updates the `heartbeat` field) every 30 seconds.
- **Handoff:** Worker writes `{worktree}/.handoff.md` before exiting. Next session reads it on startup.

### Infrastructure: 4 Components

1. **Go daemon** -- Single background process. Runs a heartbeat loop (every 60-90 seconds). Checks agent heartbeat timestamps. Restarts dead agents via tmux. Spawns new workers when tasks are unassigned. No AI reasoning; pure mechanical.

2. **tmux** -- Session manager. Each agent runs in a tmux session. Daemon creates/destroys sessions. Provides terminal access for human inspection.

3. **git** -- Worktree isolation. Each worker gets `git worktree add` for its branch. Merge operations are `git merge` commands. Standard git, no extensions.

4. **Claude Code CLI** (or equivalent) -- The AI runtime. Each agent is a Claude Code session with a system prompt defining its role.

### Go Package Estimate: 12

| Package | Purpose |
|---------|---------|
| `cmd/gt` | CLI entry point |
| `daemon` | Background heartbeat loop, agent lifecycle |
| `coordinator` | Work decomposition, task creation, assignment |
| `supervisor` | Health monitoring, merge operations, recovery |
| `worker` | Task execution setup, worktree management |
| `task` | Task data structure, CRUD operations |
| `message` | Message data structure, read/write/poll |
| `agent` | Agent identity, heartbeat, status |
| `git` | Worktree operations, merge, branch management |
| `tmux` | Session create/destroy/attach |
| `claude` | Claude Code CLI invocation and session management |
| `config` | Town/rig configuration, paths |

Twelve packages. Roughly 80% reduction from 61.

---

## Part 3: Essential vs. Accidental Complexity Matrix

| # | Feature | Essential | Accidental | Speculative | Verdict |
|---|---------|-----------|------------|-------------|---------|
| 1 | **Persistent agent identity** | Knowing which agent did what work. Required for attribution, debugging, and recovery. | -- | -- | **Essential.** An agent ID on every task and commit. |
| 2 | **Git worktree isolation** | Agents must not conflict. Worktrees provide zero-conflict parallel work. | -- | -- | **Essential.** The mechanism that makes parallelism safe. |
| 3 | **Task assignment (Hook)** | A durable record that "agent X is working on task Y." Must survive crashes. | The Hook as a separate data structure with its own lifecycle is accidental. A field on the task record suffices. | -- | **Simplify.** `task.assigned_to` replaces Hook. |
| 4 | **Merge queue** | Completed work must be integrated. Sequential merging prevents conflicts. | The Refinery as a separate AI agent is accidental. Merging is deterministic; a function, not an agent. | -- | **Simplify.** Merge is a function called by the Supervisor, not a role. |
| 5 | **Health monitoring** | Stuck or dead agents must be detected, or work is silently lost. | Three-tier watchdog chain (Daemon -> Boot -> Deacon -> Witness) is accidental. One daemon checking timestamps suffices. | -- | **Simplify.** Daemon checks heartbeats. One tier. |
| 6 | **Session handoff** | AI context windows fill. Work must survive across sessions. | The Handoff protocol with typed messages is accidental. A markdown file read on startup suffices. | -- | **Simplify.** File-based handoff. |
| 7 | **Work decomposition** | Complex tasks must be broken into parallelizable units. | -- | -- | **Essential.** The Coordinator's core function. |
| 8 | **Structured work records (Beads)** | Work must be trackable and queryable. | The full bead schema with event beads, role beads, label cache, JSONL format, and Dolt branches is accidental complexity. A JSON file with status fields suffices. | -- | **Simplify.** JSON task files. |
| 9 | **Async inter-agent messaging** | Agents must communicate without blocking each other. | Five communication channels (mail, nudge, handoff, seance, hooks) is accidental. One message directory suffices. | -- | **Simplify.** File-based messages. |
| 10 | **Worker cleanup (done means gone)** | Completed workers should release resources. | -- | -- | **Essential.** Worktree removal after merge. |
| 11 | **Escalation to human** | Some failures require human intervention. | Four severity levels with email and SMS routing is accidental. A single "help needed" notification suffices for current scale. | -- | **Simplify.** One escalation channel. |
| 12 | **Daemon process** | Something must run continuously to restart crashed agents. | -- | -- | **Essential.** Mechanical liveness guarantee. |
| 13 | **Mayor (Coordinator)** | AI reasoning needed to decompose human intent into tasks. | Always-running Mayor agent burning tokens while idle is accidental. On-demand invocation suffices. | -- | **Simplify.** Invoke on-demand, not persistent. |
| 14 | **Deacon** | -- | Middle-management relay layer between daemon and per-rig supervisors is accidental. Witnesses can report directly. | -- | **Remove.** |
| 15 | **Boot** | -- | A watchdog watching the watchdog is accidental when the daemon can check timestamps directly. | -- | **Remove.** |
| 16 | **Dog** | -- | Sub-agents for infrastructure tasks is accidental delegation. The daemon or supervisor handles these directly. | -- | **Remove.** |
| 17 | **Convoy** | -- | A "batch of tasks" abstraction is accidental. A list field or parent-child relationship on tasks suffices. | -- | **Remove.** Replace with `parent_id`. |
| 18 | **Molecule** | -- | Multi-step durable workflows are accidental at current scale. Sequential task creation by the coordinator achieves the same result. | -- | **Remove.** |
| 19 | **Wisp** | -- | Ephemeral work records are accidental. Either you need a record (use a task) or you do not. | -- | **Remove.** |
| 20 | **Formula** | -- | -- | Workflow templates require a template ecosystem. | **Remove.** Speculative. |
| 21 | **Nudge** | -- | Real-time signaling is a latency optimization over polling. Not structurally needed. | -- | **Remove.** Polling suffices. |
| 22 | **Seance** | -- | Session history querying is a convenience wrapper over reading files. | -- | **Remove.** |
| 23 | **Dolt SQL server** | -- | Full SQL server for task storage is accidental. JSON files or SQLite suffice. | -- | **Remove.** Use filesystem. |
| 24 | **Federation / HOP** | -- | -- | Cross-workspace coordination for a single-workspace system. | **Remove.** Speculative. |
| 25 | **CV chains / capability routing** | -- | -- | Agent performance history and smart routing. Requires data that does not exist yet. | **Remove.** Speculative. |
| 26 | **Events as truth / label cache pattern** | Elegant but accidental. A status field on a task record is simpler and sufficient. Event sourcing is a powerful pattern but adds complexity disproportionate to current needs. | -- | -- | **Simplify.** Direct state mutation on task records. |
| 27 | **Three-layer identity (Identity/Sandbox/Session)** | -- | Two layers suffice: identity (permanent, e.g. "worker-03") and session (ephemeral, the current tmux/Claude session). The "sandbox" layer maps to worktree, which is already tracked on the task. | -- | **Simplify.** Two layers. |
| 28 | **GUPP (Propulsion Principle)** | The principle "if you have work, execute immediately" is essential. Agents should not wait for confirmation. | The acronym and formalization are accidental. This is just "do your job." | -- | **Keep principle, drop acronym.** |
| 29 | **Attribution via BD_ACTOR** | -- | A dedicated attribution system is accidental when git commits already carry author information. Agent ID in commit author field suffices. | -- | **Simplify.** Use git author. |
| 30 | **Plugin system** | -- | -- | Extensibility framework for future integrations. Current plugin usage is minimal. | **Remove.** Build when needed. |
| 31 | **Real-time activity feed** | -- | Dashboard/monitoring convenience. Agents do not use it for decisions. | -- | **Remove.** |
| 32 | **Crew (human-directed agent)** | -- | Not part of orchestration. It is "Claude Code in tmux with a name." Orthogonal to multi-agent coordination. | -- | **Separate concern.** Not part of the orchestrator. |
| 33 | **Tmux session management** | A process manager is essential. Tmux is a reasonable choice but not the only one. | -- | -- | **Essential.** (Tmux or equivalent.) |
| 34 | **Routes file (ID prefix -> rig)** | -- | Routing table for multi-rig ID resolution. With few rigs, direct path lookup suffices. | -- | **Simplify.** Config file with rig paths. |
| 35 | **Redirects (worktree -> canonical beads DB)** | -- | Artifact of the Dolt-per-worktree architecture. Eliminated when Dolt is eliminated. | -- | **Remove.** |

**Summary: 7 essential features, 20 accidental (simplifiable or removable), 8 speculative (not yet needed).**

---

## Part 4: Three Alternative Architectures

### Architecture A: "Bare Metal"

*The absolute minimum. Under 10 Go packages. Multi-agent orchestration with nothing extra.*

**Philosophy:** The filesystem is the database. Git is the coordination layer. The CLI is the interface. Nothing runs unless work exists.

**Agent Roles: 2**

1. **Orchestrator** (on-demand, not persistent) -- Human runs `gt work "build feature X"`. The CLI invokes Claude Code once to decompose the request into task files. It assigns tasks to workers by writing task files. When all tasks are done, it is invoked again to verify integration. Between invocations, it does not run. Zero idle token burn.

2. **Worker** -- Spawned per-task. Works in a git worktree. Writes a heartbeat file. Sends "done" by updating its task file status. Exits when done. No persistent identity beyond the task assignment.

**Data Model:**
```
~/gt/
  .gt/
    tasks/          # one JSON file per task
    config.json     # rig paths, settings
  {rig}/
    .gt/
      worktrees/    # managed git worktrees
```

A task file is the only data structure. Status transitions (`open -> assigned -> in_progress -> done | failed`) are direct field mutations. No event sourcing, no message queue.

**Communication:** None between agents. Workers update task files directly. The daemon reads task files to determine system state. Workers do not talk to each other or to a supervisor.

**Infrastructure:**
- Go CLI binary
- Simple cron-like loop (or daemon) that checks task file timestamps every 60 seconds
- tmux for sessions
- git for worktrees and merging

**What is sacrificed:**
- No real-time supervisor -- stuck agents are detected only on the next cron tick
- No AI-powered recovery -- daemon restarts workers mechanically; if the task is fundamentally broken, human must intervene
- No session handoff -- workers that exhaust context windows simply fail, and the daemon restarts them from scratch (the task description is the only context)
- No merge conflict resolution -- conflicts block the task as "failed" and require human intervention
- No inter-agent communication

**Go Packages: 7**

`cmd/gt`, `task`, `worker`, `daemon`, `git`, `tmux`, `config`

**When this is enough:** Solo developer with 2-5 agents on one project. Tasks are independent and well-defined. Human is available to handle failures.

---

### Architecture B: "Sweet Spot"

*The 80/20 version. 20% of Gas Town's complexity delivering 80% of its value.*

**Philosophy:** Three AI roles is the minimum for self-sustaining orchestration. Add session handoff and merge management. Keep everything file-based. No database server.

**Agent Roles: 3**

1. **Coordinator** (on-demand) -- Decomposes work, assigns tasks, handles escalation. Invoked by `gt work` command or by Supervisor escalation. Not always-running.

2. **Supervisor** (persistent, one per rig) -- Monitors worker heartbeats (checks every 60s). Attempts merges when workers report completion. Detects stuck workers (stale heartbeat > 5 minutes). Reassigns failed tasks or escalates to Coordinator. This is an AI agent because merge conflict resolution and recovery decisions benefit from reasoning.

3. **Worker** (ephemeral per-task) -- Executes tasks in isolated worktrees. Writes heartbeat. Writes handoff file before context exhaustion. Sends task_done message. Handles rework on merge conflicts.

**Data Model:**
```
~/gt/
  .gt/
    tasks/{id}.json       # task records
    agents/{id}.json      # agent identity + heartbeat
    messages/{id}.json    # inter-agent messages
    messages/archive/     # acknowledged messages
    config.json
  {rig}/
    .gt/
      worktrees/
```

Three data structures: Task, Agent, Message. All JSON files. Git-tracked for history.

**Communication: File-based message passing**
- Workers write `task_done` messages
- Supervisor writes `merge_result` messages
- Supervisor writes `help` messages to Coordinator
- Workers write `handoff` files for session continuity
- Polling interval: 30-60 seconds

**Infrastructure:**
- Go CLI + daemon
- tmux
- git

**What is sacrificed compared to full Gas Town:**
- No cross-rig coordination (each rig is independent)
- No capability-based routing (round-robin or manual assignment)
- No performance history or agent CVs
- No real-time feed
- No plugin system
- No workflow templates
- No federation
- Simpler escalation (supervisor -> coordinator -> human; no graduated severity)

**What is gained compared to Bare Metal:**
- AI-powered merge conflict resolution
- Intelligent stuck-agent recovery
- Session handoff (work survives context exhaustion)
- Supervisor provides continuous monitoring without human attention

**Go Packages: 12**

`cmd/gt`, `coordinator`, `supervisor`, `worker`, `task`, `message`, `agent`, `daemon`, `git`, `tmux`, `claude`, `config`

**When this is enough:** Small team with 4-15 agents across 1-3 projects. Agents work for hours unattended. System self-heals from most common failures (agent crashes, context exhaustion, simple merge conflicts).

---

### Architecture C: "Different Paradigm" -- The Claim Board

*Throw out hierarchical supervision entirely. No coordinator, no supervisor, no assignment. Agents self-organize.*

**Philosophy:** Inspired by market mechanisms and ant colony stigmergy. Instead of top-down assignment, work is posted on a shared "claim board." Agents autonomously claim tasks based on their availability. Coordination emerges from simple rules, not from supervision.

**How it works:**

1. **Human posts tasks** to the claim board (a directory of task files with `status: open`).

2. **Agents self-select.** Each agent runs an identical loop:
   - Scan claim board for `status: open` tasks
   - Claim a task by atomically setting `status: claimed` + `claimed_by: {self}` (using filesystem atomic rename or lock file)
   - Create worktree, execute task
   - On completion: set `status: done`, push branch
   - Attempt merge to main. If conflict: set `status: conflict` and move on to next task
   - Loop back to scan for more work

3. **Conflict resolution is deferred.** Merge conflicts accumulate. A periodic "janitor" pass (which can be a human or an AI agent) resolves conflicts in batch. This accepts that merge conflicts are rare (with good task decomposition) and expensive to resolve (often requiring understanding of multiple changes).

4. **Health is emergent.** No heartbeat monitoring. Instead: if a task has been `claimed` for longer than a timeout (e.g., 30 minutes with no progress commit), it reverts to `open` and any agent can claim it. The abandoned agent, if it eventually finishes, will find its task already completed by another agent and simply discards its work.

**Agent Roles: 1**

**Worker** -- Every agent is identical. No hierarchy. No coordinator, no supervisor. The human is the only "manager" -- they write task descriptions and review merged results.

**Data Model:**
```
~/gt/
  board/
    {id}.json    # task: open | claimed | done | conflict
  done/
    {id}.json    # completed tasks (moved here)
```

Single data structure: Task. Two states that matter: unclaimed and claimed.

**Communication: None.** Agents do not talk to each other. They communicate indirectly through the claim board (stigmergy -- communication through the environment, like ants leaving pheromone trails).

**Coordination mechanism:** Filesystem atomicity. `mv` (rename) is atomic on all POSIX systems. To claim a task, an agent renames the file from `board/{id}.json` to `board/claimed-{agent}-{id}.json`. First rename wins. No locks, no distributed consensus.

**Infrastructure:**
- Go CLI (no daemon needed -- agents self-manage)
- tmux (or any process manager)
- git

**What is sacrificed:**
- No intelligent work decomposition (human must create well-defined tasks)
- No AI-powered recovery (timeout-based reclaim only)
- No merge conflict resolution (deferred to human or batch process)
- No session handoff (agents restart from scratch on failure)
- No escalation (human monitors output)
- No optimization of agent-to-task matching

**What is gained:**
- Extreme simplicity. Under 200 lines of orchestration code.
- No supervision overhead. Zero tokens spent on monitoring.
- Linear scaling. Adding agents adds capacity with zero coordination cost.
- No single point of failure. No coordinator or supervisor to crash.
- Naturally idempotent. Duplicate work is harmless (same result, discarded if someone finished first).
- Easy to reason about. One agent, one loop, one data structure.

**Go Packages: 5**

`cmd/gt`, `board`, `worker`, `git`, `config`

**When this is enough:** Well-defined, independent tasks. Experienced human decomposing work. Low merge conflict probability. Acceptable to waste some duplicate agent work for the benefit of zero coordination overhead. Think "open source project with many independent contributors" rather than "tightly coordinated sprint team."

---

### Architecture Comparison

| Dimension | A: Bare Metal | B: Sweet Spot | C: Claim Board | Full Gas Town |
|-----------|---------------|---------------|-----------------|---------------|
| Agent roles | 2 | 3 | 1 | 13 |
| Data structures | 1 | 3 | 1 | 10+ |
| Go packages | 7 | 12 | 5 | 61 |
| Token burn (idle) | Zero | Low (supervisor only) | Zero | High (Boot + Deacon + Witness + Mayor) |
| Self-healing | Restart only | Restart + reassign + AI recovery | Timeout-based reclaim | Full recovery pipeline |
| Merge handling | Fail on conflict | AI-assisted resolution | Defer to human | Dedicated Refinery agent |
| Max agents | ~5 | ~15 | ~50+ | ~30+ |
| Human attention needed | High | Low | Medium | Very low |
| Coordination model | Centralized assignment | Hierarchical supervision | Emergent / stigmergic | Multi-tier hierarchy |
| Infrastructure | CLI + cron + tmux + git | CLI + daemon + tmux + git | CLI + tmux + git | CLI + daemon + Dolt + tmux + git + beads |

---

## Part 5: Migration Path

How to extract Gas Town's essentials into a simpler system, starting from the current codebase.

### Strategy: Strangle Fig Pattern

Do not rewrite Gas Town. Build the new system alongside it, migrate functionality incrementally, and let the old system atrophy. This is less risky than a rewrite and allows rollback at any point.

### Phase 0: Preparation (1-2 days)

**Objective:** Understand what is actually running and what is dead code.

1. Audit which Gas Town features are actively used vs. implemented-but-unused vs. stubbed-out.
2. Identify the hot path: Human creates work -> agents execute -> work merges. Every package not on this path is a candidate for elimination.
3. Document the current file/directory layout and identify what can be reused.

**Carry over:** The `gt` CLI structure, tmux integration, git worktree logic. These are well-tested infrastructure.

**Rebuild:** Task management, messaging, agent lifecycle.

### Phase 1: File-Based Task System (2-3 days)

**Objective:** Replace beads + Dolt with JSON task files.

1. Create `task` package: Task struct, CRUD operations on JSON files.
2. Create `{rig}/.gt/tasks/` directory structure.
3. Add `gt task create`, `gt task list`, `gt task show` CLI commands.
4. Verify: Tasks can be created, listed, updated, and deleted as JSON files.

**What this replaces:** Beads (partially), Dolt server, beads CLI dependency.

**What still works alongside:** Existing agent spawning, tmux management.

### Phase 2: Worker Simplification (2-3 days)

**Objective:** Workers read task files instead of hooks/beads.

1. Create `worker` package: reads assigned task, creates worktree, writes heartbeat, updates task status.
2. Modify worker system prompt to read task JSON instead of hook/bead system.
3. Implement heartbeat: worker updates `agents/{id}.json` timestamp every 30 seconds (Claude Code hook on session activity, or a background touch).
4. Implement handoff: worker writes `.handoff.md` before exit; new session reads it.
5. Verify: Worker can receive a task file, work in a worktree, and update status to done.

**What this replaces:** Polecat lifecycle, Hook system, three-layer identity.

### Phase 3: Supervisor (3-4 days)

**Objective:** Per-rig supervisor replaces Witness + Refinery + Deacon.

1. Create `supervisor` package: patrol loop, heartbeat checking, merge operations.
2. Supervisor reads `messages/` directory for `task_done` messages.
3. Supervisor attempts merge (git merge to main branch). Reports result via `merge_result` message.
4. Supervisor checks agent heartbeats. If stale > 5 minutes, marks task as `failed` and sets `assigned_to: ""` for reassignment.
5. Supervisor writes `help` message to coordinator on repeated failures.
6. Verify: Supervisor detects stuck worker, reclaims task, successfully merges completed work.

**What this replaces:** Witness, Refinery (as agent), Deacon, Boot.

### Phase 4: Coordinator (2-3 days)

**Objective:** On-demand coordinator replaces always-running Mayor.

1. Create `coordinator` package: invoked by `gt work "description"` command.
2. Coordinator uses Claude Code to decompose work description into task files.
3. Coordinator assigns tasks to available workers (reads `agents/` directory for idle workers, or signals daemon to spawn new ones).
4. Coordinator handles escalation: reads `help` messages, attempts resolution, or notifies human.
5. Verify: Human runs `gt work "build auth system"`, coordinator creates 3 tasks, workers execute them, supervisor merges them.

**What this replaces:** Mayor (as persistent agent).

### Phase 5: Daemon Simplification (1-2 days)

**Objective:** Daemon does only mechanical work: spawn, restart, check liveness.

1. Simplify daemon to a single heartbeat loop (every 60 seconds).
2. Daemon checks: Are there unassigned tasks? If yes, spawn workers.
3. Daemon checks: Is the supervisor alive? If no, restart it.
4. Daemon checks: Are there workers with stale heartbeats? If yes, notify supervisor.
5. Remove: Boot agent, Deacon spawning, convoy watcher, feed curator, Dolt server management.
6. Verify: Daemon keeps the system alive without any AI agents dedicated to supervision of supervision.

**What this replaces:** Current daemon (simplified), Boot, Dog.

### Phase 6: Cleanup (1-2 days)

**Objective:** Remove dead code and unused packages.

1. Delete packages not referenced by the new system.
2. Remove Dolt dependency and beads CLI dependency.
3. Remove unused message types, data structures, and protocols.
4. Update CLI commands to reflect new architecture.
5. Update system prompts for all agent roles.

### Migration Sequence Diagram

```
Phase 0 ──> Phase 1 ──> Phase 2 ──> Phase 3 ──> Phase 4 ──> Phase 5 ──> Phase 6
 Audit       Tasks      Workers     Supervisor  Coordinator  Daemon      Cleanup
             (JSON)     (simple)    (merged)    (on-demand)  (simple)    (delete)

 [Old beads still work]  [Workers use new tasks]  [Old roles deprecated]  [Old code removed]
```

### Total Estimated Timeline: 12-18 days

### What Gets Carried Over vs. Rebuilt

| Carry Over (reuse) | Rebuild (new) | Delete |
|---------------------|---------------|--------|
| tmux session management | Task data structure (JSON) | Beads system |
| git worktree operations | File-based messaging | Dolt integration |
| Claude Code invocation | Supervisor patrol loop | Boot agent |
| CLI framework (cobra) | Coordinator decomposition | Deacon agent |
| Config/paths logic | Daemon heartbeat loop | Dog agent |
| Agent identity concept | Heartbeat mechanism | Convoy, Molecule, Wisp, Formula |
| System prompt patterns | Handoff file convention | Nudge, Seance |
| | | Feed curator |
| | | Plugin gate system |
| | | Refinery (as agent) |
| | | Routes, Redirects |
| | | CV chains, Digests, Epics |

### Risk Mitigation

1. **Do not delete old code until new code is verified.** Run both systems in parallel during phases 2-4.
2. **Test with one worker first.** Do not spawn 10 workers on the new system until one worker completes the full cycle (task -> work -> done -> merge).
3. **Keep the old daemon running.** It can coexist with the new daemon during migration. Use different port/directory if needed.
4. **Git is the safety net.** Everything is in git. If migration goes wrong, `git revert` restores the previous state.

---

## Conclusion: What Multi-Agent AI Orchestration Irreducibly Needs

After stripping Gas Town to bedrock, the irreducible requirements for multi-agent AI code orchestration are:

1. **A way to define work** -- a task with a description, status, and assignment.
2. **A way to isolate work** -- git worktrees (or equivalent) so agents do not conflict.
3. **A way to assign work** -- matching tasks to agents, either by human, coordinator, or self-selection.
4. **A way to integrate work** -- merging completed branches into the main codebase.
5. **A way to detect failure** -- knowing when an agent is stuck or dead.
6. **A way to recover from failure** -- reassigning work from dead agents to live ones.
7. **A way to survive session limits** -- preserving enough context for the next session to continue.

Seven requirements. Everything else -- the 13 roles, 10 data structures, 8 message types, 61 packages, Dolt server, plugin system, federation, CVs, molecules, convoys, nudges, seances, formulas, and the rest -- is either a specific implementation of one of these seven requirements, a convenience optimization, or a speculative feature for a future that has not arrived.

The child's question from the Perspective Prompting framework -- "Why do you need so many helpers to do one thing?" -- turns out to be the right question. The answer is: you do not. You need three roles (coordinate, supervise, work), seven data fields on a task record, and a filesystem.

Gas Town's essential insight is correct: AI agents working in parallel on code, with structured coordination, is extremely valuable. The extraction reveals that this insight can be delivered with roughly 80% less machinery than the current implementation carries.
