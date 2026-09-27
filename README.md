# boiler-plate-go

Boilerplate สำหรับเริ่มต้น REST API service ด้วย Go ที่จัดโครงสร้างตาม **Clean Architecture**
มี feature ตัวอย่างพร้อมใช้ 3 ตัว (auth, users, categories) เพื่อใช้เป็นแบบในการเพิ่ม feature ใหม่

- **HTTP**: [Gin](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL ผ่าน [pgx v5](https://github.com/jackc/pgx) (connection pool)
- **Auth**: JWT แบบ RS256 + รหัสผ่าน hash ด้วย bcrypt
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator) + custom tag และข้อความภาษาไทย ([pkg/validator](pkg/validator/README.md))
- **Observability**: Elastic APM, request logger
- **Tooling**: golangci-lint v2, lefthook, govulncheck, GitHub Actions, Docker

## เริ่มต้นใช้งาน

### สิ่งที่ต้องมี

- Go 1.25.14 ขึ้นไป (ตาม `go.mod`)
- PostgreSQL
- (ไม่บังคับ) [lefthook](https://github.com/evilmartians/lefthook), [golangci-lint v2](https://golangci-lint.run)

### 1. เปลี่ยนชื่อ module

boilerplate นี้ใช้ module `github.com/omkod2025-boop/omgon-notification-service` เปลี่ยนเป็นชื่อ service ของคุณก่อน

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

### 4. สร้างตาราง

repository ตัวอย่างคาดหวังตาราง `okdt_user_profiles`, `okdt_categories`, `okdt_master_categories`
และ function `oktf_category_get(user_profile_id)` — ตัวอย่าง schema ของ users อยู่ในคอมเมนต์ของ
[infrastructure/postgres/users_repository.go](infrastructure/postgres/users_repository.go)
repo นี้ยังไม่มีระบบ migration ให้ปรับตาม schema จริงของคุณ

### 5. รัน

```bash
go run .
```

```bash
curl http://localhost:21400/api/healthz
```

### รันด้วย Docker

```bash
docker build -t my-service .
docker run --env-file .env -p 21400:21400 my-service
```

## โครงสร้างโปรเจกต์

```
.
├── main.go                  # จุดเริ่มโปรแกรม: โหลด config, เริ่ม HTTP server, graceful shutdown
├── app/                     # Composition root — ประกอบ dependency ทั้งหมดเข้าด้วยกัน
│   ├── app.go               #   สร้างของที่ใช้ร่วมกัน (config, DB, JWT) แล้วเรียก wiring ของแต่ละ domain
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
│   │   └── middleware/      #   JWT, CORS, rate limit, logger, recovery, APM
│   ├── rpc/                 #   (placeholder) RPC server
│   └── websocket/           #   (placeholder) WebSocket handler
├── configs/                 # โหลด environment variable และสร้าง DB connection pool
├── pkg/                     # Library ทั่วไปที่ไม่ผูกกับ domain
│   ├── jwt/                 #   สร้าง/ตรวจ JWT, โหลด RSA key
│   ├── sql/                 #   wrapper ของ pgx pool และ helper (transaction, batch)
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

1. **middleware** (`delivery/http/middleware`) — logger, APM, recovery, CORS, rate limit แล้วตรวจ JWT และเก็บ user id ใน context
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
| GET | `/healthz` | ไม่ | liveness probe |
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
2. **infrastructure** — สร้าง `infrastructure/postgres/products_repository.go` ที่ implement `products.Repository`
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
| `APP_PORT` | `21400` | ✓ | port ของ HTTP server |
| `APP_NAME` | `fmg-auth-api` | | ชื่อ service |
| `APP_ENV` | `development` | | environment |
| `APP_VERSION` | `1.0.0` | | version |
| `APP_PREFIX` | `/api` | | prefix ของทุก route |
| `APP_LIMIT` | `100` | | rate limit ต่อ IP (request ต่อนาที) เกินจะได้ 429 |
| `WHITE_LIST_URL` | `*` | | origin ที่อนุญาตสำหรับ CORS |
| `LOG_LEVEL` | `info` | | ระดับ log (อ่านค่าแล้วแต่ยังไม่ได้ส่งให้ `logger.SetLevel`) |
| `DB_HOST` | `localhost` | ✓ | host ของ PostgreSQL |
| `DB_PORT` | `5432` | | port ของ PostgreSQL |
| `DB_USER` | `user` | ✓ | user |
| `DB_PASSWORD` | `pass` | ✓ | password |
| `DB_NAME` | `fmg_auth` | ✓ | ชื่อ database |
| `DB_MAX_CONNS` | `10` | | จำนวน connection สูงสุดใน pool |
| `DB_MIN_CONNS` | `2` | | จำนวน connection ขั้นต่ำ |
| `DB_MAX_CONN_LIFETIME` | `30m` | | อายุสูงสุดของ connection |
| `DB_MAX_CONN_IDLE_TIME` | `5m` | | เวลา idle สูงสุด |
| `DB_HEALTH_CHECK_PERIOD` | `1m` | | ความถี่ health check ของ pool |
| `JWT_PRIVATE_KEY` | | ✓ | RSA private key (PEM หรือ base64 ของ PEM) |
| `JWT_PUBLIC_KEY` | | ✓ | RSA public key (PEM หรือ base64 ของ PEM) |
| `JWT_DURATION` | `24h` | | อายุ token |
| `JWT_ISSUER` | `fmg-auth-api` | | issuer ใน token |
| `JWT_AUDIENCE` | | | audience ใน token |
| `JWT_VALIDATE_AUDIENCE` | `false` | | ตรวจ audience หรือไม่ |

- ตัวแปรที่ "บังคับ" ถ้าไม่ตั้งค่า โปรแกรมจะหยุดทำงานตอนเริ่ม
- ค่าแบบ duration ใช้รูปแบบของ Go (`time.ParseDuration`) เช่น `30m`, `24h`, `168h`
  — **ไม่รองรับหน่วย `d`** ถ้าค่าผิดรูปแบบจะใช้ค่าเริ่มต้นแทน
- รายละเอียดเรื่อง JWT ดูที่ [JWT_USAGE.md](JWT_USAGE.md)

## Development

### คำสั่งที่ใช้บ่อย

```bash
go test ./...              # unit test
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
| [ci.yml](.github/workflows/ci.yml) | PR, push ไป `main`/`master` | gofmt, golangci-lint, govulncheck |
| [deploy.yml](.github/workflows/deploy.yml) | PR | build Docker image อย่างเดียว (ไม่ push) |
| | push ไป `main`/`master`/`release*`, tag `v*` | build + push image ไป Docker Hub แล้ว deploy ขึ้น EC2 |

deploy ต้องตั้ง secret ใน repository: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`,
`EC2_HOST`, `EC2_USER`, `EC2_SSH_KEY`, `ENV_PROD` (เนื้อหาไฟล์ `.env` ของ production)
