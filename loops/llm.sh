#!/bin/bash
# Shared OpenAI-first generation adapter for Factory loops. Source, do not execute.

FACTORY_LLM_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FACTORY_REPO_ROOT="$(cd "$FACTORY_LLM_DIR/.." && pwd)"

factory_load_llm_config() {
  local config
  for config in \
    "$FACTORY_REPO_ROOT/.env.local" \
    "$FACTORY_REPO_ROOT/loops/clip-to-post/config.local.env"; do
    if [[ -f "$config" ]]; then
      set -a
      # shellcheck disable=SC1090
      source "$config"
      set +a
    fi
  done
}

factory_try_openai() {
  local prompt_file="$1"
  local out_file="$2"
  [[ -n "${OPENAI_API_KEY:-}" ]] || return 1

  : > "$out_file"
  if python3 "$FACTORY_LLM_DIR/openai_response.py" \
    --prompt-file "$prompt_file" \
    --out "$out_file" && [[ -s "$out_file" ]]; then
    FACTORY_GENERATION_PROVIDER="OpenAI ${FACTORY_OPENAI_MODEL:-gpt-5.6-sol} (${FACTORY_OPENAI_REASONING_EFFORT:-medium})"
    return 0
  fi
  return 1
}

factory_try_grok() {
  local prompt_file="$1"
  local out_file="$2"
  command -v grok >/dev/null 2>&1 || return 1
  [[ -f "$HOME/.grok/auth.json" || -n "${XAI_API_KEY:-}" ]] || return 1

  : > "$out_file"
  if grok --no-auto-update -p "$(<"$prompt_file")" --output-format plain > "$out_file" 2>/dev/null \
    && [[ -s "$out_file" ]]; then
    FACTORY_GENERATION_PROVIDER="Grok fallback"
    return 0
  fi
  return 1
}

factory_generate() {
  local prompt_file="$1"
  local out_file="$2"
  local primary="${FACTORY_LLM_PROVIDER:-openai}"
  local fallback="${FACTORY_LLM_FALLBACK_PROVIDER:-grok}"
  local provider
  local -a providers=("$primary")

  if [[ "$fallback" != "none" && "$fallback" != "$primary" ]]; then
    providers+=("$fallback")
  fi

  for provider in "${providers[@]}"; do
    case "$provider" in
      openai)
        if factory_try_openai "$prompt_file" "$out_file"; then
          return 0
        fi
        ;;
      grok)
        if factory_try_grok "$prompt_file" "$out_file"; then
          return 0
        fi
        ;;
      *)
        echo "ERROR: unsupported FACTORY_LLM_PROVIDER: $provider" >&2
        return 2
        ;;
    esac
  done

  echo "ERROR: no configured Factory LLM provider succeeded" >&2
  return 1
}

factory_load_llm_config
