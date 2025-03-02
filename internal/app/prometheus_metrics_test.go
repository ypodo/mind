package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCountRequestsMiddleware(t *testing.T) {
	// Clear the metrics before the test
	httpRequestsTotal = make(map[string]int)

	// Create a test handler that just returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the test handler with our middleware
	handler := CountRequestsMiddleware(testHandler)

	// Create test requests with different paths and methods
	testCases := []struct {
		path   string
		method string
	}{
		{"/api/pets", "GET"},
		{"/api/pets", "GET"},
		{"/api/pets", "POST"},
		{"/api/pets/123", "GET"},
	}

	// Send requests through our middleware
	for _, tc := range testCases {
		req, err := http.NewRequest(tc.method, tc.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
		}
	}

	// Check the metrics handler output
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	MetricsHandler(recorder, req)

	// Verify that metrics contain the expected labels
	metricsOutput := recorder.Body.String()
	t.Logf("Metrics output: %s", metricsOutput)

	// Check that metrics contain all expected path/method combinations
	expectedMetrics := []struct {
		path   string
		method string
		count  int
	}{
		{"/api/pets", "GET", 2},
		{"/api/pets", "POST", 1},
		{"/api/pets/123", "GET", 1},
	}
	
	for _, expected := range expectedMetrics {
		expectedString := fmt.Sprintf("petstore_http_requests_total{path=\"%s\",method=\"%s\"} %d", 
			expected.path, expected.method, expected.count)
		if !strings.Contains(metricsOutput, expectedString) {
			t.Errorf("Expected metrics to contain '%s', got:\n%s", expectedString, metricsOutput)
		}
	}
}

func TestGetMetricKey(t *testing.T) {
	key := getMetricKey("/api/pets", "GET")
	if key != "/api/pets|GET" {
		t.Errorf("Expected key to be '/api/pets|GET', got '%s'", key)
	}
}

func TestParseMetricKey(t *testing.T) {
	path, method := parseMetricKey("/api/pets|GET")
	if path != "/api/pets" || method != "GET" {
		t.Errorf("Expected path='/api/pets', method='GET', got path='%s', method='%s'", path, method)
	}
}