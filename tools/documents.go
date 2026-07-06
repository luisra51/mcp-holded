package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/server"

	mcpholded "github.com/luisra51/mcp-holded"
	"github.com/luisra51/mcp-holded/internal"
)

const maxAttachmentBytes = 10 << 20

type DocumentsListParams struct {
	ListParams
	DocType   string `json:"doc_type" jsonschema:"required,description=Document type"`
	ContactID string `json:"contact_id,omitempty" jsonschema:"description=Filter by contact ID"`
	Paid      string `json:"paid,omitempty" jsonschema:"description=Payment status: 0 not paid; 1 paid; 2 partially paid"`
	Billed    string `json:"billed,omitempty" jsonschema:"description=Billed status: 0 not billed; 1 billed"`
	Sort      string `json:"sort,omitempty" jsonschema:"description=Sort order: created-asc|created-desc"`
}

type DocumentItem struct {
	Name      string   `json:"name" jsonschema:"required,description=Product or service name"`
	Units     float64  `json:"units" jsonschema:"required,description=Quantity of units"`
	Subtotal  float64  `json:"subtotal" jsonschema:"required,description=Line subtotal before tax and discount"`
	Desc      string   `json:"desc,omitempty" jsonschema:"description=Line description"`
	SKU       string   `json:"sku,omitempty" jsonschema:"description=SKU or product reference code"`
	Tax       float64  `json:"tax,omitempty" jsonschema:"description=Tax percentage"`
	Taxes     []string `json:"taxes,omitempty" jsonschema:"description=Holded tax IDs"`
	Discount  float64  `json:"discount,omitempty" jsonschema:"description=Discount percentage"`
	ServiceID string   `json:"service_id,omitempty" jsonschema:"description=Service catalog ID"`
}

type DocumentCreateParams struct {
	DocType        string         `json:"doc_type" jsonschema:"required,description=Document type"`
	ContactID      string         `json:"contact_id" jsonschema:"required,description=Contact ID"`
	Items          []DocumentItem `json:"items" jsonschema:"required,description=Document line items"`
	Date           int64          `json:"date" jsonschema:"required,description=Document date as Unix timestamp in seconds"`
	Notes          string         `json:"notes,omitempty" jsonschema:"description=Document notes"`
	Currency       string         `json:"currency,omitempty" jsonschema:"description=Currency code"`
	InvoiceNum     string         `json:"invoice_num,omitempty" jsonschema:"description=Document reference number"`
	SalesChannelID string         `json:"sales_channel_id,omitempty" jsonschema:"description=Sales channel ID"`
	ExpAccountID   string         `json:"exp_account_id,omitempty" jsonschema:"description=Expense account ID"`
}

type DocumentUpdateParams struct {
	DocumentID string `json:"document_id" jsonschema:"required,description=Document ID"`
	DocumentCreateParams
}

type DocumentIDParams struct {
	DocType    string `json:"doc_type" jsonschema:"required,description=Document type"`
	DocumentID string `json:"document_id" jsonschema:"required,description=Document ID"`
}

type DocumentPayParams struct {
	DocType    string  `json:"doc_type" jsonschema:"required,description=Document type"`
	DocumentID string  `json:"document_id" jsonschema:"required,description=Document ID"`
	Amount     float64 `json:"amount" jsonschema:"required,description=Payment amount; must be greater than 0"`
	Date       int64   `json:"date,omitempty" jsonschema:"description=Payment date as Unix timestamp in seconds"`
	TreasuryID string  `json:"treasury_id,omitempty" jsonschema:"description=Treasury account ID"`
}

type DocumentSendParams struct {
	DocType    string   `json:"doc_type" jsonschema:"required,description=Document type"`
	DocumentID string   `json:"document_id" jsonschema:"required,description=Document ID"`
	Emails     []string `json:"emails,omitempty" jsonschema:"description=Email recipients"`
	Subject    string   `json:"subject,omitempty" jsonschema:"description=Email subject"`
	Message    string   `json:"message,omitempty" jsonschema:"description=Email message body"`
}

