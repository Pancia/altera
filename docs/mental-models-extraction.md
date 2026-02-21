# Mental Models, Prompts, and Techniques Extraction

Extracted from three articles by Michael Simmons (Blockbuster / Thought Leader School).
Context: Applying these frameworks to analyze and improve a multi-agent orchestration system (Gas Town).

---

## Article 1: "Every Mental Model You've Learned Is Wrong" (AI Command Language)

Source: https://blockbuster.thoughtleader.school/p/every-mental-model-youve-learned

### Core Thesis

Mental models as commonly taught are "vocabulary without grammar." People collect isolated models (Pareto, First Principles, etc.) but never learn how to structurally combine them into deliberate reasoning sequences. The article proposes an organized system of **Lenses**, **Operations**, and **Recipes**.

### Three Categories of Thinking Tools

| Category | Purpose | Examples |
|----------|---------|---------|
| **Lenses** | Direct attention to specific patterns | Pareto Principle, Second-Order Thinking, Chesterton's Fence, Goodhart's Law, Lindy Effect |
| **Operations** | Processing methods that transform observations | First Principles, Falsification, Distant Domain Import, Steelmanning, Systems Thinking, Abductive Reasoning |
| **Recipes** | Sequenced combinations of operations through specific lenses for particular problems | Wrong-Problem Detector, Innovation Engine, Blind Spot Finder |

### Nine Core Reasoning Operations

Organized into four functional groups:

#### GENERATE (Create Possibilities)
1. **Analogical Reasoning** -- "Import structures from distant, unrelated domains." Find parallel structures in fields that have nothing obvious in common with your problem.
2. **Abductive Reasoning** -- Generate hypotheses that explain surprising observations. Start from an anomaly and reason backward to plausible causes.
3. **Counterfactual Analysis** -- Isolate individual factors' contributions through mental experiments. Ask "What if X were removed/changed?"

#### EVALUATE (Test Against Reality)
4. **Falsification** -- Actively attempt to prove ideas wrong. Seek disconfirming evidence before confirming evidence.
5. **Bayesian Updating** -- Calibrate confidence to evidence continuously. Update beliefs proportionally as new data arrives.

#### DECONSTRUCT (Strip to Foundations)
6. **First Principles** -- Remove assumptions layer by layer until reaching bedrock truths that cannot be reduced further.

#### INTEGRATE (Combine Across Boundaries)
7. **Dialectical Synthesis** -- Hold opposing positions simultaneously to find a transcendent truth that reconciles both.
8. **Systems Thinking** -- Map relationships, feedback loops, and emergent properties. See the whole, not just parts.
9. **Perspective Simulation** -- Model others' knowledge, beliefs, and intentions with high fidelity.

### Three Complete Thinking Recipes

#### Recipe 1: Wrong-Problem Detector
Use when you suspect you are solving the wrong problem entirely.

| Step | Operation | Lens |
|------|-----------|------|
| 1 | Zeroth Principle | Inversion -- question the assumptions *behind* the assumptions about the problem definition |
| 2 | Abductive Reasoning | Anomaly Hunting -- look for surprising observations that don't fit current framing |
| 3 | Counterfactual Analysis | Removal Test -- mentally remove elements to see which ones actually matter |
| 4 | Falsification | Crucial Experiment -- design a test that would definitively prove the current framing wrong |

**Application to Gas Town:** Before optimizing agent communication protocols, apply the Wrong-Problem Detector. Are we solving the right problem? What anomalies exist in current system behavior that suggest a different root cause?

#### Recipe 2: Innovation Engine
Use when seeking novel solutions or redesigns.

| Step | Operation | Lens |
|------|-----------|------|
| 1 | Analogical Reasoning | Three distant domains -- pull structural parallels from biology, economics, urban planning, etc. |
| 2 | First Principles | Regressive Abstraction -- strip the problem to its most fundamental requirements |
| 3 | Dialectical Synthesis | Both/And Reframe -- instead of choosing between competing approaches, find a design that incorporates both |
| 4 | Falsification | Pre-Mortem -- assume the innovation failed; reason backward to discover why |

**Application to Gas Town:** Import agent coordination patterns from ant colonies (biology), market mechanisms (economics), and city infrastructure (urban planning). Strip to first principles: what does multi-agent orchestration *actually* need? Then pre-mortem the resulting design.

#### Recipe 3: Blind Spot Finder
Use when a design seems "complete" but you suspect hidden gaps.

