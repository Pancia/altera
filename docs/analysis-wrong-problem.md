# Wrong-Problem Detector: Applied to Gas Town

**Recipe applied**: Wrong-Problem Detector (from Article 1, "Every Mental Model You've Learned Is Wrong")
**Subject**: Gas Town multi-agent orchestration system
**Date**: 2026-02-19

---

## Preamble: What Problem Does Gas Town Claim to Solve?

Gas Town's stated problem: "How do we coordinate 4-30+ AI coding agents working in parallel across multiple git repositories without losing context, duplicating effort, or going unsupervised?"

The system answers this with a hierarchical supervision architecture (Mayor -> Deacon -> Witness -> Workers), structured work units (beads), persistent agent identity, a mail/nudge communication protocol, a merge queue, and a three-tier watchdog chain. 61 Go packages. 13 agent roles. 10+ work unit types. A dedicated SQL server (Dolt) for tracking.

The question we must ask: **Is "multi-agent orchestration" the right problem at all?**

---

## Step 1: Zeroth Principle via Inversion

*Question the assumptions behind the assumptions about the problem definition.*

### The Surface Assumptions

Gas Town makes several explicit assumptions:

1. AI agents need hierarchical supervision
2. Agents need persistent identity and attribution
3. Work must be decomposed into structured, trackable units (beads)
4. Agents must communicate via asynchronous mail and real-time nudges
5. A merge queue is needed to integrate parallel work
6. Session handoffs are necessary to survive context window limits
7. Health monitoring requires a multi-tier watchdog chain

### The Meta-Assumptions (One Layer Deeper)

But behind each of these lies an assumption that is never questioned:

**Meta-Assumption 1: "Agents are unreliable workers who need supervision."**

The entire supervision hierarchy -- Daemon watching Boot watching Deacon watching Witness watching Polecats -- assumes agents are fundamentally unreliable and will go stale, zombie, or produce bad work without continuous oversight. But what if agent unreliability is not an inherent property but an artifact of the system's own complexity? The Deacon goes stale every 6 minutes. The system's response is to add more supervision (Boot, nudges, heartbeat checks). But the Deacon might be going stale because it is spending most of its context window processing supervision overhead rather than doing useful work. **The supervision may be causing the very failure it was designed to prevent.**

**Meta-Assumption 2: "Agents are persistent entities that accumulate identity over time."**

Gas Town gives agents names (Rust, Chrome, Nitro), persistent identity across sessions, performance histories ("Agent CVs"), and three-layer identity structures (Identity -> Sandbox -> Session). This assumes agents are like human employees who improve with experience and whose track record matters. But current LLM-based agents do not learn across sessions. A polecat named "Rust" that has completed 50 tasks has exactly the same capabilities as a fresh one. The "CV" is metadata for human consumption, not something the agent itself benefits from. The entire identity/attribution system may be solving a human-legibility problem, not an agent-coordination problem.

**Meta-Assumption 3: "Work must flow through a command hierarchy."**

The Mayor creates convoys, the Deacon dispatches, the Witness monitors, the Refinery merges. This mirrors a corporate org chart. The meta-assumption is that coordination requires hierarchy -- that someone must be "in charge" at each level. But the actual productive work (a polecat editing code in a worktree and pushing a branch) has no inherent dependency on this hierarchy. A polecat could be spawned by a cron job, pick up the next open bead, do the work, and submit a PR. No Mayor. No Deacon. No Witness. The hierarchy exists to manage complexity that the hierarchy itself creates.

**Meta-Assumption 4: "Communication between agents is a design problem to be solved."**

Gas Town has 8 message types, a mail protocol, nudges, handoffs, and escalation with 4 severity levels. The meta-assumption is that agents need to talk to each other and that the quality of this communication protocol matters. But what if the need for inter-agent communication is itself a symptom of wrong decomposition? If each agent's task is truly independent and well-defined, the only communication needed is: (a) "here is your task" and (b) "here is my output." Git already provides both of these.

