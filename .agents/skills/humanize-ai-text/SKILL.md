---
name: humanize-ai-text
description: >-
  Edit text that carries the stylistic fingerprints of AI generation so it
  reads like clean, natural human writing. Removes AI "tells" (boilerplate
  openers, inflated diction, rhetorical devices at machine frequency,
  over-hedging, em-dash overload, restate-everything conclusions), fixes the
  smooth-but-flat AI cadence, catches prose that performs an idea rather than
  delivering one, and optionally matches a target voice. Use whenever the user
  wants to clean up an AI draft, make text sound less robotic / more human,
  remove AI giveaways or fingerprints, "de-AI" or humanize a draft, or rewrite
  an LLM output into natural prose, even if they don't use the word "skill".
  Do NOT fabricate facts, sources, or personal detail to fake authenticity;
  flag gaps instead.
---

# Humanize AI Text

The goal of this skill is to take prose that *reads* like it came out of a
language model and turn it into prose that reads like a careful person wrote it.
The two failure modes to avoid are (1) leaving the obvious tells in place and
(2) over-correcting into something choppy, gimmicky, or factually invented. Aim
for the lightest set of edits that removes the machine cadence while preserving
the author's meaning, claims, and register.

## What this skill does and does not do

It **does**: strip filler and scaffolding, deflate inflated vocabulary, break up
mechanical rhythm and structure, cut empty hedging, fix punctuation tics, and,
when a target voice is supplied, steer toward it.

It **does not**: invent facts, quotes, statistics, citations, or first-person
anecdotes to make text feel "more human." AI tells are a *style* problem;
fabrication is a different and worse problem. When a passage is vague because it
lacks real specifics, mark it (e.g. `[needs a concrete example]`) rather than
making one up.

A note on use: this improves writing quality and voice. In settings that require
disclosing AI assistance (academic work, certain editorial or professional
contexts), better prose doesn't remove that disclosure obligation. That's the
author's call, not the editor's.

## The test that sits above the catalog

Every specific tell in this document has a shelf life. Once a marker becomes
public knowledge, writers and tools route around it and the marker stops
carrying information. Word lists in particular go stale within months.

What does not go stale is the habit underneath: **a model reaching for the
appearance of depth in the place where a real idea belongs.** Before touching
any sentence, put two questions to each paragraph.

1. Is there a concrete claim here?
2. Is there something the writer has committed to that could turn out to be
   wrong?
When the answer to both is no, and the polished rhythm has done all the work on
its own, the paragraph has a hollow centre. **A hollow centre is not an editing
problem.** Deleting the em-dashes and swapping *leverage* for *use* will produce
a well-dressed void. Flag it (`[no claim here; what is the author actually
asserting?]`) and leave the surgery to the author.

This is also the check that catches non-AI writing worth distrusting: the
corporate press release, the strategically vague email, the politician's
paragraph. Same question, same answer.

## Why the tells exist

Two mechanisms, working together.

**Statistical.** A model generates text by repeatedly predicting the most
probable next token given everything before it. The most probable token is the
one that appeared most often in similar contexts during training. This pulls
prose toward the *average* of everything the model read, and that average has a
definite style: particular sentence shapes, particular transitions, particular
rhetorical moves.

**Preferential.** Models are tuned on human ratings, and when researchers have
run the comparison directly, raters score the more ornate, buzzword-heavy
version of a passage as the better writing. We reward prose that sounds
impressive, so models produce more of it.

The practical consequence: *sounds impressive* is evidence against a passage,
not for it. Treat polish as a symptom to investigate.

## Process

1. **Read for meaning first.** Identify the core claims and the intended
   register (casual blog, technical memo, marketing copy, fiction, email). Edits
   must not flatten a formal document into chat-speak or inflate a quick note.
2. **Run the hollow-centre test** on each paragraph before editing any of them.
   Paragraphs that fail get flagged, not polished.
3. **Ask for or infer a target voice** if one matters and isn't given. A single
   sentence of the author's own writing is worth more than any rule here. Mirror
   its sentence length, vocabulary level, and punctuation habits.
4. **Pass through the tell catalog below**, fixing top-down: scaffolding and
   structure first, then diction, then rhythm, then punctuation. Big structural
   tells dominate the "feels like AI" impression more than word choice does.
5. **Count, don't blacklist.** No single marker convicts. Density does. A short
   passage reaching for four or five at once is the signal; one tricolon in a
   thousand words is a writer doing their job.
6. **Re-read aloud (mentally).** Human prose has an uneven heartbeat. If every
   sentence lands at the same length and the same calm confidence, vary it.
7. **Output the edited text.** Don't append a changelog unless asked. If you had
   to flag missing specifics or hollow centres, list those flags briefly after
   the text.

## The tell catalog

