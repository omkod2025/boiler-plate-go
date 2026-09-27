package validator

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
)

type sampleRequest struct {
	Email string `json:"email" th:"อีเมล" validate:"required,email"`
	Name  string `json:"name" th:"ชื่อ" validate:"required"`
}

func TestMapValidationErrorsSeparatesMessages(t *testing.T) {
	req := sampleRequest{Email: "nope"}
	err := MapValidationErrors(Validate.Struct(req), &req)

	want := "อีเมล: " + customTagMessage("email") + ", ชื่อ: " + customTagMessage("required")
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestMapValidationErrorsSingleMessage(t *testing.T) {
	req := sampleRequest{Email: "foo@example.com"}
	err := MapValidationErrors(Validate.Struct(req), &req)

	want := "ชื่อ: " + customTagMessage("required")
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

type ginSampleRequest struct {
	Name string `json:"name" binding:"required,no_sql_inject"`
}

func TestCustomValidatorsRegisteredWithGinBinding(t *testing.T) {
	if err := binding.Validator.ValidateStruct(ginSampleRequest{Name: "x' OR 1=1 --"}); err == nil {
		t.Fatal("expected no_sql_inject to reject input through Gin binding")
	}
	if err := binding.Validator.ValidateStruct(ginSampleRequest{Name: "Food"}); err != nil {
		t.Fatalf("expected valid input to pass, got %v", err)
	}
}

func TestValidatorTags(t *testing.T) {
	cases := []struct {
		tag   string
		value any
		valid bool
	}{
		// custom validators
		{"phone", "0812345678", true},
		{"phone", "021234567", true},
		{"phone", "+66812345678", true},
		{"phone", "081234567890", false},
		{"phone", "abc", false},
		{"tel", "021234567", true},
		{"tel", "+6621234567", true},
		{"tel", "0812345678", false},
		{"thai_id", "1101700230708", true},
		{"thai_id", "3100100123451", true},
		{"thai_id", "1101700230707", false}, // check digit ผิด
		{"thai_id", "110170023070", false},  // 12 หลัก
		{"thai_id", "11017002307a8", false},
		{"string", "text", true},
		{"string", 123, false},
		{"array", []string{"a"}, true},
		{"array", "a", false},
		// aliases ไปยัง validator ของ library
		{"iso3166", "TH", true},
		{"iso3166", "THA", true},
		{"iso3166", "XX", false},
		{"currency_code", "THB", true},
		{"currency_code", "XXX1", false},
		// tag ของ library ที่เดิมถูก override ด้วย dummy ให้ผ่านเสมอ
		{"base64", "aGVsbG8=", true},
		{"base64", "not base64!", false},
		{"boolean", "true", true},
		{"boolean", "yes please", false},
		{"timezone", "Asia/Bangkok", true},
		{"timezone", "Mars/Olympus", false},
		{"iso4217", "USD", true},
		{"iso4217", "ABC", false},
		{"country_code", "TH", true},
		{"country_code", "ZZ", false},
		{"startswith=TH", "TH-01", true},
		{"startswith=TH", "US-01", false},
		{"contains=@", "a@b", true},
		{"contains=@", "ab", false},
		{"datetime=2006-01-02", "2026-09-27", true},
		{"datetime=2006-01-02", "27/09/2026", false},
		{"unique", []string{"a", "b"}, true},
		{"unique", []string{"a", "a"}, false},
	}
	for _, tc := range cases {
		err := ValidateVar(tc.value, tc.tag)
		if (err == nil) != tc.valid {
			t.Errorf("%s(%v): valid=%v, got err=%v", tc.tag, tc.value, tc.valid, err)
		}
	}
}

func TestAliasErrorUsesAliasMessage(t *testing.T) {
	type req struct {
		Country string `json:"country" th:"ประเทศ" validate:"iso3166"`
	}
	r := req{Country: "XX"}
	err := MapValidationErrors(Validate.Struct(r), &r)
	want := "ประเทศ: " + customTagMessage("iso3166")
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}
