package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"

	mcpholded "github.com/luisra51/mcp-holded"
	"github.com/luisra51/mcp-holded/holded"
	"github.com/luisra51/mcp-holded/internal"
)

var documentTypes = []string{
	"invoice",
	"salesreceipt",
	"creditnote",
	"receiptnote",
	"estimate",
	"salesorder",
	"waybill",
	"proform",
	"purchase",
	"purchaserefund",
	"purchaseorder",
}

type ListParams struct {
	Page     int      `json:"page,omitempty" jsonschema:"description=Page number starting from 1"`
	Limit    int      `json:"limit,omitempty" jsonschema:"description=Maximum number of items to return between 1 and 500"`
	Summary  bool     `json:"summary,omitempty" jsonschema:"description=Return only metadata when supported"`
	Fields   []string `json:"fields,omitempty" jsonschema:"description=Optional list of fields to keep in each returned object"`
	StartTmp string   `json:"starttmp,omitempty" jsonschema:"description=Start Unix timestamp filter"`
	EndTmp   string   `json:"endtmp,omitempty" jsonschema:"description=End Unix timestamp filter"`
}

func readOnlyOptions(title string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithTitleAnnotation(title),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithReadOnlyHintAnnotation(true),
	}
}

func writeOptions(title string) []mcp.ToolOption {
	return []mcp.ToolOption{mcp.WithTitleAnnotation(title)}
}

func destructiveOptions(title string) []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithTitleAnnotation(title),
		mcp.WithDestructiveHintAnnotation(true),
	}
}

func requireClient(ctx context.Context) (*holded.Client, error) {
	client := holded.ClientFromContext(ctx)
	if client == nil {
		return nil, &mcpholded.HardError{Err: holded.ErrMissingClient}
	}
	return client, nil
}

func doJSON(ctx context.Context, toolName string, write bool, method, path string, q url.Values, body any, meta any) (any, error) {
	if err := ensureToolAllowed(ctx, toolName, write); err != nil {
		return nil, err
	}
	client, err := requireClient(ctx)
	if err != nil {
		return nil, err
	}
	var payload any
	req, err := client.NewRequest(method, path, q, body)
	if err != nil {
		return nil, err
	}
	if err := client.DoJSON(req.WithContext(ctx), &payload); err != nil {
		return nil, err
	}
	return internal.Wrap(internal.MaskSensitive(payload), meta), nil
}

func doRawBase64(ctx context.Context, toolName, path string) (any, error) {
	if err := ensureToolAllowed(ctx, toolName, false); err != nil {
		return nil, err
	}
	client, err := requireClient(ctx)
	if err != nil {
		return nil, err
	}
	req, err := client.NewRequest(http.MethodGet, path, url.Values{}, nil)
	if err != nil {
		return nil, err
	}
	body, contentType, err := client.DoRaw(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return internal.Wrap(map[string]any{
		"content_type": contentType,
		"base64":       base64.StdEncoding.EncodeToString(body),
	}, nil), nil
}

func addListParams(q url.Values, args ListParams) (url.Values, map[string]any, error) {
	if q == nil {
		q = url.Values{}
	}
	page := internal.NormalizePage(args.Page)
	limit := args.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		return nil, nil, fmt.Errorf("limit must be between 1 and 500")
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("limit", strconv.Itoa(limit))
	if args.StartTmp != "" {
		q.Set("starttmp", args.StartTmp)
	}
	if args.EndTmp != "" {
		q.Set("endtmp", args.EndTmp)
	}
	meta := map[string]any{
		"page":    page,
		"limit":   limit,
		"summary": args.Summary,
	}
	if len(args.Fields) > 0 {
		meta["fields"] = args.Fields
	}
	return q, meta, nil
}

func validateDocumentType(docType string) error {
	return internal.RequireOneOf(docType, "doc_type", documentTypes...)
}

func compactBody(values map[string]any) map[string]any {
	out := make(map[string]any, len(values))
	for k, v := range values {
		switch t := v.(type) {
		case string:
			if t != "" {
				out[k] = t
			}
		case int:
			if t != 0 {
				out[k] = t
			}
		case float64:
			if t != 0 {
				out[k] = t
			}
		case bool:
			out[k] = t
		case []string:
			if len(t) > 0 {
				out[k] = t
			}
		case []map[string]any:
			if len(t) > 0 {
				out[k] = t
			}
		case map[string]any:
			if len(t) > 0 {
				out[k] = t
			}
		default:
			if v != nil {
				out[k] = v
			}
		}
	}
	return out
}