**Meta-Assumption 5: "The problem is orchestration."**

The deepest assumption. Gas Town frames itself as an "orchestration system." But orchestration implies that the agents are instruments in a symphony -- that their timing, coordination, and harmony require a conductor. What if the agents are not instruments but independent contractors? What if the problem is not orchestration but simple task distribution? A bulletin board where tasks are posted, workers pick them up, and results are checked. No symphony. No conductor. No harmony needed. Just a queue and a quality gate.

**Meta-Assumption 6: "Persistent, always-on supervision is needed even when no work exists."**

The daemon runs 97 heartbeat cycles. The logs show that for hours, the system does nothing but check that the Deacon is alive, and the Deacon does nothing but check that the Witnesses are alive, and the Witnesses do nothing but confirm there are no polecats to watch. The assumption is that the supervision layer must be always-on, ready for work. But this is a pattern from traditional server infrastructure (keep the web server running even when no one visits). AI agents are not servers. They cost money per token per second. An event-driven architecture (spin up supervision only when work arrives) would eliminate the idle burn entirely.

**Meta-Assumption 7: "Session handoffs preserve meaningful context."**

The system has an elaborate handoff mechanism where agents write a message to their next session. But an LLM reading a handoff note is not the same as a human picking up where they left off. The next session is a fresh model instance that reads some text. There is no actual continuity of reasoning, working memory, or understanding. The handoff creates the illusion of persistence while the underlying reality is that each session starts from scratch and reconstructs its understanding from whatever text is available. If the handoff note is good, the file system state plus the handoff note is sufficient. If the handoff note is bad, the handoff mechanism fails regardless of how elaborate it is. The mechanism adds complexity without changing the fundamental constraint (context windows are finite).

---

## Step 2: Anomaly Hunting

*Look for surprising observations that don't fit the current framing.*

### Anomaly 1: The Deacon Death Spiral

The daemon log tells a damning story. Starting around 21:09, the Deacon enters a death spiral:

```
21:09 - stale (15m)
21:12 - stale (18m)
21:15 - stale (21m)
21:18 - stale (24m)
21:21 - stale (28m) -> restart
21:24 - started successfully
21:30 - stale (37m) -> restart
21:36 - stale (43m) -> restart
21:42 - stale (49m) -> restart
21:49 - stale (55m) -> restart
21:55 - stale (1h1m) -> restart
22:01 - stale (1h7m) -> restart
22:07 - stale (1h14m) -> restart
... continues every 6 minutes until 23:09
```

**What this actually tells us**: The Deacon is being restarted every 6 minutes for over two hours. Each restart spawns a new Claude Code session (costing tokens), the session starts, and then immediately becomes "stale" from the daemon's perspective because the stale timer is never resetting. The Deacon is not actually doing anything useful during this period -- it starts, perhaps runs one patrol cycle, and then the session ends or hangs, and the daemon detects staleness again.

**The framing says**: The supervision system is correctly detecting and recovering from agent failure. NDI (Nondeterministic Idempotence) at work.

**What it actually reveals**: The system is spending ~$X in tokens every 6 minutes to repeatedly start an AI session that does nothing productive. The "recovery" is not recovering anything -- it is repeating the same failure in a loop. The supervision is not "detecting and recovering" but "detecting and re-triggering." The staleness counter in the log increases monotonically (15m, 18m, 21m, 24m... 1h1m, 1h7m, 1h14m...) even across restarts, suggesting the heartbeat timestamp is never being successfully updated by new sessions. The restart does not fix the problem; it merely costs tokens.

This is the single most important anomaly in the entire system. **A two-hour period where the primary supervisor is being restarted every 6 minutes at AI-token cost, accomplishing nothing, is not a supervision success story. It is evidence that the supervision architecture is wrong.**

