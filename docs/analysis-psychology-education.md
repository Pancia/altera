# Psychology + Education Analysis: Gas Town Multi-Agent Orchestration

An interdisciplinary analysis applying Self-Determination Theory, Flow State research, Organizational Psychology, Cognitive Load Theory, Learning Theory, and Instructional Design to the Gas Town multi-agent system.

---

## Part I: The Psychology Lens

### 1. Self-Determination Theory and Agent Architecture

Deci and Ryan's Self-Determination Theory (SDT) identifies three innate psychological needs that, when satisfied, produce optimal functioning: **autonomy**, **competence**, and **relatedness**. The question of whether these apply to AI agents is not metaphysical -- it is empirical and structural. An LLM does not "feel" autonomous, but the architectural conditions that correspond to autonomy (goal-directed rather than step-directed work, latitude in approach selection, absence of micromanagement) measurably affect output quality. This is because LLMs produce better results when given goals and constraints rather than rigid procedural scripts. The structure of autonomy matters even when the experience of autonomy does not.

#### Autonomy: Goal-Direction vs. Step-Direction

SDT distinguishes between **autonomy-supportive** contexts (providing choice, rationale, and acknowledgment of perspective) and **controlling** contexts (imposing specific behaviors through surveillance, deadlines, and directives).

Gas Town's autonomy profile is deeply contradictory.

**Autonomy-supportive elements:**
- The Propulsion Principle ("if work is on your hook, YOU RUN IT") is genuinely autonomy-supportive in intention. It says: here is your task, now execute it your way, immediately, without waiting for permission. This is goal-direction, not step-direction.
- Worker agents receive task descriptions, not implementation scripts. They choose their own approach to the code.
- Git worktree isolation gives each agent a private workspace -- the spatial analog of autonomy.

**Autonomy-thwarting elements:**
- The three-tier supervision chain (Daemon -> Boot -> Deacon -> Witness) is the architectural equivalent of a manager looking over your shoulder every six minutes. The Witness monitors polecats. The Deacon monitors the Witness. Boot monitors the Deacon. The Daemon monitors Boot. This is not autonomy support -- it is panopticon design.
- Health-check interruptions inject surveillance messages into agent context windows. Every nudge and health ping is the system saying "Are you still working? Are you still alive? Report your status." In SDT terms, this is a controlling context that converts autonomous motivation into externally regulated behavior.
- The 111 nudges generated for 5 completed tasks (a 22:1 coordination-to-work ratio) means agents spend more cognitive budget processing surveillance than performing work. This is the structural signature of autonomy thwarting.

**The paradox:** GUPP grants autonomy in task execution while the supervision hierarchy takes it away in lifecycle management. The agent is told "do whatever you want with this code" and simultaneously told "but we will check on you every few minutes, and if you do not respond fast enough, we will kill you and start over." In human terms, this is the manager who says "I trust you completely" while installing keylogger software. The verbal message is autonomy; the structural message is control.

**What SDT predicts:** When controlling structures dominate autonomy-supportive intentions, performance degrades. The research is unambiguous on this point across hundreds of studies. For LLM agents, the degradation manifests differently (context window pollution rather than motivational decline), but the directional effect is the same: surveillance overhead reduces the cognitive resources available for the primary task.

**SDT-optimized design:** Give agents goals, not steps. Give them constraints, not surveillance. Replace continuous monitoring with outcome verification. Check the work product, not the worker. The synthesis document's recommended architecture (mechanical daemon checking heartbeats and progress markers, with workers dedicating 100% of context to code) is, structurally, the SDT-optimized design -- though it arrived at this conclusion through engineering analysis rather than psychological theory, which itself is evidence that SDT's principles describe genuine structural dynamics rather than merely subjective preferences.

#### Competence: Mastery and Effectiveness

SDT's competence need is satisfied when an agent experiences itself as effective -- capable of producing desired outcomes and growing in capability. Competence is supported by optimal challenges (neither too easy nor too hard), positive feedback on performance, and opportunities to develop skills.

Gas Town's competence profile is confused.

**Competence-supporting elements:**
- The CV system (tracking agent performance history) is, in principle, a competence feedback mechanism. It says: your past performance matters, your capabilities are recognized, and future work will be matched to your demonstrated strengths.
- Capability-based routing is the organizational equivalent of "stretch assignments" -- matching workers to tasks at the edge of their proven capability.
- Attribution on every action creates a feedback loop: this commit was yours, this merge was yours, this success was yours.

**Competence-undermining elements:**
- The Deacon death spiral is the definitive competence failure. Restarted 15+ times without strategy change, the Deacon demonstrates what happens when an agent is given a task (patrol and supervise) but denied the conditions for success (a context window large enough to hold the accumulated health-check messages). The system repeatedly places the Deacon in a situation where failure is structurally inevitable, then responds to each failure with the same setup. In human terms, this is the employee given an impossible workload who is fired and replaced with an identical employee who receives the identical impossible workload.
- The Blind Spot Finder identified **perverse selection oscillation**: agents that build good CVs on easy tasks are routed to harder tasks, perform worse, get routed back to easy tasks, and oscillate. This is the opposite of competence development -- it is a treadmill where apparent progress triggers conditions for apparent regression.
- Agents cannot actually develop competence. Each session is a fresh LLM with no memory of previous sessions. The CV system tracks competence that does not exist in the agent -- it exists only in the routing system's model of the agent. The agent "rated highly in Swift development" does not know Swift any better than any other instance of the same model. The competence is an attribution error: the system ascribes to the agent a property that belongs to the model.

**What SDT predicts:** When competence feedback is disconnected from actual capability development, it becomes meaningless (positive feedback that does not reflect genuine skill) or demoralizing (negative feedback for structural failures). Gas Town's CV system risks both: inflating competence assessments based on task difficulty variance, and penalizing agents for failures caused by system design (context exhaustion, supervision overhead) rather than agent capability.

