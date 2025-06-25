Below is a cursor-friendly construction plan you can paste into Cursor’s /todo.md (or feed it task-by-task).
Each step is deliberately small, has a clear “done” signal, and builds on the previous one so Cursor’s AI-pairing works smoothly.

⸻

At-a-Glance Summary

You will bootstrap an Encore monorepo with three services (Accrual, Redemption, Admin), an immutable points ledger in Postgres (via sqlc), a rules engine (GoRules/Grule) for dynamic earn logic, and Pub/Sub + FCM to notify the mobile app.  Security is handled with JWT middleware, contracts come from OpenAPI/oapi-codegen, and feature targeting uses GO Feature Flag.  The CI/CD, cron jobs, and observability are all declared natively in Encore.  Citations are included next to the concepts they support.

⸻

0 Prerequisites (run once)
	1.	Install tooling

brew install encoredev/tap/encore
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
go install github.com/thomaspoignant/go-feature-flag/cmd/...

Confirm encore version, sqlc version, and oapi-codegen -h succeed.

	2.	Create repo & first Encore app

encore app init ev-rewards
cd ev-rewards
git init && git add . && git commit -m "feat: encore skeleton"

Encore scaffolds Go modules and a local Postgres for you  ￼.

⸻

1 Data Layer

#	Task	Acceptance
1.1	Add schema.sql under /db.  Copy the users, points_events, rewards_catalog, redemptions tables exactly as drafted.	sqlc generate succeeds with Go structs for every table.  ￼
1.2	Commit generated code in /internal/db.	CI passes, no go vet errors.


⸻

2 Accrual Service (/services/accrual)

#	Task	Acceptance
2.1	Create Go package; annotate with //encore:service and define API struct.	encore run starts and logs “accrualsvc registered”.  ￼
2.2	Add POST /v1/events/charge handler receiving {session_id, kwh, user_id}.  Write row into points_events with event_type=CHARGE_KWH.	Unit test asserts 7 kWh → 70 pts row.
2.3	Publish UserPointsUpdated after commit using Encore Pub/Sub.	Local Pub/Sub topic visible in encore run dashboard.  ￼ ￼


⸻

3 Rules Engine Integration

#	Task	Acceptance
3.1	Vendor GoRules and fallback Grule.	go get github.com/gorules/gorules & github.com/hyperjumptech/grule-rule-engine.  ￼ ￼
3.2	Create rules.yaml with default earn rules: 10 pts/kWh, 300 pts/referral, 50 pts/rating.	Unit test loads YAML and returns correct multiplier.
3.3	Move points calculation for charge/referral/rating events behind EvaluateRules(ctx, payload) helper.	Old tests still pass with engine.


⸻

4 Redemption Service (/services/redemption)

#	Task	Acceptance
4.1	New service with GET /v1/rewards.  Query rewards_catalog WHERE active=true AND segment_match(user).	Returns seeded 4 rewards.
4.2	POST /v1/redeem deducts points inside serializable transaction and publishes RedemptionCreated.	Integration test: user 600 pts can redeem 500-pt reward once.
4.3	Add Encore Cron Job /jobs/expire_pending daily to expire unfulfilled redemptions >24h.	Job visible in encore run --cron.  ￼ ￼


⸻

5 Admin Service (/services/admin)

#	Task	Acceptance
5.1	Scaffold OpenAPI file admin.yaml (paths: /rules, /rewards, /segments).	oapi-codegen --generate types,server -o admin.gen.go admin.yaml succeeds.  ￼
5.2	Implement handlers with RBAC (role=product-admin).	Unauthorized call returns 403; authorized updates rule in DB.
5.3	Wire GO Feature Flag (YAML backend) for per-segment rewards.	Toggle “early-access” flag and see catalog filtered.  ￼ ￼


⸻

6 Notifications Worker

#	Task	Acceptance
6.1	Add subscriber to UserPointsUpdated & RedemptionCreated.	Worker package logs incoming events.
6.2	Integrate Firebase Admin SDK (Go); call messaging.Client.Send(ctx, &Message{Token:…, Notification:…}).	Local unit test mocks FCM and asserts payload.  ￼


⸻

7 Security & Middleware

#	Task	Acceptance
7.1	Add jwt.Validator middleware (e.g. github.com/golang-jwt/jwt/v5).	Invalid token rejected on every endpoint.  ￼
7.2	Enforce X-Idempotency-Key header on all POST endpoints; return 409 if replay.	Duplicate POST tested and ignored.


⸻

8 Observability & Ops

#	Task	Acceptance
8.1	Emit custom metrics (encore.LogMetrics) for points earned & redeemed.	Metrics visible in Encore Cloud dashboard.
8.2	Add prometheus.yml scrape config and Grafana dashboard JSON under /ops.	make grafana-import shows dashboards locally.


⸻

9 CI/CD

#	Task	Acceptance
9.1	Push repo to GitHub; enable Encore Cloud CI.	PR triggers build, unit tests, and preview environment.  ￼
9.2	Add GitHub Action for sqlc generate && go vet && golangci-lint run.	“green” check on PR.


⸻

10 Test Harness

#	Task	Acceptance
10.1	Table-driven unit tests for each rule type.	go test ./... ≥ 85 % coverage.
10.2	End-to-end test using Encore’s et package: create user → charge 5 kWh → redeem.	Entire flow passes without panics.


⸻

11 Launch Checklist
	1.	Secrets (encore secret set --env prod FCM_KEY=…, JWT keys, DB creds).
	2.	Performance smoke: hey -n 1000 -c 50 POST /v1/events/charge < 200 ms p95.
	3.	Disaster drill: kill FCM URL, verify DLQ and refund job.
	4.	Docs: publish Swagger UI from generated admin.yaml.
	5.	Go live 🎉

⸻

Helpful Links
	•	Encore service annotation & structure  ￼
	•	Encore Pub/Sub basics  ￼
	•	Cron jobs in Encore  ￼
	•	sqlc getting-started guide  ￼
	•	GoRules no-code rules engine  ￼
	•	Grule DSL engine  ￼
	•	Firebase Admin SDK push how-to  ￼
	•	oapi-codegen for OpenAPI stubs  ￼
	•	Encore deployment & preview pipelines  ￼
	•	GO Feature-Flag library details  ￼
	•	JWT middleware pattern discussion  ￼

Copy the “# Task” blocks into Cursor one at a time (or tick them off in bulk) and you’ll have a clean, test-driven Rewards API in short order. Happy building!