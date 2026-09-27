package validator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/omkod2025-boop/omgon-notification-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

// customValidators tag ที่ลงทะเบียนเพิ่มกับทั้ง Validate และ Gin binding validator
// tag ที่ library มีให้อยู่แล้ว (datetime, boolean, base64, contains, excludes, startswith, endswith,
// unique, file, dir, iso4217, timezone, country_code ฯลฯ) ห้ามลงทะเบียนทับ เพราะจะแทนที่การตรวจของ library
var customValidators = map[string]validator.Func{
	"string":        validateString,
	"array":         validateArray,
	"phone":         validatePhone,
	"mobile":        validateMobile,
	"tel":           validateTel,
	"thai_id":       validateThaiID,
	"no_special":    validateNoSpecial,
	"no_sql_inject": validateNoSQLInjection,
}

// customAliases tag ที่เป็นชื่อย่อของ validator ใน library
var customAliases = map[string]string{
	"iso3166":       "iso3166_1_alpha2|iso3166_1_alpha3|iso3166_1_alpha_numeric",
	"currency_code": "iso4217",
}

func init() {
	registerCustomValidators(Validate)
	RegisterGinCustomValidators()
}

// registerCustomValidators ลงทะเบียน customValidators และ customAliases ทั้งหมด
// error เกิดได้เฉพาะเมื่อ tag หรือ function ไม่ถูกต้อง ซึ่งเป็นความผิดพลาดของโค้ด จึง panic ตั้งแต่ตอนเริ่มโปรแกรม
func registerCustomValidators(v *validator.Validate) {
	for tag, fn := range customValidators {
		if err := v.RegisterValidation(tag, fn); err != nil {
			panic(fmt.Sprintf("validator: register %q: %v", tag, err))
		}
	}
	for alias, tags := range customAliases {
		v.RegisterAlias(alias, tags)
	}
}

var (
	// เบอร์โทรศัพท์ไทย (บ้านหรือมือถือ) ขึ้นต้นด้วย 0 หรือ +66 เช่น 021234567, 0812345678, +66812345678
	phonePattern = regexp.MustCompile(`^(0|\+66)[1-9][0-9]{7,8}$`)
	// เบอร์โทรศัพท์บ้านไทย 9 หลัก ขึ้นต้นด้วย 02-07 หรือ +662-+667 เช่น 021234567, +6621234567
	telPattern = regexp.MustCompile(`^(0|\+66)[2-7][0-9]{7}$`)
	// เลขบัตรประชาชน 13 หลัก
	thaiIDPattern = regexp.MustCompile(`^[0-9]{13}$`)
)

// validateString field ต้องเป็นชนิด string
func validateString(fl validator.FieldLevel) bool {
	return fl.Field().Kind() == reflect.String
}

// validateArray field ต้องเป็น slice หรือ array
func validateArray(fl validator.FieldLevel) bool {
	kind := fl.Field().Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

// validatePhone เบอร์โทรศัพท์ไทย ทั้งเบอร์บ้านและมือถือ
func validatePhone(fl validator.FieldLevel) bool {
	return phonePattern.MatchString(fl.Field().String())
}

func validateMobile(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// ต้องเป็นตัวเลข 9-10 หลัก และขึ้นต้นด้วย 0
	re := regexp.MustCompile(`^0[0-9]{8,9}$`)
	return re.MatchString(value)
}

// validateTel เบอร์โทรศัพท์บ้านไทย
func validateTel(fl validator.FieldLevel) bool {
	return telPattern.MatchString(fl.Field().String())
}

// validateThaiID เลขบัตรประชาชนไทย 13 หลัก พร้อมตรวจ check digit
// check digit = (11 - (ผลรวมของหลักที่ 1-12 คูณน้ำหนัก 13 ถึง 2) mod 11) mod 10
func validateThaiID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if !thaiIDPattern.MatchString(value) {
		return false
	}
	sum := 0
	for i := 0; i < 12; i++ {
		sum += int(value[i]-'0') * (13 - i)
	}
	return (11-sum%11)%10 == int(value[12]-'0')
}

// validateNoSpecial: ห้ามมีอักษรพิเศษ (อนุญาต a-z, A-Z, 0-9, เว้นวรรค)
func validateNoSpecial(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	re := regexp.MustCompile(`^[a-zA-Z0-9 ]*$`)
	return re.MatchString(value)
}

// validateNoSQLInjection: ไม่อนุญาตอักขระที่เสี่ยง SQL Injection
func validateNoSQLInjection(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	blockPattern := regexp.MustCompile(`['";\\%<>=\(\)\|]|--|/\*|\*/`)
	return !blockPattern.MatchString(value)
}

// ValidateStruct ตรวจสอบ struct
func ValidateStruct(s interface{}) error {
	return Validate.Struct(s)
}

// ValidateVar ตรวจสอบตัวแปรเดียว
func ValidateVar(field interface{}, tag string) error {
	return Validate.Var(field, tag)
}