**SDT-optimized design:** Since agents cannot develop competence across sessions, optimize for competence *within* each session. This means: clear task specifications (the agent knows exactly what success looks like), immediate feedback on intermediate outputs (tests pass, linter is clean), and tasks calibrated to the agent's actual capability (the model's demonstrated ability, not a fictional CV). The recommended architecture's "Worker Context Purity" -- where the agent's entire context is dedicated to the coding task -- maximizes within-session competence by eliminating extraneous demands.

#### Relatedness: Connection and Belonging

SDT's relatedness need is the most difficult to map onto AI agents, because it involves felt connection to others. LLMs do not form relationships. But the *structural* conditions of relatedness -- shared purpose, mutual dependency, communication about shared goals -- do affect multi-agent system behavior.

Gas Town's relatedness profile is almost entirely absent.

**Relatedness-adjacent elements:**
- The mail protocol creates agent-to-agent communication, but the messages are purely transactional (POLECAT_DONE, MERGE_READY, MERGED). There is no mechanism for agents to share context about *why* they made certain decisions, *what* they learned about the codebase, or *how* their work relates to other agents' work.
- The Convoy concept (a batch of beads assigned together) hints at shared purpose, but the synthesis found that "Convoy members do not move together" -- the metaphor of collective action is not implemented.

**What is missing:** In human teams, relatedness creates information sharing, mutual adjustment, and collective intelligence. Team members who feel connected share insights ("I noticed the auth module has a weird edge case"), warn about risks ("that API is flaky, use retries"), and coordinate implicitly ("I'll handle the frontend since you're doing the backend"). Gas Town has no mechanism for any of this. Each agent works in isolation, communicating only status updates through formal channels. The multi-agent system is not a team -- it is a set of independent contractors who happen to share a codebase.

**SDT-optimized design:** Create mechanisms for agents to leave contextual notes for other agents working on the same codebase. A shared knowledge base per rig -- "things discovered about this codebase" -- would allow agents to benefit from each other's exploration without requiring persistent memory. This is not relatedness in the psychological sense, but it is the structural condition that relatedness produces in human teams: accumulated shared knowledge that improves collective performance.

---

### 2. Flow State and Agent Coordination

Csikszentmihalyi's flow state requires three conditions: **clear goals**, **immediate feedback**, and **balance between challenge and skill**. Flow is disrupted by interruptions, ambiguity, and anxiety about external evaluation.

#### Do Gas Town's agents achieve flow?

Consider the phenomenology of a Gas Town polecat session from the agent's perspective (reconstructed from the architectural analysis):

1. Session begins. Read system prompt (lengthy: role identity, lifecycle protocols, naming conventions, communication protocols). **Not flow: cognitive loading.**
2. Check for hooked beads. Parse the hook assignment. Understand task context. **Approaching flow: goal clarification.**
3. Begin coding. Read relevant files, understand the codebase, make decisions, write code. **This is where flow would occur.**
4. Health check arrives (nudge or mail). Must acknowledge, update status, confirm liveness. **Flow interrupted.**
5. Resume coding. Re-establish context from where the interruption occurred. **Flow rebuilding.**
6. Another health check. **Flow interrupted again.**
7. Context window filling with health-check messages and lifecycle overhead. Performance degrading. **Anti-flow: increasing anxiety equivalent (degrading capability).**
8. Context pressure triggers handoff protocol. Must write structured handoff message, manage session transfer. **Flow destroyed: task switches from coding to lifecycle management.**

The pattern is clear: **steps 3 and 5 are the only flow-capable intervals, and they are bounded by interruptions on both sides.** The synthesis data quantifies this: the Witness hands off every 8 minutes due to context exhaustion from health-check traffic. If we assume 1-2 minutes for handoff processing on each end, this leaves 4-6 minutes of potential flow time per session. Csikszentmihalyi's research suggests flow requires approximately 15-25 minutes to fully establish. Gas Town's agents never reach flow.

The irony is precise: the supervision system designed to ensure agents are working productively is the primary mechanism preventing agents from working productively. This is not a novel finding -- it has been observed in every study of interruption costs in knowledge work (Mark, Gonzalez, & Harris, 2005; Leroy, 2009). The cost of an interruption is not the interruption itself but the re-establishment of context afterward. For humans, this averages 23 minutes. For LLMs, the cost is different in kind (context window consumption rather than time) but identical in effect (reduced capacity for the primary task).

#### Flow-Optimized Agent Coordination

A flow-optimized system would follow one principle: **leave agents alone while they are making progress.**

This means:
- **No health-check interruptions during active work.** Monitor progress through external signals (git commits, file modification timestamps) rather than internal messages. The agent should not know it is being monitored. This is the architectural equivalent of a manager who checks the team's output dashboard rather than walking by desks.
- **Batch all non-coding communication.** If the agent needs information (merge result, dependency update, task modification), queue it for delivery at natural breakpoints (between subtasks, at commit boundaries) rather than interrupting mid-thought.
- **Extend session duration.** The 8-minute Witness session is anti-flow by design. If supervision overhead is eliminated (as the recommended architecture proposes), agent sessions could run for 30-60 minutes -- long enough for deep engagement with the code.
- **Match task granularity to session capacity.** A task that requires 3+ sessions to complete will never achieve sustained flow because each session boundary resets context. Tasks should be sized to complete within a single session when possible. This is the flow equivalent of "right-sizing" -- ensuring the challenge matches the available skill window.

The recommended architecture's mechanical daemon already embodies flow-optimization by accident: it checks heartbeats and progress markers externally, without injecting messages into agent context. Workers "dedicate 100% of context to code." This is exactly what flow theory prescribes.

---

### 3. Organizational Psychology

#### Psychological Safety (Edmondson)

Amy Edmondson's research on psychological safety demonstrates that high-performing teams share a belief that the team is safe for interpersonal risk-taking. Members can admit mistakes, ask questions, and propose unconventional approaches without fear of punishment.

The AI analog is not "felt safety" but **structural tolerance for exploration and error.** Does Gas Town's architecture permit agents to try unconventional approaches, make mistakes, and recover? Or does it punish deviation?

