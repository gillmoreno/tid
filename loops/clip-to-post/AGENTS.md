# Clip → Post — Agent Instructions

## Default workflow (Gil's preference)

Use the **semi-automated prepare-post flow**:

```bash
cd loops/clip-to-post
./prepare-post.sh --draft <id>
```

This copies text, opens X compose in Chrome Default, opens Finder for the clip. Gil posts manually.

**Do not** skip to full automation unless Gil explicitly asks.

## Composability rules

- `meta.json` = source of truth per draft
- `taste.md` + `draft-copy.sh` = how text is written
- `prepare-post.sh` = mechanical only (clipboard + Chrome + Finder); reads `meta.json`, no hardcoded content
- Never commit `drafts/`, `.env`, or video files

## Creating a new draft

```bash
./draft.sh --url URL --start TIME --end TIME --speaker NAME --podcast NAME [--slug slug]
```

Review `post.txt`, update `meta.json` if needed, set `status` to `approved`, then `prepare-post.sh`.

## Reference

Full docs: `README.md` in this folder.  
Project context: `../../docs/CONCEPT.md`, `../../docs/HERMES-CLIP-TO-POST.md`.