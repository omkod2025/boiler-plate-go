// Package telemetry ตั้งค่า OpenTelemetry tracing (แทน Elastic APM เดิม)
//
// ส่ง trace ด้วย OTLP/gRPC ไปที่ collector ตาม OTEL_EXPORTER_OTLP_ENDPOINT (เช่น otel-collector:4317)
// ถ้าไม่ตั้งค่า จะยังสร้าง span ได้ (propagate trace id ต่อได้) แต่ไม่ส่งออกไปไหน
// ตัวแปรมาตรฐานอื่นของ OTel SDK ใช้ได้ตามปกติ เช่น OTEL_EXPORTER_OTLP_INSECURE, OTEL_TRACES_SAMPLER
package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Options ของ tracer
type Options struct {
	ServiceName string
	Version     string
	Environment string
	Role        string
	Endpoint    string // host:port ของ collector; ว่าง = ไม่ส่งออก
}

// Init ตั้ง tracer provider และ propagator แบบ global แล้วคืนฟังก์ชัน shutdown (เรียกตอนปิดโปรแกรม
// เพื่อ flush span ที่ค้าง)
func Init(ctx context.Context, o Options) (func(context.Context) error, error) {
	res, err := resource.Merge(resource.Default(), resource.NewSchemaless(
		attribute.String("service.name", o.ServiceName),
		attribute.String("service.version", o.Version),
		attribute.String("deployment.environment.name", o.Environment),
		attribute.String("app.role", o.Role),
	))
	if err != nil {
		return nil, fmt.Errorf("telemetry: resource: %w", err)
	}
	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if o.Endpoint != "" {
		exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(o.Endpoint))
		if err != nil {
			return nil, fmt.Errorf("telemetry: otlp exporter: %w", err)
		}
		opts = append(opts, sdktrace.WithBatcher(exp))
	}
	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp.Shutdown, nil
}