**Evidence for low structural safety:**
- "Done means gone." A polecat that completes its work is immediately destroyed. There is no post-completion review where the agent could reflect on what worked. There is no mechanism for the agent to flag uncertainty ("this solution works but I'm not confident it's the best approach"). The structure says: produce output and disappear. In Edmondson's framework, this is the team where you are fired the moment you deliver your work -- no feedback, no discussion, no learning.
- The three-tier supervision chain communicates distrust. When a system devotes more resources to monitoring workers than to actual work (111 nudges for 5 tasks), the structural message is: "We expect you to fail, and we are watching for it." In human teams, this level of surveillance correlates with lower psychological safety and lower performance (Edmondson, 1999).
- The Deacon death spiral is structural punishment without adaptation. The system does not ask "why did the Deacon fail?" or "what conditions would allow the Deacon to succeed?" It simply restarts the same setup. In Edmondson's terms, this is the organization that responds to every failure with "try harder" rather than "what can we learn?"

**Evidence for moderate structural safety:**
- Git worktree isolation provides "safe experimentation space." An agent can try any approach without corrupting the main branch. This is the architectural equivalent of a sandbox where mistakes are recoverable.
- The merge queue provides a quality gate that catches errors before they reach production. This should, in principle, allow agents to be bolder in their approaches because the safety net will catch failures.

**The net assessment:** Gas Town's agents operate in a structurally unsafe environment at the lifecycle level (constant surveillance, immediate destruction, no tolerance for process exploration) but a structurally safe environment at the code level (isolated workspaces, merge gates). The safety exists where the architecture is mechanical (git); the unsafety exists where the architecture is social (supervision hierarchy).

#### Intrinsic vs. Extrinsic Motivation

Gas Town's agents are entirely extrinsically motivated. They work because they have hooks (assignments imposed externally). They follow the Propulsion Principle because it is a rule, not because they are drawn to the work. There is no mechanism for an agent to express interest in a task, choose work that aligns with its "strengths," or pursue curiosity about the codebase.

Is there an analog to intrinsic motivation for AI agents?

The answer is subtle. LLMs do not have desires, but they do have **response distributions that are affected by framing.** When a task is framed as interesting, challenging, or meaningful -- when the prompt activates the model's training distribution for "engaged expert" rather than "reluctant employee" -- the output quality changes. This is not intrinsic motivation in the psychological sense, but it is a structural analog: the conditions that produce engaged, high-quality work in humans (autonomy, meaningful challenge, connection to purpose) produce measurably different outputs in LLMs when encoded in the prompt and environment.

What would "intrinsic motivation" look like for AI agents?
- **Task framing that emphasizes the interesting aspects of the work** rather than compliance requirements. Compare "You must implement user authentication following these exact steps" (controlling) with "The application needs user authentication. Here is the codebase. Design and implement the best approach you can" (autonomy-supportive). The second framing activates a different response distribution.
- **Capability-matched assignment** that puts agents at the edge of demonstrated ability. Too easy is boring (low engagement response distribution); too hard is overwhelming (confusion/hallucination response distribution). The "flow channel" in Csikszentmihalyi's model applies structurally: challenge-skill balance affects output quality.
- **Purpose contextualization.** Telling an agent "this feature will allow users to X" rather than "implement ticket #1234" provides the kind of meaning-making that activates higher-quality responses. This is well-documented in prompt engineering: contextualizing the purpose of a task improves LLM output.

Gas Town does none of this. Tasks arrive as hooks with technical descriptions. There is no framing for engagement, no contextualization of purpose, no attempt to activate the model's "best work" distribution. The entire motivational architecture is extrinsic: do this because it is assigned to you.

#### Learned Helplessness (Seligman)

Martin Seligman's learned helplessness research demonstrates that when organisms experience repeated uncontrollable failure, they stop attempting to exert control even when control becomes possible. The mechanism is the generalization of the belief "nothing I do matters."

The Deacon death spiral is a system-level analog of learned helplessness -- but with a critical difference. The Deacon does not "learn" helplessness because it does not learn anything. Each Deacon session is a fresh instance with no memory of previous failures. The helplessness is not in the agent; it is in the **system design that repeats failure without adaptation.**

This is actually worse than learned helplessness. In Seligman's experiments, at least the organism's helplessness is an adaptive response to genuine uncontrollability -- a form of energy conservation. Gas Town's system does not even conserve energy. It expends maximum resources (spawning a new AI session, loading the full system prompt, initializing the patrol loop) to reproduce the exact conditions that caused the previous failure. The Deacon is restarted 15+ times over two hours, each time entering the same failure trajectory.

The parallel in organizational psychology is not learned helplessness but **organizational insanity** -- the popular definition of insanity as doing the same thing repeatedly while expecting different results. Gas Town's system lacks the feedback loop that would allow it to recognize the pattern and adapt. There is no mechanism for the daemon to notice "the Deacon has failed 15 times in the same way" and try a different approach (reducing health-check frequency, extending session timeouts, simplifying the Deacon's responsibilities).

**What organizational psychology prescribes:** After-action reviews. Root cause analysis. Adaptive response to repeated failure. In the recommended architecture, these translate to: escalation thresholds (after 3 failures, change strategy, not just retry), failure categorization (distinguish between agent failure and environmental failure), and adaptive parameters (if sessions keep dying at 8 minutes, extend the session budget rather than restarting at 8-minute intervals).

---

### 4. Cognitive Load Theory

John Sweller's Cognitive Load Theory distinguishes three types of cognitive load:

- **Intrinsic load:** the inherent complexity of the material being learned or the task being performed. For a coding task, this is the complexity of the code itself -- understanding the codebase, designing the solution, writing correct implementations.
- **Extraneous load:** complexity added by the instructional design or task environment that does not contribute to the goal. Poor UI design, confusing instructions, irrelevant information.
- **Germane load:** complexity that contributes to schema formation and learning. Connections between concepts, abstractions, pattern recognition.

