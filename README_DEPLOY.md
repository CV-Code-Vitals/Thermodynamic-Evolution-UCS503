README - Deploy & Local Run

This document collects the exact commands to run locally, build, and deploy the API and frontend.

Local dev - API (Go)
1) Install Go 1.25
2) From repository root, preferred quick start (Windows PowerShell):
   .\scripts\run-local.ps1

If `ADMIN_PASSKEY` is not set, the Windows local script uses `admin` as a
development-only passkey. Production deployments must set `ADMIN_PASSKEY`
through a server environment variable or GitHub/hosting secret.

The `api/` service is the Admin Portal backend. The separate `go-backend/`
service runs repository analysis and scanning; it requires `ANALYSIS_API_TOKEN`
and should not be used as a drop-in replacement for the Admin Portal API.

Or (cross-platform manual):
   # API
   cd api
   go mod download
   # Ensure ADMIN_PASSKEY is set before running the API (no default). Example in PowerShell:
   #   $env:ADMIN_PASSKEY = "your_local_passkey_here"
   # Or in bash:
   #   export ADMIN_PASSKEY=your_local_passkey_here
   $env:PORT=8080
   go run main.go

   When the frontend is hosted on a different origin, set `CORS_ALLOWED_ORIGINS`
   on the API to the frontend origin (for example, `https://zylus08.github.io`).
   Multiple origins may be comma-separated.

   If running `go-backend/`, also set `ANALYSIS_API_TOKEN`; analysis, scanning,
   deliverables, and uploaded-file routes reject requests without that token.

   # Frontend (run in separate terminal)
   cd admin-portal
   npm ci
   npm run dev

Health check:
   curl http://localhost:8080/status

Local dev - Frontend (Vite)
1) Node.js (16+)
   cd admin-portal
   npm ci
   npm run dev
2) Open browser to Vite dev URL (printed in terminal)
   The Vite dev server proxies /api to http://localhost:8080

Scripts
- Windows: .\scripts\run-local.ps1  (opens two PowerShell windows and starts API+frontend — requires ADMIN_PASSKEY to be set in the environment)
- Unix/macOS: ./scripts/run-local.sh (starts API+frontend in background — requires ADMIN_PASSKEY to be set in the environment)

Build frontend for GitHub Pages (static)
cd admin-portal
# ensure VITE_API_BASE is set (in CI this comes from GitHub Secrets)
npm run build
# output will be in admin-portal/dist

GitHub Pages setup
- In the repository, open Settings > Pages and set the Source to "GitHub Actions".
- Do not select "Deploy from a branch"; the frontend workflow uploads the built site
  as a Pages artifact.
- Add the Actions secret `VITE_API_BASE` with the full HTTPS URL of the deployed API
  (for example, `https://api.yourdomain.com/api`). Do not use `/api` here: GitHub
  Pages has no backend proxy, and production builds now fail instead of publishing
  a portal that sends login requests to the Pages site.
- Configure the API's `CORS_ALLOWED_ORIGINS` environment variable to include
  `https://cv-code-vitals.github.io`.
- After pushing to `main`, wait for the "Deploy Frontend to GitHub Pages" workflow
  to finish before opening `/Thermodynamic-Evolution-UCS503/admin-portal/`.

Build and run Docker image locally (API)
# from repo root
cd api
GOOS=linux GOARCH=amd64 go build -o api_server main.go
# build a Docker image
docker build -t ghcr.io/<owner>/ucs503-api:local -f Dockerfile .
# run locally
docker run -e ADMIN_PASSKEY="your-secure-passkey" -e PORT=8080 -p 8080:8080 ghcr.io/<owner>/ucs503-api:local

Push image to GHCR (manual)
# Tag and push
docker tag ghcr.io/<owner>/ucs503-api:local ghcr.io/<owner>/ucs503-api:latest
# Login
echo $GHCR_PAT | docker login ghcr.io -u <your-github-username> --password-stdin
docker push ghcr.io/<owner>/ucs503-api:latest

Deploy both backends on EC2 with Docker Compose

The `Publish Backend Docker Images` workflow builds and publishes both Go
services to GHCR after backend changes land on `main`. When EC2 deployment is
enabled, the follow-up workflow copies the Compose file and deploys the exact
images built for that commit.

Configure these GitHub repository secrets:
- `EC2_HOST`, `EC2_USER`, and `EC2_SSH_PRIVATE_KEY`
- `GHCR_PAT` with permission to pull the repository's GHCR packages
- Optional `EC2_SSH_PORT` (defaults to `22`)

Set the repository variable `EC2_DEPLOY_ENABLED` to `true` to turn on automatic
deployment. On the EC2 instance, create `~/ucs503/.env` with the production
values from `.env.example` before the first deployment. Use an EC2 instance
profile for S3 access instead of static AWS access keys, and allow the instance
to reach the RDS endpoint. If containers use the instance profile, require
IMDSv2 and set the instance metadata response hop limit to 2. Keep the database
and application ports closed to the public internet.

# On EC2 create folder ~/ucs503, copy docker-compose.ec2.yml and .env
# Populate .env with production values (see .env.example)
cd ~/ucs503
docker compose pull
docker compose up -d

Verify
- curl https://api.yourdomain.com/status
- On EC2, verify the analysis backend with `curl http://127.0.0.1:8081/status`.
- The analysis backend listens on `127.0.0.1:8081`; configure your host reverse
  proxy or load balancer to forward the analysis API there, and do not expose
  port 8081 publicly.
- use the Admin Portal to log in and upload archives
- check S3 console for objects
- check RDS for metadata

Rollback
- Set `IMAGE_TAG` to a previously published commit SHA, then run:
  `IMAGE_TAG=<previous-tag> docker compose -f docker-compose.ec2.yml up -d`

Notes
- Use the provided .env.example as a template on the server (do NOT commit secrets)
- For CI and deploy, use the GitHub workflows already added to this repository