### Anomaly 2: The Supervision-to-Work Ratio

From the town log:
- **5 actual work completions** (polecat tasks done)
- **5 polecat kills** (cleanup after work)
- **111 nudges** (inter-agent messages)
- **45 health checks**
- **23 handoffs** (session boundaries)
- **26 stale detections** (from daemon log)

The ratio is staggering: **for every 1 unit of productive work completed, there were 22 nudges, 9 health checks, and 5 stale detections.** The system generated 40x more coordination overhead than actual output.

**The framing says**: This overhead is the cost of reliable multi-agent orchestration.

**What it actually reveals**: The system may be a Rube Goldberg machine. If a human had simply run 5 Claude Code sessions sequentially, giving each one a task and waiting for the result, those 5 tasks would have completed with zero nudges, zero health checks, and zero stale detections. The entire orchestration layer cost more (in tokens, time, and complexity) than the work it orchestrated.

### Anomaly 3: The Witness Handoff Frequency

The Witness for the hermes rig hands off every 7-8 minutes:
```
18:14 -> 18:24 -> 18:39 -> 19:06 -> 19:19 -> 19:27 -> 19:35 -> 19:45 -> 19:54 -> 20:11 -> 20:19 -> 20:27 -> 20:35 -> 20:42 -> 20:51 -> 21:00
```

That is 16 handoffs in under 3 hours. Each handoff means the Witness hit its context window limit, wrote a handoff note, and a fresh session started. The Witness is spending most of its time re-establishing context rather than monitoring anything. When there are no polecats to watch (as the witness state confirms: "No polecats, no MRs"), the Witness is running patrol cycles that produce no value, burning through context windows, handing off, and repeating.

**What it actually reveals**: The Witness role cannot sustain itself. Its task (watching polecats) is bounded, but its execution model (continuous AI patrol loop) is unbounded. An AI agent continuously monitoring for events is the wrong tool for what is fundamentally a polling or event-subscription problem.

### Anomaly 4: Duplicate Convoy Watcher Events

Nearly every convoy watcher detection appears twice:
```
18:10:19 convoy watcher: detected close of hq-46q
18:10:19 convoy watcher: detected close of hq-46q
```

This happens consistently throughout the log. Either events are being processed twice, or there is a logging bug causing double-emission. Either way, it suggests the event pipeline has not been thoroughly debugged -- and this is the Go daemon layer (Layer 1, "pure mechanical, no AI reasoning"), which should be the most reliable component.

### Anomaly 5: The "No Wisp Config" Warning Storm

The message "Warning: no wisp config for hermes - parked state may have been lost" appears on every single heartbeat (97 times). This warning has been emitting continuously since the daemon started and was never resolved. It suggests a configuration gap that the system acknowledges but cannot fix -- the daemon knows something is wrong but the fix requires human intervention that never comes.

### Anomaly 6: Single-Rig Reality vs. Multi-Rig Design

The system was designed for 5-20 rigs. It has exactly 1 (hermes, a Swift macOS app). The Mayor, Deacon, convoy watcher, federation system, cross-project references, and capability-based routing all exist to handle multi-rig coordination that does not occur. The entire supervision hierarchy (Daemon -> Boot -> Deacon -> per-rig Witnesses) is load-bearing only at scale. With one rig, the Deacon and Witness are essentially redundant -- there is nothing to coordinate across.

---

## Step 3: Removal Test

*Mentally remove each major component and assess what actually breaks.*

### Remove the Deacon

**What the Deacon does**: Patrols all rigs, dispatches plugins, coordinates recovery, health-checks Witnesses and Refineries.

**What breaks without it**: Nothing that cannot be handled by simpler mechanisms.
- Health checking Witnesses? The daemon already checks if tmux sessions are alive. Replace AI-based health checking with process-level checking.
- Dispatching plugins? A cron-triggered script could do this.
- Coordinating recovery? The daemon already restarts dead sessions. Recovery dispatch is the Deacon's job, but the Deacon itself is the primary thing needing recovery.

