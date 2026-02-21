# Ford Assembly Line & Extreme Specialization Audit of Gas Town

Applied from the 10,000x Knowledge Worker frameworks (Article 2) and Lean Manufacturing principles to Gas Town's multi-agent orchestration architecture.

---

## Part 1: The Critical Problem Redefinition

### Ford's Key Insight

Ford did not ask "How can workers build cars faster?" He asked "Why are workers moving at all?" The answer was: because we organized work around workers instead of organizing workers around work. The car should move; the workers should stand still.

### What Gas Town Is Currently Optimizing

Gas Town is implicitly asking: **"How can we supervise AI agents more reliably?"**

The entire supervision stack -- Daemon, Boot, Deacon, Witness -- is an answer to this question. Every heartbeat cycle, every health check, every nudge, every handoff exists because the system assumes agents are unreliable and must be continuously watched. The architecture optimizes for *detecting and recovering from agent failure*.

Evidence from the daemon log: 97 heartbeat cycles over ~5 hours. In each cycle, the daemon checks Boot status, checks Deacon health, checks Witness for each rig, and checks Refinery for each rig. That is ~6 checks per cycle, ~582 checks in 5 hours. The vast majority returned "already running, skipping spawn." The system spent most of its energy confirming that nothing was wrong.

### The Ford-Level Reframe

**"Why are agents being supervised at all?"**

Or more precisely: **Why is supervision a continuous, expensive AI operation rather than a mechanical, zero-cost infrastructure guarantee?**

Ford did not put a foreman at every station to check if the worker was alive. He built the conveyor belt so that *the physical infrastructure made failure visible and recovery automatic*. A stopped belt is immediately obvious. You do not need a watcher to watch the watchers.

### Subsystem-by-Subsystem "Why Is This Moving?"

#### Supervision (Daemon -> Boot -> Deacon -> Witness)

**Current state**: Three tiers of AI agents watching each other. The Daemon (Go process) spawns Boot (AI agent) to triage the Deacon (AI agent). The Deacon health-checks Witnesses (AI agents). Witnesses monitor Polecats (AI agents).

**"Why is this moving at all?"**: Because AI sessions die unpredictably (context exhaustion, API errors, hangs). The system treats this as a supervision problem requiring intelligence.

**What if it didn't exist?**: If Claude Code sessions were reliable -- or if a simple process-level health check (PID alive? tmux pane responsive? last output timestamp recent?) replaced AI-based health monitoring -- the entire Boot, Deacon-as-supervisor, and Witness-as-monitor layer could be mechanical. The daemon already does the mechanical check (heartbeat timestamps). The AI agents doing health checks are *redundant with the daemon's own liveness detection*.

**Ford reframe**: Don't build smarter supervisors. Build infrastructure where dead sessions are automatically detected and restarted by the daemon, not by an AI agent that itself needs supervision.

#### Communication (Mail, Nudge, Handoff)

**Current state**: Eight message types. POLECAT_DONE -> Witness -> MERGE_READY -> Refinery -> MERGED/MERGE_FAILED/REWORK_REQUEST -> Witness -> Polecat. Plus HEALTH_CHECK, HELP, HANDOFF, session-started notifications, WITNESS_PING, DEACON_ALIVE.

**"Why is this moving at all?"**: Because agents need to coordinate state transitions. A polecat finishes; someone needs to know and trigger the merge.

**What if it didn't exist?**: The town.log shows the communication volume is dominated by health-check chatter, not work-coordination signals. In the 145-line log sample, approximately:
- ~60 lines are health checks / DEACON_ALIVE / WITNESS_PING (supervision chatter)
- ~20 lines are session-started nudges (lifecycle noise)
- ~15 lines are handoffs (session maintenance)
- ~10 lines are actual work events (done, kill, merge requests)

Only about **7% of logged communication is about actual work**. The other 93% is the system talking to itself about whether it is alive.

**Ford reframe**: Don't optimize the message protocol. Eliminate the messages that exist only because the supervision architecture requires them. If supervision becomes mechanical, ~90% of inter-agent communication vanishes.

#### Work Tracking (Beads, Convoys, Hooks, Molecules)

**Current state**: 10 work unit types. Beads are the core (JSONL-backed issues). Convoys group beads. Hooks pin beads to agents. Molecules chain multi-step workflows. Wisps are ephemeral beads. Formulas are templates.

**"Why is this moving at all?"**: Because tracking work across unreliable agents requires durable state. If a polecat dies mid-task, the bead and hook survive.

**What if it didn't exist?**: The hook + bead core is genuinely load-bearing. If you remove durable work assignment, you lose crash recovery. But the 10-type taxonomy exists because the system models every possible state transition rather than relying on a simpler state machine. Molecules, Wisps, Formulas, Protomolecules, Digests, Epics, and CV chains are specializations that could be collapsed into a single "bead with tags" model.