type DocumentShipLinesParams struct {
	DocType    string           `json:"doc_type" jsonschema:"required,description=Document type"`
	DocumentID string           `json:"document_id" jsonschema:"required,description=Document ID"`
	Lines      []map[string]any `json:"lines" jsonschema:"required,description=Line shipment payloads"`
}

type DocumentAttachParams struct {
	DocType    string `json:"doc_type" jsonschema:"required,description=Document type"`
	DocumentID string `json:"document_id" jsonschema:"required,description=Document ID"`
	FileBase64 string `json:"file_base64" jsonschema:"required,description=File content encoded as base64"`
	Filename   string `json:"filename" jsonschema:"required,description=Original filename"`
}

type DocumentTrackingParams struct {
	DocType        string `json:"doc_type" jsonschema:"required,description=Document type"`
	DocumentID     string `json:"document_id" jsonschema:"required,description=Document ID"`
	TrackingNumber string `json:"tracking_number,omitempty" jsonschema:"description=Tracking number"`
	Carrier        string `json:"carrier,omitempty" jsonschema:"description=Carrier name"`
}

type DocumentPipelineParams struct {
	DocType    string `json:"doc_type" jsonschema:"required,description=Document type"`
	DocumentID string `json:"document_id" jsonschema:"required,description=Document ID"`
	PipelineID string `json:"pipeline_id" jsonschema:"required,description=Pipeline ID"`
	StageID    string `json:"stage_id" jsonschema:"required,description=Pipeline stage ID"`
}

func documentsList(ctx context.Context, args DocumentsListParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	q, meta, err := addListParams(url.Values{}, args.ListParams)
	if err != nil {
		return nil, err
	}
	if args.ContactID != "" {
		q.Set("contactid", args.ContactID)
	}
	if args.Paid != "" {
		if err := internal.RequireOneOf(args.Paid, "paid", "0", "1", "2"); err != nil {
			return nil, err
		}
		q.Set("paid", args.Paid)
	}
	if args.Billed != "" {
		if err := internal.RequireOneOf(args.Billed, "billed", "0", "1"); err != nil {
			return nil, err
		}
		q.Set("billed", args.Billed)
	}
	if args.Sort != "" {
		if err := internal.RequireOneOf(args.Sort, "sort", "created-asc", "created-desc"); err != nil {
			return nil, err
		}
		q.Set("sort", args.Sort)
	}
	return doJSONList(ctx, "holded.documents.list", "/documents/"+args.DocType, q, meta, args.Fields)
}

// documentBody maps the snake_case MCP params onto the camelCase JSON body
// the Holded API expects. Required fields are always sent; optional ones only
// when set.
func documentBody(args DocumentCreateParams) map[string]any {
	items := make([]map[string]any, len(args.Items))
	for i, it := range args.Items {
		item := compactBody(map[string]any{
			"desc":      it.Desc,
			"sku":       it.SKU,
			"tax":       it.Tax,
			"taxes":     it.Taxes,
			"discount":  it.Discount,
			"serviceId": it.ServiceID,
		})
		item["name"] = it.Name
		item["units"] = it.Units
		item["subtotal"] = it.Subtotal
		items[i] = item
	}
	body := compactBody(map[string]any{
		"notes":          args.Notes,
		"currency":       args.Currency,
		"invoiceNum":     args.InvoiceNum,
		"salesChannelId": args.SalesChannelID,
		"expAccountId":   args.ExpAccountID,
	})
	body["contactId"] = args.ContactID
	body["items"] = items
	body["date"] = args.Date
	return body
}

func validateDocumentPayload(args DocumentCreateParams) error {
	if err := validateDocumentType(args.DocType); err != nil {
		return err
	}
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return err
	}
	if len(args.Items) == 0 {
		return internal.RequireID("", "items")
	}
	if args.Date <= 0 {
		return fmt.Errorf("date is required as a Unix timestamp in seconds")
	}
	return nil
}

