package health

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestReadiness(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fail := errors.New("dial tcp 10.0.0.5:5432: password authentication failed")
	r := gin.New()
	Register(r,
		Check{Name: "db", Essential: true, Probe: func(context.Context) error { return nil }},
		Check{Name: "llm", Probe: func(context.Context) error { return fail }},
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/health/ready", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"llm":"degraded"`) {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	r2 := gin.New()
	Register(r2, Check{Name: "db", Essential: true, Probe: func(context.Context) error { return fail }})
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, httptest.NewRequest("GET", "/health/ready", nil))
	if w.Code != 503 || strings.Contains(w.Body.String(), "password") {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	r2.ServeHTTP(w, httptest.NewRequest("GET", "/health/live", nil))
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
}
