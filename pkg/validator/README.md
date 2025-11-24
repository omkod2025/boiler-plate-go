# Validator Package

Package สำหรับ validation HTTP request ใน Go application

## Features

### Custom Validators

1. **Phone Number** (`phone`)
   - รองรับเบอร์โทรศัพท์ไทย
   - รูปแบบ: `0812345678`, `+66812345678`

2. **Thai ID** (`thai_id`)
   - ตรวจสอบเลขบัตรประชาชน 13 หลัก
   - ตรวจสอบ checksum algorithm

3. **Username** (`username`)
   - 3-20 ตัวอักษร
   - ตัวอักษร, ตัวเลข, underscore เท่านั้น

4. **Password** (`password`)
   - อย่างน้อย 8 ตัวอักษร
   - ต้องมีตัวพิมพ์ใหญ่, ตัวพิมพ์เล็ก, ตัวเลข, special character

5. **Thai Language** (`thai`)
   - ตรวจสอบว่ามีตัวอักษรไทยหรือไม่

6. **English Language** (`english`)
   - ตรวจสอบว่ามีตัวอักษรภาษาอังกฤษหรือไม่

7. **URL** (`url`)
   - ตรวจสอบ URL format

8. **Date Format** (`date_format`)
   - รองรับ: `YYYY-MM-DD`, `DD/MM/YYYY`, `DD-MM-YYYY`

9. **Time Format** (`time_format`)
   - รองรับ: `HH:MM:SS`, `HH:MM`

10. **JSON** (`json`)
    - ตรวจสอบ JSON string

11. **Base64** (`base64`)
    - ตรวจสอบ Base64 string

## Usage

### 1. Basic Usage

```go
import "fmg-auth-api/pkg/validator"

// ลงทะเบียน validators
validator.RegisterCustomValidators()

// ตรวจสอบ struct
type UserRequest struct {
    Username string `json:"username" validate:"required,username"`
    Email    string `json:"email" validate:"required,email"`
    Phone    string `json:"phone" validate:"required,phone"`
}

user := UserRequest{
    Username: "john_doe",
    Email:    "john@example.com",
    Phone:    "0812345678",
}

if err := validator.ValidateStruct(user); err != nil {
    // จัดการ error
}
```

### 2. Gin Middleware Usage

```go
import "fmg-auth-api/pkg/validator"

// ใน handler
func CreateUser(c *gin.Context) {
    var request UserRegistrationRequest
    
    // ตรวจสอบ request
    if !validator.ValidateRequest(c, &request) {
        return // validation failed, response already sent
    }
    
    // ทำงานต่อ...
}
```

### 3. Individual Field Validation

```go
// ตรวจสอบ email
if !validator.ValidateEmail(c, email) {
    return
}

// ตรวจสอบเบอร์โทรศัพท์
if !validator.ValidatePhone(c, phone) {
    return
}

// ตรวจสอบเลขบัตรประชาชน
if !validator.ValidateThaiID(c, thaiID) {
    return
}
```

### 4. Query Parameters Validation

```go
func SearchProducts(c *gin.Context) {
    var request SearchRequest
    
    if !validator.ValidateQuery(c, &request) {
        return
    }
    
    // ทำงานต่อ...
}
```

### 5. URI Parameters Validation

```go
func GetUserByID(c *gin.Context) {
    var request struct {
        ID string `uri:"id" binding:"required" validate:"required"`
    }
    
    if !validator.ValidateURI(c, &request) {
        return
    }
    
    // ทำงานต่อ...
}
```

## Example Structs

### User Registration
```go
type UserRegistrationRequest struct {
    Username    string `json:"username" binding:"required,username" validate:"required,username"`
    Email       string `json:"email" binding:"required,email" validate:"required,email"`
    Password    string `json:"password" binding:"required,password" validate:"required,password"`
    FirstName   string `json:"first_name" binding:"required,min=2,max=50" validate:"required,min=2,max=50"`
    LastName    string `json:"last_name" binding:"required,min=2,max=50" validate:"required,min=2,max=50"`
    Phone       string `json:"phone" binding:"required,phone" validate:"required,phone"`
    ThaiID      string `json:"thai_id" binding:"required,thai_id" validate:"required,thai_id"`
    DateOfBirth string `json:"date_of_birth" binding:"required,date_format" validate:"required,date_format"`
}
```

### Product Request
```go
type ProductRequest struct {
    Name        string  `json:"name" binding:"required,min=3,max=100" validate:"required,min=3,max=100"`
    Description string  `json:"description" binding:"required,min=10,max=500" validate:"required,min=10,max=500"`
    Price       float64 `json:"price" binding:"required,min=0" validate:"required,min=0"`
    Category    string  `json:"category" binding:"required,oneof=electronics clothing books food" validate:"required,oneof=electronics clothing books food"`
    ImageURL    string  `json:"image_url" binding:"omitempty,url" validate:"omitempty,url"`
    Stock       int     `json:"stock" binding:"required,min=0" validate:"required,min=0"`
}
```

## Error Response Format

เมื่อ validation failed จะได้ response แบบนี้:

