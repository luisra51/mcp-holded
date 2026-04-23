package internal

// Wrap returns a standard response envelope with data and optional meta.
func Wrap(data any, meta any) map[string]any {
	resp := map[string]any{"data": data}
	if meta != nil {
		resp["meta"] = meta
	}
	return resp
}
