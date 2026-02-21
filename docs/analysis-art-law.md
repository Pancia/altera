# Gas Town Through Two Lenses: Art and Constitutional Law

An interdisciplinary analysis of the Gas Town multi-agent orchestration system, applying the combined perspectives of creative direction (aesthetics, choreography, craft) and constitutional law (governance, rights, due process). These two disciplines share a deep preoccupation: the relationship between structure and freedom, constraint and expression, form and function.

---

## Part I: The Art / Creative Direction Lens

### 1. Aesthetics of Coordination

**Where is the beauty?**

There is a genuine beauty in Gas Town's ambition. The *concept* -- a town of autonomous agents, each with a name and identity, coordinating through structured messages while a hierarchy of supervisors ensures nothing falls through the cracks -- has the appeal of a Joseph Cornell box: an intricate miniature world, self-contained, internally consistent, obsessively detailed. The naming scheme alone (Town, Rig, Polecat, Refinery, Convoy, Wisp, Molecule, Seance) reveals an imagination that treats system design as world-building. This is not an accident. The designer is not merely solving an engineering problem; they are constructing a *place*, a setting with its own vocabulary, its own physics, its own culture. The beads-as-structured-data insight, the propulsion principle, the three-layer identity model -- these have the quality of good speculative fiction: internally rigorous, surprising in their implications, and genuinely illuminating about how coordination might work.

The most beautiful single idea is the Propulsion Principle: "If work is on your hook, YOU RUN IT." This has the compression of a great design rule. It is a principle that generates behavior rather than constraining it. Like the best rules in choreography ("follow the weight") or in architecture ("form follows function"), it tells you what to do in any situation without specifying the situation. It is generative. It produces motion.

**Where is it ugly?**

The ugliness is in the gap between the concept and the operation. The daemon log is the truth of the system, and the truth is monotonous. Every three minutes, the same sequence: check Boot, check Deacon, check Witness, check Refinery. "Warning: no wisp config for hermes." "Warning: no wisp config for hermes." Again. Again. 97 times. The log reads like a Philip Glass composition if Philip Glass were scoring bureaucracy rather than transcendence -- the same motif repeating, but without development, without variation, without movement toward resolution. Each heartbeat is identical to the last. Nothing changes. Nothing resolves.

The convoy watcher doubles every event. Every close is detected twice. This is not just a bug; it is an aesthetic failure. Redundancy without purpose is noise. A good designer would hear this doubled signal and be troubled -- not because it breaks anything, but because it reveals that the system does not *know itself* well enough to avoid saying the same thing twice.

**What would a choreographer see?**

A choreographer walking into a Gas Town rehearsal would see something immediately recognizable and immediately troubling: a production where the stage managers outnumber the dancers.

In a well-choreographed piece, every body on stage has a purpose visible to the audience. The choreographer would see the Mayor standing center stage, arms crossed, occasionally pointing. The Deacon would be pacing the wings, checking a clipboard, tapping dancers on the shoulder. The Boot would be standing behind the Deacon, watching the Deacon watch the dancers. The Witness would be in the front row of the audience, filling out evaluation forms. The Refinery would be at the stage door, checking IDs as dancers exit.

And the Polecats -- the actual performers -- would be in the far upstage corners, briefly dancing before being ushered off by the Witness.

The choreographer would say: "Why are there five people managing one dancer? And why is the stage manager" -- pointing at the Deacon -- "going on break every six minutes?" The answer -- that the Deacon's context window fills up -- would sound to the choreographer like a dancer who gets winded after eight bars. The response should not be to hire another stage manager to watch the first stage manager; it should be to improve the dancer's conditioning or simplify the choreography.

A great choreographic principle: **the structure should be invisible from the audience's perspective.** The audience sees the dance, not the counts, not the blocking tape on the floor, not the stage manager's cues. In Gas Town, the structure is the performance. The 22:1 coordination-to-work ratio means the audience is watching a play about stage management, not a dance.

**What would a good choreography look like?** The Polecats would be the stars. Their motion -- claiming work, writing code, committing, pushing -- would be the visible action. The coordination would be felt in the *timing* between their entrances and exits, in the seamless way one finishes and another begins, in the fact that they never collide. The audience would sense the underlying structure the way they sense a time signature -- not by counting, but by feeling the pulse. The infrastructure would be like the theater itself: the stage, the wings, the rigging. Present, load-bearing, invisible.

### 2. Rhythm

**What is the rhythm of Gas Town?**

The daemon heartbeat is the metronome: every 3 minutes, a tick. The patrol cycles for Deacon, Witness, and Refinery run on 5-minute intervals. The Deacon goes stale and restarts approximately every 6 minutes. The Witness hands off every 8 minutes.