| Step | Operation | Lens |
|------|-----------|------|
| 1 | Systems Thinking | Unintended Consequences Tracing -- follow second and third-order effects |
| 2 | Perspective Simulation | Strongest Possible Objection -- model the most sophisticated critic |
| 3 | Abductive Reasoning | Surprising Absence Detection -- what *should* be present but isn't? |
| 4 | Analogical Reasoning | Negative Analogy -- what breaks down in the analogies we're using? |

**Application to Gas Town:** Trace unintended consequences of agent autonomy levels. What would the strongest critic say about the architecture? What features are conspicuously absent? Where do our metaphors (mayor, deacon, daemon) break down?

### Key Concepts

- **"Zeroth Principle"** -- Question the assumptions behind the assumptions. One level deeper than first principles.
- **Cognitive Signature** -- Your default thinking patterns and systematic blind spots. Every thinker has them.
- **Grammar vs. Vocabulary** -- Knowing 100 mental models is useless if you can't combine them deliberately. The structure matters more than the catalog.

---

## Article 2: "10,000x Knowledge Worker" (Historical Productivity Applied to AI)

Source: https://blockbuster.thoughtleader.school/p/10000x-knowledge-worker-how-historys

### Core Thesis

The biggest productivity gains in history came not from better tools but from reimagining how work is organized (assembly line, scientific management, etc.). The same revolution is now possible with AI -- not by writing better prompts, but by building organizational structures where AI agents manage other AI agents.

### The Five Stages of AI Mastery

Each stage compounds multiplicatively (not additively) on the previous.

| Stage | Leverage | Description |
|-------|----------|-------------|
| 1. **Prompt Engineering** | 10x | Crafting better prompts with relevant data, chain-of-thought reasoning, roles, examples, formatting |
| 2. **Infinite Prompting** | 100x total | Creating "thinking structures" that allow AI to think deeply for hours or days |
| 3. **Model Management** | 1,000x total | Orchestrating multiple AIs thinking in parallel on different aspects |
| 4. **Model Leadership** | 10,000x total | AIs managing other AIs hierarchically; humans manage the AI managers |
| 5. **Autonomous Firms** | Unknown | Companies operating with minimal or no human employees |

**Application to Gas Town:** Gas Town is already operating at Stages 3-4. The question is whether its organizational structure (Mayor -> Deacon -> Daemon -> Plugins) is optimally designed for the 10,000x multiplier.

### Henry Ford's Assembly Line Principles (Applied to AI)

#### The Critical Problem Redefinition
Ford's breakthrough: Shifting from **"How can workers move faster?"** to **"Why are workers moving at all?"**

This is the single most important insight for multi-agent design. Don't optimize the current workflow -- question why agents are structured that way at all.

#### Four Key Optimizations

1. **Extreme Specialization** -- Ford decomposed one role into 29 separate jobs. Each agent should do exactly one thing exceptionally well.
2. **Radical Experimentation** -- "We try everything in a little way first." Small-scale tests before full deployment.
3. **Precise Measurement and Iteration** -- Test assembly line heights and speeds. Measure agent performance on specific metrics, not just "does it work."
4. **Compensation Alignment** -- Ford doubled wages to retain workers. For agents: ensure the reward/feedback signals align with desired behavior.

**Application to Gas Town:** How many "jobs" does each agent currently do? Can roles be decomposed further? What metrics are being tracked? Is there a systematic experimentation framework?

### Management Innovation Frameworks (Applicable to Agent Design)

These industrial-era management innovations can be directly mapped to multi-agent orchestration:

- **Scientific Management** -- Systematic study of workflows to find optimal processes
- **Lean Production** -- Eliminate waste; every step must add value
- **Six Sigma** -- Reduce variation and defects through statistical methods
- **Extreme Division of Labor** -- Each agent has a narrow, well-defined responsibility
- **Standardization** -- Common interfaces, protocols, and output formats across agents
- **Kaizen** -- Continuous incremental improvement through regular review cycles
- **Total Quality Management (TQM)** -- Quality built into every step, not inspected at the end
- **Standard Operating Procedures (SOPs)** -- Explicit, documented procedures for each agent's behavior

### Infinite Prompting Mastery Levels

Progressive capabilities for deep AI reasoning:

