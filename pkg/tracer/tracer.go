package tracer

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"roly-poly/config"
	"roly-poly/pkg/logger"
)

func InitTracerProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	log := logger.New()

	if config.AppConfig.OtelEndpoint == "" {
		log.Warn().Msg("OTEL_EXPORTER_OTLP_ENDPOINT not set, tracing disabled")
		return nil, nil
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(config.AppConfig.OtelEndpoint),
		otlptracehttp.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	serviceName := config.AppConfig.OtelServiceName
	if serviceName == "" {
		serviceName = "roly-poly"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}
