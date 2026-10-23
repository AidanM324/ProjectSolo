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

- [Done] **Phase 1** — Core API + worker logic running locally (Go, RabbitMQ, Postgres)
- [Done] **Phase 2** — Reliability: retries, dead-letter queue, idempotency keys, crash recovery
- [ ] **Phase 3** — Containerize everything (Docker + docker-compose)
- [ ] **Phase 4** — Provision infrastructure with Terraform
- [ ] **Phase 5** — Deploy to Kubernetes (k3s)
- [ ] **Phase 6** — CI/CD pipeline + Prometheus/Grafana observability

## Design decisions & trade-offs

- **Delivery is at-least-once, not exactly-once.** If a worker dies after sending an email but before acking, RabbitMQ redelivers the message and the email can be sent twice. Workers skip a job already marked `done`, which narrows this window but cannot close it; true exactly-once delivery to an external system like SMTP isn't achievable without that system's cooperation.
- **Idempotency is enforced by Postgres, not application code.** A `UNIQUE` constraint with `ON CONFLICT DO NOTHING` is race-free; an "exists, then insert" check in Go is not, because two simultaneous requests can both pass it.
- **Known gap:** if the job insert succeeds but the RabbitMQ publish fails, the job stays `pending`, and a retry with the same idempotency key will not republish it. The standard fix is the transactional outbox pattern, which is future work.
- **Retry delay uses a TTL queue (`jobs_retry`) that dead-letters back to `jobs`,** so backoff needs no timers in worker memory and survives worker restarts.
- **Messages are acked only after their replacement is safely published.** If republishing a retry or dead-letter copy fails, the message is nacked and requeued, so a broker error cannot silently drop a job.
- **Prefetch of 1.** Each worker holds at most one unacknowledged job, so work is shared fairly across workers and a crash affects at most one job.
## Benchmarks

*(to be added in Phase 6 — throughput, recovery time after a killed worker, etc.)*

## Testing

Manually tested by submitting 3 concurrent jobs via curl; verified each transitioned pending → running → done in Postgres, and all 3 emails were delivered correctly via Mailtrap, confirming the full pipeline (API → Postgres → RabbitMQ → Worker → SMTP) works end-to-end.

During multi-job testing, one job failed due to Mailtrap's SMTP rate limit being hit when submitting several jobs in rapid succession — an expected real-world failure mode that motivates the retry/backoff logic built in Phase 2.

## Phase 2 — Reliability testing
 
### Issues #8, #9 — Retry with exponential backoff and dead-letter queue
 
Tested retry/backoff logic by deliberately breaking SMTP credentials. The system correctly retried with exponential backoff (2s, 4s, 8s, 16s, 32s), and after Mailtrap temporarily locked the account following repeated failed logins, the job continued retrying against the real (correct) credentials — still failing due to the account-level lockout, not the code — until exhausting 5 retries and correctly routing to the dead-letter queue. This demonstrated the system fails gracefully and predictably even when the downstream failure mode is unexpected.
 
Retry also recovered from a genuine, unstaged failure during later testing, when Mailtrap rejected a send for exceeding its rate limit:
 
```
Failed to send email (attempt 1): failed to send email: 550 "5.7.0 Too many emails per second. ..."
Retrying in 2000ms
Processing job: a07e594a-90eb-43a1-b8bd-95ae372d9327
Email sent to: test@example.com
Job acknowledged
```
 
### Issue #10 — Idempotency keys
 
Clients may send an `Idempotency-Key` header. The key is stored in a `UNIQUE` column on the `jobs` table, and the insert uses `ON CONFLICT (idempotency_key) DO NOTHING`, so the database — not application code — decides whether a request is a duplicate. Two simultaneous requests with the same key cannot both succeed. A duplicate request returns the original job's ID with `200 OK` and does not publish a second message to RabbitMQ.
 
Test: the same request sent twice with `Idempotency-Key: test-key-1`.
 
```
Job created: aa48ca9a-f423-40c2-9b07-81182d351286
Job already exists: aa48ca9a-f423-40c2-9b07-81182d351286
```

```
 count                                                                                                                                         
-------
     1
(1 row)
```
 
### Issue #11 — Killing a worker mid-job
 
Setup: two workers, each with a simulated processing delay (`WORKER_SIMULATE_DELAY_MS`), 4 jobs submitted with unique idempotency keys, and one worker hard-killed mid-job. TODO-PASTE: state whether real SMTP was used or `WORKER_DRY_RUN=true`, which bypasses SMTP so the test isolates queue behaviour.

Results:
 
```
 idempotency_key | status | retry_count |          updated_at                                                                                  
-----------------+--------+-------------+-------------------------------
 kill-122552-2   | done   |           0 | 2026-10-23 12:26:42.592394+01
 kill-122552-3   | done   |           0 | 2026-10-23 12:26:29.334156+01
 kill-122552-4   | done   |           0 | 2026-10-23 12:26:13.224283+01
 kill-122552-1   | done   |           0 | 2026-10-23 12:26:03.082511+01
 kill-test-4     | done   |           0 | 2026-10-23 12:17:11.258796+01
 kill-test-2     | done   |           0 | 2026-10-23 12:16:59.894471+01
 kill-test-3     | done   |           0 | 2026-10-23 12:16:45.330776+01
 kill-test-1     | done   |           0 | 2026-10-23 12:16:34.28976+01
(8 rows)

```
 
The unacknowledged message from the killed worker was redelivered by RabbitMQ to the surviving worker, and all jobs completed with no loss and no duplicate sends.
 
Worker hardening done as part of this issue:
 
- Prefetch set to 1 so a crashed worker holds at most one job.
- Messages are acked only after a retry or dead-letter copy is successfully published; on a failed publish the message is nacked and requeued, so a broker error cannot drop a job.
- Workers skip a job already marked `done`, so a redelivery does not repeat completed work.


## Running locally

*(instructions to be added at the end of Phase 3)*

## Author

Aidan Malone — [GitHub](#) · [LinkedIn](#)
