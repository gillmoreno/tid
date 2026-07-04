# TID — dev and deploy commands (Roger-style)

deploy:
    docker compose -f docker-compose.yml down
    docker compose -f docker-compose.yml up -d --build

dev:
    @docker compose -f docker-compose.dev.yml up -d --build
    @BACKEND_PORT=$(docker compose -f docker-compose.dev.yml port go-slice 8080 | sed -E 's/.*:([0-9]+)$/\1/') ; \
    if [ -z "$BACKEND_PORT" ]; then \
      echo "Error: failed to resolve mapped backend port for go-slice:8080"; \
      exit 1; \
    fi ; \
    cd frontend && VITE_API_URL=/api VITE_API_PROXY_TARGET="http://localhost:$BACKEND_PORT" npm run dev

dev-down:
    docker compose -f docker-compose.dev.yml down

dev-local:
    @echo "Terminal 1: cd go-backend && go run ./cmd/server"
    @echo "Terminal 2: cd frontend && VITE_API_URL=/api VITE_API_PROXY_TARGET=http://localhost:8080 npm run dev"

deploy-cloudflare *args:
    bash scripts/deploy-cloudflare.sh {{args}}

# Clip → Post loop (semi-automated X posting prep)
prepare-post draft:
    bash loops/clip-to-post/prepare-post.sh --draft {{draft}}