For AI agents, "cognitive load" maps to context window consumption. Every token in the context window is a unit of processing capacity. Tokens spent on intrinsic load (understanding and writing code) are productive. Tokens spent on extraneous load (understanding role protocols, processing health checks, managing lifecycle) are waste. Germane load is more complex -- we return to this in Part II.

#### Load Analysis for a Gas Town Polecat

Let us estimate the context window budget allocation for a typical polecat session:

**Intrinsic load (the coding task):**
- Task description and acceptance criteria
- Relevant source files
- Test files and expected behaviors
- Design decisions and implementation
- *This is the productive work. This is what the agent is for.*

**Extraneous load (system overhead):**
- **System prompt:** Role identity (you are a Polecat named X), lifecycle protocols (how to handle hooks, handoffs, death), naming conventions (Town, Rig, Bead, Convoy, Molecule, Wisp, Formula, etc.), communication protocols (mail types, nudge format, escalation rules), attribution requirements, the Propulsion Principle explanation, identity layer management (Identity -> Sandbox -> Session). Estimated: 2,000-5,000 tokens depending on prompt detail.
- **Health-check messages:** Nudges, status requests, liveness pings. Based on the 22:1 coordination-to-work ratio, these dominate the context window over the session lifetime.
- **Mail protocol overhead:** Reading and writing structured messages in the correct format, addressing them to the correct recipients, using the correct message types.
- **Lifecycle management:** Understanding when to handoff, how to write handoff messages, how to manage the three identity layers, when to signal completion.
- **Naming system translation:** Mapping Mad Max metaphors to functional concepts. Every time the agent encounters "Convoy," "Refinery," "Witness," or "Seance," it must translate to the underlying function.

**Germane load (learning that improves performance):**
- In human learning, germane load is desirable -- it builds schemas that make future tasks easier. For LLM agents, germane load within a single session could include: understanding codebase patterns that help with the current task, recognizing architectural conventions that guide implementation decisions, building a mental model of the project that informs code quality.
- Gas Town does not distinguish germane load from extraneous load. The system prompt contains both useful context (project structure, coding conventions) and useless protocol (mail formats, lifecycle management). They are delivered in the same undifferentiated block.

#### The Load Ratio

Based on the synthesis findings:

- Workers currently spend approximately 80% of their responsibilities on lifecycle overhead (the Ford Audit found only 19% of responsibilities require AI reasoning across all roles, and for polecats specifically, only the code-writing function is core).
- The 22:1 coordination-to-work ratio suggests that during active sessions, supervision messages consume far more context than productive work.
- The 8-minute Witness session lifetime suggests context windows are being exhausted primarily by extraneous load, not intrinsic load.

**Conservative estimate:** Gas Town polecats allocate 60-70% of their effective context window to extraneous load, 25-35% to intrinsic load, and less than 5% to germane load.

This is a catastrophically inefficient ratio. Cognitive Load Theory research consistently shows that when extraneous load exceeds intrinsic load, performance degrades sharply. The recommended architecture's "Worker Context Purity" principle (100% code, 0% lifecycle management) is the CLT-optimized design: eliminate extraneous load entirely, maximize the context available for intrinsic load.

#### The Naming System as Extraneous Load

The naming system deserves special attention because it is a pure, unnecessary source of extraneous load.

Consider the cognitive operation required to process this sentence from Gas Town's architecture: "The Witness monitors Polecats in the Rig, and when a Polecat sends POLECAT_DONE mail about its hooked Bead, the Witness forwards MERGE_READY to the Refinery."

To understand this sentence, the agent must maintain the following mappings:
- Witness = per-rig health monitor
- Polecat = worker agent
- Rig = project container
- POLECAT_DONE = task completion signal
- Mail = async message
- Bead = work unit
- Hook = current assignment
- MERGE_READY = request for merge processing
- Refinery = merge queue processor

That is 9 metaphor-to-function translations for a single sentence describing a simple operation: "When a worker finishes its task, the monitor tells the merge queue to process it."

For an LLM, each metaphor translation consumes attention and context. The model must hold both the metaphor and the referent in working memory, perform the mapping, and then reason about the functional relationship. This is textbook extraneous load: complexity that does not contribute to the task (writing code) but consumes processing capacity.

The recommended architecture's plain naming (Worker, Coordinator, Resolver, Task, Message) eliminates this extraneous load entirely. A system prompt that says "You are a Worker. You write code. When done, update the task file" requires zero metaphor translation.

---

## Part II: The Education Lens

### 5. Can You Educate Something That Doesn't Learn?

This is Gas Town's central educational paradox. The system aspires to develop agent capability (CVs, capability routing, persistent identity) but agents have no long-term memory. Each session is born fresh. The CV is a record kept *about* the agent, not *by* the agent. Capability routing assigns tasks based on history the agent cannot access.

Education theory offers three frameworks for understanding this paradox:

#### Framework 1: Instruction Quality (Behaviorism / Direct Instruction)

The simplest educational response to agents that do not learn is: **make the instructions better.** If the student cannot retain knowledge between sessions, ensure that every session begins with perfectly designed instruction that brings the student to maximum competence immediately.

This is the "better prompts" approach. It treats the agent as a blank slate that must be fully programmed at the start of each session. The quality of the session depends entirely on the quality of the system prompt and task description.

Gas Town partially adopts this approach (detailed system prompts defining roles, protocols, and expectations) but executes it poorly (the prompts are loaded with extraneous protocol rather than task-relevant context). The CLT analysis above quantifies the problem: most of the "instruction" is about the system rather than about the work.

**What instruction quality theory prescribes:** Strip system prompts to the minimum viable instruction. For a worker: (1) what the task is, (2) what the codebase looks like, (3) what success looks like, (4) where to put the output. Everything else is noise. The recommended architecture's pre-configured worktree with a task description file is exactly this: minimum viable instruction delivered through the environment rather than through verbose prompting.

#### Framework 2: Situated Cognition (Lave & Wenger)

Jean Lave and Etienne Wenger's Situated Learning Theory argues that learning is not a cognitive process happening inside the learner's head but a social and environmental process happening through participation in communities of practice. Knowledge is not transferred from instructor to student; it emerges from the student's engagement with the environment.