1. **Diverse Types** -- Use open-ended, divergent, and convergent prompts in sequence
2. **Full Application** -- Apply across problem exploration, strategy, research, creation, editing
3. **Multi-Medium** -- Text, image, video generation as needed
4. **Scaffolding** -- Support AI thinking via reasoning methods arranged in ideal sequences
5. **Meta-Infinite Prompting** -- Infinite prompts that create and execute other infinite prompts in parallel
6. **Notifications** -- Completion alerts so humans don't block on AI processing
7. **Dashboard** -- Track executing prompts with estimated completion times

**Application to Gas Town:** Are there equivalent layers in Gas Town? Is there meta-orchestration (agents spawning agent workflows)?

### Prompt Engineering Best Practices

Essential components for high-performance prompts:
- Chain of thought reasoning
- Relevant contextual data
- Clear goals
- Role specification
- Examples
- Formatting (XML, JSON)
- Variables
- Output formats

Software development practices adapted for prompts:
- Domain-specific instruction libraries
- Evaluation systems (evals)
- Version control for prompts
- A/B testing by user segments

### Compounding Gains Pattern

**Gains multiply rather than add: 10x x 10x x 10x = 1,000x (not 30x)**

This is critical for multi-agent design. Each layer of improvement compounds. A 2x improvement in agent specialization, combined with 2x improvement in communication, combined with 2x improvement in error handling = 8x total improvement, not 6x.

### The "Mega Steve" Model

An AI CEO concept where one intelligence experiences all company activities simultaneously through millions of specialized copies, with knowledge flowing back instantly across the system. No information loss between departments. No communication overhead.

**Application to Gas Town:** Does Gas Town have a unified knowledge/context layer that all agents share? Or does each agent operate with its own siloed context? The Mega Steve model suggests maximum value from a shared consciousness layer.

### Management Bandwidth Hierarchy

Progressive autonomy levels for AI agents, measured by how often human verification is needed:

- Every 15 minutes (low autonomy)
- Every hour
- Every 5 hours
- Until AI manages AI verification (full autonomy)

**Application to Gas Town:** What is the current "check-in frequency" at each level of the hierarchy? Can it be reduced?

### Key Paradigm Shifts

1. **From Tool to Workforce** -- Stop perfecting individual prompts; start building organizational charts where agents manage agents.
2. **Thinking Time as Currency** -- "AI thinking time is the new currency of intellectual work."
3. **Capital -> Compute -> Labor** -- "For the first time in history, you can just turn capital into compute and compute into labor."

### Coding Productivity Waves

Each modality approximately 5-10x more productive than predecessor:

| Modality | Multiplier | Paradigm |
|----------|-----------|----------|
| Manual coding | Baseline | Human writes everything |
| Chat (Copilot, etc.) | ~5x | Human asks, AI generates snippets |
| Agents (Claude Code, etc.) | ~5x over chat | AI executes multi-step tasks autonomously |
| Agent Clusters | ~5x over agents | Multiple agents working in parallel |

**The human becomes "an orchestra conductor" rather than a solo contributor.**

---

## Article 3: "Perspective Prompting" (Reid Hoffman's Multi-Perspective Method)

Source: https://blockbuster.thoughtleader.school/p/perspective-prompting-how-reid-hoffmans

### Core Thesis

The highest-quality thinking comes not from a single brilliant perspective but from systematically summoning multiple expert viewpoints, designing productive disagreement between them, and synthesizing insights no single perspective could reach. AI makes this practical at scale.

### The Perspective Prompting Framework

Three-phase process:

1. **Summon** diverse expert viewpoints (10+ perspectives from different disciplines)
2. **Design Disagreement** -- structure productive conflict between perspectives
3. **Synthesize** -- combine insights into higher-order understanding

### Wisdom of Crowds (Scientific Foundation)

Diverse, independent perspectives mathematically compound accuracy. When aggregated correctly, individual errors cancel while accurate judgments reinforce.

**Conditions for Wisdom of Crowds to work:**
1. **Independent perspectives** -- extract rare insights
2. **Diverse perspectives** -- cancel biases
3. **Synthesis** -- generate emergent insight
4. **Aggregation mechanism** -- market, vote, algorithm, or structured review

### The Diversity Prediction Theorem (Scott Page)

**Collective Error = Average Individual Error - Prediction Diversity**

Groups solve more problems correctly when members think differently, even if individual members are less capable. Diversity of thought > individual brilliance.

