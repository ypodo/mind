package app

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var (
	mutex sync.Mutex
	// Modified to include method as part of the key
	httpRequestsTotal = make(map[string]int)
)

// getMetricKey creates a unique key for each path+method combination
func getMetricKey(path, method string) string {
	return fmt.Sprintf("%s|%s", path, method)
}

// parseMetricKey extracts path and method from a metric key
func parseMetricKey(key string) (path string, method string) {
	parts := strings.Split(key, "|")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return key, ""
}

func CountRequestsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the path and method from the request
		path := r.URL.Path
		method := r.Method

		// Create a unique key for this path+method combination
		key := getMetricKey(path, method)

		// Increment the counter for the path+method
		mutex.Lock()
		httpRequestsTotal[key]++
		mutex.Unlock()

		next.ServeHTTP(w, r)
	})
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	mutex.Lock()
	defer mutex.Unlock()

	// Write the metrics in Prometheus format
	for key, count := range httpRequestsTotal {
		path, method := parseMetricKey(key)
		fmt.Fprintf(w, "petstore_http_requests_total{path=\"%s\",method=\"%s\"} %d\n", path, method, count)
	}
}