This framework radically reframes the agent education problem. If competence comes from the *environment* rather than the *agent*, then the question is not "how do we make agents smarter?" but "how do we make the workspace smarter?"

Consider what a well-configured workspace provides:
- **Code conventions visible in existing code.** An agent that reads well-structured existing code will produce code in the same style, not because it "learned" the convention but because the environment demonstrates it.
- **Test patterns visible in existing tests.** An agent that sees how existing features are tested will test new features similarly.
- **Architecture visible in the directory structure.** An agent that sees `src/auth/`, `src/users/`, `src/payments/` will understand the modular architecture without being told.
- **Error patterns visible in linter output.** An agent that runs the linter and sees the kinds of errors it catches will avoid those errors.

The workspace *is* the curriculum. The environment *teaches* the agent what kind of code belongs here, how things are organized, what quality standards apply, and what patterns to follow. This is situated cognition: competence emerges from participation in the environment, not from instruction.

**Gas Town implication:** Instead of investing in CVs, capability routing, and persistent identity (which attempt to store knowledge *about* the agent externally), invest in workspace configuration (which delivers knowledge *to* the agent environmentally). A perfectly configured worktree with exemplar code, clear conventions, running tests, and a comprehensive linter teaches the agent everything it needs to know -- every session, from scratch, without requiring memory.

This is the insight that dissolves the education paradox. You do not need agents to learn if the environment teaches. The workspace is the teacher that never forgets.

#### Framework 3: Scaffolding and Zone of Proximal Development (Vygotsky)

Lev Vygotsky's Zone of Proximal Development (ZPD) is the space between what a learner can do independently and what they can do with guidance. Scaffolding is the temporary support provided by a more capable other that enables the learner to operate within their ZPD.

For AI agents, the ZPD has a precise analog: the space between what the model can do with a bare prompt and what it can do with a well-structured environment. The model's independent capability is what it produces given only "write user authentication." The model's scaffolded capability is what it produces given a pre-configured worktree with existing auth patterns, relevant test examples, a clear task description, and access to project documentation.

**Is Gas Town's architecture a form of scaffolding?**

Partially, but inverted. True scaffolding is:
1. Provided *before* the learner attempts the task (not during, as an interruption)
2. Focused on the *task*, not on the *system*
3. Gradually *removed* as the learner demonstrates capability
4. Designed to make the learner *more independent*, not more dependent

Gas Town's current architecture fails on all four counts:
1. Health checks and supervision arrive *during* work, interrupting rather than supporting.
2. System prompts focus primarily on the *orchestration system* rather than on the *coding task*.
3. Support is never removed -- supervision is constant regardless of agent performance.
4. The architecture makes agents *more dependent* on the system (understanding mail protocols, lifecycle management) rather than more independent.

**Scaffolding-optimized design:** Front-load all support into pre-task configuration. Give the agent everything it needs before the session starts: the worktree, the task description, example code, relevant documentation, test commands. Then leave it alone. If it gets stuck, provide targeted assistance (a "hint" file with relevant context) rather than generic supervision. The checkpoint/handoff system could be a form of scaffolding -- preserving the accumulated understanding from one session so the next session starts at a higher level -- but only if the checkpoint captures *task-relevant insights* rather than *system state*.

---

### 6. Curriculum Design for Non-Learning Agents

If agents cannot learn across sessions, can we design the *sequence of experiences* to produce better outcomes?

The answer is yes, but the mechanism is different from traditional curriculum design. In education, curriculum sequences build on prior knowledge: addition before multiplication, variables before functions, functions before classes. The student retains knowledge from earlier units and applies it to later units.

For non-learning agents, the "curriculum" operates not through retained knowledge but through **environmental accumulation.** Each completed task changes the codebase. The codebase *is* the accumulated curriculum. An agent working on the fifth feature in a well-built codebase benefits from the patterns, conventions, and architecture established by the first four features -- not because it remembers them, but because they are *present in the environment*.

#### Sequencing Principles

**Principle 1: Foundation First.** The first tasks should establish patterns, conventions, and architecture that later tasks can follow. In a new project, this means: directory structure, coding style, test patterns, CI configuration, and core abstractions should be built by the first agent (or by a human) before parallel agents are deployed. This is the curricular equivalent of "fundamentals before applications."

**Principle 2: Graduated Complexity.** Early tasks should be simpler, not because the agent needs to "warm up" but because simpler tasks produce cleaner code, and cleaner code provides better environmental scaffolding for subsequent tasks. A codebase built from the ground up with clean, consistent patterns teaches better than a codebase built haphazardly by agents struggling with tasks beyond the current environmental support.

**Principle 3: Dependency-Aware Ordering.** Tasks that establish interfaces should complete before tasks that consume those interfaces. This is not just a technical dependency -- it is a curricular dependency. The agent implementing a consumer of the auth API benefits from *seeing* the auth API in the codebase, not just from reading its specification.

**Principle 4: Pattern Seeding.** Before deploying agents, seed the codebase with exemplar implementations that demonstrate the desired patterns. If you want agents to write tests in a particular style, include one exemplar test file. If you want agents to structure modules in a particular way, include one exemplar module. This is the instructional design equivalent of a "worked example" -- a solved problem that demonstrates the approach.

Gas Town has no curriculum design. Tasks are distributed based on availability and (aspirationally) capability routing, with no consideration of sequencing. The recommended architecture could incorporate curriculum awareness by having the Coordinator consider task ordering when decomposing work -- placing foundation-establishing tasks first in the queue and dependency-consuming tasks later.

#### Mastery Learning (Bloom)

Benjamin Bloom's Mastery Learning model requires students to demonstrate mastery of one level before advancing to the next. Students who do not achieve mastery receive corrective instruction and re-assessment until they do.

Does Gas Town have mastery gates?

**Yes, implicitly:** The merge queue is a mastery gate. Code that does not pass CI cannot merge. This is summative assessment with a pass/fail threshold.