This creates a polyrhythm, but not a good one. In music, polyrhythm is the layering of different time signatures to create tension and release. The best polyrhythms (think West African drumming, or Steve Reich's "Music for 18 Musicians") create a sense of inevitability through layered patterns that interlock. The listener cannot predict the exact moment of convergence, but can feel it approaching.

Gas Town's polyrhythm has none of this. The 3-minute heartbeat and 5-minute patrols and 6-minute Deacon restarts do not interlock; they simply overlap chaotically. At minute 15, the heartbeat fires (tick #5), the Deacon has been restarted twice, the Witness has handed off once, and the Refinery may or may not have processed anything. At minute 30, the heartbeat fires (tick #10), the Deacon has been restarted four or five times, and the system's state is indistinguishable from minute 0. There is no phrase structure. No verse, no chorus, no bridge. Just the same measure repeating.

**The rhythm should be:** idle silence -- punctuated by bursts of coordinated activity when work arrives -- resolving into merge completion -- returning to silence. The rhythm of good work is not metronomic; it is *phrased*. There should be tension (work in progress), climax (merge attempt), and resolution (integration or rework). The current design imposes a constant pulse regardless of whether anything is happening. A quiet town hums at the same volume as a busy one. This is the temporal equivalent of a fluorescent light that can never be turned off.

**What would good rhythm sound like?** Silence when idle. A clear downbeat when work is assigned: the daemon detects a new task, spawns a worker, and the worker begins. A sustained middle section where the only sound is the worker's commits -- the rhythmic scratch of code being written. A crescendo as the worker pushes and the merge pipeline triggers. A clear resolution: merge success or failure, logged and done. Then silence again. The daemon as a resting musician: present, attentive, but not playing until the score calls for it.

### 3. Negative Space

**What does Gas Town leave out?**

In visual art, negative space -- the empty area around and between subjects -- is as compositionally important as the subject itself. Matisse's cutouts, the enso circle in Japanese calligraphy, the pause in music: these are defined by what is *not* there.

Gas Town's negative space is revealing:

**No user experience.** The system has 13 agent roles, 10 work types, 61 Go packages -- and no consideration of what it feels like to use. There is no dashboard. There is no visualization. The TUI exists but is subordinate to the CLI. The human interacts with Gas Town through terminal commands and log files. The *experience* of running a multi-agent coding team -- the satisfaction, the anxiety, the sense of control or its absence -- is entirely absent from the design. This is a system designed for agents, not for the humans who depend on them. The negative space is the human.

**No error aesthetics.** When something fails, the system logs it and retries. But there is no *grammar* of failure. Is a merge conflict a problem or a normal event? Is a Deacon restart a crisis or routine? The system treats all failure identically: detect, restart, log. A good design would differentiate failures the way a good dashboard differentiates severity -- through visual weight, color, urgency, escalation cadence. Right now, a catastrophic bug and a routine heartbeat generate the same log format. Everything is monochrome.

**No celebration of completion.** When a polecat finishes work, it sends POLECAT_DONE and is destroyed. "Done means gone." There is no record of craft, no appreciation of output, no moment of review. The work product vanishes into a merge queue and the worker vanishes into nothing. Compare this with how a theater company operates: at the end of a performance, there is applause, a curtain call, notes from the director. These rituals serve a function -- they create closure, enable reflection, and provide feedback. Gas Town's negative space around completion is the absence of any feedback loop that might improve future performance.

**No play.** The system is entirely serious. Every concept is functional. There is no room for experimentation, no sandbox mode, no "what if we tried this differently?" The naming is playful (Mad Max theme), but the architecture is grimly utilitarian. A creative director would say: where is the space for happy accidents? Where is the R&D lab? Where do agents get to try things that might not work?

### 4. Craft vs. Manufacturing

**Is agent work craft or manufacturing?**

This is the central aesthetic tension in Gas Town, and the system cannot decide.

On one hand, the Propulsion Principle, the "done means gone" lifecycle, the structured data approach, and the merge queue all treat agent work as *manufacturing*: standardized, repeatable, interchangeable. A polecat is a piston. Any polecat can do any task. The output is measured, tracked, attributed. Quality is enforced through gates. This is the factory floor.

On the other hand, coding is *craft*. Good code requires judgment, taste, context, understanding of how a change fits into the larger whole. A function that passes tests but is poorly structured, a feature that works but is implemented in a way that makes future changes harder, an API that is correct but confusing -- these are failures that no automated gate catches. They require the eye of a craftsperson.

**The Arts and Crafts parallel is exact.** William Morris and John Ruskin argued that factory production destroyed meaning in two ways: first, by separating the worker from the whole (the factory worker makes a wheel, not a carriage; the polecat implements a task, not a feature); second, by optimizing for throughput over quality (the factory measures output per hour; Gas Town measures beads completed per token spent).

The bead abstraction -- reducing complex coding work to "atomic issue/task in JSONL" -- is the coding equivalent of piecework. It enables measurement and tracking, but at the cost of decomposing an inherently holistic activity (software design) into fragments. A craftsperson building a table considers the whole table -- the proportions, the materials, the joinery, the finish -- at every moment. A factory worker cutting table legs considers only the leg. Gas Town's polecats are leg-cutters. They implement tasks. They do not design systems.

**The tension is not resolvable, but it can be managed.** The answer is not "all craft" (no structure, no coordination, no measurement) or "all manufacturing" (pure piecework, interchangeable agents, beads in and beads out). The answer is a system that provides manufacturing's structure (clear tasks, quality gates, merge coordination) while preserving craft's conditions (enough context to understand the whole, time to make good judgments, feedback that distinguishes quality from mere correctness). Currently, Gas Town is tilted heavily toward manufacturing. The worker's context is consumed by lifecycle management rather than codebase understanding. The 75-second startup overhead before a single line of code is written is the equivalent of a craftsperson spending the first 20 minutes of their workday filling out timesheets.

### 5. The Role of the Director

**Is the Mayor a good director?**

A good director has four qualities:

1. **Clear vision.** The director knows what the production should be. They can articulate it. Every decision serves the vision.
2. **Trust in performers.** The director hires well and then gets out of the way. They intervene only when necessary.
3. **Economy of direction.** Every note is essential. The director does not give notes that change nothing. They do not repeat themselves.
4. **Taste.** The director can distinguish good from good enough from not good enough. They know when to push and when to accept.

The Mayor has *ambition* rather than vision. It has a system prompt that tells it to coordinate, decompose, and distribute. But vision requires understanding the destination. The Mayor knows the *process* (decompose goals into convoys, assign polecats, monitor progress) but not the *product* (what does good hermes code look like? what does the finished app need to be?). It is a project manager, not a creative director. There is nothing wrong with project management, but calling it "Mayor" implies governance and direction rather than logistics.

**Over-directing in Gas Town** looks like this: the Witness monitors polecats, the Deacon patrols rigs, the Mayor receives escalations, and the Boot watches the Deacon watch the Witness watch the polecats. Four layers of supervision for a single worker. This is the directorial equivalent of giving notes on every line reading, adjusting every blocking mark, second-guessing every choice. The performer has no room to perform. Their cognitive space is occupied not by their work but by the weight of observation.

**Under-directing** would be the opposite extreme: throw tasks at agents with no coherent sequence, no architectural vision, no quality standards beyond "does it compile?" The Propulsion Principle without a director's vision is just motion without direction -- efficient, fast, and possibly building the wrong thing.

**What Gas Town needs is neither more nor less direction, but better-timed direction.** A great director gives intensive notes in early rehearsals and then progressively withdraws. By opening night, the director is in the back of the house, watching. Their work was front-loaded: setting the vision, establishing standards, training the performers. The ongoing supervision is minimal because the preparation was thorough. Gas Town inverts this: the preparation is thin (a system prompt and a bead description), but the ongoing supervision is constant.

### 6. Elegance

**Where is Gas Town elegant?**

Elegance in design means maximum effect from minimum means. It means the right constraint at the right level, producing emergent order rather than enforced compliance.

**Elegant:**
- **Git worktree isolation.** One decision -- give each agent its own worktree -- eliminates an entire category of problems (file conflicts, state corruption, merge chaos during parallel work). This is a *generative constraint*: it produces safety without requiring ongoing enforcement. It is like the rule in a fugue that each voice must be independent -- a single principle that generates infinite variety within a safe structure.
- **Hook-based assignment.** "Your hook has work on it? Execute." Two concepts (hook + propulsion) replace what might otherwise require a ticket system, a notification pipeline, a scheduling framework, and an acknowledgment protocol. Compressed, powerful, clear.
- **Attribution via git author.** Instead of building a separate attribution system, embed identity in the existing version control metadata. Use what is already there. Do not invent when you can reuse. This is elegance as restraint.

**Inelegant:**
- **The three-tier watchdog chain.** Daemon watches Boot watches Deacon watches Witness watches polecats. Each tier exists because the previous tier might fail. But each tier also fails. The response to unreliability is more unreliability. This is the opposite of elegance -- it is accretion. An elegant solution to "agents might die" is: check if they are alive, and if not, restart them. One operation, one tier. The current design is a Rube Goldberg machine for process management.
- **Ten work unit types.** Bead, Convoy, Hook, Molecule, Protomolecule, Wisp, Formula, Digest, Epic, CV chain. Each has a rationale; taken together, they form a taxonomy that is more complex than the domain it models. Elegant data modeling means finding the *smallest* set of abstractions that covers all cases. One work type with a status field and a parent pointer can represent tasks, subtasks, epics, and workflows. The proliferation of types is not richness; it is entropy.
- **The naming system.** Elegance in naming means the name teaches you what the thing does. "Witness" implies observation; the Witness actively intervenes. "Refinery" implies transformation of crude input into refined output; the Refinery runs `git merge`. "Convoy" implies synchronized travel; convoy members work independently. When names mislead, they create negative elegance -- they make the system harder to understand than unnamed abstractions would be. Ironic naming is acceptable in poetry; in infrastructure, it is a maintenance cost.

### 7. The Unfinished Work

**Is Gas Town an unfinished masterpiece or an abandoned sketch?**

Gas Town exists in the state of Gaudi's Sagrada Familia circa 1926 -- after the architect's death, with the nave barely started, the towers incomplete, and the plans existing partly as drawings, partly as plaster models, and partly in the minds of collaborators who did not fully agree on the vision.

Consider what is built: a daemon that heartbeats, agents that spawn, a mail protocol, a bead system, worktree management, tmux orchestration. One rig. One project. A handful of completed tasks. The system runs. It runs imperfectly, but it runs.

Consider what is sketched but not built: Federation, HOP (cross-workspace coordination), capability-based routing from agent CVs, A/B testing between models, Formula templates, the Mol Mall marketplace, multi-rig orchestration at scale. These are towers of the Sagrada Familia -- visible in the plans, implied by the foundations, but years of construction away.

There is a critical difference between the Sagrada Familia and Gas Town: the Sagrada Familia's completed portions (the crypt, the Nativity facade) are themselves masterworks. Each finished section justifies the project's existence independent of the grand vision. Gas Town's completed portions are infrastructure -- they have no value independent of the whole. You cannot appreciate a heartbeat loop the way you appreciate a stone facade. The daemon's 97 heartbeat cycles are not beautiful in isolation; they are valuable only if they eventually support a system that produces meaningful work at scale.

**The question is whether Gas Town is incomplete like the Sagrada Familia (the vision is sound, execution continues) or incomplete like Kafka's The Castle (the incompleteness is the point, because the protagonist can never reach the Castle, and the system can never be fully built).** The evidence is ambiguous. The core data model insight is sound. The ambition is genuine. But the gap between the current state (one rig, supervision overhead exceeding productive work) and the envisioned state (20 rigs, capability routing, federation) is not a construction schedule -- it is a category change. And the synthesis document's proposed simplification (from 61 packages to 15, from 13 roles to 4) suggests that the vision itself may need to be reimagined, not merely continued.

Gas Town may be most honestly understood as a *study* -- like a painter's study for a larger work. The compositional ideas are visible. Some passages are fully realized. The proportions are experimental. The final work, if it comes, will look quite different from the study. And the study has value in itself -- not as a finished work, but as a record of thought, a space where ideas were tested and evaluated.

---

## Part II: The Constitutional Law Lens

### 8. Governance Architecture: Mapping Gas Town to a Constitution

Every constitution answers four questions: Who has power? What limits exist on that power? How does power transfer? How are disputes resolved?

**The Executive Branch: The Mayor**

The Mayor is the executive. It has the power to decompose work (legislation by decree -- turning a vague human directive into specific tasks), distribute assignments (patronage -- choosing which agent gets which work), and handle escalation (emergency powers -- making decisions when the normal process fails).

But this is a weak executive, not a strong one. The Mayor does not *command* agents; it creates convoys and assigns beads. The agents operate under the Propulsion Principle, which is effectively a standing order: "execute whatever is on your hook." The Mayor's power is front-loaded -- it shapes the work, but once the work is assigned, the Mayor's influence drops to zero until something goes wrong. This is closer to a constitutional monarchy than a presidency: the Mayor reigns but does not rule. Day-to-day governance is handled by the Deacon and Witness (the civil service) while the Mayor handles ceremony (convoys) and crisis (escalation).

The problem: the executive has no legislative agenda. The Mayor has no way to say "the priority has changed" or "stop working on that" or "this approach is wrong." Once work is distributed, it flows through the system like water through pipes. The Mayor cannot redirect it. This is a constitution that defines the power to *start* things but not the power to *stop* or *change* things. In constitutional terms, there is no veto power.

**The Judiciary: The Merge Queue**

The Refinery, operating the merge queue, is the judiciary. It renders judgment on whether work product is acceptable for integration into the canonical codebase (the "law of the land"). Like a court, it does not initiate action -- it responds to submissions. Like a court, its decisions are consequential: MERGED (acquittal -- the work is accepted), MERGE_FAILED (conviction -- the work is rejected with cause), or REWORK_REQUEST (remand -- sent back for correction with instructions).

**Is the judiciary fair?** Fairness requires three things: clear standards, consistent application, and the ability to appeal.

- *Clear standards*: The merge criteria are implicit. The Refinery checks for clean git state and runs merges. But what makes a merge "ready"? The standards are embedded in the process (does `git merge` succeed? do tests pass?) rather than articulated as principles. A court that says "you will know the law when we apply it" is not a fair court.
- *Consistent application*: Because the Refinery is an AI agent, its behavior is nondeterministic. The same merge might be handled differently in different sessions. A human court has precedent; the Refinery has no memory across sessions. Every case is a case of first impression.
- *Appeal*: There is no appellate process. MERGE_FAILED is final (the polecat is instructed to rework). REWORK_REQUEST is a directed verdict (rebase and try again). There is no mechanism for the polecat to say "the merge criteria are wrong" or "this test failure is a false positive" or "the conflict was introduced by another agent's flawed work." The judiciary is absolute. There is no Supreme Court to which one can appeal a Refinery decision.

**The Legislature: Who Makes the Rules?**

In Gas Town, rules come from three sources:

1. **System prompts** -- the equivalent of a constitution. They define roles, powers, and procedures. They are written by the human designer and injected at session start.
2. **Go code** -- the equivalent of statutory law. The daemon's behavior, the mail protocol, the heartbeat intervals, the staleness thresholds -- all codified in compiled code.
3. **Configuration files** -- the equivalent of regulations. `daemon.json`, `escalation.json`, `config.json` -- these parameterize behavior within the limits set by code.

This is a legislature of one: the human designer. There is no legislative process. Agents cannot propose rule changes. They cannot vote. They cannot even comment on the rules. The "legislature" is an absolute monarchy. The rules are handed down from above, and the agents' only recourse is compliance or failure.

Is this appropriate? For an AI orchestration system, perhaps. But consider: in a system designed for scale (20 rigs, 100+ agents), the inability of the system to adapt its own rules creates rigidity. A constitution that cannot be amended becomes irrelevant when circumstances change. Gas Town's rules are frozen in code and config. The living constitution question -- should rules evolve with practice? -- is answered firmly on the originalist side: the rules are the designer's original intent, and they do not change unless the designer changes them.

### 9. Due Process

**When a polecat is killed for being "stale," is there due process?**

Due process, at minimum, requires: notice (you are told what you are accused of), hearing (you have the opportunity to respond), and proportionality (the punishment fits the offense).

In Gas Town, a polecat is deemed "stale" when its heartbeat timestamp exceeds a threshold. The Witness detects this, and the polecat's bead is recovered for reassignment. There is:

- *No notice.* The polecat does not know it has been flagged. It may still be working -- perhaps on a long-running operation that does not update the heartbeat (a build, a large test suite, a complex reasoning chain). It receives no warning that its time is running out.
- *No hearing.* The polecat cannot explain why it appears stale. It cannot say "I am working on something that takes longer than your threshold" or "my heartbeat mechanism failed but I am alive." The decision is made unilaterally by the Witness based on a single metric.
- *Questionable proportionality.* The response to staleness is termination and work reassignment. But staleness is not necessarily failure. An agent deep in a complex reasoning chain is indistinguishable from a dead agent -- both have stale heartbeats. Killing the former to protect against the latter is a due process violation: punishing productive work because it resembles failure.

**The habeas corpus problem.** Can an agent be stuck indefinitely? Yes. Consider a polecat that receives a bead with insufficient information to complete the task. It cannot complete the work. It may not know how to escalate (the HELP mechanism requires the agent to recognize it is stuck, which requires a sophistication that an agent in a confused state may not have). If it writes code that is wrong and submits it, the Refinery rejects it and sends a REWORK_REQUEST. The polecat reworks and resubmits. The Refinery rejects again. This cycle can repeat indefinitely -- a Kafkaesque loop where the agent is neither freed (task completed) nor released (task reassigned) but perpetually detained in a rework cycle.

The maximum re-escalation limit (2, per `escalation.json`) provides some protection, but only for explicit escalation paths. The rework loop between polecat and Refinery has no circuit breaker. This is indefinite detention without judicial review.

**When work is rejected at the merge queue:** The MERGE_FAILED message tells the polecat what happened (test failure, merge failure) but the information's quality depends on the Refinery's ability to communicate. The polecat's "right" to understand why its work was rejected is technically supported by the protocol but practically undermined by the AI's nondeterministic communication quality. Sometimes the error is clear. Sometimes it is not. There is no standardized format for rejection reasons, no structured error taxonomy, no checklist of what the polecat should fix. The judgment is rendered in natural language, which may be precise or vague depending on the session.

### 10. Rights Framework

This sounds absurd -- do *agents* have *rights*? -- but constitutional thinking applied to system design is powerful precisely because it forces you to consider every component's needs and guarantees.

**Right to clear instructions.**

A citizen has the right to know the law. An agent has the right to know its task. Does Gas Town provide this?

Partially. The bead system provides structured task descriptions. But the quality of the description depends on the Mayor's decomposition, which is AI-generated and variable. A polecat might receive a bead that says "Implement authentication endpoint" -- clear enough -- or one that says "Fix the thing from the last convoy" -- opaque, dependent on context the polecat may not have. The PRIME.md and CLAUDE.md files provide role instructions, but they describe process (how to use `gt` commands), not substance (what good code looks like for this project, what architectural patterns to follow, what constraints exist).

The right to clear instructions is *constitutional* (system prompts define the protocol) but not *statutory* (no enforceable minimum standard for task description quality).

**Right to necessary resources.**

Does every agent have what it needs to succeed? The answer is structured no. Polecats receive a git worktree, a bead, and a system prompt. They do *not* reliably receive: architectural context about the project, information about concurrent work by other polecats, knowledge of previous decisions that constrain current choices, or understanding of the human's intent beyond the bead description. They work in isolation -- which is correct for file-level safety but corrosive for design-level coherence.

This is like giving a citizen the right to a trial but not the right to a lawyer. The formal mechanism exists; the substantive support does not.

**Right to fair evaluation.**

The merge queue evaluates work on a single axis: does it merge cleanly and pass tests? This is a narrow definition of "fair." Code that is technically correct but architecturally wrong will pass. Code that is brilliant but conflicts with concurrent work will fail. The evaluation criteria privilege mechanical correctness over design quality. This is like a justice system that can determine whether a law was technically violated but cannot assess mitigating circumstances, context, or intent.

**Right to appeal.**

Non-existent. As discussed above, there is no appellate mechanism. The Refinery's judgment is final within its tier. Escalation to the Mayor exists for emergencies, but there is no structured process for a polecat to say "this evaluation is wrong." The HELP message is a plea for assistance, not a formal appeal.

**Protection from arbitrary termination.**

Staleness-based killing, as discussed under due process, is arbitrary by the standard of any rights framework. The conditions for termination (heartbeat exceeds threshold) are well-defined but overbroad. A more rights-respecting design would provide: warning before termination ("your heartbeat is stale; are you alive?"), grace periods for long operations, and differentiation between "process dead" and "process alive but busy."

### 11. Federalism

**The Town-Rig structure is federal.**

The Town level holds: Mayor (executive), Deacon (civil service), Boot (infrastructure oversight), and town-level beads (`hq-*`) for cross-rig coordination. The Rig level holds: Witness (local supervisor), Refinery (local judiciary), Polecats (local workers), and rig-level beads for implementation work.

This maps closely to a federal system where the central government handles coordination, defense, and inter-state relations while states handle local governance, law enforcement, and implementation. The question is whether the balance is right.

**The subsidiarity argument:** Subsidiarity -- the principle that decisions should be made at the lowest level capable of making them -- suggests that most of Gas Town's current Town-level functions should be pushed down to the Rig level. Why does the Mayor decompose work for a specific rig? The Rig knows its own codebase. Why does the Deacon patrol all rigs? Each Rig's Witness knows its own agents. The Town level should handle only what requires cross-rig perspective: dependency management between rigs, resource allocation when multiple rigs compete for agents, and escalation that no single rig can resolve.

Currently, the balance is inverted. The Town level (Mayor, Deacon) is *more* active than the Rig level (Witness, Refinery), even though most work happens at the Rig level. This is like a federal government that manages local school districts, inspects local restaurants, and directs local traffic -- technically within its power, but violating the spirit of subsidiarity and creating overhead that local governance would handle more efficiently.

**The "states' rights" argument for rig autonomy:** Each rig wraps a distinct git repository with its own codebase, its own conventions, its own testing infrastructure. A rig should be largely self-governing. It should have the autonomy to decide: how many polecats it needs, how to evaluate merge readiness (which tests to run, which quality gates to apply), and how to handle local failures. The Town level should set policy (budget limits, escalation thresholds), not manage operations.

The current design reverses this. The daemon's patrol config (`daemon.json`) sets intervals and enables/disables patrols centrally. The Witness has local knowledge but limited autonomy -- it reports to the Deacon, which reports to the Mayor. Decisions flow up for approval rather than being made locally. This is centralism dressed as federalism.

### 12. Amendment and Evolution

**How does Gas Town's "constitution" change?**

It does not, autonomously. System prompts are files. Go code requires recompilation. Configuration files require manual editing. There is no mechanism within the running system to propose, debate, or enact rule changes. Every amendment requires the designer to stop the system, modify code or configuration, and restart.

This is the equivalent of a constitution that can only be amended by its original drafter. It works when the drafter is available and attentive. It fails when the drafter is absent (agents running overnight), when the drafter's assumptions are wrong (the heartbeat interval is too short for certain workloads), or when the drafter cannot anticipate future conditions (a new rig with different needs than hermes).

**A "living constitution" approach** would mean the system evolves its own rules based on experience. If polecats consistently hit the staleness threshold during build operations, the threshold should automatically increase for tasks tagged with "build." If certain task descriptions consistently produce rework cycles, the decomposition rules should tighten. This is what machine learning systems do: adapt parameters based on outcomes. Gas Town does not do this. Its rules are static.

**An "originalist" approach** -- rules reflect the designer's original intent -- is what Gas Town currently implements. The GUPP principle, the "done means gone" lifecycle, the three-tier watchdog chain, the 13 agent roles -- all are decisions frozen at design time. Whether the designer intended this rigidity is unclear; the system may simply not have reached the point where self-adaptation was implemented. But the *effect* is originalist: the system operates according to its founding principles regardless of whether those principles serve the current reality.

**The US Constitution's difficulty of amendment is a feature, not a bug -- stability against hasty change.** But Gas Town is not a nation; it is a software system. The cost of a bad amendment (a broken build) is low and recoverable (roll back the commit). The cost of inability to amend (persistent operational dysfunction) is high (the Deacon restart loop continued for 97 heartbeats -- approximately 5 hours -- without self-correction). Gas Town should be easier to amend than a constitution, not harder.

---

## Part III: The Synthesis -- Structure and Freedom

### 13. The Fundamental Question

Both art and constitutional law grapple with the same tension: **How do you create structure that enables freedom rather than constraining it?**

A great constitution creates a framework within which citizens can freely pursue their own goals. It constrains *power*, not *people*. The Bill of Rights does not tell citizens what to do; it tells the government what it cannot do.

A great choreography creates a structure within which dancers can express themselves. The counts, the formations, the spatial patterns -- these are constraints. But within those constraints, each dancer brings their own quality of movement, their own interpretation, their own presence. The constraints make the art *possible* because they solve the coordination problem, freeing the dancer to focus on expression.

Gas Town aspires to this -- a framework that coordinates autonomous agents. But it inverts the relationship.

**Gas Town constrains the agents rather than constraining the infrastructure.**

The system prompts tell agents what to do at every step. Check your hook. Run the work. Call `gt done`. Follow the session close protocol. The agents are choreographed to the smallest gesture. Meanwhile, the infrastructure is unconstrained: the daemon heartbeats forever, the Deacon restarts without limit, the convoy watcher fires duplicate events without correction, the "no wisp config" warning repeats 194 times without suppression.

A constitutional reframing would invert this: constrain the *infrastructure* (the daemon must not heartbeat when idle, the Deacon must not restart more than twice before escalating, the convoy watcher must not emit duplicate events) and *free the agents* (the worker receives a task and a worktree; how it approaches the code is its own decision).

### 14. Gas Town as Constitution

What would Gas Town look like if designed as a constitution?

**Preamble:** We, the designers of this system, in order to coordinate parallel coding agents, establish fair work distribution, ensure productive merge integration, provide for failure recovery, and secure the benefits of structured attribution, do ordain and establish this framework.

**Article I: The Legislature (Rule-Making)**
- Section 1: Rules are encoded as configuration, not as code. They are human-readable, version-controlled, and auditable.
- Section 2: Rules may be amended by the human overseer at any time. Amendments take effect at the next daemon cycle without restart.
- Section 3: No rule shall be enforced that has not been published in the configuration directory. (Transparency requirement: no hidden constraints in compiled code that agents cannot inspect.)
- Section 4: The system may propose rule amendments (e.g., "staleness threshold too short for build tasks") based on operational data. The human overseer ratifies or rejects.

**Article II: The Executive (Coordination)**
- Section 1: The Coordinator has the power to decompose work and distribute tasks.
- Section 2: The Coordinator shall not be a persistent agent. It shall be invoked on demand and dissolved upon completion of its directive.
- Section 3: The Coordinator's power is limited to task creation and assignment. It may not interfere with a worker's execution, modify a worker's output, or override the merge queue.
- Section 4: In emergency (systemic failure, budget exhaustion), the daemon may invoke the Coordinator for crisis management. This power shall not be exercised more than once per hour absent human authorization.

**Article III: The Judiciary (Evaluation)**
- Section 1: The merge queue shall evaluate work according to published criteria: clean merge, passing tests, and any rig-specific quality gates.
- Section 2: Evaluation criteria shall be enumerated in the rig's configuration file. No work shall be rejected on criteria not published in advance.
- Section 3: A worker whose submission is rejected shall receive a structured explanation: which criterion failed, what the expected behavior was, and what the actual behavior was.
- Section 4: A worker may appeal a rejection by flagging the task for human review. The human review is final.

**Article IV: Federalism (Town and Rig)**
- Section 1: Powers not explicitly reserved to the Town belong to the Rig.
- Section 2: Town powers: cross-rig dependency management, global budget enforcement, human escalation.
- Section 3: Rig powers: worker spawning, merge evaluation, local quality gates, local failure recovery.
- Section 4: No Town-level agent shall directly manage Rig-level workers. Coordination between levels flows through task records, not commands.

**Bill of Rights:**
- *First*: Every agent shall receive a task description sufficient to complete the work without external clarification. Insufficient task descriptions are a failure of the Coordinator, not the worker.
- *Second*: Every worker shall have access to the full codebase context within its worktree, including relevant documentation, test infrastructure, and architectural guidance documents.
- *Third*: No agent shall be terminated without warning. A staleness detection shall trigger a liveness query before termination. An agent that responds to the query shall be granted an extension.
- *Fourth*: No agent shall be trapped in a rework loop exceeding three iterations. After three rejections, the task shall be escalated to the Coordinator or flagged for human review.
- *Fifth*: Every agent shall have the right to escalate. The HELP mechanism shall be available at all times, and the response time for escalation shall not exceed the daemon heartbeat interval.
- *Sixth*: Agent evaluations shall be based on published, objective criteria. No agent shall be penalized for factors outside its control (API latency, model degradation, insufficient task description).

### 15. Gas Town as Choreography

What would Gas Town look like if designed as a choreographic score?

A dance score specifies structure (formations, counts, spatial relationships) while leaving room for individual expression (quality of movement, interpretive choices, personal style). The score coordinates without dictating.

**The Score:**

*Movement I: Stillness (Idle State)*
The stage is empty. The daemon breathes -- a slow, inaudible pulse. No light. No motion. The system costs nothing. It waits.

*Movement II: Intention (Work Arrives)*
A human speaks. The Coordinator enters, alone, in a single pool of light. It reads the intention, divides it into parts, places each part on the stage as a marked position. It exits. The parts glow softly: available work.

*Movement III: Emergence (Workers Claim)*
Workers enter from the wings. Each approaches a glowing mark. When a worker touches a mark, the mark transfers its light to the worker. The worker moves to its own area of the stage -- its worktree. The marks dim as they are claimed. When all marks are dark, no new workers enter.

*Movement IV: The Work (Parallel Execution)*
Each worker dances alone in its area. The choreography within each area is improvised -- the worker decides how to approach the code, what to write, when to test. The constraint is spatial: each worker stays in its worktree. The freedom is expressive: how they work is their own.

This is the heart of the piece. The audience sees three, four, five dancers working simultaneously, each in their own idiom, each pursuing their own task. The beauty is in the parallelism -- the sense that multiple creative acts are happening at once, coordinated not by a conductor but by the spatial structure of the stage itself.

*Movement V: Convergence (Merge)*
One by one, workers complete their phrases and move to the center of the stage -- the merge point. They arrive in sequence, not simultaneously. Each deposits their work. The merge happens mechanically: the work either fits into the whole (clean merge -- the dancer exits cleanly) or conflicts with another's work (merge conflict -- the dancer pauses, adjusts, and tries again). A Resolver may enter briefly to mediate a conflict, then exits.

*Movement VI: Resolution*
The center of the stage now contains the integrated work. The marks have all been claimed, the workers have all exited, the merges are complete. A final moment of stillness. Then the stage returns to Movement I. The cycle can repeat.

**What this score teaches us:**

1. **The idle state is not a failure; it is a structural element.** The pause between movements is part of the piece. A system that costs nothing when idle is like a dancer who can hold stillness -- it demonstrates control, not absence.

2. **The Coordinator's role is compositional, not supervisory.** It enters once, creates the structure, and exits. It does not hover. It does not watch. Its work is in the arrangement, not the oversight.

3. **Worker autonomy is the artistic content.** The interesting part of the system is the work itself -- the code being written. Everything else (coordination, merging, cleanup) is stagecraft. Stagecraft should be excellent and invisible.

4. **The merge is a group phrase, not a judgment.** In the score, workers converge voluntarily, and the merge is a collaborative act. The current Refinery framing (judiciary, judgment, accept/reject) could be replaced with a more collaborative framing: workers' contributions are *composed* into a whole, with conflicts resolved through adjustment rather than rejection.

5. **The score has silence.** Between movements, nothing happens. This silence is earned and intentional. Gas Town's current design has no silence -- the daemon pulses constantly, supervisors patrol endlessly. A score that never rests exhausts its performers and its audience.

### 16. Design Principles from Both Lenses

Taking both the constitutional and choreographic perspectives seriously, the following principles emerge for multi-agent coordination:

**Principle 1: Constrain power, not agents.**
*Constitutional:* The Bill of Rights constrains the government, not the people.
*Choreographic:* The score constrains the structure, not the movement.
*Applied:* The infrastructure should be tightly constrained (budget limits, heartbeat intervals, merge criteria, escalation thresholds). The agents should be loosely constrained (here is your task, here is your workspace, here are the quality standards -- how you work is your decision).

**Principle 2: Silence is a feature.**
*Constitutional:* The government that governs least governs best. Powers not exercised are not wasted; they are held in reserve.
*Choreographic:* The rest is part of the music. Stillness on stage is a choice, not a failure.
*Applied:* Idle systems should cost nothing. Supervision should activate only when triggered, not patrol continuously. The default state is quiet readiness, not busy vigilance.

**Principle 3: Due process before termination.**
*Constitutional:* No person shall be deprived of life, liberty, or property without due process of law.
*Choreographic:* A director does not pull a dancer off stage mid-performance without cause and communication.
*Applied:* Before killing an agent, query it. Before rejecting work, explain why with specificity. Before reassigning a task, confirm the original agent has actually failed rather than merely being slow.

**Principle 4: Subsidiarity -- decide at the lowest capable level.**
*Constitutional:* Federalism. Powers not delegated to the central government are reserved to the states.
*Choreographic:* The dancer interprets the score. The choreographer does not control the dancer's muscles.
*Applied:* Rigs should self-govern. Workers should self-direct within their tasks. The Town level should handle only what requires cross-rig perspective. Every decision made at a higher level than necessary is overhead.

**Principle 5: The work is the art; the coordination is the frame.**
*Constitutional:* Government exists to serve the people, not the reverse.
*Choreographic:* The audience comes to see the dance, not the stage management.
*Applied:* Maximize the percentage of tokens, time, and context devoted to actual coding. Minimize the percentage devoted to coordination, supervision, and lifecycle management. The 22:1 coordination-to-work ratio should be inverted: 1:22 would be closer to correct.

**Principle 6: Published standards, not arbitrary judgment.**
*Constitutional:* Ex post facto laws are prohibited. Citizens must be able to know the law before they act.
*Choreographic:* The score is published before rehearsal. Dancers know the structure before they begin.
*Applied:* Merge criteria must be enumerated before work begins. Quality gates must be defined in configuration, not discovered at evaluation time. An agent should be able to predict whether its work will be accepted by checking it against the published criteria before submission.

**Principle 7: The right to appeal.**
*Constitutional:* Every court decision can be appealed. Even the Supreme Court can be overridden by constitutional amendment.
*Choreographic:* A dancer can question a direction. The rehearsal process includes dialogue, not just dictation.
*Applied:* Every rejection should be appealable -- either to a higher authority (Coordinator) or to the human overseer. Rework loops must have circuit breakers. No agent should be trapped in a cycle with no exit.

**Principle 8: Amendment is normal, not exceptional.**
*Constitutional:* A constitution that cannot be amended becomes a dead letter.
*Choreographic:* A score that cannot be revised in rehearsal produces a rigid, lifeless performance.
*Applied:* Gas Town's rules should be easily modifiable configuration, not compiled code. The system should propose its own amendments based on operational data ("the staleness threshold caused 14 false positives last week; recommend increasing to 10 minutes"). The human overseer ratifies.

**Principle 9: Elegance through generative constraint.**
*Constitutional:* The First Amendment -- sixteen words that generate an entire body of law. Maximum effect from minimum specification.
*Choreographic:* "Follow the weight." Two words that guide infinite movement choices.
*Applied:* Seek rules that *generate* correct behavior rather than rules that *enumerate* correct behavior. "If you have work, execute immediately" is generative. "Check your hook, then check your mail, then check bd ready, then wait" is enumerative. The former scales; the latter accumulates.

**Principle 10: Beauty matters.**
*Constitutional:* The eloquence of the Constitution is not incidental -- it is part of its authority. "We hold these truths to be self-evident" compels by its rhetoric, not just its logic.
*Choreographic:* Beauty is not a luxury; it is the purpose.
*Applied:* A well-designed system should be pleasurable to observe in operation. Clean logs, clear status displays, satisfying completion signals, elegant error messages. The aesthetic quality of a system's interface and output is evidence of (and contributor to) its functional quality. A system that produces ugly logs is probably not well-understood by its creator. A system that produces beautiful ones probably is.

---

## Coda: What Pure Engineering Cannot See

The engineering analyses of Gas Town (all seven of them) correctly identified the operational pathologies: supervision overhead, AI misallocation, concept bloat, polling waste. They proposed solutions: mechanize supervision, reduce roles, simplify data models, event-driven dispatch.

But engineering analysis, by its nature, asks "does it work?" and "how can it work better?" It does not ask:

- **"What is it for?"** -- the art question. What experience does this system create for its human? What does it feel like to run a town of coding agents? Is that feeling one of control, of creation, of collaboration? Or is it one of anxiety, of opacity, of managing managers? The aesthetic lens reveals that Gas Town's human experience is entirely undesigned. The system has no *felt quality* -- no sense of place, pace, or presence. Adding this is not polish; it is purpose.

- **"Is it just?"** -- the constitutional question. Does the system treat its components fairly? Not in a moral sense (agents are not moral patients), but in a *functional* sense: does every component receive what it needs to succeed? Is every evaluation based on clear criteria? Is every termination warranted? The constitutional lens reveals that Gas Town is an autocracy with ambitions of governance. It has hierarchy without accountability, judgment without appeal, and rules without amendment.

- **"Is it beautiful?"** -- the combined question. A system that works perfectly but is ugly -- whose logs are noise, whose architecture is accidental, whose concepts do not illuminate -- is a system that will not inspire its creator or its users to invest in its future. A system that is beautiful -- whose design reveals its purpose, whose operation has rhythm and clarity, whose constraints generate freedom -- is a system that people want to build, maintain, and use. Gas Town's core insight is beautiful. Its implementation, currently, is not. The path from here to there is not just engineering. It is direction, composition, and governance. It is art and law applied to code.

The deepest insight from this combined lens: **Gas Town is a system that has been over-engineered and under-designed.** It has too much machinery and not enough meaning. The fix is not more machinery, or less machinery, but the right machinery -- chosen with a director's eye for what serves the performance, a choreographer's sense of when to specify and when to let go, and a constitutional designer's commitment to structure that enables rather than constrains.
