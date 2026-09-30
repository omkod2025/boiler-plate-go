# boiler-plate-go

Boilerplate สำหรับเริ่มต้น REST API service ด้วย Go ที่จัดโครงสร้างตาม **Clean Architecture**
มี feature ตัวอย่างพร้อมใช้ 3 ตัว (auth, users, categories) เพื่อใช้เป็นแบบในการเพิ่ม feature ใหม่

- **HTTP**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL ผ่าน [pgx v5](https://github.com/jackc/pgx) (connection pool)
- **Auth**: JWT แบบ RS256 + รหัสผ่าน hash ด้วย bcrypt
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator) + custom tag และข้อความภาษาไทย ([pkg/validator](pkg/validator/README.md))
- **Roles**: binary/image เดียว เลือกหน้าที่ด้วย `-role` หรือ `APP_ROLE`: `api`, `stream`, `worker`, `receiver`, `sandbox`, `migrate` ([ดูหัวข้อ Roles](#roles))
- **Migration**: [goose](https://github.com/pressly/goose) (SQL ฝังใน binary, role `migrate`)
- **Messaging**: RabbitMQ ผ่าน [amqp091-go](https://github.com/rabbitmq/amqp091-go) — publisher confirm, consumer ack หลังทำงานเสร็จ ([pkg/amqp](pkg/amqp/amqp.go))
- **Contract**: OpenAPI ใน `api/openapi.yaml` → type ด้วย [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) (`make generate`)
- **Observability**: OpenTelemetry tracing (OTLP) แทน Elastic APM, request logger, `/health/live` + `/health/ready`
- **Test**: unit test + integration test กับ PostgreSQL/RabbitMQ จริงด้วย [testcontainers-go](https://golang.testcontainers.org) ([pkg/testhelper](pkg/testhelper/testhelper.go))
- **Tooling**: golangci-lint v2, lefthook, govulncheck, GitHub Actions, Docker

## เริ่มต้นใช้งาน

### สิ่งที่ต้องมี

- Go 1.26.8 ขึ้นไป (ตาม `go.mod`)
- PostgreSQL (18 ตรงกับ image ที่ test ใช้)
- Docker (สำหรับ integration test)
- (ไม่บังคับ) [lefthook](https://github.com/evilmartians/lefthook), [golangci-lint v2](https://golangci-lint.run)

### 1. เปลี่ยนชื่อ module

boilerplate นี้ใช้ module `github.com/omkod2025/boiler-plate-go` เปลี่ยนเป็นชื่อ service ของคุณก่อน

```bash
make change-module NEW_MODULE=github.com/your-org/your-service
```

คำสั่งนี้แก้ `go.mod`, import ในทุกไฟล์ `.go` แล้วรัน `go mod tidy`

### 2. สร้าง RSA key สำหรับ JWT

```bash
go run scripts/generate_rsa_keys.go
```

จะได้ไฟล์ key ทั้งแบบ PEM และ base64 และพิมพ์ค่า `JWT_PRIVATE_KEY` / `JWT_PUBLIC_KEY` สำหรับใส่ใน `.env`
(ไฟล์ key และ `.env` อยู่ใน `.gitignore` แล้ว อย่า commit)

### 3. ตั้งค่า `.env`

```env
APP_PORT=21400
APP_PREFIX=/api

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=app

JWT_PRIVATE_KEY=<base64 ของ private key>
JWT_PUBLIC_KEY=<base64 ของ public key>
```

ดูตัวแปรทั้งหมดได้ที่ [Configuration](#configuration)

### 4. สร้างตาราง (migration)

```bash
make migrate-up             # = go run . -role=migrate
make migrate-create NAME=add_products
```

migration อยู่ใน [migrations/](migrations/) (goose, ฝังใน binary) — `00001_init.sql` สร้างตารางของตัวอย่าง,
`00002_routines.sql` สร้าง function/procedure ที่ repository เรียก ([การเข้าถึงข้อมูล](#การเข้าถึงข้อมูล))

### 5. รัน

```bash
make run                    # role api (ค่าเริ่มต้น)
make run ROLE=worker        # หรือ go run . -role=worker
```

```bash
curl http://localhost:21400/health/ready
```

### รันด้วย Docker

```bash
docker build -t my-service .
docker run --env-file .env -p 21400:21400 my-service                 # api
docker run --env-file .env my-service -role=migrate                  # migration แล้วจบ
docker run --env-file .env -e APP_ROLE=worker my-service              # worker
```

## Roles

| Role | ใช้ DB | ใช้ JWT | RabbitMQ | HTTP บน `APP_PORT` | หน้าที่ |
|---|---|---|---|---|---|
| `api` | ✓ | ✓ | — | route ทั้งหมด + health | REST API หลัก |
| `stream` | ✓ | ✓ | — | WebSocket/SSE ([delivery/websocket](delivery/websocket/websocket.go)) + health, ไม่มี write timeout | connection ยาว |
| `worker` | ✓ | — | consume `AMQP_QUEUE` | health เท่านั้น | งานเบื้องหลัง ([delivery/amqp](delivery/amqp/handler.go)) |
| `receiver` | — | — | publish ไป `AMQP_EXCHANGE` | `POST /hooks/:source` + health | รับ webhook: publish (รอ confirm) ก่อนตอบ 202 ([delivery/receiver](delivery/receiver/receiver.go)) |
| `sandbox` | — | — | — | — | อ่านงาน JSON ทีละบรรทัดจาก stdin เขียนผลลง stdout ให้รันด้วย `--network none` ([delivery/sandbox](delivery/sandbox/sandbox.go)) |
| `migrate` | ✓ | — | — | — | `goose up` แล้วจบ — รันก่อน deploy role อื่น |

role ที่ไม่ได้ใช้ DB/JWT ไม่ต้องตั้งค่าตัวแปรเหล่านั้น role ที่ต้องใช้แต่ไม่ได้ตั้งจะหยุดตอนเริ่มพร้อมบอกชื่อตัวแปรที่ขาด
ทุก role หยุดรับงานใหม่เมื่อได้ SIGTERM แล้วรองานที่ค้าง (`SHUTDOWN_TIMEOUT`) ก่อนปิด

## การเข้าถึงข้อมูล

runtime **ไม่อ่าน/เขียนตารางตรง**:

- อ่าน (GET/list/count/exists) → function `oktf_*` ด้วย `SELECT ... FROM public.oktf_x($1)`
- เขียน (INSERT/UPDATE/DELETE, upsert, soft delete) → procedure `oktp_*` ด้วย `CALL public.oktp_x(..., NULL)`
  ค่าที่ต้องได้กลับ (id, เวลา, ผลว่าพบหรือไม่) เป็น `OUT` parameter ส่ง `NULL` ในตำแหน่งนั้นแล้ว `Scan` จากแถวผลของ `CALL`
- procedure ไม่ `COMMIT`/`ROLLBACK` เอง caller รวมหลายคำสั่งเป็น transaction เดียวได้ ([pkg/sql/example.go](pkg/sql/example.go))
- routine ตั้ง `SET search_path = pg_catalog, pg_temp` และอ้างตารางแบบระบุ schema
- SQL ที่แตะตารางอยู่ใน `migrations/` เท่านั้น — ตัวอย่างการเรียกอยู่ใน [infrastructure/postgres](infrastructure/postgres/)
- ใน production ให้ runtime role ได้แค่ `EXECUTE` บน routine ที่ใช้ (routine เป็น `SECURITY DEFINER` ของ owner ที่ไม่ login)
  ส่วนตัวอย่างใน boilerplate เป็น `SECURITY INVOKER` เพื่อให้รันได้โดยไม่ต้องสร้าง role เพิ่ม
- prefix ของ object ตามมาตรฐานของโครงการที่ใช้ boilerplate (ตัวอย่างนี้ `okdt_` ตาราง, `oktf_` function, `oktp_` procedure)

## โครงสร้างโปรเจกต์

```
.
├── main.go                  # จุดเริ่มโปรแกรม: เลือก role, โหลด config, tracing, สัญญาณ shutdown
├── api/openapi.yaml         # contract ของ HTTP API → make generate
├── migrations/              # goose migration (SQL) ฝังใน binary
├── oapi-codegen.yaml        # config ของ oapi-codegen
├── app/                     # Composition root — ประกอบ dependency ทั้งหมดเข้าด้วยกัน
│   ├── app.go               #   สร้างของที่ใช้ร่วมกัน (config, DB, JWT) แล้วเรียก wiring ของแต่ละ domain
│   ├── roles.go             #   runner ของแต่ละ role (api, stream, worker, receiver, sandbox, migrate)
│   ├── auth.go              #   wiring ของ domain auth
│   ├── users.go             #   wiring ของ domain users
│   └── categories.go        #   wiring ของ domain categories
├── domain/                  # Business logic — ไม่รู้จัก HTTP, SQL หรือ framework ใด ๆ
│   ├── auth/                #   login: ตรวจรหัสผ่านแล้วออก token
│   ├── users/               #   สมัครสมาชิก, ดึงข้อมูลผู้ใช้
│   └── categories/          #   CRUD หมวดหมู่ของผู้ใช้
│       ├── *_entity.go      #     entity และ value object
│       ├── *_port.go        #     interface ที่ domain ต้องการจากภายนอก + domain error
│       ├── *_use_case.go    #     use case
│       └── *_test.go        #     unit test ด้วย fake implementation
├── infrastructure/          # Implementation ของ port ที่ domain ประกาศไว้
│   ├── postgres/            #   repository ที่คุยกับ PostgreSQL (SQL และ DB row model อยู่ที่นี่)
│   └── security/            #   bcrypt password hasher, JWT token issuer
├── delivery/                # ช่องทางรับ request จากภายนอก
│   ├── http/
│   │   ├── http.go          #   สร้าง Gin router และลำดับ middleware
│   │   ├── handler/         #   HTTP handler + request/response DTO (tag json/binding อยู่ที่นี่)
│   │   ├── routes/          #   ผูก path กับ handler ของแต่ละ feature
│   │   ├── gen/             #   type ที่ generate จาก api/openapi.yaml (ห้ามแก้ด้วยมือ)
│   │   └── middleware/      #   JWT, CORS, rate limit, logger, recovery, OpenTelemetry tracing
│   ├── amqp/                #   handler ของข้อความ RabbitMQ (role worker)
│   ├── receiver/            #   รับ webhook แล้ว publish (role receiver)
│   ├── sandbox/             #   ประมวลผลงานจาก stdin (role sandbox)
│   ├── rpc/                 #   (placeholder) RPC server
│   └── websocket/           #   (placeholder) WebSocket handler
├── configs/                 # โหลด environment variable และสร้าง DB connection pool
├── pkg/                     # Library ทั่วไปที่ไม่ผูกกับ domain
│   ├── jwt/                 #   สร้าง/ตรวจ JWT, โหลด RSA key
│   ├── sql/                 #   wrapper ของ pgx pool และ helper (transaction, batch)
│   ├── amqp/                #   RabbitMQ: connection ที่ต่อใหม่เอง, publisher confirm, consumer
│   ├── migrate/             #   รัน goose migration
│   ├── health/              #   /health/live, /health/ready
│   ├── telemetry/           #   OpenTelemetry tracer provider (OTLP)
│   ├── testhelper/          #   PostgreSQL/RabbitMQ ใน container สำหรับ integration test
│   ├── validator/           #   validator + ข้อความ error ภาษาไทย
│   ├── response/            #   รูปแบบ JSON response มาตรฐาน
│   ├── logger/              #   logger แบบมี level
│   ├── exception/           #   AppError (ยังไม่มีที่ใช้)
│   └── gin/                 #   helper สร้าง gin.Engine (ยังไม่มีที่ใช้)
├── scripts/                 # สร้าง RSA key, เปลี่ยนชื่อ module
├── .github/workflows/       # CI (gofmt, lint, govulncheck) และ deploy
├── .golangci.yml            # config ของ golangci-lint v2
├── .lefthook.yml            # git hooks
└── Dockerfile               # multi-stage build → alpine
```

## Architecture

### Dependency rule

dependency ของ source code ชี้เข้าหา domain เสมอ — domain ไม่ import package ชั้นนอก

```
            ┌──────────────────────────────────────────┐
            │  app/  (composition root: new ทุกอย่าง)    │
            └───────┬───────────────┬──────────────────┘
                    │               │
          ┌─────────▼──────┐  ┌─────▼──────────────┐
          │   delivery/    │  │  infrastructure/   │
          │ handler, DTO   │  │ postgres, security │
          └─────────┬──────┘  └─────┬──────────────┘
                    │  เรียก use case  │  implement port
                    └───────┬───────┘
                    ┌───────▼────────┐
                    │    domain/     │  entity, port (interface), use case
                    └────────────────┘
```

- **domain** ประกาศ interface (port) ที่ต้องการ เช่น `categories.Repository`, `auth.TokenIssuer`
  และทำงานกับ entity ของตัวเองเท่านั้น
- **infrastructure** implement port เหล่านั้น เช่น `postgres.CategoryRepository` แปลง DB row เป็น entity
- **delivery** แปลง HTTP request เป็น input ของ use case และแปลง entity/domain error เป็น HTTP response
- **app** เป็นที่เดียวที่รู้จักทุกชั้น ทำหน้าที่ new และส่ง dependency ให้กัน
- domain แต่ละตัวไม่ import กันเอง เช่น `auth` ประกาศ `CredentialRepository` ของตัวเอง
  แทนการ import `users`

### Flow ของ request

ตัวอย่าง `PUT /api/categories/:id`

1. **middleware** (`delivery/http/middleware`) — logger, tracing, recovery, CORS, rate limit แล้วตรวจ JWT และเก็บ user id ใน context
2. **handler** (`delivery/http/handler`) — อ่าน user id และ `:id`, bind + validate JSON เข้า `UpdateCategoryRequest`, แปลงเป็น `categories.UpdateInput`
3. **use case** (`domain/categories`) — ตรวจ business rule (เช่น `Type` ต้องถูกต้อง) แล้วเรียก `Repository.Update(userID, id, input)`
4. **repository** (`infrastructure/postgres`) — รัน SQL ที่จำกัดด้วย `user_profile_id` และแปลงผลเป็น entity หรือคืน `categories.ErrNotFound`
5. **handler** — แปลง entity เป็น `CategoryResponse` หรือแปลง domain error เป็น status (`ErrNotFound` → 404, `ErrInvalidType` → 400)

### รูปแบบ response

ทุก endpoint ตอบด้วยรูปแบบเดียวกันจาก `pkg/response`

```json
{
  "success": true,
  "status": "success",
  "statusCode": 200,
  "message": "ok",
  "result": { "id": 1, "name": "Food" }
}
```

## API

prefix ตั้งค่าได้ด้วย `APP_PREFIX` (ค่าเริ่มต้น `/api`)

| Method | Path | ต้องมี token | คำอธิบาย |
|---|---|---|---|
| GET | `/health/live` (ไม่มี prefix) | ไม่ | liveness probe — ไม่แตะ DB |
| GET | `/health/ready` (ไม่มี prefix) | ไม่ | readiness — ตรวจ DB (และ RabbitMQ ของ worker/receiver) ตอบ 503 ถ้าไม่พร้อม |
| GET | `/healthz` | ไม่ | liveness probe (แบบเดิม) |
| GET | `/readiness` | ไม่ | readiness probe |
| POST | `/auth/login` | ไม่ | เข้าสู่ระบบด้วยอีเมล/รหัสผ่าน ได้ `accessToken` |
| POST | `/users` | ไม่ | สมัครสมาชิก (อีเมลซ้ำ → 409) |
| GET | `/users/me` | ใช่ | ข้อมูลผู้ใช้ที่ login อยู่ |
| GET | `/categories` | ใช่ | หมวดหมู่ของผู้ใช้ |
| POST | `/categories` | ใช่ | สร้างหมวดหมู่ |
| PUT | `/categories/:id` | ใช่ | แก้ไขหมวดหมู่ของตัวเอง (ของคนอื่น → 404) |
| DELETE | `/categories/:id` | ใช่ | ลบหมวดหมู่ของตัวเอง (ของคนอื่น → 404) |

ส่ง token ด้วย header `Authorization: Bearer <accessToken>`

```bash
curl -X POST localhost:21400/api/users -H 'Content-Type: application/json' \
  -d '{"email":"a@example.com","name":"A","password":"secret123"}'

curl -X POST localhost:21400/api/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"a@example.com","password":"secret123"}'
```

## การเพิ่ม feature ใหม่

ตัวอย่างการเพิ่ม domain `products`

1. **domain** — สร้าง `domain/products/`
   - `products_entity.go`: entity (ไม่มี tag `json`/`db`)
   - `products_port.go`: `Repository` interface และ domain error เช่น `ErrNotFound`
   - `products_use_case.go`: use case ที่รับ `Repository` ผ่าน constructor
   - `products_use_case_test.go`: test ด้วย fake repository
0. **contract + migration** — เพิ่ม path/schema ใน `api/openapi.yaml` แล้ว `make generate` (ใช้ type ใน `delivery/http/gen` เป็น DTO)
   และ `make migrate-create NAME=products` สำหรับตาราง + function อ่าน + procedure เขียน
2. **infrastructure** — สร้าง `infrastructure/postgres/products_repository.go` ที่ implement `products.Repository` โดยเรียก routine เท่านั้น
   (ใส่ `var _ products.Repository = (*ProductRepository)(nil)` เพื่อให้ compiler ตรวจ)
3. **delivery** — สร้าง `delivery/http/handler/products_dto.go`, `products_handler.go`
   และ `delivery/http/routes/products_routes.go`
4. **wiring** — สร้าง `app/products.go` ที่มี `newProductsHandler(d dependencies)`
5. **ลงทะเบียน** — เพิ่ม field `Products` ใน `routes.Handlers` และเรียก `RegisterProductsRoutes`
   ใน `routes.RegisterRoutes` แล้วเพิ่ม `Products: newProductsHandler(d)` ใน `InitApp`

ถ้า route ไม่ต้องใช้ token ให้เพิ่มใน `ignorePaths` ของ `delivery/http/http.go`

## Configuration

อ่านจาก environment variable (และ `.env` ถ้ามี) ใน [configs/config.go](configs/config.go)

| ตัวแปร | ค่าเริ่มต้น | บังคับ | คำอธิบาย |
|---|---|---|---|
| `APP_ROLE` | `api` | | role ของ process (หรือ flag `-role`) |
| `APP_PORT` | `21400` | | port ของ HTTP server (ทุก role ที่มี HTTP รวม health ของ worker) |
| `SHUTDOWN_TIMEOUT` | `20s` | | เวลาที่รอ request/งานที่ค้างตอนปิด |
| `APP_NAME` | `fmg-auth-api` | | ชื่อ service |
| `APP_ENV` | `development` | | environment |
| `APP_VERSION` | `1.0.0` | | version |
| `APP_PREFIX` | `/api` | | prefix ของทุก route |
| `APP_LIMIT` | `100` | | rate limit ต่อ IP (request ต่อนาที) เกินจะได้ 429 |
| `WHITE_LIST_URL` | `*` | | origin ที่อนุญาตสำหรับ CORS |
| `LOG_LEVEL` | `info` | | ระดับ log: `debug`, `info`, `warn`, `error` (ค่าอื่นใช้ `info`) |
| `DB_HOST` | | role ที่ใช้ DB | host ของ PostgreSQL |
| `DB_PORT` | `5432` | | port ของ PostgreSQL |
| `DB_USER` | | role ที่ใช้ DB | user |
| `DB_PASSWORD` | | role ที่ใช้ DB | password |
| `DB_NAME` | | role ที่ใช้ DB | ชื่อ database |
| `DB_MAX_CONNS` | `10` | | จำนวน connection สูงสุดใน pool |
| `DB_MIN_CONNS` | `2` | | จำนวน connection ขั้นต่ำ |
| `DB_MAX_CONN_LIFETIME` | `30m` | | อายุสูงสุดของ connection |
| `DB_MAX_CONN_IDLE_TIME` | `5m` | | เวลา idle สูงสุด |
| `DB_HEALTH_CHECK_PERIOD` | `1m` | | ความถี่ health check ของ pool |
| `JWT_PRIVATE_KEY` | | api, stream | RSA private key (PEM หรือ base64 ของ PEM) |
| `JWT_PUBLIC_KEY` | | api, stream | RSA public key (PEM หรือ base64 ของ PEM) |
| `JWT_DURATION` | `24h` | | อายุ token |
| `JWT_ISSUER` | `fmg-auth-api` | | issuer ใน token |
| `JWT_AUDIENCE` | | | audience ใน token |
| `JWT_VALIDATE_AUDIENCE` | `false` | | ตรวจ audience หรือไม่ |
| `AMQP_URL` | | worker/receiver ที่ใช้ RabbitMQ | เช่น `amqps://user:pass@host:5671/vhost` (ถูกปิดค่าใน log) |
| `AMQP_QUEUE` | | worker เมื่อมี `AMQP_URL` | queue ที่ worker consume |
| `AMQP_EXCHANGE` | `""` | | exchange ที่ receiver publish (routing key `hooks.<source>`) |
| `AMQP_PREFETCH` | `10` | | จำนวนข้อความที่ประมวลผลพร้อมกัน |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | | | `host:port` ของ OpenTelemetry collector (ว่าง = ไม่ส่ง trace ออก) |

- ตัวแปรที่บังคับตาม role ถ้าไม่ตั้งค่า role นั้นจะหยุดทำงานตอนเริ่มพร้อมบอกชื่อตัวแปร
- ค่าแบบ duration ใช้รูปแบบของ Go (`time.ParseDuration`) เช่น `30m`, `24h`, `168h`
  — **ไม่รองรับหน่วย `d`** ถ้าค่าผิดรูปแบบจะใช้ค่าเริ่มต้นแทน
- ตอนเริ่มโปรแกรมจะ log config ที่ระดับ `info` โดยปิดค่า `JWT_PRIVATE_KEY` / `JWT_PUBLIC_KEY` เป็น `[REDACTED]`
- รายละเอียดเรื่อง JWT ดูที่ [JWT_USAGE.md](JWT_USAGE.md)

## Development

### คำสั่งที่ใช้บ่อย

```bash
make test                  # unit + integration test (ต้องมี Docker; go test -short ./... ข้าม integration)
make generate              # regenerate type จาก api/openapi.yaml
go vet ./...
gofmt -l .                 # ไฟล์ที่ยังไม่ได้ format
golangci-lint run ./...    # lint (ต้องใช้ golangci-lint v2)
go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...   # ตรวจ vulnerability
```

### Git hooks

```bash
lefthook install
```

| Hook | ตรวจ |
|---|---|
| pre-commit | gofmt (เฉพาะไฟล์ที่ stage), `go vet`, golangci-lint |
| pre-push | govulncheck |

### CI/CD

| Workflow | ทำงานเมื่อ | ทำอะไร |
|---|---|---|
| [ci.yml](.github/workflows/ci.yml) | PR, push ไป `main`/`master` | gofmt, golangci-lint, test (รวม integration กับ PostgreSQL/RabbitMQ), generated code ตรง contract, govulncheck |
| [deploy.yml](.github/workflows/deploy.yml) | PR | build Docker image อย่างเดียว (ไม่ push) |
| | push ไป `main`/`master`/`release*`, tag `v*` | build + push image ไป Docker Hub แล้ว deploy ขึ้น EC2 |

deploy ต้องตั้ง secret ใน repository: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`,
`EC2_HOST`, `EC2_USER`, `EC2_SSH_KEY`, `ENV_PROD` (เนื้อหาไฟล์ `.env` ของ production)