### 1. Boilerplate and scaffolding (highest priority, cut these)

These are the loudest giveaways. Delete them outright; they almost never carry
information.

- Openers: "Certainly!", "Of course!", "Great question", "I hope this helps",
  "Let's dive in", "Let's explore", "In today's fast-paced world", "In the realm
  of", "When it comes to ...", "Buckle up".
- Closers: "In conclusion", "All in all", "At the end of the day", and the
  conclusion paragraph that restates everything already said. End on the last
  real point instead.
- Self-narration: "As an AI", "It's important to note that", "It's worth noting
  that", "It's crucial to understand", "Rest assured". Usually the sentence works
  better with the throat-clearing removed entirely.

### 2. Inflated diction (deflate to plain words)

**Read this list as examples of a habit, not as a blacklist.** It was accurate
when written and will decay. A draft that scrupulously avoids *delve* while
still reaching for impressive-sounding words is the same draft.

Two clusters, and the second is the one people miss.

**The Latinate cluster.** AI reaches for "impressive" words where plain ones read
as more human. Replace when the plain word loses nothing:

- leverage → use; utilize → use; delve into → look at / dig into;
  foster → build / encourage; harness → use; underscore → show / stress;
  embark on → start; facilitate → help.
- robust, seamless, comprehensive, holistic, myriad, plethora, multifaceted,
  intricate, meticulous → usually deletable or replaceable with a specific
  number or detail.
- crucial, pivotal, vital, key, essential → reserve for when something truly is.
  AI sprays these on everything, which drains them of force.
- realm, landscape, tapestry, testament, journey, ever-evolving, game-changer,
  cutting-edge → near-automatic red flags. Rewrite the sentence.

**The sincerity cluster.** These sound warm, plain and candid, which is exactly
why they slip past. They perform frankness in place of saying something frank:

- genuinely, honestly, truly, meaningful, straightforward, deeply, profound
- *signal* as a verb ("this signals a shift"), *speaks to*, *at its core*,
  *the real question is*, *what's striking here is*

Almost always deletable. "Honestly, this is a meaningful and straightforward way
to look at a rich subject" says nothing that "This is a good way to look at the
subject" doesn't, and the shorter version doesn't ask for credit.

### 3. Rhetorical devices at the wrong frequency

The devices below are not faults. Every one of them is a mark of good prose in
skilled hands, and each has a long history. **The device has not changed. The
frequency has, and frequency is what kills it.** A skilled writer deploys one at
a chosen moment, against deliberately uneven surrounding prose, so it lands with
force. A model reaches for it in every paragraph, and the reader stops feeling
any single instance.

So the edit is almost never "remove the device." It is: **keep the best one, cut
the rest, and make sure the one you kept sits where it earns its landing.**

- **Tricolon** (the rule of three). Three carries a completeness that two does
  not, and four begins to feel like a list, which is why it has been a mark of
  good oratory since the Romans. A model defaults to three whether or not the
  content divides into three parts. Test the third item: if it adds no new idea
  and is there because the rhythm wanted it, cut it and let the pair stand.
- **Antithesis** (opposed ideas in matched grammar). "Ask not what your country
  can do for you." Real force, once. When every paragraph pivots on a tidy
  opposition, none of them land. Watch for the escalation variants: "It's not
  just X, it's Y", "X isn't about Y, it's about Z", "Not only ... but also".
  One per document at most.
- **Parallelism and isocolon** (matched structures, clauses mirrored in length
  and rhythm). This one is nearly invisible and worth learning to see. Adjectives
  cluster in pairs (*careful and considered*, *clear and compelling*). Clauses
  come out the same length. The prose settles into a relentless evenness that is
  pleasant the way background music is pleasant, meaning it doesn't ask you to
  feel anything in particular. Fix by breaking a limb: cut one adjective from
  each pair, and let one clause run long or stop short.

### 4. Structural tics

- **Uniform blocks.** Every paragraph three sentences, every sentence medium
  length. Merge some, split others, leave a one-line paragraph for emphasis.
- **Reflexive bulleting and bold labels.** Not every list should be bullets, and
  "**Term:** explanation" run down a page is a dead giveaway. Convert to prose
  where the content flows.
- **Signposting:** "Firstly / Secondly / Finally", "Now, let's turn to". Cut;
  good order makes signposts unnecessary.

### 5. Hedging and the balanced non-answer

AI hedges to avoid being wrong and both-sides to avoid taking a position, because
the safest output on a hard question is one that grants something to every side.
Humans writing with intent commit.

- Strip stacked qualifiers: "It could potentially perhaps sometimes ..." Say it.
- Kill the reflexive "While it's true that A, it's also important to consider B"
  when the author actually has a view. Let them have it.
- Remove empty universals: "a wide range of", "various", "numerous", "countless".
  Give the actual number or example, or cut.