func documentCreate(ctx context.Context, args DocumentCreateParams) (any, error) {
	if err := validateDocumentPayload(args); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.documents.create", true, http.MethodPost, "/documents/"+args.DocType, url.Values{}, documentBody(args), nil)
}

func documentGet(ctx context.Context, args DocumentIDParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.documents.get", false, http.MethodGet, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID), url.Values{}, nil, nil)
}

func documentUpdate(ctx context.Context, args DocumentUpdateParams) (any, error) {
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	if err := validateDocumentPayload(args.DocumentCreateParams); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.documents.update", true, http.MethodPut, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID), url.Values{}, documentBody(args.DocumentCreateParams), nil)
}

func documentDelete(ctx context.Context, args DocumentIDParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.documents.delete", true, http.MethodDelete, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID), url.Values{}, nil, nil)
}

func documentPay(ctx context.Context, args DocumentPayParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	if args.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}
	body := compactBody(map[string]any{"amount": args.Amount, "date": args.Date, "treasuryId": args.TreasuryID})
	return doJSON(ctx, "holded.documents.pay", true, http.MethodPost, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID)+"/pay", url.Values{}, body, nil)
}

func documentSend(ctx context.Context, args DocumentSendParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	body := compactBody(map[string]any{"emails": args.Emails, "subject": args.Subject, "message": args.Message})
	return doJSON(ctx, "holded.documents.send", true, http.MethodPost, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID)+"/send", url.Values{}, body, nil)
}

func documentPDFGet(ctx context.Context, args DocumentIDParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	return doRawBase64(ctx, "holded.documents.pdf.get", "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID)+"/pdf")
}

func documentShipAll(ctx context.Context, args DocumentIDParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.documents.ship.all", true, http.MethodPost, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID)+"/ship", url.Values{}, nil, nil)
}

func documentShipLines(ctx context.Context, args DocumentShipLinesParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	if len(args.Lines) == 0 {
		return nil, internal.RequireID("", "lines")
	}
	return doJSON(ctx, "holded.documents.ship.lines", true, http.MethodPost, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID)+"/ship", url.Values{}, map[string]any{"lines": args.Lines}, nil)
}

func documentShippedGet(ctx context.Context, args DocumentIDParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.documents.shipped.get", false, http.MethodGet, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID)+"/shipped", url.Values{}, nil, nil)
}

func documentAttach(ctx context.Context, args DocumentAttachParams) (any, error) {
	if err := ensureToolAllowed(ctx, "holded.documents.attach", true); err != nil {
		return nil, err
	}
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.Filename, "filename"); err != nil {
		return nil, err
	}
	if base64.StdEncoding.DecodedLen(len(args.FileBase64)) > maxAttachmentBytes {
		return nil, fmt.Errorf("file exceeds the maximum attachment size of %d MB", maxAttachmentBytes>>20)
	}
	file, err := base64.StdEncoding.DecodeString(args.FileBase64)
	if err != nil {
		return nil, err
	}
	client, err := requireClient(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := client.UploadFile(ctx, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID)+"/attach", file, args.Filename)
	if err != nil {
		return nil, err
	}
	return internal.Wrap(internal.MaskSensitive(payload), nil), nil
}

func documentTrackingUpdate(ctx context.Context, args DocumentTrackingParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.DocumentID, "document_id"); err != nil {
		return nil, err
	}
	body := compactBody(map[string]any{"trackingNumber": args.TrackingNumber, "carrier": args.Carrier})
	return doJSON(ctx, "holded.documents.tracking.update", true, http.MethodPost, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID)+"/tracking", url.Values{}, body, nil)
}

