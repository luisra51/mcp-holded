package tools

import (
	"net/url"
	"reflect"
	"testing"
)

func TestCompactBody(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name: "discard empty string",
			input: map[string]any{
				"name":  "",
				"email": "test@example.com",
			},
			expected: map[string]any{
				"email": "test@example.com",
			},
		},
		{
			name: "discard zero int",
			input: map[string]any{
				"count":  0,
				"amount": 100,
			},
			expected: map[string]any{
				"amount": 100,
			},
		},
		{
			name: "discard zero float64",
			input: map[string]any{
				"price": 0.0,
				"tax":   5.5,
			},
			expected: map[string]any{
				"tax": 5.5,
			},
		},
		{
			name: "preserve bool true",
			input: map[string]any{
				"active": true,
				"name":   "test",
			},
			expected: map[string]any{
				"active": true,
				"name":   "test",
			},
		},
		{
			name: "preserve bool false",
			input: map[string]any{
				"active": false,
				"name":   "test",
			},
			expected: map[string]any{
				"active": false,
				"name":   "test",
			},
		},
		{
			name: "preserve non-empty string",
			input: map[string]any{
				"name": "John",
			},
			expected: map[string]any{
				"name": "John",
			},
		},
		{
			name: "preserve non-zero int",
			input: map[string]any{
				"count": 42,
			},
			expected: map[string]any{
				"count": 42,
			},
		},
		{
			name: "preserve non-zero float64",
			input: map[string]any{
				"price": 9.99,
			},
			expected: map[string]any{
				"price": 9.99,
			},
		},
		{
			name: "discard empty slice",
			input: map[string]any{
				"tags": []string{},
				"name": "test",
			},
			expected: map[string]any{
				"name": "test",
			},
		},
		{
			name: "preserve non-empty slice",
			input: map[string]any{
				"tags": []string{"a", "b"},
				"name": "test",
			},
			expected: map[string]any{
				"tags": []string{"a", "b"},
				"name": "test",
			},
		},
		{
			name: "discard empty map",
			input: map[string]any{
				"metadata": map[string]any{},
				"name":     "test",
			},
			expected: map[string]any{
				"name": "test",
			},
		},
		{
			name: "preserve non-empty map",
			input: map[string]any{
				"metadata": map[string]any{"key": "value"},
				"name":     "test",
			},
			expected: map[string]any{
				"metadata": map[string]any{"key": "value"},
				"name":     "test",
			},
		},
		{
			name: "complex scenario with multiple types",
			input: map[string]any{
				"name":     "Product",
				"price":    0.0,
				"count":    0,
				"active":   true,
				"inactive": false,
				"tags":     []string{},
				"attrs":    map[string]any{"key": "val"},
				"empty":    "",
			},
			expected: map[string]any{
				"name":     "Product",
				"active":   true,
				"inactive": false,
				"attrs":    map[string]any{"key": "val"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compactBody(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("compactBody() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAddListParams(t *testing.T) {
	tests := []struct {
		name          string
		q             url.Values
		args          ListParams
		expectedQuery map[string][]string
		expectedMeta  map[string]any
		expectError   bool
	}{
		{
			name: "page 0 becomes page 1, limit 0 becomes 50",
			q:    nil,
			args: ListParams{Page: 0, Limit: 0},
			expectedQuery: map[string][]string{
				"page":  {"1"},
				"limit": {"50"},
			},
			expectedMeta: map[string]any{
				"page":  1,
				"limit": 50,
			},
			expectError: false,
		},
		{
			name: "limit 501 returns error",
			q:    nil,
			args: ListParams{Page: 1, Limit: 501},
			expectedQuery: nil,
			expectedMeta:  nil,
			expectError:   true,
		},
		{
			name: "starttmp and endtmp added to query",
			q:    nil,
			args: ListParams{Page: 1, Limit: 10, StartTmp: "1000", EndTmp: "2000"},
			expectedQuery: map[string][]string{
				"page":     {"1"},
				"limit":    {"10"},
				"starttmp": {"1000"},
				"endtmp":   {"2000"},
			},
			expectedMeta: map[string]any{
				"page":  1,
				"limit": 10,
			},
			expectError: false,
		},
		{
			name: "only starttmp added when endtmp empty",
			q:    nil,
			args: ListParams{Page: 2, Limit: 20, StartTmp: "1500", EndTmp: ""},
			expectedQuery: map[string][]string{
				"page":     {"2"},
				"limit":    {"20"},
				"starttmp": {"1500"},
			},
			expectedMeta: map[string]any{
				"page":  2,
				"limit": 20,
			},
			expectError: false,
		},
		{
			name: "meta includes fields when provided",
			q:    nil,
			args: ListParams{Page: 1, Limit: 50, Fields: []string{"id", "name"}},
			expectedQuery: map[string][]string{
				"page":  {"1"},
				"limit": {"50"},
			},
			expectedMeta: map[string]any{
				"page":   1,
				"limit":  50,
				"fields": []string{"id", "name"},
			},
			expectError: false,
		},
		{
			name: "limit 1 is valid",
			q:    nil,
			args: ListParams{Page: 1, Limit: 1},
			expectedQuery: map[string][]string{
				"page":  {"1"},
				"limit": {"1"},
			},
			expectedMeta: map[string]any{
				"page":  1,
				"limit": 1,
			},
			expectError: false,
		},
		{
			name: "limit 500 is valid",
			q:    nil,
			args: ListParams{Page: 1, Limit: 500},
			expectedQuery: map[string][]string{
				"page":  {"1"},
				"limit": {"500"},
			},
			expectedMeta: map[string]any{
				"page":  1,
				"limit": 500,
			},
			expectError: false,
		},
		{
			name: "existing query values are preserved",
			q:    url.Values{"existing": []string{"value"}},
			args: ListParams{Page: 1, Limit: 50},
			expectedQuery: map[string][]string{
				"existing": {"value"},
				"page":     {"1"},
				"limit":    {"50"},
			},
			expectedMeta: map[string]any{
				"page":  1,
				"limit": 50,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, meta, err := addListParams(tt.q, tt.args)
			if (err != nil) != tt.expectError {
				t.Errorf("addListParams() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if tt.expectError {
				return
			}
			if !reflect.DeepEqual(q, url.Values(tt.expectedQuery)) {
				t.Errorf("addListParams() query = %v, want %v", q, tt.expectedQuery)
			}
			if !reflect.DeepEqual(meta, tt.expectedMeta) {
				t.Errorf("addListParams() meta = %v, want %v", meta, tt.expectedMeta)
			}
		})
	}
}

func TestValidateDocumentType(t *testing.T) {
	tests := []struct {
		name        string
		docType     string
		expectError bool
	}{
		{
			name:        "invoice is valid",
			docType:     "invoice",
			expectError: false,
		},
		{
			name:        "salesreceipt is valid",
			docType:     "salesreceipt",
			expectError: false,
		},
		{
			name:        "creditnote is valid",
			docType:     "creditnote",
			expectError: false,
		},
		{
			name:        "receiptnote is valid",
			docType:     "receiptnote",
			expectError: false,
		},
		{
			name:        "estimate is valid",
			docType:     "estimate",
			expectError: false,
		},
		{
			name:        "salesorder is valid",
			docType:     "salesorder",
			expectError: false,
		},
		{
			name:        "waybill is valid",
			docType:     "waybill",
			expectError: false,
		},
		{
			name:        "proform is valid",
			docType:     "proform",
			expectError: false,
		},
		{
			name:        "purchase is valid",
			docType:     "purchase",
			expectError: false,
		},
		{
			name:        "purchaserefund is valid",
			docType:     "purchaserefund",
			expectError: false,
		},
		{
			name:        "purchaseorder is valid",
			docType:     "purchaseorder",
			expectError: false,
		},
		{
			name:        "notatype is invalid",
			docType:     "notatype",
			expectError: true,
		},
		{
			name:        "empty string is invalid",
			docType:     "",
			expectError: true,
		},
		{
			name:        "random string is invalid",
			docType:     "randomtype",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDocumentType(tt.docType)
			if (err != nil) != tt.expectError {
				t.Errorf("validateDocumentType(%q) error = %v, expectError %v", tt.docType, err, tt.expectError)
			}
		})
	}
}
