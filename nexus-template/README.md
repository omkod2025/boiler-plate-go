# Nexus service template (FND-01.5)

ใช้จาก `nexus-backend/tools/new-service.mjs` เท่านั้น ตัว script แทนค่า placeholders,
สร้าง role/port/DB/RabbitMQ config จาก service registry และเพิ่ม module ใน go.work
template นี้ใช้ Go 1.26.8 + Gin + pgx ผ่าน nexus-kit, ไม่ใช้ Elastic APM

- `main.go`: role entrypoint และคำสั่ง `migrate` แยกจาก API startup
- `app/app.go`: composition root สำหรับ domain routes/consumer runners ของ service
- `configs/service.go`: role/port/dependency definition
- `delivery/http/`, `domain/`, `infrastructure/postgres/`: ชั้นเดียวกับ boilerplate
- `oapi-codegen.yaml` + `make generate`: generated Go models/Gin server จาก pinned contract
- `migrations/`: Goose Up เท่านั้น, migration credentials แยก, advisory lock, audit ledger
- `.air.toml`: hot reload; Dockerfile runtime ใช้ user 10001, binary เดียวต่อ service

shared packages อยู่ใน `nexus-backend/packages/nexus-kit`: `config`/`config/awssecrets`,
`amqpx` (amqp091-go), `telemetry` (OpenTelemetry), `lifecycle` (drain/shutdown),
`health`, `migration`, `contracttest`, `testkit` (testcontainers PostgreSQL/RabbitMQ)
service tests import `testkit` จาก `_test.go` เท่านั้น ตั้ง `NEXUS_REQUIRE_DOCKER=1`
เพื่อให้ CI fail เมื่อ Docker ใช้ไม่ได้ และใช้ `contracttest` ตรวจ frozen OpenAPI

scaffold ยังไม่มี domain endpoint/consumer handler จึงไม่ผ่าน Endpoint DoD โดยอัตโนมัติ
ห้าม tick งาน service implementation จากการ build template ผ่าน

`go.mod` replace ชี้ shared kit ภายใน backend repo; template ไม่ใช่ standalone Go module
สำหรับ test/build ให้ generate ใน nexus-backend ก่อน แล้วรัน `go vet`, `golangci-lint`,
`go test` ใน module นั้น ใช้ implement-api + postgres-pro + Database policy ทุกงาน backend