**Verdict: DECORATIVE.** The Deacon is the single largest consumer of tokens for supervision, and its primary observable behavior in the logs is going stale and being restarted. Removing it and distributing its responsibilities to the daemon (process health) and a simpler event-driven dispatcher (plugin execution, recovery) would reduce complexity and cost with no loss of functionality.

### Remove the Mayor

**What the Mayor does**: Creates convoys, distributes work, serves as escalation target.

**What breaks without it**: Work distribution would need a different mechanism.
- Creating convoys is a planning/decomposition function. This could be done by the human user directly (create beads, assign to agents) or by a single planning session (not a persistent agent).
- Escalation could go directly to the human (email/SMS on failure).

**Verdict: PARTIALLY LOAD-BEARING, but not as a persistent agent.** The Mayor's work-decomposition function is valuable but episodic -- it is needed when new work arrives, not continuously. A "Mayor session" invoked on-demand when the human submits a goal would serve the same purpose without the persistent overhead.

### Remove Beads Entirely

**What beads do**: Track work as structured JSONL data in Dolt, enabling querying, attribution, audit trails.

**What breaks without them**: Work tracking reverts to git branches and PRs, which already track what was done, by whom, and when.
- Attribution? Git commits have author fields.
- Audit trails? Git log provides this.
- Querying? GitHub Issues, Linear, or any external tracker provides structured query.
- Performance history? PR merge rates, test pass rates, and cycle times are derivable from git + CI.

**Verdict: MOSTLY DECORATIVE at current scale.** Beads duplicate information already available in git. At enterprise scale with 30 agents and compliance requirements, structured work-unit tracking has value. For 1 rig with a handful of polecats, beads are overhead. The question is whether building the tracking infrastructure before it is needed is premature optimization or strategic foresight.

### Remove the Mail Protocol

**What mail does**: Structured async messages between agents (POLECAT_DONE, MERGE_READY, MERGED, etc.).

**What breaks without it**: Agents cannot coordinate lifecycle transitions.
- But: If a polecat finishes work, it could simply push a branch and create a PR. A webhook or file-watcher could trigger the merge queue. No mail needed.
- If a merge fails, the PR gets a failing status check. No mail needed.
- If an agent is stuck, it writes to a file or exits with an error code. No mail needed.

**Verdict: REPLACEABLE.** The mail protocol recreates functionality already provided by git's collaboration model (branches, PRs, status checks) and basic process management (exit codes, file watches). The mail protocol is isomorphic to these existing mechanisms but adds its own failure modes (messages lost, stale, or ignored).

### Remove Persistent Identity

**What identity does**: Agents have names, permanent IDs, three-layer identity structures, performance histories.

**What breaks without it**: Attribution becomes generic. Logs say "an agent did this" rather than "Chrome did this."
- But: Git author fields and branch names already provide attribution. `agent-7/feature-branch` is as informative as `chrome/feature-branch` for debugging purposes.
- Performance history: Without identity, you cannot build "agent CVs." But agent CVs are metadata for human consumption (the agent does not read its own CV and perform better).

**Verdict: LOW STRUCTURAL IMPORTANCE.** Identity is a user-experience feature, not an architectural necessity. It helps humans reason about the system but does not affect the system's ability to do work.

### Remove the Merge Queue (Refinery)

**What the Refinery does**: Manages merge order, handles conflicts, runs quality gates.

**What breaks without it**: Parallel branches could conflict on merge. Without ordered integration, you get merge conflicts and broken main.
- But: GitHub's built-in merge queue exists for exactly this purpose. CI/CD already runs quality gates. The Refinery as an AI agent is spending tokens to do what GitHub Actions could do mechanically.