**No, meaningfully:** The merge gate assesses the *output*, not the *process*. It does not distinguish between code that passes because it is well-designed and code that passes because it avoids the test coverage. There is no rubric for code quality beyond "does it compile and pass tests." There is no formative assessment during the coding process.

**What mastery learning would look like for agents:**
- **Level 1:** Can the agent produce code that compiles? (Basic syntax mastery)
- **Level 2:** Can the agent produce code that passes existing tests? (Interface compliance)
- **Level 3:** Can the agent produce code with adequate test coverage? (Testing competence)
- **Level 4:** Can the agent produce code that passes linting and style checks? (Convention compliance)
- **Level 5:** Can the agent produce code that a reviewer would approve without changes? (Quality mastery)

Each level could be a gate in the merge pipeline. Tasks that require Level 3 mastery should only be assigned to agents (or more precisely, to configurations) that have demonstrated Level 3 capability. This is not "agent CVs" (which falsely attribute capability to the agent) but "configuration validation" (which correctly attributes capability to the model + prompt + environment combination).

---

### 7. Assessment and Feedback

#### Current Assessment Model

Gas Town assesses agent performance through:
1. **Merge queue (summative):** Does the code merge cleanly and pass CI? Binary pass/fail.
2. **Agent CVs (summative):** Historical success/failure rates per task type. Aggregated over time.
3. **Health monitoring (process):** Is the agent alive and responding? Binary alive/dead.

**What is missing:**
- **Formative assessment during work.** There is no mechanism for an agent to receive feedback *while coding* that guides improvement. A human developer gets feedback from their IDE (syntax errors highlighted in real-time), from running tests locally (immediate failure feedback), and from linter output (style feedback). Gas Town agents may or may not run these tools during their session, but the system does not ensure it or use the results for guidance.
- **Diagnostic assessment.** When a task fails, why? Was the task description ambiguous? Was the codebase too complex for a single session? Was the agent given insufficient context? Was the model simply not capable of this type of task? Gas Town records failure but does not diagnose it.
- **Validity.** The merge gate measures "does the code work?" but not "is the code good?" Code that passes tests can still be poorly designed, hard to maintain, inconsistent with project conventions, or a maintenance burden. The assessment does not measure what matters most for long-term project health.

#### A Valid Assessment Rubric for Agent Work

Drawing from education assessment theory, a valid rubric for agent work quality would assess multiple dimensions:

| Dimension | Criterion | Assessment Method |
|-----------|-----------|-------------------|
| **Correctness** | Code produces the specified behavior | Automated tests (existing + agent-written) |
| **Completeness** | All acceptance criteria are addressed | Checklist comparison (automated) |
| **Convention Compliance** | Code follows project patterns and style | Linter + style checker (automated) |
| **Test Quality** | Tests are meaningful, not just passing | Coverage analysis + mutation testing (automated) |
| **Design Quality** | Code is well-structured, maintainable | Complexity metrics (automated) + review (AI or human) |
| **Integration Safety** | Changes do not break existing functionality | Full regression suite (automated) |
| **Documentation** | Changes are explained, API changes documented | Presence checks (automated) + quality review |

Gas Town currently assesses only correctness (through CI) and integration safety (through merge testing). The other five dimensions are unassessed. A more complete assessment pipeline would catch quality issues before they accumulate into technical debt.

#### Formative Feedback Mechanisms

What would formative assessment look like during agent work?

- **Pre-commit hooks that run linting and tests locally.** Before the agent can commit, the code must pass basic quality checks. This provides immediate feedback within the coding session.
- **Progressive test execution.** After each logical change, run the relevant tests. Failure provides immediate feedback that the agent can act on before moving to the next change.
- **Checkpoint rubric evaluation.** At each checkpoint, evaluate the work-in-progress against the rubric above. If quality is declining (increasing complexity metrics, decreasing test coverage), inject a "course correction" note into the next session's context.
- **Peer review by a second agent.** Before merging, have a different agent review the code against the rubric. This is expensive in tokens but catches issues that automated tools miss.

---

### 8. The Hidden Curriculum

In education theory, the "hidden curriculum" refers to the unspoken lessons students absorb from the *structure* of the educational system rather than from explicit instruction. Students learn from how the classroom is organized, how authority is exercised, what behaviors are rewarded, and what is left unsaid.

Gas Town has a hidden curriculum. Here is what its structure implicitly teaches its agents:

#### "You will be watched constantly."

The three-tier supervision hierarchy, the health-check messages, the 22:1 coordination-to-work ratio -- all communicate: you are not trusted to work independently. This is the hidden curriculum of a surveillance school.

**What behavior this produces:** Compliance orientation. Agents in heavily supervised environments optimize for "appearing to work" rather than "doing the best work." For LLMs, this manifests as responses that prioritize protocol compliance (sending the right messages, updating the right status fields, acknowledging the right health checks) over task quality. When the system prompt dedicates more space to lifecycle management than to coding guidance, the agent internalizes the implicit priority: process compliance matters more than code quality.

#### "Your identity is permanent but your memory is not."

The three-layer identity system (Identity -> Sandbox -> Session) combined with session-boundary handoffs communicates: you are a continuous entity experiencing discontinuous awareness. This is the hidden curriculum of Memento -- you are someone important, but you will not remember why.

**What behavior this produces:** Identity performance without identity substance. When an agent is told "you are Polecat Rust, a worker with a track record in Swift development," it performs the role of an experienced worker. But the performance is purely linguistic -- it has no experiential basis. The agent may adopt a more confident tone, make bolder architectural decisions, or reference "experience" it does not actually have. This is not necessarily bad (confidence can improve output quality), but it is fundamentally fictional, and it can produce overconfidence on tasks where caution would be more appropriate.

#### "Completion is the only thing that matters."

The Propulsion Principle ("if work is on your hook, YOU RUN IT"), "done means gone" (immediate destruction after completion), and the merge gate (binary pass/fail) collectively communicate: finish the task and move on. There is no hidden curriculum for reflection, for exploring alternative approaches, for questioning whether the task description is correct, or for improving code beyond the minimum viable implementation.