// customTagMessage คืนข้อความตาม tag (ภาษาไทย)
func customTagMessage(tag string) string {
	switch tag {
	case "required":
		return "กรุณาระบุข้อมูล"
	case "email":
		return "รูปแบบอีเมลไม่ถูกต้อง"
	case "max":
		return "ข้อมูลยาวเกินไป"
	case "min":
		return "ข้อมูลสั้นเกินไป"
	case "len":
		return "ความยาวข้อมูลไม่ถูกต้อง"
	case "eq":
		return "ข้อมูลต้องตรงกับค่าที่กำหนด"
	case "ne":
		return "ข้อมูลต้องไม่ตรงกับค่าที่กำหนด"
	case "gt":
		return "ค่าต้องมากกว่า"
	case "gte":
		return "ค่าต้องมากกว่าหรือเท่ากับ"
	case "lt":
		return "ค่าต้องน้อยกว่า"
	case "lte":
		return "ค่าต้องน้อยกว่าหรือเท่ากับ"
	case "oneof":
		return "ข้อมูลต้องเป็นหนึ่งในค่าที่กำหนด"
	case "numeric":
		return "ต้องเป็นตัวเลขเท่านั้น"
	case "alphanum":
		return "ต้องเป็นตัวอักษรหรือตัวเลขเท่านั้น"
	case "url":
		return "รูปแบบ URL ไม่ถูกต้อง"
	case "uuid":
		return "รูปแบบ UUID ไม่ถูกต้อง"
	case "datetime":
		return "รูปแบบวันเวลาไม่ถูกต้อง"
	case "time":
		return "รูปแบบเวลาไม่ถูกต้อง"
	case "boolean":
		return "ต้องเป็น true หรือ false"
	case "base64":
		return "ต้องเป็น base64 เท่านั้น"
	case "ip":
		return "รูปแบบ IP address ไม่ถูกต้อง"
	case "hostname":
		return "รูปแบบ hostname ไม่ถูกต้อง"
	case "contains":
		return "ข้อมูลต้องมีค่าที่กำหนด"
	case "excludes":
		return "ข้อมูลต้องไม่มีค่าที่กำหนด"
	case "startswith":
		return "ข้อมูลต้องขึ้นต้นด้วยค่าที่กำหนด"
	case "endswith":
		return "ข้อมูลต้องลงท้ายด้วยค่าที่กำหนด"
	case "unique":
		return "ข้อมูลต้องไม่ซ้ำกัน"
	case "number":
		return "ต้องเป็นตัวเลข"
	case "integer":
		return "ต้องเป็นจำนวนเต็ม"
	case "string":
		return "ต้องเป็นข้อความ"
	case "array":
		return "ต้องเป็น array"
	case "file":
		return "ต้องเป็นไฟล์"
	case "dir":
		return "ต้องเป็นโฟลเดอร์"
	case "credit_card":
		return "หมายเลขบัตรเครดิตไม่ถูกต้อง"
	case "isbn":
		return "หมายเลข ISBN ไม่ถูกต้อง"
	case "iso3166":
		return "รหัสประเทศไม่ถูกต้อง (ISO3166)"
	case "iso4217":
		return "รหัสสกุลเงินไม่ถูกต้อง (ISO4217)"
	case "timezone":
		return "รหัส timezone ไม่ถูกต้อง"
	case "country_code":
		return "รหัสประเทศไม่ถูกต้อง"
	case "currency_code":
		return "รหัสสกุลเงินไม่ถูกต้อง"
	case "phone":
		return "เบอร์โทรศัพท์ไม่ถูกต้อง"
	case "mobile":
		return "เบอร์มือถือไม่ถูกต้อง"
	case "tel":
		return "เบอร์โทรศัพท์ไม่ถูกต้อง"
	case "thai_id":
		return "เลขบัตรประชาชนไทยไม่ถูกต้อง"
	case "no_special":
		return "ห้ามมีอักษรพิเศษ"
	case "no_sql_inject":
		return "ห้ามมีอักษรพิเศษ"
	default:
		return tag
	}
}

// MapValidationErrors แปลง validation error เป็น custom message
// หากมีหลาย field จะคั่นแต่ละข้อความด้วย ", "
func MapValidationErrors(err error, obj interface{}) error {
	var msgs []string
	if errs, ok := err.(validator.ValidationErrors); ok {
		typ := reflect.TypeOf(obj)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		for _, e := range errs {
			field, tag := e.Field(), e.Tag()
			fieldName := field // fallback
			if f, ok := typ.FieldByName(field); ok {
				// ถ้ามี tag th ให้ใช้ก่อน, ถ้าไม่มีก็ใช้ json, ถ้าไม่มีทั้งคู่ใช้ชื่อ field
				if th, ok := f.Tag.Lookup("th"); ok && th != "" {
					fieldName = th
				} else if jsonTag, ok := f.Tag.Lookup("json"); ok && jsonTag != "" {
					fieldName = jsonTag
				}
			}
			msgs = append(msgs, fmt.Sprintf("%s: %s", fieldName, customTagMessage(tag)))
		}
	}
	return errors.New(strings.Join(msgs, ", "))
}

func ResponseValidationError(c *gin.Context, err error, obj interface{}) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		errors := MapValidationErrors(ve, obj)
		response.BadRequest(c, errors.Error())
		return true
	}
	return false
}

// RegisterGinCustomValidators สำหรับ Gin binding validator
// ตัวอย่างการเรียก: validator.RegisterGinCustomValidators() ใน main หรือก่อน init Gin
func RegisterGinCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		registerCustomValidators(v)
	}
}