- **The balanced non-answer** is the hardest tell to catch on a quick read, and
  the most important. Several paragraphs of careful, even-handed analysis,
  acknowledging complexity, granting merit to each perspective, arriving at a
  conclusion so hedged it commits to nothing. On a first pass it reads like
  considered judgement. Go looking for the position the writer has taken and
  there isn't one. The giveaway sentence: *"Ultimately, whether this is the
  right choice depends on a range of factors, and reasonable people may
  disagree."* This is the hollow centre wearing a suit. Flag it. Do not smooth it.

### 6. Punctuation and typography

- **Em-dash overload**, especially the paired "—like this—" interruption, two or
  three times a paragraph. Keep at most one for genuine effect; otherwise use
  commas, parentheses, or a full stop.
- The "...; however, ..." semicolon-pivot construction, repeated. Break into two
  sentences.
- Title-Casing Every Header, emoji bullets, and decorative arrows are tells in
  most registers.

### 7. Rhythm and stance

This is the subtle one and matters most once the obvious tells are gone. AI prose
is *smooth, even, and uniformly confident*. Human prose has texture:

- Vary sentence length hard. Follow a long, clause-laden sentence with a short
  one. Three words is a sentence.
- Allow a fragment. Allow a sentence to start with "And" or "But".
- Let the author sound like they have a specific, slightly uneven point of view
  rather than a balanced encyclopedia entry.

## Examples

**Example 1 — scaffolding + inflation**

Input: "In today's fast-paced world, it's important to note that leveraging data
analytics is crucial for businesses looking to navigate the ever-evolving
landscape of modern commerce."

Output: "Data analytics gives businesses an edge, if they actually use it."

**Example 2 — tricolon + hedging**

Input: "The new policy could potentially improve efficiency, reduce costs, and
enhance overall employee satisfaction, though it's worth noting that results may
vary depending on a wide range of factors."

Output: "The policy should cut costs and speed things up. Whether people are
happier with it is less clear."

*Note the third limb went first. Efficiency and cost are one idea stated twice;
satisfaction was the rhythm talking.*

**Example 3 — bullet/bold reflex collapsed to prose**

Input:
"**Speed:** The engine is fast.
**Reliability:** It rarely fails.
**Cost:** It is affordable."

Output: "The engine is fast, rarely fails, and doesn't cost much."

**Example 4 — parallelism, unbalanced**

Input: "The framework is powerful and flexible, the documentation is clear and
thorough, and the community is active and welcoming."

Output: "The framework is flexible. Documentation is unusually clear for a
project this young, and the community answers questions within a day."

*Three matched pairs became one short sentence and one specific one. Note that
the specificity is real, or would need a flag: `[is a day right?]`*

**Example 5 — hollow centre, flagged rather than edited**

Input: "The rise of AI in the workplace represents a profound shift, and while
the opportunities are significant, the challenges are equally real. Ultimately,
organisations that thoughtfully navigate this transition will be best positioned
for success."

Output: `[No claim here. Both sentences are true of any technology and commit to
nothing that could be wrong. What does the author actually think happens to
these jobs, at this company, by when?]`

*Cutting "profound", "significant" and "ultimately" would leave a shorter
paragraph that still says nothing. Don't polish a void.*

**Example 6 — preserve register (don't over-humanize)**

A formal compliance memo should lose its AI tells but stay formal. Cut "it's
important to note" and the em-dashes; do **not** add fragments, "And"-openers, or
casual asides. Naturalness is register-appropriate, not uniformly breezy.

## Restraint check before returning

- Did I keep every real claim and fact intact?
- Did I avoid inventing any specific, quote, or anecdote? (If a gap remained, did
  I flag it instead of filling it?)
- Did I flag hollow centres rather than dressing them up?
- Of the rhetorical devices I kept, is each one sitting where it earns its
  landing, with uneven prose around it?
- Is the register still appropriate to the document's purpose?
- Did I stop at natural, or did I overshoot into choppy/quirky-for-its-own-sake?

If the edit reads like a *different person with a tic* rather than *the same
content written by a clear human*, dial it back.

---

*On the decay of word lists.* Kobak et al., "Delving into LLM-assisted writing in
biomedical publications through excess vocabulary" (arXiv 2406.07016; Science
Advances), tracked style-word frequency across roughly 14 million PubMed
abstracts and found an abrupt post-2022 jump in words like *delve* and
*underscore*. Geng and Trotta, "Human-LLM Coevolution: Evidence from Academic
Writing" (Findings of ACL 2025), then found *delve* falling away in arXiv
abstracts soon after it was named in early 2024, while less conspicuous
model-favoured words such as *significant* kept climbing. Authors adapted; the
habit did not. Assume the same of every list above.
