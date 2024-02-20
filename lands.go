// Package lands is a ListenAndServe helper that chooses the first listening
// address from:
//
// 1. Inherited via systemd LISTEN_FDS.
// 2. PORT number via environment variable.
// 3. Default address passed in as argument.
//
// Additionally it enables graceful shutdown via the Context, as well as the
// SIGTERM and SIGINT signals.
package lands

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ListenAndServe sets up the listerer and starts serving the given handler.
func ListenAndServe(ctx context.Context, addr string, handler http.Handler) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var err error
	var l net.Listener
	if os.Getenv("LISTEN_FDS") == "1" {
		f := os.NewFile(uintptr(3), "http")
		defer f.Close()
		l, err = net.FileListener(f)
	} else {
		port := os.Getenv("PORT")
		if port != "" {
			addr = ":" + port
		}
		l, err = net.Listen("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("lands: error setting up listener: %w", err)
	}

	hs := &http.Server{Handler: handler}

	errShutdown := make(chan error, 1)
	go func() {
		defer close(errShutdown)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := hs.Shutdown(shutdownCtx); err != nil {
			errShutdown <- err
		}
	}()

	fmt.Println("Serving on http://" + l.Addr().String() + "/")
	switch err := hs.Serve(l); err {
	case nil:
		return nil
	case http.ErrServerClosed:
		if err := <-errShutdown; err != nil {
			return fmt.Errorf("lands: error shutting down: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("lands: error serving: %w", err)
	}
}
