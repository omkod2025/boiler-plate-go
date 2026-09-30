package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/omkod2025/boiler-plate-go/configs"
)

func TestSandboxRole(t *testing.T) {
	in := strings.NewReader(`{"id":"1","kind":"echo","input":{"a":1}}` + "\n" + `not json` + "\n" + `{"id":"3","kind":"pdf"}` + "\n")
	var out bytes.Buffer
	if err := runSandbox(context.Background(), in, &out); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	want := []string{`{"id":"1","output":{"a":1}}`, `{"id":"","error":"invalid task"}`, `{"id":"3","error":"unsupported kind"}`}
	if len(lines) != len(want) {
		t.Fatalf("got %q", lines)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("line %d: %s want %s", i, lines[i], want[i])
		}
	}
}

func TestUnknownRoleAndMissingDependencies(t *testing.T) {
	ctx := context.Background()
	if err := Run(ctx, &configs.Config{Env: configs.EnvConfig{APP_ROLE: "cron"}}); err == nil {
		t.Fatal("unknown role must fail")
	}
	// api ต้องมี JWT key
	if err := Run(ctx, &configs.Config{Env: configs.EnvConfig{APP_ROLE: RoleAPI}}); err == nil || !strings.Contains(err.Error(), "JWT") {
		t.Fatalf("api without keys: %v", err)
	}
	// migrate/worker ต้องมี database
	t.Setenv("DB_HOST", "")
	for _, role := range []string{RoleMigrate, RoleWorker} {
		if err := Run(ctx, &configs.Config{Env: configs.EnvConfig{APP_ROLE: role}}); err == nil || !strings.Contains(err.Error(), "needs a database") {
			t.Fatalf("%s without DB: %v", role, err)
		}
	}
}
