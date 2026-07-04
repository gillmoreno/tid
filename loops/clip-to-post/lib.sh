#!/bin/bash
# Shared helpers for clip-to-post scripts. Source, do not execute.

set -euo pipefail

CLIP_TO_POST_ROOT="$(cd "$(dirname "${BASH_SOURCE[1]:-${BASH_SOURCE[0]}}")" && pwd)"

draft_dir() {
  local id="$1"
  printf '%s/drafts/%s' "$CLIP_TO_POST_ROOT" "$id"
}

require_draft() {
  local id="$1"
  local dir
  dir="$(draft_dir "$id")"
  if [[ ! -f "$dir/meta.json" ]]; then
    echo "ERROR: draft not found: $dir/meta.json" >&2
    exit 1
  fi
  printf '%s' "$dir"
}

meta_field() {
  local meta="$1"
  local field="$2"
  jq -r --arg f "$field" '.[$f] // empty' "$meta"
}

resolve_clip_path() {
  local dir="$1"
  local meta="$dir/meta.json"
  local clip
  clip="$(meta_field "$meta" "clip_path")"
  if [[ -n "$clip" && -f "$clip" ]]; then
    printf '%s' "$clip"
    return 0
  fi
  if [[ -f "$dir/clip.mp4" ]]; then
    printf '%s/clip.mp4' "$dir"
    return 0
  fi
  echo "ERROR: clip.mp4 not found for draft in $dir" >&2
  exit 1
}

default_draft_id() {
  local url="$1"
  local slug="${2:-}"
  local video_id
  video_id="$(echo "$url" | sed -n 's/.*[?&]v=\([^&]*\).*/\1/p')"
  if [[ -z "$video_id" ]]; then
    video_id="$(echo "$url" | sed 's|.*youtu.be/||' | cut -c1-11)"
  fi
  if [[ -n "$slug" ]]; then
    printf '%s-%s' "$(date +%Y%m%d)" "$slug"
  else
    printf '%s-%s' "$(date +%Y%m%d)" "${video_id:0:8}"
  fi
}