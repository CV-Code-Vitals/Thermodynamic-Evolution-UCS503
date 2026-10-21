Final Deployment Checklist — Thermodynamic-Evolution-UCS503

This checklist gathers the final items to confirm before pushing to GitHub and deploying to AWS (EC2 + RDS + S3). Follow each step and only push when everything in the "Pre-push verification" section is complete.

1) Pre-push verification (local)
- [ ] Run Go static checks and build
  - cd api
  - go mod download
  - go vet ./...
  - go build ./...
- [ ] Run analysis backend checks
  - cd go-backend
  - go vet ./...
  - go test ./...
  - go build ./...
- [ ] Run frontend checks and build
  - cd admin-portal
  - npm ci
  - npm run build
- [ ] Run linter/tests if you have them (add as needed)

2) Repository hygiene
- [ ] Remove any sensitive data from the working tree (no .env with secrets, no AWS keys, no PEM files)
- [ ] Ensure archives.json does not contain secrets. If it contains paths/URLs that are private, consider removing before push or migrating to DB.
- [ ] Update top-level README with a short note about deployment (optional)

3) GitHub secrets to add
- GHCR_PAT
- VITE_API_BASE
- EC2_HOST
- EC2_USER
- EC2_SSH_PRIVATE_KEY (SSH private key contents)
- Optional EC2_SSH_PORT (defaults to 22)
- Repository variable `EC2_DEPLOY_ENABLED=true` (required to run the EC2 deploy job)
- (Optional) S3_PRESIGN_EXPIRE_HOURS

4) CI / Workflows
- [ ] Confirm .github/workflows/ci-go.yml exists and triggers on api changes
- [ ] Confirm .github/workflows/publish-api-image.yml builds and pushes both backend images to GHCR
- [ ] Confirm .github/workflows/deploy-to-ec2.yml has EC2 secrets configured and `EC2_DEPLOY_ENABLED=true`
- [ ] Confirm .github/workflows/deploy-frontend.yml contains VITE_API_BASE secret usage

5) Server (EC2) preparation checklist
- [ ] EC2 instance up (Ubuntu recommended) with Docker & docker compose installed
- [ ] `.env` file created on server (see `.env.example`) and owned by appropriate user with permissions 600
- [ ] Configure `ADMIN_PASSKEY`, `ANALYSIS_API_TOKEN`, database settings, and S3 settings in the server `.env`
- [ ] Attach an EC2 instance role with least-privilege access to the S3 bucket; do not store AWS access keys on the instance
- [ ] If containers use the EC2 instance role, require IMDSv2 and set the instance metadata response hop limit to 2
- [ ] Allow EC2 to connect to RDS, and restrict the database security group to the EC2 security group
- [ ] Nginx or ALB configured to reverse-proxy to the API on `127.0.0.1:8080` and analysis backend on `127.0.0.1:8081`
- [ ] Keep ports 8080, 8081, and 5432 closed to public ingress
- [ ] SSL cert obtained with certbot or via ACM/load balancer

6) Post-deploy checks (after push & deploy)
- [ ] Curl the health endpoint:
  curl https://api.yourdomain.com/status
- [ ] Upload test file via Admin Portal and confirm S3 object appears (or local uploads dir)
- [ ] Confirm metadata record in RDS (psql) or archives.json updated
- [ ] Confirm download link works (presigned URL or direct object URL)

7) Rollback plan
- [ ] Keep the previous Docker image tag available on GHCR to roll back quickly
- [ ] Backup archives.json (if still using local storage) and a DB snapshot for RDS before migration

Notes
- Do not commit .env or any private keys. Use GitHub Secrets for CI and server-side management tools (AWS Secrets Manager, Parameter Store) for production.
- If using EC2 instance roles, you can avoid storing AWS keys on the server.

If you want, I can run through this checklist with you interactively and help set each secret and run the first push. Tell me which step to begin with.