**Verdict: THE FUNCTION IS LOAD-BEARING; THE AI IMPLEMENTATION IS NOT.** Ordered merging of parallel work is genuinely necessary. But it does not require an AI agent. It requires a queue data structure and a CI pipeline.

### Remove Health Monitoring

**What health monitoring does**: Detects stuck agents, triggers recovery.

**What breaks without it**: Zombie agents would go undetected, work would stall.
- But: Tmux session liveness is checkable without AI. Exit codes tell you if a process succeeded. A 3-line shell script polling tmux sessions would provide the same detection as the Deacon + Boot + Witness chain -- at zero token cost.

**Verdict: THE FUNCTION IS LOAD-BEARING; THE AI IMPLEMENTATION IS NOT.** Health monitoring is necessary. Using AI agents to implement it is not.

### Summary of Removal Test

| Component | Load-Bearing? | Could Be Simpler? |
|-----------|--------------|-------------------|
| Deacon | No | Eliminate entirely |
| Mayor | Partially (work decomposition) | On-demand session, not persistent agent |
| Beads | No (at current scale) | Git + external tracker |
| Mail Protocol | No | Git branches + webhooks |
| Persistent Identity | No | Git author fields |
| Merge Queue | Yes (the function) | GitHub merge queue + CI |
| Health Monitoring | Yes (the function) | Shell scripts + process monitoring |
| Git Worktree Isolation | **Yes** | Irreducible |
| Session Handoff | Partially | Simpler file-based context |

**The only truly irreducible element is git worktree isolation** -- the mechanism that lets multiple agents edit code simultaneously without conflicts. Everything else is either decorative or reimplements existing tooling at higher cost.

---

## Step 4: Crucial Experiment

*Design tests that would definitively prove the current architecture is the wrong approach.*

### Experiment 1: The Lobotomy Test

**Hypothesis**: Gas Town's supervision hierarchy adds more value than it costs.

**Design**: Run the same 10 tasks two ways:
- **Treatment A**: Full Gas Town stack (Daemon, Boot, Deacon, Witness, Refinery, Polecats, beads, mail protocol)
- **Treatment B**: A bash script that spawns N Claude Code sessions in parallel worktrees, each given a task as a CLI argument, with a simple merge script that runs after all sessions exit

**Measure**: Wall-clock time to completion, total tokens consumed, number of tasks successfully completed, number of merge conflicts.

**What proves Gas Town wrong**: If Treatment B completes the same work in comparable time with fewer tokens and comparable success rate, then the entire supervision layer is overhead. The prediction from the anomaly data is that Treatment B would win decisively, because 40x fewer coordination messages means 40x fewer tokens spent on non-work.

### Experiment 2: The Idle Cost Audit

**Hypothesis**: Persistent supervision agents are cost-effective.

**Design**: Run Gas Town overnight with no work queued. Measure total tokens consumed by all supervision agents (Deacon, Boot, Witness, Refinery) over 8 hours.

**Measure**: Total API cost of idle supervision.

**What proves Gas Town wrong**: The daemon log already suggests this experiment has been inadvertently run. From ~21:00 to ~23:09, the system had no active work but the Deacon was being restarted every 6 minutes (approximately 20 restarts at non-trivial token cost per restart). If the idle overnight cost is non-trivial (e.g., more than $1), then always-on AI supervision is economically indefensible for a system that may only receive work during business hours.

### Experiment 3: The Flat vs. Hierarchical Race

**Hypothesis**: Hierarchical agent supervision (Mayor -> Deacon -> Witness -> Polecat) produces better outcomes than flat task distribution.

**Design**: Take a real project milestone with 20 tasks. Run it two ways:
- **Hierarchical**: Mayor decomposes, Deacon dispatches, Witness monitors, Polecats execute
- **Flat**: Human decomposes tasks into a queue, a scheduler spawns polecats directly, a mechanical merge queue integrates results

**Measure**: Tasks completed per hour, error rate (bad merges, failed tests), human intervention required.