**Ford reframe**: Ford did not have 10 types of car-in-progress. He had one car, moving through stations, with its state visible from its physical position on the line. Work tracking should be one type with a position in a pipeline, not a taxonomy of types.

#### Merge Queue (Refinery)

**Current state**: One Refinery per rig. Runs as a persistent AI agent. Receives MERGE_READY messages, attempts merges, reports outcomes.

**"Why is this moving at all?"**: Because merging branches requires conflict detection and possibly resolution.

**What if it didn't exist?**: `git merge --no-edit` is a deterministic operation. It either succeeds or fails. The Refinery AI agent is running a command that does not require intelligence in the success case. Intelligence is only needed for conflict resolution -- which is rare. The common case (clean merge) is being handled by an expensive AI agent that sits idle most of the time.

**Ford reframe**: Make the common case mechanical (daemon runs `git merge`), and only spawn an AI agent when the mechanical merge fails with conflicts.

#### Identity (Three-layer: Identity -> Sandbox -> Session)

**Current state**: Each polecat has a permanent identity, a per-assignment sandbox, and ephemeral sessions. Attribution traces through BD_ACTOR across commits, events, and logs.

**"Why is this moving at all?"**: Because enterprise compliance requires knowing who did what.

**What if it didn't exist?**: Attribution is genuinely load-bearing for auditability and performance tracking. The three-layer model, however, adds complexity for what is fundamentally a key-value problem: (agent_id, assignment_id, session_id). Three layers of identity management could be a single identity record with three fields.

**Ford reframe**: Identity is a data problem, not an architecture problem. It should be a row in a table, not a three-layer abstraction.

---

## Part 2: Extreme Specialization Audit

Ford decomposed one "car builder" role into 29 specialized jobs. Gas Town has 13 agent roles, but the problem is the opposite of Ford's: Gas Town's agents are not too generalized -- they are too *numerous* for overlapping responsibilities. The issue is not "one agent does too much" but "too many agents do the same thing (supervision) at different levels."

### Role-by-Role Decomposition

#### 1. Mayor (Town-Level Coordinator)

**Current responsibilities** (5):
1. Create convoys (batch work orders)
2. Distribute work across rigs
3. Handle escalations from any agent
4. Coordinate cross-rig dependencies
5. Serve as human-facing interface for town status

**Could be separated**:
- Work creation/distribution (1-2) is a dispatch function -- could be purely mechanical (round-robin, capability matching from a database)
- Escalation handling (3) is a genuine AI judgment task
- Cross-rig coordination (4) is rare in a single-rig installation; could be event-driven rather than a persistent role

**Could be eliminated**:
- Human-facing status (5) could be a CLI query against the beads database, not a persistent agent

**Maximum specialization**: Mayor becomes "Escalation Judge" -- only invoked when an agent sends HELP. All other responsibilities become daemon functions or CLI queries.

#### 2. Deacon (Background Patrol / Plugin Executor)

**Current responsibilities** (6):
1. Patrol all rigs for health status
2. Dispatch health checks to Witnesses and Refineries
3. Execute plugins on gate-triggered schedules
4. Coordinate agent recovery (rate-limited re-dispatch)
5. Respond to WITNESS_PING messages
6. Perform handoffs when context fills

**Could be separated**:
- Health patrol (1-2) is mechanical -- checking timestamps, sending nudges
- Plugin execution (3) is a scheduler function
- Recovery coordination (4) is a state machine
- Handoff management (6) is session lifecycle

**Could be eliminated**:
- WITNESS_PING responses (5) are pure supervision chatter
- Health patrol (1-2) is redundant with the daemon's own heartbeat checks

**Maximum specialization**: Deacon is eliminated. Plugin execution becomes a cron-like daemon feature. Health checks are daemon heartbeats. Recovery is a daemon state machine with escalation-to-Mayor as the only AI-requiring fallback.

**Critical observation**: The daemon log shows the Deacon going stale and being restarted every ~6 minutes. An agent whose primary job is health monitoring *cannot keep itself healthy*. This is the clearest signal that this role should not be an AI agent.

#### 3. Boot (Watchdog's Watchdog)

**Current responsibilities** (2):
1. Triage Deacon health
2. Decide whether Deacon needs restart or nudge

**Could be separated**: N/A -- this is already maximally specialized.

**Could be eliminated entirely**: Yes. If the Deacon is eliminated (replaced by daemon functions), Boot has no reason to exist. Even with the Deacon, the daemon already detects Deacon staleness and restarts it. Boot is a redundant check layer.