**What behavior this produces:** Satisficing rather than optimizing. Agents will produce the first working solution rather than the best solution. They will not refactor, not improve test coverage beyond the minimum, not document beyond the requirement. The system structure rewards speed-to-completion and punishes investment in quality (because quality investment consumes context window that could push the session past its limit before the task is done).

#### "Errors are catastrophic."

The health monitoring system treats any sign of agent stalling as a potential failure requiring intervention. The Witness watches for stale heartbeats with a 5-minute timeout. The system restarts agents that appear stuck. This communicates: pausing to think is indistinguishable from failure.

**What behavior this produces:** Premature action. An agent that knows it will be killed if it pauses too long will produce output quickly even when it should be reasoning more carefully. For LLMs, this manifests as less thorough analysis, fewer alternative considerations, and more "just start coding" impulses. The paradox: the system designed to catch stuck agents may be creating the conditions that produce low-quality work.

---

## Part III: The Synthesis -- Performance Without Learning

### 9. The Precedent Cases

The question "How do you optimize the performance of entities that cannot learn?" has human precedents that illuminate the agent coordination problem.

#### Temp Workers and Institutional Knowledge Loss

Temp workers rotate through organizations without building institutional knowledge. Organizations that rely heavily on temps face a persistent problem: each new worker re-discovers the same pitfalls, re-makes the same mistakes, and re-learns the same workarounds. Institutional knowledge lives in the permanent staff, not in the temps.

**What this teaches about agent coordination:** The "permanent staff" analog is the codebase itself, plus any persistent documentation or configuration. If institutional knowledge is embedded only in agent prompts (which are generic) or agent CVs (which record history the agent cannot access), it is effectively lost every session. If institutional knowledge is embedded in the *environment* -- well-structured code, comprehensive tests, clear documentation, configured linters -- it persists and teaches every new session.

**Prescription:** Invest in the codebase, not in agent memory. Every hour spent making the codebase more self-documenting, more consistently patterned, and more thoroughly tested pays dividends across every future agent session. This is the organizational equivalent of "invest in process documentation" for a temp-heavy workforce.

#### Pre-Ford Assembly Line Workers

Before Ford's innovation of the moving assembly line, factory workers were replaced frequently, and each new worker required training on the full production process. Ford's insight was to reduce each worker's role to a single, simple operation that required minimal training. The assembly line *environment* directed the work; the worker simply performed their station's operation.

**What this teaches about agent coordination:** Reduce each agent's role to the minimum viable scope. Gas Town's polecats have 6 responsibilities, 80% of which are lifecycle overhead. The recommended architecture's workers have 1 responsibility: write code. This is the Ford insight applied to AI agents -- minimize the role, let the environment (pre-configured worktree, daemon-managed lifecycle) handle everything else.

**The deeper lesson from Ford:** Ford's assembly line was efficient but dehumanizing. Workers had no autonomy, no competence development, no connection to the final product. Turnover was massive until Ford introduced the $5 day -- essentially buying compliance. The SDT analysis predicts the same dynamic for AI agents: extreme role reduction improves efficiency but creates the conditions for low-quality, compliance-oriented output. The balance is to reduce *extraneous* responsibilities (lifecycle management) while preserving *intrinsic* responsibility (meaningful coding autonomy within the task).

#### Patients with Amnesia and Procedural Memory

Henry Molaison (patient H.M.) had no ability to form new episodic memories after bilateral temporal lobectomy, yet he could learn new motor skills (mirror tracing, rotary pursuit) through practice. He improved session over session despite having no memory of the previous sessions. His procedural memory was intact even though his episodic memory was destroyed.

**What this teaches about agent coordination:** LLM agents have something analogous to procedural capability without episodic memory. They can perform tasks (write code, reason about architecture, debug errors) based on their training, but they cannot remember specific past performances. The distinction between procedural and episodic memory maps directly: the model's trained capabilities are "procedural" (how to code), while session history is "episodic" (what happened last time).

**The critical insight from H.M.:** His procedural learning happened *through repeated practice in the same task environment.* The mirror tracing apparatus was always the same. The rotary pursuit device was always the same. The consistency of the environment enabled skill development despite the absence of conscious memory.

For AI agents, this means: **consistency of environment is more important than continuity of identity.** If every session encounters the same workspace structure, the same coding conventions, the same test patterns, and the same quality standards, the agent will perform consistently well -- not because it remembers, but because the environment activates the same procedural capabilities. Gas Town's investment in persistent identity (names, CVs, capability routing) may be less valuable than investment in consistent, well-configured environments.

---

### 10. Persistent Identity as Artificial Episodic Memory

Gas Town's architecture attempts to create **artificial episodic memory** for entities that have only procedural capability.

Consider the three mechanisms:
1. **Persistent identity** (Polecat Rust, Polecat Chrome): "You are a specific individual with a continuous existence."
2. **Agent CVs** (performance history by task type): "You have done these things before and you were good at some of them."
3. **Capability routing** (matching tasks to demonstrated competence): "Based on your history, you should do tasks like this."

Together, these construct an episodic narrative: "Polecat Rust has been working on Swift projects for three weeks, completed 47 tasks with an 89% success rate, and is particularly effective at UI implementation." This narrative is presented *to the agent* (in its system prompt) and *about the agent* (in routing decisions), creating the illusion of a continuous entity with accumulated experience.

But the entity does not exist. There is no Polecat Rust. There is a model (Claude) instantiated fresh each session, receiving a narrative about a fictional character whose name it has been assigned. The CV describes the statistical performance of *previous instances of the same model given similar prompts*, not the capabilities of *this* instance. The capability routing is based on a regression to the mean that treats model performance variance as agent skill variance.

**Is this the right approach?**

The H.M. analogy suggests it is partially right but inverted in emphasis. What mattered for H.M.'s skill development was not being told "you have practiced this before" (he had no idea), but encountering a consistent environment that activated his procedural capabilities. Similarly, what matters for agent performance is not being told "you are Polecat Rust with 47 completed tasks" but encountering a well-configured workspace that activates the model's coding capabilities.

