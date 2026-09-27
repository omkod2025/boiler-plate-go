package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/v2"
)

func APMTracerMiddleware(ctx *context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		tx := apm.DefaultTracer().StartTransaction(c.Request.Method+" "+c.Request.URL.Path, "request")
		*ctx = apm.ContextWithTransaction(*ctx, tx)
		c.Request = c.Request.WithContext(*ctx)
		defer tx.End()

		tx.Context.SetHTTPRequest(c.Request)
		tx.Context.SetHTTPResponseHeaders(c.Writer.Header())

		c.Set("tx", tx)
		c.Next()
	}
}