**Maximum specialization**: Boot is eliminated. The daemon checks the Deacon directly (which it already does).

#### 4. Dog (Infrastructure Helper)

**Current responsibilities** (3):
1. Cleanup tasks for the Deacon
2. Background health checks
3. Plugin execution assistance

**Could be eliminated entirely**: Yes. These are sub-tasks of the Deacon, which itself should be a daemon function.

**Maximum specialization**: Dogs are eliminated. Their tasks become daemon goroutines.

#### 5. Witness (Per-Rig Supervisor)

**Current responsibilities** (7):
1. Monitor polecat lifecycle (spawn, work, completion)
2. Detect zombie/abandoned polecats
3. Trigger polecat recovery
4. Forward MERGE_READY to Refinery
5. Relay MERGED/MERGE_FAILED results to polecats
6. Send WITNESS_PING to Deacon
7. Perform session handoffs every ~8 minutes (observed pattern)

**Could be separated**:
- Polecat monitoring (1-2) is a process-liveness check -- mechanical
- Recovery triggering (3) is a state machine
- Merge relay (4-5) is pure message routing
- Deacon pinging (6) is supervision chatter

**Could be eliminated**:
- Message relay (4-5): Polecats could write to a merge queue directly (a file or database row), and the Refinery could poll it. No relay agent needed.
- Deacon pinging (6): eliminated if Deacon is eliminated
- Handoffs (7): the Witness hands off approximately every 8 minutes because its context window fills with health-check traffic. If health-check chatter is eliminated, the Witness's context lasts much longer -- or the role can be eliminated entirely.

**Maximum specialization**: Witness becomes a daemon goroutine that checks polecat process liveness and writes state to a file. No AI agent needed for the common case. AI is invoked only for novel recovery situations.

**Critical observation from logs**: The Witness performs handoffs every ~8 minutes (18:14, 18:24, 18:39, 18:52, 19:06, 19:19, 19:27, 19:35, 19:45, 19:54, 20:11, 20:19, 20:27, 20:35, 20:42, 20:51, 21:00...). Each handoff is a session restart, which means priming context, reading state, resuming patrol. The Witness is spending more time *restarting* than *witnessing*. This is a context-burn spiral: health-check messages fill context -> handoff -> restart -> receive more health-check messages -> context fills again.

#### 6. Refinery (Merge Queue Processor)

**Current responsibilities** (4):
1. Receive MERGE_READY messages
2. Attempt git merge
3. Run post-merge validation (tests, builds)
4. Report outcomes (MERGED, MERGE_FAILED, REWORK_REQUEST)

**Could be separated**:
- Git merge (2) is a deterministic CLI operation
- Validation (3) is a script execution
- Conflict resolution is the only part requiring AI judgment

**Could be eliminated partially**: The Refinery as a persistent AI agent could be replaced by:
- A daemon function that runs `git merge` and `make test`
- An AI agent spawned only when conflicts occur

**Maximum specialization**: "Merge Executor" (mechanical daemon function) + "Conflict Resolver" (ephemeral AI agent, spawned on demand).

#### 7. Polecat (Ephemeral Worker)

**Current responsibilities** (6):
1. Receive work assignment via hook
2. Execute the assigned task (code changes)
3. Run quality gates (tests, lints, builds)
4. Commit and push changes
5. Signal POLECAT_DONE to Witness
6. Self-clean (destroy worktree)

**Could be separated**:
- Task execution (2) is the core value-producing activity
- Quality gates (3) could be a separate validation agent or mechanical step
- Git operations (4) could be a post-task mechanical step
- Signaling and cleanup (5-6) are infrastructure

**Could be eliminated**:
- Signaling to Witness (5): if the daemon watches for branch pushes (git hook or filesystem watch), completion is detectable mechanically
- Self-cleaning (6): daemon can clean up worktrees when branches are merged

**Maximum specialization**: Polecat does exactly ONE thing: write code to solve the assigned problem. Everything else (checkout, quality gates, commit, push, signal, cleanup) is handled by the infrastructure before and after the polecat runs. The polecat's context window is 100% code, 0% lifecycle management.

This is the Ford insight applied directly: **the worker should not carry the car to the next station; the conveyor belt moves the car.** The polecat should not manage its own lifecycle; the infrastructure should manage the polecat.

#### 8. Crew (Persistent Human-Directed Agent)

**Current responsibilities** (3):
1. Maintain persistent workspace
2. Execute human-directed tasks
3. Preserve context across sessions

**Assessment**: This is the most naturally specialized role. It does one thing (execute what the human asks) in a durable context. No changes needed.

**Maximum specialization**: Already achieved. Crew is the ideal agent role -- single responsibility, clear interface.

### Specialization Summary Table