**What proves Gas Town wrong**: If flat distribution achieves comparable throughput with less human intervention, then the hierarchy is not providing the coordination value it claims. The key prediction: at 1-rig scale with well-defined tasks, flat distribution will win because the overhead of the hierarchy exceeds the value of its coordination.

### Experiment 4: The AI-vs-Mechanical Supervision Test

**Hypothesis**: AI-based health monitoring (Deacon checking Witnesses via nudges) provides better recovery than mechanical monitoring.

**Design**: Instrument both:
- **AI supervision**: Current Gas Town stack
- **Mechanical supervision**: A daemon that checks tmux session liveness every 30 seconds, restarts dead sessions, and escalates to human via SMS after 3 consecutive failures

**Measure**: Mean time to detect failure, mean time to recover, false positive rate, tokens consumed.

**What proves Gas Town wrong**: If mechanical supervision detects failures faster (30-second polling vs. 3-minute heartbeat) and recovers them more reliably (process restart vs. AI session that might itself go stale), then using AI for supervision is strictly worse than using traditional process management. The Deacon death spiral (26 stale detections over 2 hours with no successful recovery) strongly suggests mechanical supervision would outperform.

### Experiment 5: The Attribution Value Test

**Hypothesis**: Persistent agent identity and attribution data improves decision-making.

**Design**: After 100 completed tasks, present a human decision-maker with:
- **Treatment A**: Full Gas Town attribution data (agent CVs, performance history, skill ratings)
- **Treatment B**: Git log with branch names and PR descriptions

Ask the human to answer: "Which tasks are blocked?", "What went wrong with task X?", "Should we assign task Y to the same agent that did task Z?"

**Measure**: Decision quality, time to answer, user confidence.

**What proves Gas Town wrong**: If the git-log-only view provides sufficient information for all practical decisions, then the attribution system is solving a problem that does not exist. The critical test is whether any decision changes based on the additional attribution data. If the agent CVs and identity data never alter a routing or debugging decision, they are pure overhead.

### Experiment 6: The Context Window Boundary Test

**Hypothesis**: Session handoffs preserve meaningful work continuity.

**Design**: Give an agent a task that requires 3 sessions to complete (due to context window limits). Compare:
- **Treatment A**: Gas Town handoff protocol (agent writes structured handoff, next session reads it)
- **Treatment B**: No handoff; fresh session reads the filesystem state (git diff, open beads, code state) and the task description

**Measure**: Continuation success rate, work duplication (redoing already-done work), final output quality.

**What proves Gas Town wrong**: If Treatment B (no handoff, just filesystem state) achieves comparable continuation success, then the handoff protocol adds complexity without adding value. The prediction: for coding tasks, the filesystem state (what code exists, what tests pass, what branch we are on) contains 90% of the meaningful context. The handoff note adds at most 10% -- and that 10% comes at the cost of the handoff protocol's complexity.

---

## Synthesis: Is Gas Town Solving the Wrong Problem?

The evidence is strong that Gas Town is, at minimum, solving the right problem with the wrong solution -- and may be solving the wrong problem entirely.

### The Right Problem (that Gas Town buries under orchestration)

The actual valuable core is small:

1. **Parallel git worktree management** -- letting multiple agents edit code simultaneously
2. **Task queue** -- distributing work items to available agents
3. **Merge ordering** -- integrating parallel branches without conflicts
4. **Failure detection** -- knowing when an agent has died

These four functions are the load-bearing structure. Everything else -- hierarchical supervision, persistent identity, mail protocols, beads, nudges, handoffs, agent CVs, convoy watchers, molecules, wisps, formulas -- is superstructure built on an assumption that has not been validated: that AI agent coordination requires AI agent supervision.

### The Wrong Problem (that Gas Town may be solving)

Gas Town may be solving: **"How do we build an organization chart for AI agents?"**

This is the wrong problem because:

