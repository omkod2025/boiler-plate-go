package configs

import (
	"fmt"
	"strings"
	"testing"
)

func TestEnvConfigRedactsKeys(t *testing.T) {
	env := EnvConfig{
		APP_NAME:        "svc",
		JWT_PRIVATE_KEY: "-----BEGIN PRIVATE KEY-----secret",
		JWT_PUBLIC_KEY:  "-----BEGIN PUBLIC KEY-----public",
	}
	for _, format := range []string{"%v", "%+v", "%#v", "%s"} {
		out := fmt.Sprintf(format, env)
		if strings.Contains(out, "BEGIN") {
			t.Errorf("%s leaks key: %s", format, out)
		}
		if !strings.Contains(out, "svc") || !strings.Contains(out, "[REDACTED]") {
			t.Errorf("%s: expected other fields and redaction marker, got %s", format, out)
		}
	}
	// logger ใช้ fmt.Sprintln
	if out := fmt.Sprintln("env", env); strings.Contains(out, "BEGIN") {
		t.Errorf("Sprintln leaks key: %s", out)
	}
	if out := fmt.Sprintf("%v", &env); strings.Contains(out, "BEGIN") {
		t.Errorf("pointer leaks key: %s", out)
	}
}