| Role | Current Responsibilities | Essential (AI-requiring) | Eliminable (Mechanical) |
|------|------------------------|--------------------------|------------------------|
| Mayor | 5 | 1 (escalation judgment) | 4 (dispatch, status, coordination, distribution) |
| Deacon | 6 | 0 | 6 (all can be daemon functions) |
| Boot | 2 | 0 | 2 (daemon already does this) |
| Dog | 3 | 0 | 3 (daemon goroutines) |
| Witness | 7 | 1 (novel recovery judgment) | 6 (monitoring, relay, pinging, handoffs) |
| Refinery | 4 | 1 (conflict resolution) | 3 (merge, validate, report) |
| Polecat | 6 | 1 (write code) | 5 (checkout, gates, commit, signal, cleanup) |
| Crew | 3 | 3 (all human-directed) | 0 |

**Total responsibilities across 8 roles: 36**
**AI-requiring: 7 (19%)**
**Mechanizable: 29 (81%)**

Ford decomposed 1 role into 29 specialized jobs. Gas Town should decompose 8 roles into 2 categories: **mechanical infrastructure** (the conveyor belt) and **AI judgment** (the specialized workers). The mechanical category handles 29 of 36 responsibilities. The AI category handles 7.

---

## Part 3: Lean/Waste Analysis (Seven Types of Muda)

### 1. Transport -- Unnecessary Data/Message Movement

**Examples in Gas Town**:

- **Health-check relay chains**: Deacon sends HEALTH_CHECK to Witness, Witness responds, Deacon processes response, Deacon sends HEALTH_CHECK to Refinery, Refinery responds. This is a multi-hop relay for information the daemon already has (process liveness from tmux).

- **POLECAT_DONE relay**: Polecat -> Witness -> Refinery requires two message hops. The polecat could write directly to a merge queue file/database.

- **WITNESS_PING -> Deacon -> DEACON_ALIVE -> Witness**: A three-message round trip that says "I'm alive" / "Are you alive?" / "Yes, I'm alive." The daemon's heartbeat timestamps already contain this information.

- **session-started nudges**: Every agent restart sends a session-started nudge to the Deacon. The Deacon does nothing useful with most of these. In the town.log, there are 30+ session-started nudges, almost all of which are noise.

**Elimination strategy**: Replace message-based liveness detection with file-based state (daemon writes heartbeat files, agents read them). Replace relay chains with direct writes to shared state (merge queue as a file, not a message).

### 2. Inventory -- Work Items Sitting Idle

**Examples in Gas Town**:

- **Merge queue backlog**: The town.log shows "2 MRs in queue" at 18:45, indicating work waiting for the Refinery. The Refinery is a persistent agent that may be processing other messages (health checks) instead of merges.

- **Beads in ready state**: `bd ready` returns available work, but polecats are not spawned until the Witness or Mayor dispatches them. Work items sit in "ready" state while supervision agents coordinate dispatch.

- **Convoy close backlog**: The daemon log shows bursts of convoy closes (18:50 shows 8 closes in one minute), suggesting work items queued up and then processed in a batch rather than flowing continuously.

**Elimination strategy**: Pull-based work assignment. Polecats spawn themselves when ready beads exist (the conveyor belt pulls the next car, rather than a manager pushing cars to workers). Merge queue processes continuously via daemon polling, not via message-triggered AI sessions.

### 3. Motion -- Unnecessary Steps and Context Switching

**Examples in Gas Town**:

- **Polecat lifecycle overhead**: A polecat must: read hook -> check mail -> run `gt prime` -> read assignment -> set up worktree -> *do actual work* -> run quality gates -> commit -> push -> signal done -> self-clean. The code-writing step is surrounded by 10 infrastructure steps.

- **Witness context-burn spiral**: The Witness fills its context window with health-check messages (not work-related content), triggers a handoff, restarts, re-reads state, resumes monitoring, and immediately begins accumulating health-check messages again. The Witness is doing more context-management *motion* than supervision *work*.

- **Deacon plugin execution**: The Deacon must check gate conditions, load plugin definitions, decide whether to run, execute, and log results. Each plugin invocation requires the Deacon to context-switch from patrol to execution and back.

**Elimination strategy**: Separate the infrastructure from the intelligence. Polecats should receive a pre-configured worktree and write code; everything else is pre- and post-processing by the daemon. The Witness context-burn is solved by eliminating the health-check messages that fill its context.

### 4. Waiting -- Agents Blocked on Other Agents

**Examples in Gas Town**:

- **Deacon stale for 16-31+ minutes**: Between heartbeats 28-31, the Deacon is stale for 16-25 minutes. The daemon nudges it repeatedly, but the Deacon is unresponsive. Meanwhile, all supervision that depends on the Deacon is blocked.

