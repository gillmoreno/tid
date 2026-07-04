#!/usr/bin/env bash
# Build TID frontend and deploy to Cloudflare Pages on the-idea-guy.com
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG="${ROOT}/deploy/cloudflare.json"
ENV_FILE="${ROOT}/deploy/.env"

usage() {
  cat <<'EOF'
Usage: ./scripts/deploy-cloudflare.sh [options]

Deploy the TID frontend to Cloudflare Pages on the-idea-guy.com.

Options:
  --init          Create Pages project and attach apex custom domain (first time)
  --skip-build    Upload existing frontend/dist without rebuilding
  --dry-run       Print actions without calling Cloudflare
  -h, --help      Show this help

Environment (deploy/.env or shell):
  CLOUDFLARE_API_TOKEN   Required
  CLOUDFLARE_ACCOUNT_ID  Required for --init and domain API calls

Examples:
  ./scripts/deploy-cloudflare.sh --init
  ./scripts/deploy-cloudflare.sh
EOF
}

log() { printf '→ %s\n' "$*"; }
die() { printf '✘ %s\n' "$*" >&2; exit 1; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "Missing required command: $1"
}

cf_api() {
  local method="$1"
  local path="$2"
  local data="${3:-}"
  local url="https://api.cloudflare.com/client/v4${path}"
  if [[ -n "$data" ]]; then
    curl -sfS -X "$method" "$url" \
      -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
      -H "Content-Type: application/json" \
      --data "$data"
  else
    curl -sfS -X "$method" "$url" \
      -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
      -H "Content-Type: application/json"
  fi
}

expand_env_value() {
  local template="$1"
  local result="$template"
  local var
  while [[ "$result" =~ \$\{([A-Za-z_][A-Za-z0-9_]*)\} ]]; do
    var="${BASH_REMATCH[1]}"
    result="${result//\$\{${var}\}/${!var:-}}"
  done
  printf '%s' "$result"
}

get_production_branch() {
  if git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git -C "$ROOT" branch --show-current
  else
    printf '%s' "main"
  fi
}

get_zone_id() {
  local apex="$1"
  if [[ -n "${CLOUDFLARE_ZONE_ID:-}" ]]; then
    printf '%s' "$CLOUDFLARE_ZONE_ID"
    return 0
  fi
  cf_api GET "/zones?name=${apex}&status=active" | jq -r '.result[0].id // empty'
}

get_pages_subdomain() {
  local project="$1"
  cf_api GET "/accounts/${CLOUDFLARE_ACCOUNT_ID}/pages/projects/${project}" \
    | jq -r '.result.subdomain // empty'
}

ensure_dns_cname() {
  local zone_id="$1"
  local record_name="$2"
  local target="$3"
  local existing
  existing="$(cf_api GET "/zones/${zone_id}/dns_records?type=CNAME&name=${record_name}" \
    | jq -r '.result[0].id // empty')"
  if [[ -n "$existing" ]]; then
    log "Updating DNS CNAME ${record_name} → ${target}"
    cf_api PUT "/zones/${zone_id}/dns_records/${existing}" \
      "{\"type\":\"CNAME\",\"name\":\"${record_name}\",\"content\":\"${target}\",\"proxied\":true}" \
      >/dev/null
    return 0
  fi

  local a_existing
  a_existing="$(cf_api GET "/zones/${zone_id}/dns_records?type=A&name=${record_name}" \
    | jq -r '.result[0].id // empty')"
  if [[ -n "$a_existing" ]]; then
    log "Removing old A record for ${record_name}"
    cf_api DELETE "/zones/${zone_id}/dns_records/${a_existing}" >/dev/null || true
  fi

  log "Creating DNS CNAME ${record_name} → ${target}"
  cf_api POST "/zones/${zone_id}/dns_records" \
    "{\"type\":\"CNAME\",\"name\":\"${record_name}\",\"content\":\"${target}\",\"proxied\":true}" \
    >/dev/null
}

DO_INIT=0
SKIP_BUILD=0
DRY_RUN=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --init) DO_INIT=1; shift ;;
    --skip-build) SKIP_BUILD=1; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    *) die "Unknown option: $1" ;;
  esac
done

require_cmd jq
require_cmd curl
require_cmd wrangler
[[ -f "$CONFIG" ]] || die "Config not found: $CONFIG"

if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  set -a
  source "$ENV_FILE"
  set +a
fi

[[ -n "${CLOUDFLARE_API_TOKEN:-}" ]] || die "Set CLOUDFLARE_API_TOKEN (see deploy/.env.example)"
export CLOUDFLARE_API_TOKEN