**Application to Gas Town:** Agent diversity matters more than individual agent capability. Having agents with different "cognitive approaches" (different system prompts, different models, different specializations) should outperform homogeneous agents.

### Constructive Controversy (Johnson & Johnson)

Structured intellectual disagreement produces better outcomes than:
- Consensus-seeking (groupthink)
- Simple averaging (median quality)
- Debate (adversarial, seeks to win rather than learn)

**The key is structured disagreement with synthesis, not conflict.**

**Application to Gas Town:** Build disagreement into the agent workflow. Have agents explicitly challenge each other's outputs before synthesis. Not just "review" but "constructively oppose."

### Superforecasting Principles (Philip Tetlock)

- **Active open-mindedness** -- treat beliefs as testable hypotheses, not treasures to guard
- Superforecasters consistently outperform intelligence analysts by ~30%
- The skill is updateability, not initial accuracy

### Perspective Types Framework

A taxonomy for ensuring comprehensive perspective coverage:

| Category | Examples |
|----------|---------|
| **Expertise** | Venture capitalist, product manager, CEO |
| **Field** | Artistic, scientific, engineering |
| **Mental Model** | 80/20 Rule, Blue Ocean Strategy |
| **Thinking Style** | Devil's advocate, systems thinker |
| **Mindset** | Optimist, pessimist, realist |
| **Stakeholder** | Boss, subordinate, partner, board member, end user |
| **Temporal** | Historian, futurist, present-focused operator |
| **Failure Mode** | Post-mortem analyst, disappointed customer |
| **Outsider** | Intelligent newcomer, child, cross-industry observer |
| **Methodological** | Qualitative researcher, quantitative analyst |
| **Cross-Cultural/Disciplinary** | Japanese quality engineer, Danish designer, physicist |

**Application to Gas Town:** When evaluating agent architecture, systematically rotate through these perspective categories. Each reveals blind spots the others miss.

### Seven Quick-Win AI Prompts

#### Prompt 1: Identifying Missing Perspectives
```
List all perspectives I take in this [document/design/plan] grouped into categories.
What new categories and viewpoints are in my blindspot that would drastically
increase quality?
```

#### Prompt 2: Finding Cognitive Blind Spots
Add to AI system settings:
```
Identify which cognitive operations/frames/mental models/paradigms I'm employing
and which ones I'm not yet aware I could be using.
```

#### Prompt 3: Stakeholder Translation
```
How would an experienced [their role] interpret what I'm about to say?
What would they hear unintentionally?
What concerns aren't on my radar?
```

#### Prompt 4: Anti-Fragile Ideas
```
Generate the 10 most sophisticated critiques from different disciplinary
perspectives -- philosopher, empiricist, historian, systems theorist,
practitioner, etc.
```

#### Prompt 5: Infinite Audience Testing
Structure: For each audience segment, rate the output 1-10, recommend the #1 improvement, provide detailed response, identify problem areas, highlight what works.

#### Prompt 6: Creative Synthesis
```
Synthesize [current trend/design] with:
(1) A Stoic philosopher's perspective on unchanging human nature
(2) A sci-fi visionary's view on what we'll look back on as quaintly mistaken
Extract the timeless principle.
```

#### Prompt 7: Developmental Growth
```
Identify my "internal logic" or "subject-object trap" in this thinking.
Present the most sophisticated version of the perspective I'm avoiding.
End with one "killer question" that would shift my frame.
```

### Five Types of Synthesis

| Type | Purpose | Application |
|------|---------|-------------|
| **Conceptual Synthesis** | Combine different models/frameworks into comprehensive view | Merge multiple architectural patterns into unified design |
| **Empirical Synthesis** | Combine findings/evidence into aggregate conclusions | Aggregate performance data across agent types |
| **Practical Synthesis** | Combine recommendations into actionable guidance | Turn analysis into concrete next steps |
| **Value/Interest Synthesis** | Find integrative solutions across different values/priorities | Balance reliability vs. speed vs. cost in agent design |
| **Dialectical Synthesis** | Create higher-order view transcending opposing positions | Resolve tension between agent autonomy and control |

### Mixture-of-Agents (MoA) Research

Scientific finding: Multiple ordinary AI models synthesized through rounds outperform a single superior model working alone.

**Self-MoA Alternative:** Generating multiple diverse outputs from a single high-quality model and then synthesizing often outperforms mixing different models.

