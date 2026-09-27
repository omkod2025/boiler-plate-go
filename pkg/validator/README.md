# Validator Package

Package สำหรับ validation HTTP request ใน Go application

## Features

### Custom Validators

| tag | ตรวจสอบ | ตัวอย่างที่ผ่าน |
|---|---|---|
| `phone` | เบอร์โทรศัพท์ไทย (บ้านหรือมือถือ) ขึ้นต้นด้วย `0` หรือ `+66` | `0812345678`, `021234567`, `+66812345678` |
| `mobile` | ตัวเลข 9-10 หลัก ขึ้นต้นด้วย `0` | `0812345678` |
| `tel` | เบอร์โทรศัพท์บ้านไทย 9 หลัก ขึ้นต้นด้วย `02`-`07` | `021234567`, `+6621234567` |
| `thai_id` | เลขบัตรประชาชน 13 หลัก พร้อมตรวจ check digit | `1101700230708` |
| `string` | field ต้องเป็นชนิด string | |
| `array` | field ต้องเป็น slice หรือ array | |
| `no_special` | อนุญาตเฉพาะ a-z, A-Z, 0-9 และเว้นวรรค | |
| `no_sql_inject` | ห้ามมีอักขระเสี่ยง SQL injection เช่น `'`, `;`, `--` | |

Alias ไปยัง validator ของ library:

| tag | เทียบเท่า |
|---|---|
| `iso3166` | `iso3166_1_alpha2\|iso3166_1_alpha3\|iso3166_1_alpha_numeric` |
| `currency_code` | `iso4217` |

tag อื่นใช้ของ [go-playground/validator](https://github.com/go-playground/validator) โดยตรง เช่น `datetime=2006-01-02`,
`boolean`, `base64`, `contains=@`, `startswith=TH`, `unique`, `timezone`, `iso4217`, `country_code`
(`datetime` ต้องระบุ layout เสมอ)

## Usage

import package แล้วใช้ได้เลย ไม่ต้องเรียกฟังก์ชันลงทะเบียนเอง — `init()` ลงทะเบียน custom validator
ให้ทั้ง `validator.Validate` และ Gin binding validator อัตโนมัติ

```go
import "github.com/omkod2025-boop/omgon-notification-service/pkg/validator"
```

### 1. ใช้กับ Gin handler (แนะนำ)

ใส่กฎใน tag `binding` แล้วส่ง error จาก `ShouldBindJSON` ให้ `ResponseValidationError`
ถ้าเป็น validation error ฟังก์ชันจะตอบ 400 ให้เองและคืน `true`

```go
type RegisterUserRequest struct {
	Email    string `json:"email" th:"อีเมล" binding:"required,email,max=255"`
	Name     string `json:"name" th:"ชื่อ" binding:"required,min=1,max=100,no_sql_inject"`
	Phone    string `json:"phone" th:"เบอร์โทรศัพท์" binding:"omitempty,phone"`
	Password string `json:"password" th:"รหัสผ่าน" binding:"required,min=8,max=72"`
}

func (h *UsersHandler) Register(c *gin.Context) {
	var dto RegisterUserRequest
	if err := c.ShouldBindJSON(&dto); err != nil {
		if validator.ResponseValidationError(c, err, &dto) {
			return // ตอบ 400 พร้อมข้อความ validation แล้ว
		}
		response.BadRequest(c, err.Error()) // JSON ผิดรูปแบบ ฯลฯ
		return
	}
	// ...
}
```

ใช้แบบเดียวกันได้กับ `c.ShouldBindQuery` (tag `form`) และ `c.ShouldBindUri` (tag `uri`)

### 2. ตรวจ struct นอก HTTP handler

`ValidateStruct` ใช้ tag `validate` (ไม่ใช่ `binding`) แปลง error เป็นข้อความด้วย `MapValidationErrors`

```go
type ImportRow struct {
	ThaiID string `th:"เลขบัตรประชาชน" validate:"required,thai_id"`
	Phone  string `th:"เบอร์โทรศัพท์" validate:"required,phone"`
}

row := ImportRow{ThaiID: "1101700230707", Phone: "abc"}
if err := validator.ValidateStruct(row); err != nil {
	fmt.Println(validator.MapValidationErrors(err, &row))
	// เลขบัตรประชาชน: เลขบัตรประชาชนไทยไม่ถูกต้อง, เบอร์โทรศัพท์: เบอร์โทรศัพท์ไม่ถูกต้อง
}
```

### 3. ตรวจค่าเดี่ยว

```go
if err := validator.ValidateVar(c.Param("id"), "required,number"); err != nil {
	response.BadRequest(c, "id: ต้องเป็นตัวเลข")
	return
}
```

### ชื่อ field ในข้อความ error

เลือกตามลำดับ: tag `th` → tag `json` → ชื่อ field ใน struct

```go
Name string `json:"name" th:"ชื่อ" binding:"required"`
// ไม่ผ่าน → "ชื่อ: กรุณาระบุข้อมูล"
```

ถ้าผิดหลาย field ข้อความจะคั่นด้วย `, `

### รูปแบบ error response

```json
{
  "success": false,
  "status": "bad_request",
  "statusCode": 400,
  "message": "อีเมล: รูปแบบอีเมลไม่ถูกต้อง, ชื่อ: กรุณาระบุข้อมูล"
}
```

### ข้อความ error

ข้อความของแต่ละ tag อยู่ในฟังก์ชัน `customTagMessage` ใน validator.go
tag ที่ไม่มีข้อความจะแสดงชื่อ tag แทน

## ฟังก์ชัน

| ฟังก์ชัน | ใช้ทำอะไร |
|---|---|
| `ResponseValidationError(c, err, obj) bool` | ถ้า `err` เป็น validation error ตอบ 400 พร้อมข้อความ แล้วคืน `true` |
| `MapValidationErrors(err, obj) error` | แปลง validation error เป็นข้อความภาษาไทยตาม tag `th`/`json` |
| `ValidateStruct(s) error` | ตรวจ struct ด้วย tag `validate` |
| `ValidateVar(value, tag) error` | ตรวจค่าเดี่ยวด้วยกฎที่ระบุ |
| `Validate` | instance ของ `*validator.Validate` ที่ลงทะเบียน custom validator แล้ว |

## การเพิ่ม custom validation

เพิ่มฟังก์ชันและ tag ใน `customValidators` (หรือ `customAliases`) ใน validator.go
ระบบจะลงทะเบียนให้ทั้ง `Validate` และ Gin binding validator และเพิ่มข้อความใน `customTagMessage`

```go
var customValidators = map[string]validator.Func{
	// ...
	"postcode": validatePostcode,
}
```

ห้ามใช้ชื่อ tag ที่ library มีอยู่แล้ว เพราะจะไปแทนที่การตรวจของ library

## อ้างอิง
- [go-playground/validator](https://github.com/go-playground/validator)
