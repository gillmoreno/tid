# MCP Harness for Custom Room Generation

**Goal:** Users connect their own AI (Claude via Cursor/Claude Desktop, Grok, local models, etc.) so the *builder* never pays for LLM generation of rooms.

## Interface (tool)

Expose a tool:

- Name: `generate_custom_room`
- Description: Turn a natural language description into a custom room bundle (inner HTML + script) that works with `initRoomBridge()` inside the Rooms meta shell (and can run standalone with the real kit).
- Input:
  ```json
  {
    "description": "string — what the user wants (e.g. 'beautiful shared grocery list with who bought what and running totals')",
    "style_hints": "optional string",
    "previous_bundle": "optional — for iterative edits"
  }
  ```
- Output (structured):
  ```json
  {
    "title": "Suggested room title",
    "bundle": "<the inner markup + script for the room>",
    "notes": "any warnings or suggestions"
  }
  ```

The system prompt + few-shot examples are exactly the content of `LLM-HARNESS-PROMPT.md`.

## How a user uses it today (MVP)

1. User has Claude/Cursor/etc. connected to an MCP server that exposes this tool (or pastes the prompt).
2. They say: "use the custom-rooms generator: a movie club queue with votes and next-watch picker"
3. Model calls the tool (or user pastes the full prompt + description).
4. They copy the returned `bundle`.
5. Inside the meta shell (or a future "paste bundle" flow) they create a new room and paste the bundle, or the client UI has a "generate" that talks to their local MCP.

## Future

- Real MCP server binary or stdio server that hosts the prompt.
- Optional hosted generation behind auth + metering (secondary path).
- The meta shell can have a "Bring your own key / MCP" settings screen.

This keeps the economics sane: the person who wants the room pays for the tokens (or runs local).

See also: PLAN.md, LLM-HARNESS-PROMPT.md