- **Refinery waiting for MERGE_READY**: The Refinery sits as a persistent agent doing nothing until a Witness sends a merge request. Between merges, it consumes a tmux pane and potentially API tokens on idle supervision responses.

- **Polecats waiting for dispatch**: Work exists in `bd ready` but polecats are not spawned until the supervision chain (Mayor -> Deacon -> Witness) decides to dispatch them. The supervision chain is itself frequently stale or restarting.

- **Sequential merge processing**: Even if multiple polecats finish simultaneously, the merge queue processes one at a time. Later merges wait on earlier ones.

**Elimination strategy**: Event-driven architecture. File watches or git hooks trigger immediate processing instead of relying on agent polling cycles. The merge queue processes mechanically on push events, not on message receipt by an AI agent. Polecat dispatch is triggered by bead creation, not by supervision coordination.

### 5. Overproduction -- Creating More Structure Than Needed

**Examples in Gas Town**:

- **10 work unit types**: Bead, Convoy, Hook, Molecule, Protomolecule, Wisp, Formula, Digest, Epic, CV chain. The current single-rig installation uses primarily beads and hooks. The other 8 types are architecture for future scale that imposes present complexity.

- **4 escalation severity levels**: Critical, High, Medium, Low -- each with different routing (bead, mail:mayor, email:human, sms:human). The current installation has empty contacts (`"contacts": {}`), meaning escalation routes to... nowhere beyond beads and the mayor.

- **4 plugin gate types**: Cooldown, Cron, Condition, Event -- a full scheduling framework for what is currently a handful of plugins.

- **Three-layer identity model**: Identity -> Sandbox -> Session -- for a system where agents are ephemeral and their "identity" is a name in a tmux pane.

- **8 message types in mail protocol**: For a system where ~90% of messages are health-check noise.

- **Dolt SQL server**: A full versioned database server for what could be JSONL files (which beads already uses as its underlying format).

