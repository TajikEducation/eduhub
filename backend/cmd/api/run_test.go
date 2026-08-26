package main

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/abdulhalim/eduhub/backend/internal/platform/config"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
	"github.com/abdulhalim/eduhub/backend/internal/platform/logger"
)

// fakePinger — дублёр Pinger для тестов, не открывает реальное соединение к БД.
type fakePinger struct {
	closed bool
}

func (f *fakePinger) Ping(ctx context.Context) error { return nil }
func (f *fakePinger) Close()                         { f.closed = true }

func TestRun_ServesHealthzOnEphemeralPort(t *testing.T) {
	log := logger.New("info", "test", io.Discard)

	router := httpx.NewRouter(log)
	router.Handle("GET /healthz", httpx.Healthz(log))

	cfg := config.Config{
		HTTPAddr:        ":0",
		ShutdownTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addrCh := make(chan string, 1)
	deps := Deps{
		Logger:  log,
		Pool:    &fakePinger{},
		Handler: router,
		Ready:   func(addr string) { addrCh <- addr },
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, cfg, deps)
	}()

	var addr string
	select {
	case addr = <-addrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server address")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+"/healthz", nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			t.Errorf("resp.Body.Close: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for run to return")
	}
}

func TestRun_GracefulShutdownWaitsForInFlightRequest(t *testing.T) {
	log := logger.New("info", "test", io.Discard)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	cfg := config.Config{
		HTTPAddr:        ":0",
		ShutdownTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addrCh := make(chan string, 1)
	deps := Deps{
		Logger:  log,
		Pool:    &fakePinger{},
		Handler: mux,
		Ready:   func(addr string) { addrCh <- addr },
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, cfg, deps)
	}()

	var addr string
	select {
	case addr = <-addrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server address")
	}

	statusCh := make(chan int, 1)
	getErrCh := make(chan error, 1)
	go func() {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+addr+"/slow", nil)
		if err != nil {
			getErrCh <- err
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			getErrCh <- err
			return
		}
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				getErrCh <- closeErr
			}
		}()
		statusCh <- resp.StatusCode
		getErrCh <- nil
	}()

	// Даём запросу реально начаться (in-flight), прежде чем отменять контекст.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(cfg.ShutdownTimeout + time.Second):
		t.Fatal("timed out waiting for run to return")
	}

	select {
	case err := <-getErrCh:
		if err != nil {
			t.Fatalf("GET /slow: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for /slow request to complete")
	}

	select {
	case status := <-statusCh:
		if status != http.StatusOK {
			t.Fatalf("status = %d, want %d", status, http.StatusOK)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for /slow status")
	}
}

func TestRun_ClosesPoolAfterShutdown(t *testing.T) {
	log := logger.New("info", "test", io.Discard)

	cfg := config.Config{
		HTTPAddr:        ":0",
		ShutdownTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pinger := &fakePinger{}
	addrCh := make(chan string, 1)
	deps := Deps{
		Logger:  log,
		Pool:    pinger,
		Handler: http.NewServeMux(),
		Ready:   func(addr string) { addrCh <- addr },
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, cfg, deps)
	}()

	select {
	case <-addrCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server address")
	}

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(cfg.ShutdownTimeout + time.Second):
		t.Fatal("timed out waiting for run to return")
	}

	if !pinger.closed {
		t.Fatal("expected pool to be closed after run returns")
	}
}