func documentPipelineUpdate(ctx context.Context, args DocumentPipelineParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	for field, value := range map[string]string{"document_id": args.DocumentID, "pipeline_id": args.PipelineID, "stage_id": args.StageID} {
		if err := internal.RequireID(value, field); err != nil {
			return nil, err
		}
	}
	body := map[string]any{"pipelineId": args.PipelineID, "stageId": args.StageID}
	return doJSON(ctx, "holded.documents.pipeline.update", true, http.MethodPost, "/documents/"+args.DocType+"/"+url.PathEscape(args.DocumentID)+"/pipeline", url.Values{}, body, nil)
}

func paymentMethodsList(ctx context.Context, _ struct{}) (any, error) {
	return doJSON(ctx, "holded.payment_methods.list", false, http.MethodGet, "/paymentmethods", url.Values{}, nil, nil)
}

var (
	DocumentsList          = mcpholded.MustTool("holded.documents.list", "List documents by document type.", documentsList, readOnlyOptions("List documents")...)
	DocumentCreate         = mcpholded.MustTool("holded.documents.create", "Create a document (write).", documentCreate, writeOptions("Create document")...)
	DocumentGet            = mcpholded.MustTool("holded.documents.get", "Retrieve a document by ID.", documentGet, readOnlyOptions("Get document")...)
	DocumentUpdate         = mcpholded.MustTool("holded.documents.update", "Update a document (write).", documentUpdate, writeOptions("Update document")...)
	DocumentDelete         = mcpholded.MustTool("holded.documents.delete", "Delete a document (write).", documentDelete, destructiveOptions("Delete document")...)
	DocumentPay            = mcpholded.MustTool("holded.documents.pay", "Register a document payment (write).", documentPay, writeOptions("Pay document")...)
	DocumentSend           = mcpholded.MustTool("holded.documents.send", "Send a document by email (write).", documentSend, writeOptions("Send document")...)
	DocumentPDFGet         = mcpholded.MustTool("holded.documents.pdf.get", "Retrieve a document PDF as base64.", documentPDFGet, readOnlyOptions("Get document PDF")...)
	DocumentShipAll        = mcpholded.MustTool("holded.documents.ship.all", "Ship all document items (write).", documentShipAll, writeOptions("Ship all items")...)
	DocumentShipLines      = mcpholded.MustTool("holded.documents.ship.lines", "Ship selected document lines (write).", documentShipLines, writeOptions("Ship selected lines")...)
	DocumentShippedGet     = mcpholded.MustTool("holded.documents.shipped.get", "Retrieve shipped units for a document.", documentShippedGet, readOnlyOptions("Get shipped units")...)
	DocumentAttach         = mcpholded.MustTool("holded.documents.attach", "Attach a file to a document (write).", documentAttach, writeOptions("Attach document file")...)
	DocumentTrackingUpdate = mcpholded.MustTool("holded.documents.tracking.update", "Update document tracking information (write).", documentTrackingUpdate, writeOptions("Update tracking")...)
	DocumentPipelineUpdate = mcpholded.MustTool("holded.documents.pipeline.update", "Update document pipeline stage (write).", documentPipelineUpdate, writeOptions("Update pipeline")...)
	PaymentMethodsList     = mcpholded.MustTool("holded.payment_methods.list", "List available payment methods.", paymentMethodsList, readOnlyOptions("List payment methods")...)
)

func AddDocumentTools(m *server.MCPServer) {
	DocumentsList.Register(m)
	DocumentCreate.Register(m)
	DocumentGet.Register(m)
	DocumentUpdate.Register(m)
	DocumentDelete.Register(m)
	DocumentPay.Register(m)
	DocumentSend.Register(m)
	DocumentPDFGet.Register(m)
	DocumentShipAll.Register(m)
	DocumentShipLines.Register(m)
	DocumentShippedGet.Register(m)
	DocumentAttach.Register(m)
	DocumentTrackingUpdate.Register(m)
	DocumentPipelineUpdate.Register(m)
	PaymentMethodsList.Register(m)
}
