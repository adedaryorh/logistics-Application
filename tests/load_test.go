package tests

import (
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGatewayHealthLoad(t *testing.T) {
	integrationEnabled(t)

	const workers = 10
	const requestsPerWorker = 20
	var successCount atomic.Int64

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				resp, err := http.Get(gatewayBaseURL() + "/healthz")
				if err != nil {
					continue
				}
				_, _ = io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					successCount.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	total := int64(workers * requestsPerWorker)
	if successCount.Load() < total*9/10 {
		t.Fatalf("expected at least 90%% success, got %d/%d", successCount.Load(), total)
	}
	if elapsed > 10*time.Second {
		t.Fatalf("expected load test under 10s, got %s", elapsed)
	}
}

func BenchmarkGatewayHealth(b *testing.B) {
	if testing.Short() {
		b.Skip("skip benchmark in short mode")
	}
	integrationEnabledForBenchmark(b)
	if gatewayBaseURL() == "" {
		b.Skip("gateway url unavailable")
	}
	for i := 0; i < b.N; i++ {
		resp, err := http.Get(gatewayBaseURL() + "/healthz")
		if err != nil {
			b.Fatalf("get healthz: %v", err)
		}
		_ = resp.Body.Close()
	}
}

func integrationEnabledForBenchmark(b *testing.B) {
	if os.Getenv("TEST_INTEGRATION") != "1" {
		b.Skip("set TEST_INTEGRATION=1 to run integration benchmarks")
	}
}