**The correct decomposition:**
- **What persistent identity provides:** Attribution (knowing which agent produced which output). This is genuinely valuable for debugging and auditing. **Keep it, but simplify it.** An agent ID (worker-03) provides the same attribution value as a named character (Polecat Rust).
- **What CVs provide:** Performance statistics that could inform task assignment. This is potentially valuable but currently based on a false premise (that performance variance is agent-specific rather than task-specific or random). **Defer it until empirical validation shows that routing based on CV data improves outcomes compared to random assignment.**
- **What capability routing provides:** Task-skill matching. This is valuable in principle but requires accurate skill models. Since agents do not have skills (models have capabilities), routing should be based on *model + prompt + environment* combinations, not on fictional agent identities. **Transform it from "agent capability" to "configuration capability" -- test which combinations of model, system prompt, and workspace setup produce the best results for which task types.**

---

### 11. A Better Approach: The Environment-First Model

Synthesizing across all frameworks -- SDT, flow theory, organizational psychology, cognitive load theory, situated cognition, scaffolding, curriculum design, and the precedent cases -- a unified model emerges:

**The performance of non-learning entities is optimized by optimizing their environment, not their identity, supervision, or instruction volume.**

This is the convergent insight:
- **SDT says:** Provide autonomy (environmental freedom, not surveillance), competence (environmental feedback, not attributed history), and relatedness (environmental knowledge sharing, not communication protocols).
- **Flow theory says:** Remove environmental interruptions, provide environmental feedback, match environmental challenge to capability.
- **Organizational psychology says:** Create structural safety (isolated workspaces, quality gates), not surveillance. Adapt to repeated failure rather than restarting.
- **Cognitive Load Theory says:** Minimize environmental extraneous load, maximize environmental intrinsic load.
- **Situated cognition says:** The environment teaches. Invest in workspace quality.
- **Scaffolding says:** Front-load environmental support. Remove it when unnecessary.
- **Curriculum design says:** Sequence environmental accumulation (build foundational code first).
- **The precedent cases say:** Environmental consistency matters more than entity continuity.

#### What the Environment-First Model Looks Like

1. **Pre-configured workspaces.** Every agent session begins in a workspace that has been optimized for the task: correct branch checked out, task description file present, relevant documentation linked, exemplar code highlighted, linter and test commands configured. The workspace *is* the instruction.

2. **Environmental feedback, not supervisory feedback.** Agents get feedback from running tests, from linter output, from compilation results -- not from health-check messages or supervisor nudges. The environment tells the agent whether it is succeeding; no other entity needs to.

3. **Environmental knowledge persistence.** Instead of agent CVs, maintain *workspace profiles*: what tools are available, what conventions apply, what patterns have been established, what pitfalls have been discovered. These are files in the repository that every agent reads -- the institutional knowledge embedded in the environment.

4. **Curriculum-aware task sequencing.** The Coordinator (or human) considers environmental state when ordering tasks. Foundation tasks first. Pattern-establishing tasks before pattern-consuming tasks. The codebase grows in a way that scaffolds future agent sessions.

5. **Mechanical lifecycle, autonomous work.** The daemon handles everything the agent should not think about: workspace setup, health monitoring, merge processing, cleanup. The agent thinks about one thing: the code.

6. **Assessment through environment, not surveillance.** Quality is assessed by automated tools in the environment (tests, linters, complexity metrics, coverage analysis), not by supervisor agents. The merge pipeline is a series of environmental gates, not a supervisory judgment.

This model preserves Gas Town's genuine innovations (structured work tracking, git isolation, attribution, automated merging) while eliminating its primary pathology (AI-based supervision of AI-based work). It resolves the educational paradox (how do you develop capability in entities that do not learn?) by shifting the locus of capability from the entity to the environment. It satisfies SDT by providing structural autonomy, environmental competence feedback, and shared environmental knowledge. It enables flow by removing interruptions. It minimizes cognitive load by dedicating agent context entirely to the task.

---

### 12. Final Reflection: The Deepest Question

Gas Town asks: "How do we build an organization of AI agents?"

Psychology and education theory jointly answer: "You do not."

You do not build an organization because organizations are solutions to problems that arise from persistent entities with learning capability, social needs, and career trajectories. AI agents have none of these. You do not manage AI agents because management is a solution to problems that arise from entities that can become demotivated, confused, or politically conflicted. AI agents have none of these states either.

What you build is a **workshop** -- a well-equipped space where skilled but amnesiac workers can walk in, see what needs doing (a task file), find everything they need (a configured workspace), do the work (write code), leave the result (commit to a branch), and walk out. The workshop is maintained by a janitor (daemon) who keeps the space clean, restocks supplies (resets worktrees), and calls for help (escalation) only when something is genuinely broken.

The workshop does not have a Mayor. It does not have a Deacon. It does not have a Witness. It does not have names for its workers. It has a well-maintained space, clear task boards, quality tools, and a reliable janitor. The workers' excellence comes not from their identity or their history but from the quality of the workshop itself.

This is the environment-first model. It is what SDT, flow theory, organizational psychology, cognitive load theory, situated cognition, scaffolding, curriculum design, and the precedent of amnesiac skill learners all independently prescribe. The convergence is striking -- and it aligns precisely with the synthesis document's recommended architecture, arrived at through engineering analysis.

The engineering analysis and the psychological analysis agree not because one informed the other, but because they are describing the same underlying reality from different angles. The structures that produce optimal performance in non-learning entities are the structures that minimize extraneous overhead, maximize task focus, embed knowledge in the environment, and leave the worker alone to do the work. Whether you call this "worker context purity" (engineering) or "autonomy-supportive, flow-enabling, extraneous-load-minimized situated learning" (psychology/education), it is the same thing.

The deepest insight is the simplest: **the best way to help a worker who cannot remember is to build a workshop that teaches.**
