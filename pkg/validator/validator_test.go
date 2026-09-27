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
