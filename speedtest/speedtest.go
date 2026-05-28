package speedtest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// maxParallel caps worker fan-out so a malformed config can't spawn an
// unbounded number of goroutines/connections.
const maxParallel = 64

// uploadChunk is the size of each POST body during the upload test.
const uploadChunk = 1_000_000

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
	Endpoints    []string
	Parallel     int
	Timeout      time.Duration
	UploadTarget string // optional; empty disables the upload phase
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
	if cfg.Parallel > maxParallel {
		cfg.Parallel = maxParallel
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}

	tctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	var totalBytes atomic.Int64
	var errCount atomic.Int64
	var chunkIdx atomic.Int64
	start := time.Now()

	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: false}}

	// Latency to the download endpoint itself (headers only), measured before
	// saturating the link — this is the speed test's own RTT, not the
	// connectivity monitor's ping latency.
	latencyMs := measureLatency(tctx, client, cfg.Endpoints[0])

	// Fixed worker pool: each worker fetches chunks in a loop until the context
	// ends. Every wg.Add happens here, before wg.Wait — so Add never races with
	// Wait (the previous design grew the WaitGroup from the ticker goroutine).
	var wg sync.WaitGroup
	worker := func(id int) {
		defer wg.Done()
		for tctx.Err() == nil {
			endpoint := cfg.Endpoints[id%len(cfg.Endpoints)]
			url := endpoint + "?bytes=" + strconv.FormatInt(chunkSizes[chunkIdx.Load()], 10)
			req, err := http.NewRequestWithContext(tctx, "GET", url, nil)
			if err != nil {
				errCount.Add(1)
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				errCount.Add(1)
				continue
			}
			n, _ := io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			totalBytes.Add(n)
		}
	}
	for i := 0; i < cfg.Parallel; i++ {
		wg.Add(1)
		go worker(i)
	}

	// Progress reporter + adaptive chunk advancement. Single goroutine, reads
	// shared counters atomically, never touches the WaitGroup.
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-tctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(start)
				if elapsed.Seconds() <= 0 {
					continue
				}
				mbps := float64(totalBytes.Load()*8) / elapsed.Seconds() / 1_000_000
				if progress != nil {
					progress(mbps, elapsed)
				}
				if elapsed.Seconds() < 3 && mbps > 200 {
					if i := chunkIdx.Load(); i < int64(len(chunkSizes))-1 {
						chunkIdx.Store(i + 1)
					}
				}
			}
		}
	}()

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
	downloadMbps = math.Round(downloadMbps*10) / 10

	return &Result{
		DownloadMbps:    downloadMbps,
		LatencyMs:       latencyMs,
		DurationSeconds: elapsed.Seconds(),
		Endpoints:       cfg.Endpoints,
		ParallelConns:   cfg.Parallel,
	}, nil
}

// MeasureUpload saturates the upload link by POSTing fixed-size chunks with a
// pool of workers until the timeout, returning the measured Mbps. Returns
// (0, err) if cancelled before any bytes were sent or all requests failed.
func MeasureUpload(ctx context.Context, cfg Config) (float64, error) {
	if cfg.UploadTarget == "" {
		return 0, errors.New("no upload target configured")
	}
	if cfg.Parallel < 1 {
		cfg.Parallel = 1
	}
	if cfg.Parallel > maxParallel {
		cfg.Parallel = maxParallel
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}

	tctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	var totalBytes atomic.Int64
	var errCount atomic.Int64
	start := time.Now()

	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: false}}

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		payload := make([]byte, uploadChunk)
		for tctx.Err() == nil {
			req, err := http.NewRequestWithContext(tctx, "POST", cfg.UploadTarget, bytes.NewReader(payload))
			if err != nil {
				errCount.Add(1)
				return
			}
			req.Header.Set("Content-Type", "application/octet-stream")
			resp, err := client.Do(req)
			if err != nil {
				errCount.Add(1)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			totalBytes.Add(int64(len(payload)))
		}
	}
	for i := 0; i < cfg.Parallel; i++ {
		wg.Add(1)
		go worker()
	}

	<-tctx.Done()
	wg.Wait()

	elapsed := time.Since(start)
	sent := totalBytes.Load()
	if ctx.Err() != nil && sent == 0 {
		return 0, ctx.Err()
	}
	if errCount.Load() > 0 && sent == 0 {
		return 0, errors.New("upload_failed")
	}

	mbps := 0.0
	if elapsed.Seconds() > 0 {
		mbps = float64(sent*8) / elapsed.Seconds() / 1_000_000
	}
	return math.Round(mbps*10) / 10, nil
}

// measureLatency does a single small request to the endpoint and returns the
// round-trip in ms, or 0 if it fails.
func measureLatency(ctx context.Context, client *http.Client, endpoint string) int64 {
	lctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(lctx, "GET", endpoint+"?bytes=0", nil)
	if err != nil {
		return 0
	}
	t0 := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return time.Since(t0).Milliseconds()
}
