// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Intel Corporation

package tracing

import (
	"net"
	"net/http"
	"os"

	"gitlab.com/project-emco/core/emco-base/src/orchestrator/pkg/infra/config"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

func createTracerProvider() (*tracesdk.TracerProvider, error) {
	var opts []tracesdk.TracerProviderOption
	if config.GetConfiguration().ZipkinIP != "" {
		endpoint := "http://" + net.JoinHostPort(config.GetConfiguration().ZipkinIP, config.GetConfiguration().ZipkinPort) + "/api/v2/spans"
		exp, err := zipkin.New(endpoint)
		if err != nil {
			return nil, err
		}
		opts = append(opts, tracesdk.WithBatcher(exp))
	}
	var ok bool
	var name, namespace string
	if name, ok = os.LookupEnv("APP_NAME"); !ok {
		name = "unknown"
	}
	if namespace, ok = os.LookupEnv("POD_NAMESPACE"); !ok {
		namespace = "default"
	}
	opts = append(opts, tracesdk.WithResource(resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(name+"."+namespace),
	)))
	return tracesdk.NewTracerProvider(opts...), nil
}

func InitializeTracer() error {
	tp, err := createTracerProvider()
	if err != nil {
		return err
	}
	otel.SetTracerProvider(tp)
	// Istio uses b3 propagation
	otel.SetTextMapPropagator(b3.New())
	return nil
}

func Middleware(next http.Handler) http.Handler {
	var ok bool
	var name string
	if name, ok = os.LookupEnv("APP_NAME"); !ok {
		name = "unknown"
	}
	tracer := otel.Tracer(name)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path)
		defer span.End()
		next.ServeHTTP(w, r.Clone(ctx))
	})
}