**Application to Gas Town:** This directly validates multi-agent architectures. But the key insight is that the synthesis mechanism matters as much as the agents themselves. How Gas Town aggregates and synthesizes agent outputs is a critical design choice.

### The 15 Drafts Rule

Michael Simmons' editing framework applying Wisdom of Crowds through 15 rounds of revision, each from a different perspective:

1. **Intuitive Perspective** -- vomit draft, raw output
2. **Hook Perspective** -- grab attention, lead with value
3. **Layman's Perspective** -- beginner/outsider view, strip jargon
4. **Trademark Ideas Perspective** -- naming opportunities, memorable concepts
5. **Comprehensive Perspective** -- include all ideas without trimming
6. (Perspectives 6-15 for deeper refinement)

**Application to Gas Town:** Agent outputs could go through multiple refinement passes, each applying a different quality lens, before being considered final.

### Key Insight

"You came to work on a problem. But the problem starts working on you."

The process of systematically taking multiple perspectives doesn't just improve the answer -- it transforms the person (or system) asking the question.

---

## Cross-Article Synthesis: Applicable Techniques for Gas Town

### Immediate Application Techniques

1. **Wrong-Problem Detector** (Article 1) -- Before any optimization, verify the system is solving the right problem.

2. **Ford's Problem Redefinition** (Article 2) -- Shift from "How can agents communicate faster?" to "Why are agents communicating at all?" Eliminate unnecessary coordination.

3. **Perspective Prompting on Architecture** (Article 3) -- Evaluate Gas Town through 10+ perspectives: systems theorist, end user, Japanese quality engineer, failure analyst, historian of distributed systems, etc.

4. **Extreme Specialization Audit** (Article 2) -- How many distinct jobs does each agent do? Can they be decomposed into more specialized sub-agents? Ford went from 1 role to 29.

5. **Constructive Controversy in Agent Pipelines** (Article 3) -- Build explicit disagreement into agent workflows. Don't just have agents pass output forward; have them challenge each other.

6. **Compounding Gains** (Article 2) -- Small improvements at each layer multiply. Focus on improvements that compound across the whole pipeline.

### Strategic Design Principles

1. **Diversity > Individual Capability** (Article 3, Diversity Prediction Theorem) -- Different agent approaches matter more than one perfect agent.

2. **Synthesis is the Bottleneck** (Article 3, MoA Research) -- The aggregation mechanism is as important as the agents.

3. **Management Innovation > Tool Innovation** (Article 2) -- The organizational structure of agents matters more than the model powering them.

4. **Recipes over Individual Models** (Article 1) -- Design agent workflows as explicit sequences of operations, not ad-hoc chains.

5. **Blind Spot Finder as Continuous Process** (Article 1) -- Regularly run the Blind Spot Finder recipe against the architecture to surface hidden assumptions.

### Meta-Prompts for System Analysis

Use these prompts to analyze Gas Town itself:

```
Apply the Wrong-Problem Detector to Gas Town's multi-agent orchestration:
1. Zeroth Principle via Inversion: What assumptions sit behind the assumption
   that we need Mayor -> Deacon -> Daemon -> Plugin hierarchy?
2. Anomaly Hunting: What surprising behaviors or failures suggest we've
   framed the problem incorrectly?
3. Removal Test: What happens if we remove the Deacon layer entirely?
   The Mayor layer? What is actually load-bearing?
4. Crucial Experiment: What test would definitively prove this architecture
   is wrong?
```

```
Apply the Innovation Engine to Gas Town:
1. Three distant domain analogies for multi-agent orchestration:
   - Biological: How do ant colonies / immune systems / neural networks
     coordinate without central control?
   - Economic: How do markets achieve coordination through price signals
     rather than commands?
   - Urban: How do cities self-organize through zoning, infrastructure,
     and emergent behavior?
2. First Principles: What does agent orchestration irreducibly need?
3. Both/And: Can we have both hierarchical control AND emergent behavior?
4. Pre-Mortem: We redesigned Gas Town and it failed. Why?
```

```
Apply Perspective Prompting to Gas Town:
- Systems Architect: Is the coupling between components appropriate?
- Reliability Engineer: What are the failure modes?
- End User: Does this complexity serve the user or the developer?
- Economist: What are the transaction costs between agents?
- Biologist: What would natural selection do to this system over 1000 generations?
- Historian: What precedents exist for this kind of coordination?
- Child: Why do you need so many helpers to do one thing?
- Competitor: How would I build this to beat Gas Town?
```
