# TID — dev and deploy commands (Roger-style)

deploy:
    docker compose -f docker-compose.yml down
    docker compose -f docker-compose.yml up -d --build

# Local Go API + Vite. Post Factory runs here (needs host scripts: yt-dlp, Chrome, Finder).
dev:
    bash scripts/dev.sh

dev-down:
    @lsof -ti :8080 | xargs kill -9 2>/dev/null || true
    @lsof -ti :5180 | xargs kill -9 2>/dev/null || true
    @docker compose -f docker-compose.dev.yml down 2>/dev/null || true

dev-docker:
    @docker compose -f docker-compose.dev.yml up -d --build
    @BACKEND_PORT=$(docker compose -f docker-compose.dev.yml port go-slice 8080 | sed -E 's/.*:([0-9]+)$/\1/') ; \
    if [ -z "$BACKEND_PORT" ]; then \
      echo "Error: failed to resolve mapped backend port for go-slice:8080"; \
      exit 1; \
    fi ; \
    cd frontend && VITE_API_URL=/api VITE_API_PROXY_TARGET="http://localhost:$BACKEND_PORT" npm run dev

deploy-cloudflare *args:
    bash scripts/deploy-cloudflare.sh {{args}}

# Clip → Post loop (semi-automated X posting prep)
prepare-post draft:
    bash loops/clip-to-post/prepare-post.sh --draft {{draft}}

# Post Factory CLI (same backend as /factory UI)
factory *args:
    cd go-backend && go run ./cmd/factory {{args}}

factory-ingest url title="" podcast="":
    cd go-backend && go run ./cmd/factory ingest --url "{{url}}" --title "{{title}}" --podcast "{{podcast}}"

factory-analyze source:
    cd go-backend && go run ./cmd/factory analyze --source {{source}}

factory-tick:
    cd go-backend && go run ./cmd/factory tick

# Agent one-shot: URL → analyze → candidates in SQLite + /factory UI
factory-run url title="" podcast="":
    bash loops/clip-to-post/run-factory.sh --url "{{url}}" --title "{{title}}" --podcast "{{podcast}}"