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

const maxParallel = 64

const uploadChunk = 1_000_000

type Result struct {
	DownloadMbps    float64
	LatencyMs       int64
	DurationSeconds float64
	Endpoints       []string
	ParallelConns   int
}

type Config struct {
	Endpoints    []string
	Parallel     int
	Timeout      time.Duration
	UploadTarget string
}

const maxConsecutiveFailures = 6

func newClient() *http.Client {
	return &http.Client{Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConnsPerHost:   maxParallel,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
	}}
}

type failGate struct {
	consecutive int
	transferred *atomic.Int64
	errors      *atomic.Int64
}

func (g *failGate) fail(ctx context.Context) bool {
	g.errors.Add(1)
	g.consecutive++
	if g.consecutive >= maxConsecutiveFailures && g.transferred.Load() == 0 {
		return false
	}
	backoff := min(time.Duration(g.consecutive)*200*time.Millisecond, 2*time.Second)
	select {
	case <-ctx.Done():
		return false
	case <-time.After(backoff):
		return true
	}
}

func (g *failGate) ok() { g.consecutive = 0 }

func waitWorkers(ctx context.Context, wg *sync.WaitGroup) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-ctx.Done():
		<-done
	case <-done:
	}
}

var chunkSizes = []int64{
	1_000_000,
	5_000_000,
	10_000_000,
	25_000_000,
	50_000_000,
}

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

	client := newClient()
	defer client.CloseIdleConnections()

	latencyMs := measureLatency(ctx, client, cfg.Endpoints[0])

	tctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	var totalBytes atomic.Int64
	var errCount atomic.Int64
	var chunkIdx atomic.Int64
	start := time.Now()

	var wg sync.WaitGroup
	worker := func(id int) {
		defer wg.Done()
		gate := &failGate{transferred: &totalBytes, errors: &errCount}
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
				if !gate.fail(tctx) {
					return
				}
				continue
			}
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				resp.Body.Close()
				if !gate.fail(tctx) {
					return
				}
				continue
			}
			n, _ := io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			totalBytes.Add(n)
			gate.ok()
		}
	}
	for i := 0; i < cfg.Parallel; i++ {
		wg.Add(1)
		go worker(i)
	}

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

	waitWorkers(tctx, &wg)
	cancel()

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

func MeasureUpload(ctx context.Context, cfg Config, progress func(float64, time.Duration)) (float64, error) {
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

	client := newClient()
	defer client.CloseIdleConnections()

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		gate := &failGate{transferred: &totalBytes, errors: &errCount}
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
				if !gate.fail(tctx) {
					return
				}
				continue
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				if !gate.fail(tctx) {
					return
				}
				continue
			}
			totalBytes.Add(int64(len(payload)))
			gate.ok()
		}
	}
	for i := 0; i < cfg.Parallel; i++ {
		wg.Add(1)
		go worker()
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-tctx.Done():
				return
			case <-ticker.C:
				el := time.Since(start)
				if progress != nil && el.Seconds() > 0 {
					progress(float64(totalBytes.Load()*8)/el.Seconds()/1_000_000, el)
				}
			}
		}
	}()

	waitWorkers(tctx, &wg)
	cancel()

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
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return time.Since(t0).Milliseconds()
}
