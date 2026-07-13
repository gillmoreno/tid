#!/bin/sh
set -eu

umask 077
capacity=""
ttl="24h"

usage() {
  printf '%s\n' 'Usage: mint-token.sh --capacity 2..50 [--ttl 24h]' >&2
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --capacity)
      [ "$#" -ge 2 ] || { usage; exit 2; }
      capacity="$2"
      shift 2
      ;;
    --ttl)
      [ "$#" -ge 2 ] || { usage; exit 2; }
      ttl="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage
      exit 2
      ;;
  esac
done

case "$capacity" in
  ''|*[!0-9]*) usage; exit 2 ;;
esac
[ "$capacity" -ge 2 ] && [ "$capacity" -le 50 ] || {
  printf '%s\n' 'Capacity must be between 2 and 50.' >&2
  exit 2
}

command -v go >/dev/null 2>&1 || { printf '%s\n' 'Go is required.' >&2; exit 1; }
command -v pbcopy >/dev/null 2>&1 || { printf '%s\n' 'pbcopy is required on macOS.' >&2; exit 1; }

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
default_repo_root=$(CDPATH= cd -- "$script_dir/../../../.." && pwd)
repo_root=${ROOMWORKS_REPO_ROOT:-$default_repo_root}
seed_file=${ROOMWORKS_CREATOR_SIGNING_SEED:-"$HOME/.config/roomworks/creator-signing.seed"}

[ -f "$repo_root/signaling/go.mod" ] || {
  printf 'Roomworks signaling module not found at %s\n' "$repo_root/signaling" >&2
  exit 1
}

cd "$repo_root/signaling"
exec go run ./cmd/roomworks-token mint \
  --capacity "$capacity" \
  --ttl "$ttl" \
  --seed-file "$seed_file"
