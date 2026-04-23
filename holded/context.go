package holded

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/server"
)

type configKey struct{}
type clientKey struct{}

type httpContextFunc func(ctx context.Context, req *http.Request) context.Context

func WithConfig(ctx context.Context, cfg Config) context.Context {
	return context.WithValue(ctx, configKey{}, cfg)
}

func ConfigFromContext(ctx context.Context) Config {
	if cfg, ok := ctx.Value(configKey{}).(Config); ok {
		return cfg
	}
	return Config{}
}

func WithClient(ctx context.Context, client *Client) context.Context {
	return context.WithValue(ctx, clientKey{}, client)
}

func ClientFromContext(ctx context.Context) *Client {
	c, ok := ctx.Value(clientKey{}).(*Client)
	if !ok {
		return nil
	}
	return c
}

func ExtractInfoFromEnv(ctx context.Context) context.Context {
	cfg := ConfigFromContext(ctx)
	cfg.URL = baseURLFromEnv()
	cfg.APIKey = apiKeyFromEnv()
	cfg.AllowWrite = envBool(allowWriteEnvVar)
	cfg.AllowedTools = allowedToolsFromEnv()
	cfg.Timeout = timeoutFromEnv()
	cfg.Debug = envBool(debugEnvVar)
	cfg.DisableRateLimit = envBool(disableRateLimitEnvVar)
	return WithConfig(ctx, cfg)
}

func ExtractInfoFromHeaders(ctx context.Context, req *http.Request) context.Context {
	cfg := ConfigFromContext(ctx)
	u := strings.TrimRight(req.Header.Get(urlHeader), "/")
	if u == "" {
		u = baseURLFromEnv()
	}
	apiKey := strings.TrimSpace(req.Header.Get(apiKeyHeader))
	if apiKey == "" {
		apiKey = apiKeyFromEnv()
	}
	cfg.URL = u
	cfg.APIKey = apiKey
	cfg.AllowWrite = envBool(allowWriteEnvVar)
	cfg.AllowedTools = allowedToolsFromEnv()
	cfg.Timeout = timeoutFromEnv()
	cfg.Debug = envBool(debugEnvVar)
	cfg.DisableRateLimit = envBool(disableRateLimitEnvVar)
	return WithConfig(ctx, cfg)
}

func ExtractClientFromEnv(ctx context.Context) context.Context {
	cfg := ConfigFromContext(ctx)
	if cfg.URL == "" || cfg.APIKey == "" {
		slog.Warn("missing config", "url_set", cfg.URL != "", "api_key_set", cfg.APIKey != "")
	}
	return WithClient(ctx, NewClient(cfg))
}

func ExtractClientFromHeaders(ctx context.Context, _ *http.Request) context.Context {
	cfg := ConfigFromContext(ctx)
	if cfg.URL == "" || cfg.APIKey == "" {
		slog.Warn("missing config", "url_set", cfg.URL != "", "api_key_set", cfg.APIKey != "")
	}
	return WithClient(ctx, NewClient(cfg))
}

func ComposeStdioContextFuncs(funcs ...server.StdioContextFunc) server.StdioContextFunc {
	return func(ctx context.Context) context.Context {
		for _, f := range funcs {
			ctx = f(ctx)
		}
		return ctx
	}
}

func ComposeSSEContextFuncs(funcs ...httpContextFunc) server.SSEContextFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		for _, f := range funcs {
			ctx = f(ctx, req)
		}
		return ctx
	}
}

func ComposeHTTPContextFuncs(funcs ...httpContextFunc) server.HTTPContextFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		for _, f := range funcs {
			ctx = f(ctx, req)
		}
		return ctx
	}
}

func ComposedStdioContextFunc() server.StdioContextFunc {
	return ComposeStdioContextFuncs(
		ExtractInfoFromEnv,
		ExtractClientFromEnv,
	)
}

func ComposedSSEContextFunc() server.SSEContextFunc {
	return ComposeSSEContextFuncs(
		ExtractInfoFromHeaders,
		ExtractClientFromHeaders,
	)
}

func ComposedHTTPContextFunc() server.HTTPContextFunc {
	return ComposeHTTPContextFuncs(
		ExtractInfoFromHeaders,
		ExtractClientFromHeaders,
	)
}