```json
{
    "success": false,
    "status": "bad_request",
    "statusCode": 400,
    "message": "Validation failed",
    "data": {
        "errors": [
            {
                "field": "username",
                "tag": "username",
                "value": "invalid_username",
                "message": "username must be 3-20 characters, letters, numbers and underscore only"
            },
            {
                "field": "phone",
                "tag": "phone",
                "value": "123",
                "message": "phone must be a valid phone number"
            }
        ],
        "count": 2
    }
}
```

## Available Validation Tags

### Built-in Tags
- `required` - ต้องมีค่า
- `email` - รูปแบบ email
- `min=X` - ความยาวขั้นต่ำ
- `max=X` - ความยาวสูงสุด
- `oneof=value1 value2` - ต้องเป็นหนึ่งในค่าที่กำหนด

### Custom Tags
- `phone` - เบอร์โทรศัพท์ไทย
- `thai_id` - เลขบัตรประชาชน
- `username` - username format
- `password` - password strength
- `thai` - ภาษาไทย
- `english` - ภาษาอังกฤษ
- `url` - URL format
- `date_format` - รูปแบบวันที่
- `time_format` - รูปแบบเวลา
- `json` - JSON string
- `base64` - Base64 string

## Best Practices

1. **ใช้ binding และ validate tags ร่วมกัน**
   ```go
   type Request struct {
       Field string `json:"field" binding:"required" validate:"required"`
   }
   ```

2. **ตรวจสอบ validation ใน handler**
   ```go
   func Handler(c *gin.Context) {
       var request Request
       if !validator.ValidateRequest(c, &request) {
           return
       }
       // ทำงานต่อ...
   }
   ```

3. **ใช้ custom error messages**
   ```go
   // ใน validator.go สามารถปรับ error messages ได้
   ```

4. **ตรวจสอบ required fields แยก**
   ```go
   fields := map[string]interface{}{
       "username": username,
       "email":    email,
   }
   if !validator.ValidateRequiredFields(c, fields) {
       return
   }
   ```

## Installation

```bash
go get github.com/go-playground/validator/v10
```

## Dependencies

- `github.com/go-playground/validator/v10` - Validation library
- `github.com/gin-gonic/gin` - Web framework
- `fmg-auth-api/pkg/response` - Response package 

# Validator Usage Guide

## 1. การ validate struct

```go
import (
    "fmg-auth-api/pkg/validator"
)

type CreateCustomerRequest struct {
    Name  string `json:"name" th:"ชื่อ" binding:"required" validate:"required,max=255"`
    Phone string `json:"phone" th:"เบอร์โทรศัพท์" binding:"required" validate:"required,max=20"`
    Email string `json:"email" th:"อีเมล" binding:"required,email,max=255"`
}

req := CreateCustomerRequest{
    Name:  "",
    Phone: "0812345678",
    Email: "not-an-email",
}

err := validator.Validate.Struct(req)
if err != nil {
    errors := validator.MapValidationErrorsWithFieldLabel(err, req)
    fmt.Println(errors)
    // Output: map[Name:ชื่อ: กรุณาระบุข้อมูล Email:อีเมล: รูปแบบอีเมลไม่ถูกต้อง]
}
```

---

## 2. ใช้งานกับ Gin Handler

```go
import (
    "fmg-auth-api/pkg/validator"
    "github.com/gin-gonic/gin"
)

func CreateCustomerHandler(c *gin.Context) {
    var req CreateCustomerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        errors := validator.MapValidationErrorsWithFieldLabel(err, req)
        c.JSON(400, gin.H{"status": "error", "errors": errors})
        return
    }
    // ... ดำเนินการต่อ ...
    c.JSON(200, gin.H{"status": "success"})
}
```

---

## 3. Custom Message ภาษาไทย (หรือภาษาอื่น)
- กำหนดได้ในฟังก์ชัน `customTagMessage` ใน validator.go
- สามารถเพิ่ม/แก้ไขข้อความแต่ละ tag ได้เอง

---

## 4. การใช้ tag json หรือ th
- ถ้า struct field มี tag `th` จะใช้เป็นชื่อ field ใน error message
- ถ้าไม่มี tag `th` จะใช้ tag `json`
- ถ้าไม่มีทั้งคู่ จะใช้ชื่อ field จริง

```go
type Example struct {
    Name string `json:"name" th:"ชื่อ" validate:"required"`
}
// ถ้า validate ไม่ผ่าน จะได้ error: ชื่อ: กรุณาระบุข้อมูล
```

---

## 5. ตัวอย่าง response error

```json
{
  "status": "error",
  "errors": {
    "Name": "ชื่อ: กรุณาระบุข้อมูล",
    "Email": "อีเมล: รูปแบบอีเมลไม่ถูกต้อง"
  }
}
```

---

## 6. ฟังก์ชันที่สำคัญ
- `MapValidationErrors(err error) map[string]string` : custom message ตาม tag
- `MapValidationErrorsWithFieldLabel(err error, obj interface{}) map[string]string` : custom message + ใช้ label จาก tag json/th
- `ResponseValidationError(c *gin.Context, err error) bool` : response error อัตโนมัติใน Gin handler

---

## 7. การเพิ่ม custom validation

```go
// ใน init() ของ validator.go
validate.RegisterValidation("datetime", validateDateFormat)
validate.RegisterValidation("time", validateTimeFormat)
```

---

## 8. อ้างอิง
- [go-playground/validator](https://github.com/go-playground/validator) 