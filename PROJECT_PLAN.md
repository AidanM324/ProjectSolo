# Project Plan — Distributed Job Queue & Scheduler

## Goal

Build a fault-tolerant, cloud-native distributed job processing platform as a portfolio
project demonstrating backend systems depth: concurrency, reliability under failure,
containerization, orchestration, infrastructure-as-code, CI/CD, and observability.

## How this is tracked

- One GitHub Issue per task below. Reference the issue number in commits,
  e.g. `git commit -m "Add exponential backoff retry logic (#12)"`.
- One GitHub Projects board (Kanban: To Do / In Progress / Done) with each issue as a card.
- This file's checklist is updated as phases complete — treat it as the source of truth
  for overall progress at a glance.

## Phase 1 — Core logic locally
- [Done] Scaffold Go module structure (api/, worker/, shared/)
- [Done] API: HTTP endpoint to accept a job submission
- [Done] API: publish job to RabbitMQ
- [Done] Worker: consume job from RabbitMQ
- [ ] Worker: execute a real task (pick one: image resize / email send / file processing)
- [ ] Persist job state (pending/running/done/failed) to Postgres
- [ ] Manual end-to-end test: submit a job, see it processed, check DB state

## Phase 2 — Reliability logic
- [ ] Manual ack/nack so an in-flight job isn't lost on worker crash
- [ ] Retry with exponential backoff on failure
- [ ] Dead-letter queue for jobs that exceed max retries
- [ ] Idempotency keys to prevent duplicate job creation
- [ ] Test: kill a worker mid-job, confirm no job loss or duplication

## Phase 3 — Containerize
- [ ] Dockerfile for API service
- [ ] Dockerfile for worker service
- [ ] docker-compose.yml: API + workers + RabbitMQ + Postgres
- [ ] Verify full stack starts with one command

## Phase 4 — Infrastructure as Code
- [ ] Terraform config to provision a VPS (Hetzner or similar)
- [ ] `terraform apply` produces a running server from nothing
- [ ] Document the Terraform setup in README

## Phase 5 — Kubernetes deployment
- [ ] Install k3s on the VPS
- [ ] Write Deployment manifests (API, workers, RabbitMQ, Postgres)
- [ ] Write Service manifests
- [ ] ConfigMaps/Secrets for configuration
- [ ] Verify worker Deployment can scale replicas up/down

## Phase 6 — CI/CD & observability
- [ ] GitHub Actions: run tests on push
- [ ] GitHub Actions: build Docker images, push to GitHub Container Registry
- [ ] GitHub Actions: deploy to k3s cluster automatically
- [ ] Prometheus scraping key metrics (queue depth, success/failure rate, latency)
- [ ] Grafana dashboard for live visualization

## Wrap-up
- [ ] Architecture diagram in README
- [ ] Design decisions & trade-offs section written up
- [ ] Benchmarks recorded (throughput, recovery time after killed worker)
- [ ] Demo video/GIF recorded
- [ ] CV line finalized