APEX_DOMAIN="$(jq -r '.apex_domain' "$CONFIG")"
PROJECT_NAME="$(jq -r '.project_name' "$CONFIG")"
SUBDOMAIN="$(jq -r '.subdomain' "$CONFIG")"
BUILD_DIR="$(jq -r '.build_dir' "$CONFIG")"
OUTPUT_DIR="$(jq -r '.output_dir' "$CONFIG")"
BUILD_COMMAND="$(jq -r '.build_command' "$CONFIG")"

if [[ -n "$SUBDOMAIN" && "$SUBDOMAIN" != "null" ]]; then
  SITE_FQDN="${SUBDOMAIN}.${APEX_DOMAIN}"
else
  SITE_FQDN="${APEX_DOMAIN}"
fi

BUILD_DIR_ABS="${ROOT}/${BUILD_DIR}"
OUTPUT_DIR_ABS="${ROOT}/${OUTPUT_DIR}"

log "Project: $PROJECT_NAME"
log "Target: https://${SITE_FQDN}/"

if [[ "$SKIP_BUILD" -eq 0 ]]; then
  log "Building in ${BUILD_DIR}…"
  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] would run: $BUILD_COMMAND"
  else
    cd "$BUILD_DIR_ABS"
    while IFS=$'\t' read -r key template; do
      [[ -n "$key" && "$key" != "null" ]] || continue
      value="$(expand_env_value "$template")"
      export "$key=$value"
    done < <(jq -r '.build_env // {} | to_entries[] | [.key, .value] | @tsv' "$CONFIG")
    eval "$BUILD_COMMAND"
    cd "$ROOT"
  fi
else
  log "Skipping build (--skip-build)"
fi

[[ "$DRY_RUN" -eq 1 || -d "$OUTPUT_DIR_ABS" ]] || die "Build output missing: $OUTPUT_DIR_ABS"

if [[ "$DO_INIT" -eq 1 ]]; then
  [[ -n "${CLOUDFLARE_ACCOUNT_ID:-}" ]] || die "Set CLOUDFLARE_ACCOUNT_ID for --init"

  PRODUCTION_BRANCH="$(get_production_branch)"
  [[ -n "$PRODUCTION_BRANCH" ]] || die "Could not detect git branch"

  if wrangler pages project list 2>/dev/null | grep -qF "${PROJECT_NAME}"; then
    log "Pages project '$PROJECT_NAME' already exists"
  elif [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] would create Pages project: $PROJECT_NAME"
  else
    log "Creating Pages project '${PROJECT_NAME}' (branch: ${PRODUCTION_BRANCH})…"
    wrangler pages project create "$PROJECT_NAME" --production-branch="$PRODUCTION_BRANCH"
  fi

  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] would attach custom domain: $SITE_FQDN"
  else
    log "Attaching custom domain ${SITE_FQDN}…"
    if cf_api POST "/accounts/${CLOUDFLARE_ACCOUNT_ID}/pages/projects/${PROJECT_NAME}/domains" \
      "{\"name\":\"${SITE_FQDN}\"}" >/dev/null 2>&1; then
      log "Custom domain registered on Pages project"
    else
      log "Domain may already be attached (or check token permissions)"
    fi

    ZONE_ID="$(get_zone_id "$APEX_DOMAIN")"
    [[ -n "$ZONE_ID" ]] || die "Cloudflare zone not found for ${APEX_DOMAIN}"
    PAGES_TARGET="$(get_pages_subdomain "$PROJECT_NAME")"
    [[ -n "$PAGES_TARGET" ]] || die "Could not resolve Pages subdomain for ${PROJECT_NAME}"
    ensure_dns_cname "$ZONE_ID" "$SITE_FQDN" "$PAGES_TARGET"
  fi
fi

if [[ "$DRY_RUN" -eq 1 ]]; then
  log "[dry-run] would deploy: wrangler pages deploy $OUTPUT_DIR --project-name=$PROJECT_NAME"
  exit 0
fi

PRODUCTION_BRANCH="$(get_production_branch)"
[[ -n "$PRODUCTION_BRANCH" ]] || die "Could not detect git branch"

log "Deploying to Cloudflare Pages (branch: ${PRODUCTION_BRANCH})…"
CLOUDFLARE_API_TOKEN="$CLOUDFLARE_API_TOKEN" wrangler pages deploy "$OUTPUT_DIR_ABS" \
  --project-name="$PROJECT_NAME" \
  --branch="$PRODUCTION_BRANCH" \
  --commit-dirty=true

log "Done."
log "  https://${PROJECT_NAME}.pages.dev"
log "  https://${SITE_FQDN}/"