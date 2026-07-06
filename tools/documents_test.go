package tools

import (
	"reflect"
	"testing"
)

func TestDocumentBodyMapsSnakeCaseParamsToHoldedCamelCase(t *testing.T) {
	args := DocumentCreateParams{
		DocType:        "invoice",
		ContactID:      "c1",
		Date:           1700000000,
		Notes:          "a note",
		InvoiceNum:     "F-001",
		SalesChannelID: "sc1",
		Items: []DocumentItem{
			{Name: "Widget", Units: 2, Subtotal: 10.5, Tax: 21, ServiceID: "svc1"},
			{Name: "Free line", Units: 1, Subtotal: 0},
		},
	}

	body := documentBody(args)

	if _, ok := body["doc_type"]; ok {
		t.Error("doc_type must not leak into the API body")
	}
	if body["contactId"] != "c1" {
		t.Errorf("contactId = %v, want c1", body["contactId"])
	}
	if body["date"] != int64(1700000000) {
		t.Errorf("date = %v, want 1700000000", body["date"])
	}
	if body["invoiceNum"] != "F-001" {
		t.Errorf("invoiceNum = %v, want F-001", body["invoiceNum"])
	}
	if body["salesChannelId"] != "sc1" {
		t.Errorf("salesChannelId = %v, want sc1", body["salesChannelId"])
	}
	if _, ok := body["expAccountId"]; ok {
		t.Error("unset optional expAccountId must be omitted")
	}

	items, ok := body["items"].([]map[string]any)
	if !ok || len(items) != 2 {
		t.Fatalf("items = %v, want 2 mapped items", body["items"])
	}
	want := map[string]any{"name": "Widget", "units": 2.0, "subtotal": 10.5, "tax": 21.0, "serviceId": "svc1"}
	if !reflect.DeepEqual(items[0], want) {
		t.Errorf("items[0] = %v, want %v", items[0], want)
	}
	if items[1]["subtotal"] != 0.0 {
		t.Errorf("zero subtotal must still be sent, got %v", items[1])
	}
}

func TestValidateDocumentPayload(t *testing.T) {
	valid := DocumentCreateParams{
		DocType:   "invoice",
		ContactID: "c1",
		Date:      1700000000,
		Items:     []DocumentItem{{Name: "x", Units: 1, Subtotal: 1}},
	}
	if err := validateDocumentPayload(valid); err != nil {
		t.Fatalf("valid payload rejected: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*DocumentCreateParams)
	}{
		{"bad doc type", func(p *DocumentCreateParams) { p.DocType = "notatype" }},
		{"missing contact", func(p *DocumentCreateParams) { p.ContactID = "" }},
		{"missing items", func(p *DocumentCreateParams) { p.Items = nil }},
		{"missing date", func(p *DocumentCreateParams) { p.Date = 0 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := valid
			tc.mutate(&p)
			if err := validateDocumentPayload(p); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