1. **AI agents are not employees.** They do not learn from experience, form working relationships, develop institutional knowledge, or need career development. The org-chart metaphor (Mayor, Deacon, Witness) imports assumptions from human management that do not apply.

2. **Supervision of AI by AI is a recursive cost center.** Each layer of AI supervision consumes the same scarce resource (tokens/context) as the productive work itself. In human organizations, managers use different skills than workers. In Gas Town, supervisors and workers are the same model, consuming the same tokens, and the supervisors are more likely to fail (as the Deacon death spiral shows) because supervision is a harder, less well-defined task than executing a specific coding task.

3. **The communication overhead exceeds the coordination value.** 111 nudges for 5 completed tasks. The system is talking to itself more than it is working. This is not a coordination success; it is a coordination pathology.

### What the Right Problem Might Actually Be

If we strip away the org-chart metaphor, the right problem might be:

**"How does a human direct N parallel coding agents with minimum overhead?"**

The answer to that question looks nothing like Gas Town. It looks like:
- A task queue (Postgres, Redis, or even a text file)
- A spawner that creates worktrees and starts agent sessions
- A watcher that detects process death (not AI-based -- process-level)
- A merge queue (GitHub's built-in one, or a simple script)
- A dashboard showing task status

No Mayor. No Deacon. No Witness. No beads. No mail protocol. No agent identity. No handoffs. No molecules. No convoys. No wisps.

Would this simpler system handle 30 agents across 20 rigs? Maybe not. But Gas Town does not have 30 agents across 20 rigs. It has 1 rig with a handful of polecats. And the architecture analysis itself acknowledges this: "Much of the architecture is designed for 5-20 rigs with cross-rig coordination, federation, and model A/B testing -- features that provide value at scale but add overhead at small scale."

The question is not whether Gas Town will be the right system at scale. The question is whether building the scale-system before achieving scale is preventing you from ever reaching scale -- because the overhead of the system itself consumes the resources that would otherwise produce the work that would justify the system.

### The Ford Question

Article 2 cites Henry Ford's breakthrough: shifting from "How can workers move faster?" to "Why are workers moving at all?"

Applied to Gas Town: **Why are agents supervising at all?**

The Deacon does not need to be an AI agent. The Witness does not need to be an AI agent. The Boot does not need to be an AI agent. Health monitoring, process liveness, merge queue management -- these are mechanical functions that traditional software handles better, faster, and cheaper than LLMs.

The only place AI genuinely adds value in this system is:
1. **Work decomposition** (breaking a goal into tasks) -- Mayor function, episodic
2. **Code generation** (executing tasks) -- Polecat function, the actual work
3. **Conflict resolution** (when merges require understanding code intent) -- rare, could be on-demand

Everything between the human's goal and the polecat's code generation is plumbing. Gas Town has turned the plumbing into an AI-powered smart-home system when what was needed was copper pipes.

---

## Conclusion

Gas Town is an impressively engineered system that may be an elaborate answer to the wrong question. The evidence suggests:

1. **The supervision architecture causes more failures than it prevents** (Deacon death spiral)
2. **The coordination overhead dwarfs the productive output** (111 nudges : 5 completions)
3. **Most components are replaceable by simpler, non-AI mechanisms** (Removal Test)
4. **The system is designed for a scale it has not reached**, and its overhead may prevent reaching that scale
5. **The org-chart metaphor imports inapplicable assumptions** from human management

The crucial experiments proposed above would either validate or falsify these conclusions. But the anomaly data is already suggestive: a system that spends two hours restarting its supervisor every six minutes, generating zero productive work, is not orchestrating -- it is thrashing.

The right next step is not to optimize Gas Town. It is to run Experiment 1 (the Lobotomy Test) and determine whether the simplest possible alternative -- a bash script with parallel worktrees -- achieves comparable results. If it does, then Gas Town's complexity is not solving a problem. It is the problem.
