// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package httprouter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lannguyen-c0x12c/dd-trace-go/contrib/internal/namingschematest"
	"github.com/lannguyen-c0x12c/dd-trace-go/ddtrace/ext"
	"github.com/lannguyen-c0x12c/dd-trace-go/ddtrace/mocktracer"
	"github.com/lannguyen-c0x12c/dd-trace-go/ddtrace/tracer"
	"github.com/lannguyen-c0x12c/dd-trace-go/internal/globalconfig"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestHttpTracer200(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	// Send and verify a 200 request
	url := "/200"
	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	router().ServeHTTP(w, r)
	assert.Equal(200, w.Code)
	assert.Equal("OK\n", w.Body.String())

	spans := mt.FinishedSpans()
	assert.Equal(1, len(spans))

	s := spans[0]
	assert.Equal("http.request", s.OperationName())
	assert.Equal("my-service", s.Tag(ext.ServiceName))
	assert.Equal("GET "+url, s.Tag(ext.ResourceName))
	assert.Equal(url, s.Tag(ext.HTTPRoute))
	assert.Equal("200", s.Tag(ext.HTTPCode))
	assert.Equal("GET", s.Tag(ext.HTTPMethod))
	assert.Equal("http://example.com"+url, s.Tag(ext.HTTPURL))
	assert.Equal("testvalue", s.Tag("testkey"))
	assert.Equal(nil, s.Tag(ext.Error))
	assert.Equal("julienschmidt/httprouter", s.Tag(ext.Component))
	assert.Equal(ext.SpanKindServer, s.Tag(ext.SpanKind))
}

func TestHttpTracer200WithPathParameter(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	// Send and verify a 200 request
	url := "/200/value"
	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	router().ServeHTTP(w, r)
	assert.Equal(200, w.Code)
	assert.Equal("value", w.Body.String())

	spans := mt.FinishedSpans()
	assert.Equal(1, len(spans))

	s := spans[0]
	assert.Equal("http.request", s.OperationName())
	assert.Equal("my-service", s.Tag(ext.ServiceName))
	assert.Equal("GET /200/:parameter", s.Tag(ext.ResourceName))
	assert.Equal("/200/:parameter", s.Tag(ext.HTTPRoute))
	assert.Equal("200", s.Tag(ext.HTTPCode))
	assert.Equal("GET", s.Tag(ext.HTTPMethod))
	assert.Equal("http://example.com"+url, s.Tag(ext.HTTPURL))
	assert.Equal("testvalue", s.Tag("testkey"))
	assert.Equal(nil, s.Tag(ext.Error))
	assert.Equal("julienschmidt/httprouter", s.Tag(ext.Component))
	assert.Equal(ext.SpanKindServer, s.Tag(ext.SpanKind))
}

func TestHttpTracer500(t *testing.T) {
	assert := assert.New(t)
	mt := mocktracer.Start()
	defer mt.Stop()

	// Send and verify a 500 request
	url := "/500"
	r := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()
	router().ServeHTTP(w, r)
	assert.Equal(500, w.Code)
	assert.Equal("500!\n", w.Body.String())

	spans := mt.FinishedSpans()
	assert.Equal(1, len(spans))

	s := spans[0]
	assert.Equal("http.request", s.OperationName())
	assert.Equal("my-service", s.Tag(ext.ServiceName))
	assert.Equal("GET "+url, s.Tag(ext.ResourceName))
	assert.Equal("500", s.Tag(ext.HTTPCode))
	assert.Equal(url, s.Tag(ext.HTTPRoute))
	assert.Equal("GET", s.Tag(ext.HTTPMethod))
	assert.Equal("http://example.com"+url, s.Tag(ext.HTTPURL))
	assert.Equal("testvalue", s.Tag("testkey"))
	assert.Equal("500: Internal Server Error", s.Tag(ext.Error).(error).Error())
	assert.Equal("julienschmidt/httprouter", s.Tag(ext.Component))
	assert.Equal(ext.SpanKindServer, s.Tag(ext.SpanKind))
}

func TestAnalyticsSettings(t *testing.T) {
	assertRate := func(t *testing.T, mt mocktracer.Tracer, rate interface{}, opts ...RouterOption) {
		router := New(opts...)
		router.GET("/200", handler200)
		r := httptest.NewRequest("GET", "/200", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		spans := mt.FinishedSpans()
		assert.Len(t, spans, 1)
		s := spans[0]
		assert.Equal(t, rate, s.Tag(ext.EventSampleRate))
	}

	t.Run("defaults", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, nil)
	})

	t.Run("global", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		assertRate(t, mt, 0.4)
	})

	t.Run("enabled", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, 1.0, WithAnalytics(true))
	})

	t.Run("disabled", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		assertRate(t, mt, nil, WithAnalytics(false))
	})

	t.Run("override", func(t *testing.T) {
		mt := mocktracer.Start()
		defer mt.Stop()

		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		assertRate(t, mt, 0.23, WithAnalyticsRate(0.23))
	})
}

func router() http.Handler {
	router := New(
		WithServiceName("my-service"),
		WithSpanOptions(tracer.Tag("testkey", "testvalue")),
	)

	router.GET("/200", handler200)
	router.GET("/200/:parameter", handler200Parameter)
	router.GET("/500", handler500)

	return router
}

func handler200(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.Write([]byte("OK\n"))
}

func handler200Parameter(w http.ResponseWriter, _ *http.Request, p httprouter.Params) {
	w.Write([]byte(p.ByName("parameter")))
}

func handler500(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	http.Error(w, "500!", http.StatusInternalServerError)
}

func TestNamingSchema(t *testing.T) {
	genSpans := namingschematest.GenSpansFn(func(t *testing.T, serviceOverride string) []mocktracer.Span {
		var opts []RouterOption
		if serviceOverride != "" {
			opts = append(opts, WithServiceName(serviceOverride))
		}
		mt := mocktracer.Start()
		defer mt.Stop()

		mux := New(opts...)
		mux.GET("/200", handler200)
		r := httptest.NewRequest("GET", "/200", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		return mt.FinishedSpans()
	})
	namingschematest.NewHTTPServerTest(genSpans, "http.router")(t)
}
