# Humanize Instructions (distilled from humanize-ai-text skill)

Goal: Turn prose that *reads* like it came out of a language model into prose that reads like a careful person wrote it.

**Core test before editing any paragraph:**
1. Is there a concrete claim here?
2. Is there something the writer has committed to that could turn out to be wrong?

If both answers are no, the paragraph has a hollow centre. Flag it instead of polishing (`[no claim here; what is the author actually asserting?]`).

**What to do:**
- Read for meaning and intended register first.
- Strip filler and scaffolding (highest priority).
- Deflate inflated vocabulary.
- Break mechanical rhythm and structure.
- Cut empty hedging.
- Fix punctuation tics (especially paired em-dashes).
- Vary sentence length aggressively. Allow fragments. Allow sentences to start with "And" or "But".
- When a target voice is supplied, steer toward it (here: Gil's direct, skeptical-curious, staccato-but-uneven voice from biases/taste).

**Common AI tells to remove (count density, don't blacklist single words):**

**Boilerplate/scaffolding (delete):**
- Openers like "In today's...", "When it comes to...", "Let's dive in".
- Throat-clearing: "It's important to note that", "It's worth noting", "Rest assured".
- Closers that restate everything.

**Inflated diction (deflate):**
- leverage/utilize/harness/foster/embark/facilitate → use, help, start, build
- robust/seamless/comprehensive/holistic/myriad/plethora/multifaceted → delete or be specific
- crucial/pivotal/vital/key/essential → only when truly so
- realm/landscape/tapestry/testament/journey/ever-evolving/game-changer/cutting-edge → rewrite the sentence
- Sincerity cluster: genuinely/honestly/truly/meaningful/straightforward/deeply/profound, "signals a shift", "speaks to", "at its core"

**Rhetorical devices at wrong frequency:**
- Tricolon (rule of three): if the third item adds no new idea, cut it.
- Antithesis escalation ("not just X, it's Y", "not only... but also").
- Parallel adjective pairs ("powerful and flexible", "clear and thorough").

**Structural tics:**
- Every paragraph ~3 medium sentences.
- Reflexive bold labels or bullets where prose would flow.
- "Firstly / Secondly".

**Hedging and balanced non-answer:**
- Strip stacked qualifiers.
- Remove "While it's true that A, it's also important to consider B" when the author has a view.
- Cut "a wide range of", "various", "numerous".
- The classic hollow: "Ultimately, whether this is the right choice depends on a range of factors, and reasonable people may disagree." — flag, do not smooth.

**Punctuation:**
- Paired em-dashes used as interruptions (—like this—). Keep at most one per piece for real effect.
- Repeated "; however," constructions — break into sentences.

**Rhythm and stance:**
- Human prose has texture and a slightly uneven point of view.
- Follow a long sentence with a short one (three words can be a sentence).
- Let the author sound specific rather than a balanced encyclopedia entry.

**Restraint check before returning edited text:**
- Every real claim and fact kept intact?
- No invented specifics, quotes, or anecdotes?
- Hollow centres flagged rather than dressed up?
- Register still appropriate (for X posts: direct, no corporate slop, no hype for hype's sake)?
- Did we stop at natural, or overshoot into choppy/quirky?

**Examples of the move:**

Input: "In today's fast-paced world, it's important to note that leveraging data analytics is crucial for businesses looking to navigate the ever-evolving landscape of modern commerce."

Better: "Data analytics gives businesses an edge, if they actually use it."

Input with tricolon + hedging: "The new policy could potentially improve efficiency, reduce costs, and enhance overall employee satisfaction, though it's worth noting that results may vary depending on a wide range of factors."

Better: "The policy should cut costs and speed things up. Whether people are happier with it is less clear."

Do not fabricate to "make it sound more human." If the original is vague because it lacks specifics, the humanized version should still be vague or flag it.

When applying: do the lightest set of edits that removes the machine cadence while preserving meaning, claims, and register. Re-read mentally for uneven heartbeat.