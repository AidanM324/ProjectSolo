# Distributed Job Queue & Scheduler

A fault-tolerant, cloud-native distributed job processing platform built to explore
production-grade backend and infrastructure engineering — from application logic to
containerization, orchestration, infrastructure-as-code, CI/CD, and observability.

## What it does

Clients submit jobs via an HTTP API. Jobs are queued in RabbitMQ and picked up by a pool
of Go worker processes for execution. The system guarantees no job is lost or silently
duplicated, even if a worker crashes mid-job — via manual acknowledgements, retries with
backoff, a dead-letter queue, and idempotency keys.

## Why this project

Most portfolio projects prove you can build a feature. This one is aimed at proving
something different: that the system keeps working correctly when things go wrong —
worker crashes, retries, backpressure — which is closer to what real backend/platform
engineering roles actually care about.

## Architecture

*(diagram to be added once Phase 1–3 are complete)*

- **API service** (Go) — accepts job submissions, validates, publishes to RabbitMQ
- **Worker service** (Go) — consumes jobs, executes them, reports status
- **RabbitMQ** — job transport, manual ack/nack, dead-letter queue
- **PostgreSQL** — job state and history
- **Kubernetes (k3s)** — orchestration, running on a Hetzner VPS
- **Terraform** — provisions the VPS from scratch
- **GitHub Actions** — CI/CD: test → build → push image → deploy
- **Prometheus + Grafana** — live metrics: queue depth, throughput, failure rate

## Build phases

- [ ] **Phase 1** — Core API + worker logic running locally (Go, RabbitMQ, Postgres)
- [ ] **Phase 2** — Reliability: retries, dead-letter queue, idempotency keys, crash recovery
- [ ] **Phase 3** — Containerize everything (Docker + docker-compose)
- [ ] **Phase 4** — Provision infrastructure with Terraform
- [ ] **Phase 5** — Deploy to Kubernetes (k3s)
- [ ] **Phase 6** — CI/CD pipeline + Prometheus/Grafana observability

## Design decisions & trade-offs

*(to be filled in as the project progresses — this is often the most-read section
by anyone technical reviewing the repo)*

## Benchmarks

*(to be added in Phase 6 — throughput, recovery time after a killed worker, etc.)*

## Testing

Manually tested by submitting 3 concurrent jobs via curl; verified each transitioned pending → running → done in Postgres, and all 3 emails were delivered correctly via Mailtrap, confirming the full pipeline (API → Postgres → RabbitMQ → Worker → SMTP) works end-to-end.

During multi-job testing, one job failed due to Mailtrap's SMTP rate limit being hit when submitting several jobs in rapid succession — an expected real-world failure mode that motivates the retry/backoff logic built in Phase 2.

##Phase 2 - Issue 8, 9

Tested retry/backoff logic by deliberately breaking SMTP credentials. The system correctly retried with exponential backoff (2s, 4s, 8s, 16s, 32s), and after Mailtrap temporarily locked the account following repeated failed logins, the job continued retrying against the real (correct) credentials — still failing due to the account-level lockout, not the code — until exhausting 5 retries and correctly routing to the dead-letter queue. This demonstrated the system fails gracefully and predictably even when the downstream failure mode is unexpected.'

## Running locally

*(instructions to be added at the end of Phase 3)*

## Author

Aidan Malone — [GitHub](#) · [LinkedIn](#)
