package validator

import "testing"

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
