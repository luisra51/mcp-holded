package tools

import (
	"context"
	"net/http"
	"net/url"
	"strings"

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
	PostalCode string `json:"postalCode,omitempty" jsonschema:"description=Postal code"`
	Province   string `json:"province,omitempty" jsonschema:"description=Province or region"`
	Country    string `json:"country,omitempty" jsonschema:"description=Country"`
}

type ContactPerson struct {
	Name  string `json:"name" jsonschema:"description=Contact person name"`
	Phone string `json:"phone,omitempty" jsonschema:"description=Contact person phone"`
	Email string `json:"email,omitempty" jsonschema:"description=Contact person email"`
}

type ContactCreateParams struct {
	Name           string          `json:"name" jsonschema:"description=Contact name"`
	Email          string          `json:"email,omitempty" jsonschema:"description=Contact email"`
	Phone          string          `json:"phone,omitempty" jsonschema:"description=Contact phone number"`
	Code           string          `json:"code,omitempty" jsonschema:"description=NIF CIF VAT or tax identification code"`
	Type           string          `json:"type,omitempty" jsonschema:"description=Contact type: client|supplier|lead|debtor|creditor"`
	BillAddress    *ContactAddress `json:"billAddress,omitempty" jsonschema:"description=Billing address"`
	Tradename      string          `json:"tradename,omitempty" jsonschema:"description=Trade name"`
	Note           string          `json:"note,omitempty" jsonschema:"description=Contact notes"`
	ContactPersons []ContactPerson `json:"contactPersons,omitempty" jsonschema:"description=Associated contact persons"`
}

type ContactUpdateParams struct {
	ContactID string `json:"contact_id" jsonschema:"description=Contact ID"`
	ContactCreateParams
}

type ContactIDParams struct {
	ContactID string `json:"contact_id" jsonschema:"description=Contact ID"`
}

type ContactAttachmentParams struct {
	ContactID    string `json:"contact_id" jsonschema:"description=Contact ID"`
	AttachmentID string `json:"attachment_id" jsonschema:"description=Attachment ID"`
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
	if len(args.CustomID) > 0 {
		q.Set("customId[]", strings.Join(args.CustomID, ","))
	}
	return doJSON(ctx, "holded.contacts.list", false, http.MethodGet, "/contacts", q, nil, meta)
}

func contactCreate(ctx context.Context, args ContactCreateParams) (any, error) {
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	if args.Type != "" {
		if err := internal.RequireOneOf(args.Type, "type", "client", "supplier", "lead", "debtor", "creditor"); err != nil {
			return nil, err
		}
	}
	return doJSON(ctx, "holded.contacts.create", true, http.MethodPost, "/contacts", url.Values{}, args, nil)
}

func contactGet(ctx context.Context, args ContactIDParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.get", false, http.MethodGet, "/contacts/"+args.ContactID, url.Values{}, nil, nil)
}

func contactUpdate(ctx context.Context, args ContactUpdateParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.update", true, http.MethodPut, "/contacts/"+args.ContactID, url.Values{}, args.ContactCreateParams, nil)
}

func contactDelete(ctx context.Context, args ContactIDParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.delete", true, http.MethodDelete, "/contacts/"+args.ContactID, url.Values{}, nil, nil)
}

func contactAttachmentsList(ctx context.Context, args ContactIDParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.attachments.list", false, http.MethodGet, "/contacts/"+args.ContactID+"/attachments", url.Values{}, nil, nil)
}

func contactAttachmentGet(ctx context.Context, args ContactAttachmentParams) (any, error) {
	if err := internal.RequireID(args.ContactID, "contact_id"); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.AttachmentID, "attachment_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contacts.attachments.get", false, http.MethodGet, "/contacts/"+args.ContactID+"/attachments/"+args.AttachmentID, url.Values{}, nil, nil)
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
