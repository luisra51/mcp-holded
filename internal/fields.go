package internal

// FilterFields keeps only the requested top-level keys in each object of a
// decoded JSON payload. Arrays are filtered element by element; an empty
// fields list returns the payload unchanged.
func FilterFields(v any, fields []string) any {
	if len(fields) == 0 {
		return v
	}
	keep := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		keep[f] = struct{}{}
	}
	return filterValue(v, keep)
}

func filterValue(v any, keep map[string]struct{}) any {
	switch t := v.(type) {
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = filterValue(item, keep)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(keep))
		for k, v2 := range t {
			if _, ok := keep[k]; ok {
				out[k] = v2
			}
		}
		return out
	default:
		return v
	}
}
