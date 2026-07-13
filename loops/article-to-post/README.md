# Article Factory Loop

Article URL → clean text → AI analysis → standalone X post candidates.

Feeds the **Article Factory** app at http://localhost:5180/factory/articles

## Pipeline

```
Article URL
  → fetch-article.sh   (extract_article.py → drafts/{source-id}/article.txt + article.json)
  → analyze.sh         (biases + prompt + mentions → analysis.json: post candidates, no timestamps)
  → Go API inserts candidates into SQLite
  → React /factory/articles UI lists them for edit / refine / schedule
  → prepare-post.sh    (clipboard + Chrome — text only, no clip)
```

Biases + the `@` mentions dictionary are shared with the clip factory (edit in the
**Sources** page / clip factory settings). The article **analysis prompt** is
separate and editable in the Article Factory settings panel (stored in SQLite under
the `article-analysis` prompt template).

## Setup

Extraction uses [trafilatura](https://trafilatura.readthedocs.io):

```bash
pip3 install -r loops/article-to-post/requirements.txt
```

If trafilatura is missing, `extract_article.py` falls back to a basic
urllib + tag-stripping extractor (lower quality).

## Scripts

| Script | Role |
|--------|------|
| `fetch-article.sh` | URL → `drafts/{id}/article.txt` + `article.json` (`{title,text}`) |
| `extract_article.py` | trafilatura extraction (with fallback) |
| `analyze.sh` | article + biases + prompt + mentions → `analysis.json` (grok, dev fallback) |
| `build_system_prompt.py` | Assembles the analyzer system prompt |
| `rewrite.sh` | Apply Gil's lens + an instruction to one post |
| `prepare-post.sh` | Copy post text to clipboard + open X compose in Chrome |
| `prompt.default.md` | Seed analysis instruction |

Never commit `drafts/` (gitignored).
