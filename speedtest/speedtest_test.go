package speedtest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestMeasureUpload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	mbps, err := MeasureUpload(context.Background(),
		Config{UploadTarget: srv.URL + "/__up", Parallel: 2, Timeout: time.Second})
	if err != nil {
		t.Fatalf("MeasureUpload error: %v", err)
	}
	if mbps <= 0 {
		t.Fatalf("upload mbps = %v, want > 0", mbps)
	}
}

func TestMeasureUploadNoTarget(t *testing.T) {
	if _, err := MeasureUpload(context.Background(), Config{Timeout: time.Second}); err == nil {
		t.Error("expected error when upload target is empty")
	}
}

func sizedServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int64(64_000)
		if v := r.URL.Query().Get("bytes"); v != "" {
			if p, err := strconv.ParseInt(v, 10, 64); err == nil {
				n = p
			}
		}
		w.Write(make([]byte, n))
	}))
}

func TestRunDownloads(t *testing.T) {
	srv := sizedServer()
	defer srv.Close()

	res, err := Run(context.Background(),
		Config{Endpoints: []string{srv.URL + "/__down"}, Parallel: 4, Timeout: time.Second}, nil)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if res == nil || res.DownloadMbps <= 0 {
		t.Fatalf("expected download > 0, got %+v", res)
	}
	if res.ParallelConns != 4 {
		t.Errorf("ParallelConns = %d, want 4", res.ParallelConns)
	}
}

func TestRunClampsParallel(t *testing.T) {
	srv := sizedServer()
	defer srv.Close()

	res, err := Run(context.Background(),
		Config{Endpoints: []string{srv.URL}, Parallel: 100000, Timeout: 500 * time.Millisecond}, nil)
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if res.ParallelConns != maxParallel {
		t.Errorf("ParallelConns = %d, want clamp to %d", res.ParallelConns, maxParallel)
	}
}

func TestRunCancelledReturnsError(t *testing.T) {
	srv := sizedServer()
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled before Run

	if _, err := Run(ctx,
		Config{Endpoints: []string{srv.URL}, Parallel: 2, Timeout: 2 * time.Second}, nil); err == nil {
		t.Error("expected error when context is cancelled before any data, got nil")
	}
}
