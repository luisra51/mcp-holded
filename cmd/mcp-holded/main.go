// Entry point for the MCP server. Supports stdio, SSE, and streamable-http
// transports. Config comes from environment variables (see holded/config.go)
// or HTTP headers (SSE / streamable-http only, see holded/context.go).
//
// File path: cmd/mcp-holded/main.go
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"

	"github.com/luisra51/mcp-holded/holded"
	"github.com/luisra51/mcp-holded/tools"
)

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

type httpServer interface {
	Start(addr string) error
	Shutdown(ctx context.Context) error
}

func runHTTPServer(ctx context.Context, srv httpServer, addr, transportName string) error {
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Start(addr); err != nil {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info(fmt.Sprintf("%s server shutting down...", transportName))
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown error: %v", err)
		}
		select {
		case err := <-serverErr:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("server error during shutdown: %v", err)
			}
		case <-shutdownCtx.Done():
			slog.Warn(fmt.Sprintf("%s server did not stop gracefully within timeout", transportName))
		}
	}
	return nil
}

func newServer() *server.MCPServer {
	s := server.NewMCPServer(
		"mcp-holded",
		"0.1.0",
		server.WithInstructions(`
This server provides access to the Holded Invoicing API.

Note: Write tools are disabled unless HOLDED_ALLOW_WRITE=true.
`),
	)

	tools.AddAllTools(s)
	return s
}

func run(transport, addr, basePath, endpointPath string, logLevel slog.Level) error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))
	s := newServer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		<-sigChan
		slog.Info("Received shutdown signal")
		cancel()
		if transport == "stdio" {
			_ = os.Stdin.Close()
		}
	}()

	switch transport {
	case "stdio":
		srv := server.NewStdioServer(s)
		srv.SetContextFunc(holded.ComposedStdioContextFunc())
		slog.Info("Starting MCP server using stdio transport")
		err := srv.Listen(ctx, os.Stdin, os.Stdout)
		if err != nil && err != context.Canceled {
			return fmt.Errorf("server error: %v", err)
		}
		return nil
	case "sse":
		httpSrv := &http.Server{Addr: addr}
		srv := server.NewSSEServer(s,
			server.WithSSEContextFunc(holded.ComposedSSEContextFunc()),
			server.WithStaticBasePath(basePath),
			server.WithHTTPServer(httpSrv),
		)
		mux := http.NewServeMux()
		if basePath == "" {
			basePath = "/"
		}
		mux.Handle(basePath, srv)
		mux.HandleFunc("/healthz", handleHealthz)
		httpSrv.Handler = mux
		slog.Info("Starting MCP server using SSE transport", "address", addr, "basePath", basePath)
		return runHTTPServer(ctx, srv, addr, "SSE")
	case "streamable-http":
		httpSrv := &http.Server{Addr: addr}
		srv := server.NewStreamableHTTPServer(s,
			server.WithHTTPContextFunc(holded.ComposedHTTPContextFunc()),
			server.WithEndpointPath(endpointPath),
			server.WithStreamableHTTPServer(httpSrv),
		)
		mux := http.NewServeMux()
		mux.Handle(endpointPath, srv)
		mux.HandleFunc("/healthz", handleHealthz)
		httpSrv.Handler = mux
		slog.Info("Starting MCP server using StreamableHTTP transport", "address", addr, "endpointPath", endpointPath)
		return runHTTPServer(ctx, srv, addr, "StreamableHTTP")
	default:
		return fmt.Errorf("invalid transport type: %s. Must be 'stdio', 'sse' or 'streamable-http'", transport)
	}
}

func main() {
	var transport string
	flag.StringVar(&transport, "t", "stdio", "Transport type (stdio, sse or streamable-http)")
	flag.StringVar(&transport, "transport", "stdio", "Transport type (stdio, sse or streamable-http)")
	addr := flag.String("address", "localhost:8000", "The host and port to start the sse/http server on")
	basePath := flag.String("base-path", "", "Base path for the sse server")
	endpointPath := flag.String("endpoint-path", "/mcp", "Endpoint path for the streamable-http server")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	if err := run(transport, *addr, *basePath, *endpointPath, parseLevel(*logLevel)); err != nil {
		panic(err)
	}
}

func parseLevel(level string) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return slog.LevelInfo
	}
	return l
}
