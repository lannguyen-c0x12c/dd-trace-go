// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package vault

import (
	"math"

	"github.com/lannguyen-c0x12c/dd-trace-go/dd-trace-go.v1/internal"
	"github.com/lannguyen-c0x12c/dd-trace-go/dd-trace-go.v1/internal/globalconfig"
	"github.com/lannguyen-c0x12c/dd-trace-go/dd-trace-go.v1/internal/namingschema"
)

type config struct {
	analyticsRate float64
	serviceName   string
	operationName string
}

const defaultServiceName = "vault"

// Option can be passed to NewHTTPClient and WrapHTTPClient to configure the integration.
type Option func(*config)

func defaults(cfg *config) {
	cfg.serviceName = namingschema.NewServiceNameSchema(
		"",
		defaultServiceName,
		namingschema.WithVersionOverride(namingschema.SchemaV0, defaultServiceName),
	).GetName()
	cfg.operationName = namingschema.NewDBOutboundOp(
		"vault",
		namingschema.WithVersionOverride(namingschema.SchemaV0, "http.request"),
	).GetName()

	if internal.BoolEnv("DD_TRACE_VAULT_ANALYTICS_ENABLED", false) {
		cfg.analyticsRate = 1.0
	} else {
		cfg.analyticsRate = globalconfig.AnalyticsRate()
	}
}

// WithAnalytics enables or disables Trace Analytics for all started spans.
func WithAnalytics(on bool) Option {
	if on {
		return WithAnalyticsRate(1.0)
	}
	return WithAnalyticsRate(math.NaN())
}

// WithAnalyticsRate sets the sampling rate for Trace Analytics events correlated to started spans.
func WithAnalyticsRate(rate float64) Option {
	return func(c *config) {
		c.analyticsRate = rate
	}
}

// WithServiceName sets the given service name for the http.Client.
func WithServiceName(name string) Option {
	return func(c *config) {
		c.serviceName = name
	}
}
