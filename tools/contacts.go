package tools

import (
	"context"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/server"

	mcpholded "github.com/luisra51/mcp-holded"
	"github.com/luisra51/mcp-holded/internal"
)

type ContactsListParams struct {
	ListParams
	Phone    string   `json:"phone,omitempty" jsonschema:"description=Filter by exact phone number"`
	Mobile   string   `json:"mobile,omitempty" jsonschema:"description=Filter by exact mobile number"`
	CustomID []string `json:"custom_id,omitempty" jsonschema:"description=Filter by one or more custom IDs"`
}

type ContactAddress struct {
	Address    string `json:"address,omitempty" jsonschema:"description=Street address"`
	City       string `json:"city,omitempty" jsonschema:"description=City"`
	PostalCode string `json:"postal_code,omitempty" jsonschema:"description=Postal code"`
	Province   string `json:"province,omitempty" jsonschema:"description=Province or region"`
	Country    string `json:"country,omitempty" jsonschema:"description=Country"`
}

type ContactPerson struct {
	Name  string `json:"name" jsonschema:"required,description=Contact person name"`
	Phone string `json:"phone,omitempty" jsonschema:"description=Contact person phone"`
	Email string `json:"email,omitempty" jsonschema:"description=Contact person email"`
}

type ContactCreateParams struct {
	Name           string          `json:"name" jsonschema:"required,description=Contact name"`
	Email          string          `json:"email,omitempty" jsonschema:"description=Contact email"`
	Phone          string          `json:"phone,omitempty" jsonschema:"description=Contact phone number"`
	Code           string          `json:"code,omitempty" jsonschema:"description=NIF CIF VAT or tax identification code"`
	Type           string          `json:"type,omitempty" jsonschema:"description=Contact type: client|supplier|lead|debtor|creditor"`
	BillAddress    *ContactAddress `json:"bill_address,omitempty" jsonschema:"description=Billing address"`
	Tradename      string          `json:"tradename,omitempty" jsonschema:"description=Trade name"`
	Note           string          `json:"note,omitempty" jsonschema:"description=Contact notes"`
	ContactPersons []ContactPerson `json:"contact_persons,omitempty" jsonschema:"description=Associated contact persons"`
}

type ContactUpdateParams struct {
	ContactID string `json:"contact_id" jsonschema:"required,description=Contact ID"`
	ContactCreateParams
}

type ContactIDParams struct {
	ContactID string `json:"contact_id" jsonschema:"required,description=Contact ID"`
}

type ContactAttachmentParams struct {
	ContactID    string `json:"contact_id" jsonschema:"required,description=Contact ID"`
	AttachmentID string `json:"attachment_id" jsonschema:"required,description=Attachment ID"`
}

func contactsList(ctx context.Context, args ContactsListParams) (any, error) {
	q, meta, err := addListParams(url.Values{}, args.ListParams)
	if err != nil {
		return nil, err
	}
	if args.Phone != "" {
		q.Set("phone", args.Phone)
	}
	if args.Mobile != "" {
		q.Set("mobile", args.Mobile)
	}
	for _, id := range args.CustomID {
		q.Add("customId[]", id)
	}
	return doJSONList(ctx, "holded.contacts.list", "/contacts", q, meta, args.Fields)
}

// contactBody maps the snake_case MCP params onto the camelCase JSON body the
// Holded API expects.
func contactBody(args ContactCreateParams) map[string]any {
	body := compactBody(map[string]any{
		"email":     args.Email,
		"phone":     args.Phone,
		"code":      args.Code,
		"type":      args.Type,
		"tradename": args.Tradename,
		"note":      args.Note,
	})
	body["name"] = args.Name
	if args.BillAddress != nil {
		body["billAddress"] = compactBody(map[string]any{
			"address":    args.BillAddress.Address,
			"city":       args.BillAddress.City,
			"postalCode": args.BillAddress.PostalCode,
			"province":   args.BillAddress.Province,
			"country":    args.BillAddress.Country,
		})
	}
	if len(args.ContactPersons) > 0 {
		persons := make([]map[string]any, len(args.ContactPersons))
		for i, p := range args.ContactPersons {
			person := compactBody(map[string]any{"phone": p.Phone, "email": p.Email})
			person["name"] = p.Name
			persons[i] = person
		}
		body["contactPersons"] = persons
	}
	return body
}

func validateContactPayload(args ContactCreateParams) error {
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return err
	}
	if args.Type != "" {
		return internal.RequireOneOf(args.Type, "type", "client", "supplier", "lead", "debtor", "creditor")
	}
	return nil
}

func contactCreate(ctx context.Context, args ContactCreateParams) (any, error) {
	if err := validateContactPayload(args); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.create", true, http.MethodPost, "/contacts", url.Values{}, contactBody(args), nil)
}

func contactGet(ctx context.Context, args ContactIDParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.get", false, http.MethodGet, "/contacts/"+url.PathEscape(args.ContactID), url.Values{}, nil, nil)
}

func contactUpdate(ctx context.Context, args ContactUpdateParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	if err := validateContactPayload(args.ContactCreateParams); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.update", true, http.MethodPut, "/contacts/"+url.PathEscape(args.ContactID), url.Values{}, contactBody(args.ContactCreateParams), nil)
}

func contactDelete(ctx context.Context, args ContactIDParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.delete", true, http.MethodDelete, "/contacts/"+url.PathEscape(args.ContactID), url.Values{}, nil, nil)
}

func contactAttachmentsList(ctx context.Context, args ContactIDParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.attachments.list", false, http.MethodGet, "/contacts/"+url.PathEscape(args.ContactID)+"/attachments", url.Values{}, nil, nil)
}

func contactAttachmentGet(ctx context.Context, args ContactAttachmentParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.AttachmentID, "attachment_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.attachments.get", false, http.MethodGet, "/contacts/"+url.PathEscape(args.ContactID)+"/attachments/"+url.PathEscape(args.AttachmentID), url.Values{}, nil, nil)
}

var (
	ContactsList           = mcpholded.MustTool("holded.contacts.list", "List contacts with optional filters.", contactsList, readOnlyOptions("List contacts")...)
	ContactCreate          = mcpholded.MustTool("holded.contacts.create", "Create a contact (write).", contactCreate, writeOptions("Create contact")...)
	ContactGet             = mcpholded.MustTool("holded.contacts.get", "Retrieve a contact by ID.", contactGet, readOnlyOptions("Get contact")...)
	ContactUpdate          = mcpholded.MustTool("holded.contacts.update", "Update a contact (write).", contactUpdate, writeOptions("Update contact")...)
	ContactDelete          = mcpholded.MustTool("holded.contacts.delete", "Delete a contact (write).", contactDelete, destructiveOptions("Delete contact")...)
	ContactAttachmentsList = mcpholded.MustTool("holded.contacts.attachments.list", "List contact attachments.", contactAttachmentsList, readOnlyOptions("List contact attachments")...)
	ContactAttachmentGet   = mcpholded.MustTool("holded.contacts.attachments.get", "Retrieve a contact attachment.", contactAttachmentGet, readOnlyOptions("Get contact attachment")...)
)

func AddContactTools(m *server.MCPServer) {
	ContactsList.Register(m)
	ContactCreate.Register(m)
	ContactGet.Register(m)
	ContactUpdate.Register(m)
	ContactDelete.Register(m)
	ContactAttachmentsList.Register(m)
	ContactAttachmentGet.Register(m)
}