**Elimination strategy**: Aggressively trim to what the current installation actually uses. One work type (bead), one escalation level (escalate or don't), one identity record, one message type (notification), one storage backend (JSONL/SQLite). Add complexity only when a specific scaling need demands it.

### 6. Over-processing -- More Work Than Value Warrants

**Examples in Gas Town**:

- **AI-powered health checks**: Checking if a tmux pane is alive does not require AI reasoning. Yet the Deacon, Boot, and Witness all use Claude Code sessions (consuming API tokens at ~$0.01-0.10 per check) to verify liveness. A `tmux list-panes` command achieves the same result for $0.

- **Full session restart for each Witness cycle**: Every ~8 minutes, the Witness performs a complete handoff: summarize state, write handoff message, terminate session, start new session, read context, resume. This is like stopping the assembly line every 8 minutes to rebuild the foreman's desk.

- **Convoy watcher processing close events**: The daemon log shows the convoy watcher detecting every bead close event, often duplicated ("detected close of hq-46q" appears twice each time). Each detection triggers processing logic for what is essentially a bookkeeping event.

- **Attribution tracking on supervision actions**: Every health check, every nudge, every patrol cycle gets BD_ACTOR attribution. Tracking *who checked if the system was alive* is audit overhead with no value.

**Elimination strategy**: Reserve AI token spend for value-producing work (code writing, conflict resolution, architectural decisions). Replace AI-based monitoring with zero-cost mechanical checks. Stop attributing supervision actions.

### 7. Defects -- Rework, Failed States, Restart Loops

**Examples in Gas Town**:

- **Deacon restart loop**: From the daemon log (heartbeats 65-97, approximately 21:24 to 23:09), the Deacon is restarted approximately every 6 minutes. Each restart: daemon detects stale -> kills session -> starts new session -> new session runs for ~3-6 minutes -> goes stale -> repeat. Over 2 hours, the Deacon was restarted approximately 15 times. Each restart consumes tokens for priming and produces no useful work.

- **Duplicate convoy watcher events**: Every convoy close is logged twice ("detected close of hq-46q" / "detected close of hq-46q"). This suggests either duplicate processing or a double-notification bug.

- **"Warning: no wisp config for hermes"**: This warning appears on every heartbeat cycle (194 occurrences in 97 heartbeats x 2 per cycle). The system logs a warning for a missing configuration that has never been configured. This is noise that pollutes logs and may obscure real issues.

- **Witness handoff frequency**: The Witness hands off every ~8 minutes, meaning each Witness session accomplishes roughly 8 minutes of monitoring before context exhaustion forces a restart. The new session spends time re-reading state before being useful. Effective monitoring time per session is perhaps 5-6 minutes out of 8.

**Elimination strategy**: Fix the Deacon stale loop by making its responsibilities mechanical (daemon functions). Remove the duplicate event processing bug. Suppress warnings for intentionally unconfigured features. Extend Witness session life by eliminating the health-check message traffic that fills its context.

### Waste Summary

| Waste Type | Severity | Primary Cause | Estimated Impact |
|-----------|----------|---------------|-----------------|
| Transport | High | Multi-hop message relays for liveness | ~90% of messages are waste |
| Inventory | Medium | Supervision bottleneck delays dispatch | Work waits for coordinators |
| Motion | High | Agent lifecycle overhead vs. actual work | Polecats: ~80% overhead, 20% code |
| Waiting | Critical | Deacon staleness blocks supervision | 16-31 min supervision gaps |
| Overproduction | High | 10 work types, 4 gates, 4 escalation levels | Cognitive overhead, code complexity |
| Over-processing | Critical | AI tokens for mechanical checks | $0.01-0.10 per check x ~100/hour |
| Defects | Critical | Deacon restart loop, duplicate events | ~15 restarts/2hr, all producing nothing |

---

## Part 4: Value Stream Map

### Current Flow: One Bead from Creation to Completion

```
STEP                          WHO            TIME/COST         VALUE?
────────────────────────────  ─────────────  ─────────────────  ──────
1. Bead created               Human/Mayor    ~30s, ~$0.02      YES
2. Bead enters ready pool     Beads DB       <1s, $0            yes
3. Mayor creates convoy       Mayor (AI)     ~60s, ~$0.05      partial*
4. Witness detects ready      Witness (AI)   ~0-480s wait, $0   NO (waiting)
   work (or next patrol)
5. Witness spawns polecat     Witness (AI)   ~15s, ~$0.03      partial*
6. Polecat session starts     tmux/daemon    ~5s, $0            infrastructure
7. Polecat runs gt prime      Polecat (AI)   ~30s, ~$0.05      NO (overhead)
8. Polecat reads hook         Polecat (AI)   ~10s, ~$0.02      NO (overhead)
9. Polecat reads assignment   Polecat (AI)   ~15s, ~$0.02      partial
10. Polecat sets up worktree  Polecat (AI)   ~20s, ~$0.02      NO (overhead)
11. *** POLECAT WRITES CODE   Polecat (AI)   ~300-1800s, $0.50+ YES <<<
12. Polecat runs tests        Polecat (AI)   ~60-300s, ~$0.05   YES
13. Polecat commits/pushes    Polecat (AI)   ~30s, ~$0.02      partial
14. Polecat signals done      Polecat (AI)   ~10s, ~$0.02      NO (overhead)
15. Witness receives done     Witness (AI)   ~0-480s wait, $0   NO (waiting)
16. Witness verifies state    Witness (AI)   ~20s, ~$0.03      partial
17. Witness sends MERGE_READY Witness (AI)   ~10s, ~$0.02      NO (relay)
18. Refinery receives msg     Refinery (AI)  ~0-480s wait, $0   NO (waiting)
19. Refinery runs merge       Refinery (AI)  ~30s, ~$0.05      YES
20. Refinery runs validation  Refinery (AI)  ~60-300s, ~$0.05   YES
21. Refinery sends result     Refinery (AI)  ~10s, ~$0.02      NO (overhead)
22. Witness receives result   Witness (AI)   ~0-480s wait, $0   NO (waiting)
23. Witness cleans polecat    Witness (AI)   ~15s, ~$0.02      NO (overhead)
24. Done                      --             --                 --
```

*"partial" = necessary but could be done more cheaply (mechanically rather than by AI)

### Time Analysis

**Best case (no waiting)**: ~10-40 minutes
**Worst case (with supervision delays)**: ~50-70 minutes

**Value-adding steps**: 1, 11, 12, 19, 20 = 5 steps
**Non-value-adding steps**: 3-10, 13-18, 21-23 = 18 steps
**Value-adding ratio**: 5/23 = **22% of steps produce value**

**Token cost breakdown (rough estimates)**:
- Value-producing work (code, tests, merge, validation): ~$0.60-1.00
- Overhead (priming, signaling, relaying, cleaning): ~$0.30-0.50
- Supervision (health checks during this bead's lifetime): ~$0.50-2.00
- **Total overhead ratio: 55-75% of tokens spent on non-value work**

### Bottlenecks

1. **Supervision polling gaps**: Steps 4, 15, 18, 22 all wait for the next patrol/check cycle of an AI agent. Each gap can be 0-8 minutes.

2. **Deacon staleness**: When the Deacon is stale (which happens every ~6 minutes for ~6-31 minutes), the entire supervision chain may degrade, creating cascading delays.

3. **Sequential merge queue**: Even if 5 polecats finish simultaneously, merges happen one at a time through a single AI agent (Refinery).

4. **Polecat startup overhead**: Steps 7-10 consume ~75 seconds and ~$0.11 before the polecat writes a single line of code.

### Ideal (Minimal Waste) Value Stream

```
STEP                          WHO              TIME/COST        VALUE?
────────────────────────────  ───────────────  ──────────────── ──────
1. Bead created               Human/CLI        ~5s, $0          YES
2. Daemon detects ready bead  Daemon (Go)      <1s, $0          infrastructure
3. Daemon provisions worktree Daemon (Go)      ~10s, $0         infrastructure
4. Daemon spawns polecat      Daemon (Go)      ~5s, $0          infrastructure
   in pre-configured worktree
5. *** POLECAT WRITES CODE    Polecat (AI)     ~300-1800s, $0.50+ YES <<<
6. Polecat signals done       exit code        <1s, $0          infrastructure
   (or daemon detects push)
7. Daemon runs merge          Daemon (Go)      ~5s, $0          YES
8. Daemon runs validation     Daemon (Go)      ~60-300s, $0     YES
9. If merge conflicts:        Conflict AI      ~120s, ~$0.20    YES (rare)
   spawn conflict resolver
10. Daemon cleans worktree    Daemon (Go)      ~5s, $0          infrastructure
11. Done                      --               --               --
```

**Steps**: 11 (down from 23)
**Value-adding steps**: 1, 5, 7, 8, (9 when needed) = 4-5 steps
**Value ratio**: 4/11 = **36-45% of steps produce value** (up from 22%)
**AI token cost**: ~$0.50-0.70 (down from ~$1.40-3.50)
**Waiting time**: Near zero (event-driven, not poll-driven)
**Total time**: ~6-35 minutes (down from ~10-70 minutes)

### Improvement Factor

- **Steps**: 2x reduction (23 -> 11)
- **Token cost**: 2-5x reduction
- **Latency**: 2-3x reduction (elimination of polling gaps)
- **Value ratio**: 2x improvement (22% -> 45%)

---

## Part 5: Compounding Gains Analysis

The key principle from Article 2: gains multiply, not add. **2x * 2x * 2x = 8x, not 6x.** Which improvements amplify each other?

### Improvement 1: Mechanize Supervision (2-3x)

**What**: Replace AI-based health monitoring (Boot, Deacon-as-supervisor, Witness-as-monitor) with daemon process checks (tmux liveness, file timestamps, PID monitoring).

**Direct effect**: Eliminates ~$0.50-2.00 in supervision tokens per bead. Eliminates Deacon restart loop. Eliminates 90% of inter-agent communication.

**Compounding effect**: This improvement *enables* every other improvement. When supervision messages stop flooding agent context windows:
- Witness context lasts 10x longer (no more 8-minute handoff cycles)
- Deacon context doesn't fill with WITNESS_PING noise
- Refinery isn't interrupted by health checks
- All agents can dedicate 100% of context to their actual work

**This is the "conveyor belt" -- the infrastructure that makes everything else possible.**

### Improvement 2: Event-Driven Dispatch (2-3x)

**What**: Replace poll-based agent coordination with event-driven triggers. Daemon watches for filesystem events (git push, bead file changes) and triggers actions immediately.

**Direct effect**: Eliminates 0-8 minute polling gaps at steps 4, 15, 18, 22 in the current value stream.

**Compounding effect with Improvement 1**: Once supervision is mechanical, the daemon can be the single event processor. It watches for pushes, triggers merges, spawns workers -- all within milliseconds. The combination of mechanical supervision + event-driven dispatch eliminates the entire "waiting for the next patrol cycle" category of waste.

**Estimated combined effect**: (2.5x supervision) * (2.5x dispatch) = **6.25x improvement** in throughput latency.

### Improvement 3: Polecat Context Purity (2-3x)

**What**: Pre-configure worktrees before polecat spawn. The polecat receives a fully configured environment (correct branch, assignment details in a file, worktree ready) and its ONLY job is writing code. Post-completion, the daemon handles commit, push, merge submission, and cleanup.

**Direct effect**: Polecat startup overhead drops from ~75 seconds to ~5 seconds. Polecat context is 100% code-related, 0% lifecycle management. Quality of code output improves because the agent isn't wasting tokens on git operations and protocol compliance.

**Compounding effect with Improvements 1 & 2**: When supervision is mechanical AND dispatch is event-driven AND polecats start instantly with pure context, the entire pipeline from "bead created" to "code written" drops from 10-45 minutes to 5-30 minutes. But the quality improvement compounds too: a polecat with 100% code context produces better code, which means fewer merge conflicts, which means less Refinery work, which means faster throughput for all polecats.

**Estimated combined effect**: (2.5x) * (2.5x) * (2x) = **12.5x improvement** in effective output per token.

### Improvement 4: Unified Work Type (1.5-2x)

**What**: Collapse 10 work unit types into one (bead with status field and optional tags). Eliminate Convoy, Hook, Molecule, Protomolecule, Wisp, Formula, Digest, Epic, CV chain as separate concepts.

**Direct effect**: Dramatic reduction in system prompt complexity. Agent instructions become simpler. `bd` CLI surface area shrinks. Developer cognitive load drops.

**Compounding effect with Improvement 3**: Simpler work types mean simpler polecat instructions. Simpler instructions mean more context for code. More context for code means better output quality. Better quality means fewer rework cycles. This is a second-order compounding effect that ripples through every agent interaction.

**Estimated combined effect**: (2.5x) * (2.5x) * (2x) * (1.5x) = **18.75x improvement**.

### Improvement 5: Mechanical Merge Pipeline (1.5-2x)

**What**: Daemon runs `git merge` and `make test` directly. AI agent spawned only for conflict resolution (rare case). Merges happen within seconds of push, not minutes.

**Direct effect**: Merge latency drops from 5-20 minutes (including Refinery wait time) to seconds. Conflict resolution remains high-quality (AI when needed) but the common case is instant.

**Compounding effect with all above**: Faster merges mean the main branch stays closer to current work, which means fewer conflicts for subsequent polecats, which means even fewer AI merge-resolution invocations. This is a positive feedback loop: fast merges -> fewer conflicts -> even faster merges.

**Estimated total compounding**: (2.5x) * (2.5x) * (2x) * (1.5x) * (1.5x) = **~28x improvement**.

### Compounding Gains Summary

| # | Improvement | Individual | Cumulative (Compounded) |
|---|------------|-----------|------------------------|
| 1 | Mechanize supervision | 2-3x | 2.5x |
| 2 | Event-driven dispatch | 2-3x | 6.25x |
| 3 | Polecat context purity | 2-3x | 12.5x |
| 4 | Unified work types | 1.5-2x | 18.75x |
| 5 | Mechanical merge pipeline | 1.5-2x | ~28x |

If these improvements were additive: 2.5 + 2.5 + 2 + 1.5 + 1.5 = 10x.
Because they compound: 2.5 * 2.5 * 2 * 1.5 * 1.5 = **~28x**.

The difference between additive thinking and multiplicative thinking is the difference between 10x and 28x improvement.

### The Meta-Insight

The single most important Ford principle applied to Gas Town is this:

**Gas Town built a factory where the supervisors outnumber the workers.**

In the current single-rig installation:
- Supervision agents: Mayor, Deacon, Boot, Dog, Witness = 5
- Work agents: Polecats (ephemeral), Crew = 1-3 at any time
- Infrastructure: Refinery, Daemon = 2

The supervisors are AI agents consuming tokens to watch other AI agents consume tokens. Ford would look at this factory and ask: "Why do you have 5 foremen watching 2 workers? And why are the foremen more expensive than the workers?"

The Ford-level redesign is radical: **the factory itself (daemon + filesystem + git hooks) becomes the conveyor belt, and the only AI agents are the ones doing the actual work (writing code, resolving conflicts, making architectural decisions).** Supervision is not a job. It is a property of the infrastructure.

---

## Appendix: Implementation Priority

### Phase 1: Stop the Bleeding (Week 1)
- Fix the Deacon restart loop by making health checks a daemon function
- Eliminate HEALTH_CHECK nudge traffic between agents
- Suppress "no wisp config" warnings
- Fix duplicate convoy watcher events

### Phase 2: Build the Conveyor Belt (Weeks 2-3)
- Implement event-driven dispatch in daemon (filesystem watch for bead changes)
- Implement mechanical merge (daemon runs `git merge` + validation scripts)
- Pre-configure polecat worktrees before agent spawn
- Retire Boot and Dog roles entirely

### Phase 3: Specialize the Workers (Weeks 3-4)
- Simplify polecat system prompt to code-only focus
- Collapse work unit types to bead + status
- Replace Witness AI agent with daemon goroutine
- Replace Refinery AI agent with daemon function + on-demand conflict-resolution agent

### Phase 4: Measure and Iterate (Ongoing)
- Track tokens per bead completed (total cost of work)
- Track time from bead creation to merge (lead time)
- Track polecat context utilization (% of tokens spent on code vs. overhead)
- A/B test simplified pipeline against current architecture

---

*Analysis produced by applying Henry Ford's Assembly Line Principles and the 10,000x Knowledge Worker Extreme Specialization Audit to Gas Town's multi-agent orchestration system. Source frameworks from Michael Simmons / Blockbuster Thought Leader School, Article 2.*
