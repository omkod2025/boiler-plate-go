package receiver

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakePub struct {
	err  error
	keys []string
}

func (f *fakePub) Publish(_ context.Context, _, key, _ string, _ []byte) error {
	f.keys = append(f.keys, key)
	return f.err
}

func TestReceiver(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pub := &fakePub{}
	r := gin.New()
	RegisterRoutes(r.Group("/api"), pub, "hooks")
	post := func(path, body string) int {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(body)))
		return w.Code
	}
	if c := post("/api/hooks/line", `{"events":[]}`); c != 202 || pub.keys[0] != "hooks.line" {
		t.Fatalf("%d %v", c, pub.keys)
	}
	if c := post("/api/hooks/Bad_Source", `{}`); c != 404 {
		t.Fatal(c)
	}
	if c := post("/api/hooks/line", strings.Repeat("a", MaxBody+1)); c != 413 {
		t.Fatal(c)
	}
	pub.err = errors.New("broker down")
	if c := post("/api/hooks/line", `{}`); c != 503 {
		t.Fatalf("broker failure must be 5xx so the provider retries: %d", c)
	}
}
