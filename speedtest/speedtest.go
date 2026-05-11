package speedtest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Result holds the outcome of a completed speed test.
type Result struct {
	DownloadMbps    float64
	LatencyMs       int64
	DurationSeconds float64
	Endpoints       []string
	ParallelConns   int
}

// Config holds the parameters for a single test run.
type Config struct {
	Endpoints []string
	Parallel  int
	Timeout   time.Duration
}

var chunkSizes = []int64{
	1_000_000,
	5_000_000,
	10_000_000,
	25_000_000,
	50_000_000,
}

// Run executes the adaptive download speed test.
// ctx cancellation stops the test immediately.
// progress is called approximately every 1 second with (currentMbps, elapsed).
// Returns (nil, ctx.Err()) if cancelled before any data was received.
func Run(ctx context.Context, cfg Config, progress func(float64, time.Duration)) (*Result, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, errors.New("no endpoints configured")
	}
	if cfg.Parallel < 1 {
		cfg.Parallel = 1
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}

	tctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	var totalBytes atomic.Int64
	var errCount atomic.Int64
	chunkIdx := 0
	start := time.Now()

	var wg sync.WaitGroup
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: false}}

	fetchChunk := func(endpoint string, chunkBytes int64) {
		defer wg.Done()
		url := endpoint
		if chunkBytes > 0 {
			url = endpoint + "?bytes=" + itoa(chunkBytes)
		}
		req, err := http.NewRequestWithContext(tctx, "GET", url, nil)
		if err != nil {
			errCount.Add(1)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			errCount.Add(1)
			return
		}
		defer resp.Body.Close()
		n, _ := io.Copy(io.Discard, resp.Body)
		totalBytes.Add(n)
	}

	// Launch initial parallel workers
	for i := 0; i < cfg.Parallel; i++ {
		wg.Add(1)
		endpoint := cfg.Endpoints[i%len(cfg.Endpoints)]
		go fetchChunk(endpoint, chunkSizes[chunkIdx])
	}

	// Progress ticker + adaptive chunk advancement
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-tctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(start)
				bytes := totalBytes.Load()
				if elapsed.Seconds() > 0 {
					mbps := float64(bytes*8) / elapsed.Seconds() / 1_000_000
					if progress != nil {
						progress(mbps, elapsed)
					}
					// Advance chunk size if on a fast link
					if elapsed.Seconds() < 3 && mbps > 200 && chunkIdx < len(chunkSizes)-1 {
						chunkIdx++
					}
					// Launch another wave of workers to keep saturation
					if elapsed < cfg.Timeout-1*time.Second {
						for i := 0; i < cfg.Parallel; i++ {
							wg.Add(1)
							endpoint := cfg.Endpoints[i%len(cfg.Endpoints)]
							go fetchChunk(endpoint, chunkSizes[chunkIdx])
						}
					}
				}
			}
		}
	}()

	// Wait for timeout or cancellation
	<-tctx.Done()
	wg.Wait()

	elapsed := time.Since(start)
	bytes := totalBytes.Load()

	if ctx.Err() != nil && bytes == 0 {
		return nil, ctx.Err()
	}

	if errCount.Load() > 0 && bytes == 0 {
		return nil, errors.New("all_endpoints_failed")
	}

	downloadMbps := 0.0
	if elapsed.Seconds() > 0 {
		downloadMbps = float64(bytes*8) / elapsed.Seconds() / 1_000_000
	}
	// Round to 1 decimal
	downloadMbps = float64(int(downloadMbps*10+0.5)) / 10

	return &Result{
		DownloadMbps:    downloadMbps,
		DurationSeconds: elapsed.Seconds(),
		Endpoints:       cfg.Endpoints,
		ParallelConns:   cfg.Parallel,
	}, nil
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
