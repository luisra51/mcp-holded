package tools

import (
	"context"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/server"

	mcpholded "github.com/luisra51/mcp-holded"
	"github.com/luisra51/mcp-holded/internal"
)

type NameCodeParams struct {
	Name string `json:"name" jsonschema:"description=Resource name"`
	Code string `json:"code,omitempty" jsonschema:"description=Resource code"`
}

type NameOnlyParams struct {
	Name string `json:"name" jsonschema:"description=Resource name"`
}

type TreasuryCreateParams struct {
	Name    string  `json:"name" jsonschema:"description=Treasury account name"`
	IBAN    string  `json:"iban,omitempty" jsonschema:"description=IBAN"`
	BIC     string  `json:"bic,omitempty" jsonschema:"description=BIC or SWIFT code"`
	Balance float64 `json:"balance,omitempty" jsonschema:"description=Initial balance"`
}

type TreasuryIDParams struct {
	TreasuryID string `json:"treasury_id" jsonschema:"description=Treasury account ID"`
}

type ExpenseAccountIDParams struct {
	AccountID string `json:"account_id" jsonschema:"description=Expense account ID"`
}

type ExpenseAccountUpdateParams struct {
	AccountID string `json:"account_id" jsonschema:"description=Expense account ID"`
	Name      string `json:"name,omitempty" jsonschema:"description=Expense account name"`
	Code      string `json:"code,omitempty" jsonschema:"description=Expense account code"`
}

type SalesChannelIDParams struct {
	ChannelID string `json:"channel_id" jsonschema:"description=Sales channel ID"`
}

type SalesChannelUpdateParams struct {
	ChannelID string `json:"channel_id" jsonschema:"description=Sales channel ID"`
	Name      string `json:"name" jsonschema:"description=Sales channel name"`
}

type ContactGroupIDParams struct {
	GroupID string `json:"group_id" jsonschema:"description=Contact group ID"`
}

type ContactGroupUpdateParams struct {
	GroupID string `json:"group_id" jsonschema:"description=Contact group ID"`
	Name    string `json:"name" jsonschema:"description=Contact group name"`
}

type RemittanceIDParams struct {
	RemittanceID string `json:"remittance_id" jsonschema:"description=Remittance ID"`
}

func listSimple(ctx context.Context, toolName, path string, args ListParams) (any, error) {
	q, meta, err := addListParams(url.Values{}, args)
	if err != nil {
		return nil, err
	}
	return doJSON(ctx, toolName, false, http.MethodGet, path, q, nil, meta)
}

func createNameCode(ctx context.Context, toolName, path string, args NameCodeParams) (any, error) {
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	return doJSON(ctx, toolName, true, http.MethodPost, path, url.Values{}, compactBody(map[string]any{"name": args.Name, "code": args.Code}), nil)
}

func createNameOnly(ctx context.Context, toolName, path string, args NameOnlyParams) (any, error) {
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	return doJSON(ctx, toolName, true, http.MethodPost, path, url.Values{}, args, nil)
}

func treasuriesList(ctx context.Context, args ListParams) (any, error) {
	return listSimple(ctx, "holded.treasuries.list", "/treasury", args)
}

func treasuryCreate(ctx context.Context, args TreasuryCreateParams) (any, error) {
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.treasuries.create", true, http.MethodPost, "/treasury", url.Values{}, args, nil)
}

func treasuryGet(ctx context.Context, args TreasuryIDParams) (any, error) {
	if err := internal.RequireID(args.TreasuryID, "treasury_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.treasuries.get", false, http.MethodGet, "/treasury/"+args.TreasuryID, url.Values{}, nil, nil)
}

func AddTreasuryTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.treasuries.list", "List treasury accounts.", treasuriesList, readOnlyOptions("List treasuries")...).Register(m)
	mcpholded.MustTool("holded.treasuries.create", "Create a treasury account (write).", treasuryCreate, writeOptions("Create treasury")...).Register(m)
	mcpholded.MustTool("holded.treasuries.get", "Retrieve a treasury account.", treasuryGet, readOnlyOptions("Get treasury")...).Register(m)
}

func expenseAccountsList(ctx context.Context, args ListParams) (any, error) {
	return listSimple(ctx, "holded.expense_accounts.list", "/expensesaccounts", args)
}

func expenseAccountCreate(ctx context.Context, args NameCodeParams) (any, error) {
	return createNameCode(ctx, "holded.expense_accounts.create", "/expensesaccounts", args)
}

func expenseAccountGet(ctx context.Context, args ExpenseAccountIDParams) (any, error) {
	if err := internal.RequireID(args.AccountID, "account_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.expense_accounts.get", false, http.MethodGet, "/expensesaccounts/"+args.AccountID, url.Values{}, nil, nil)
}

func expenseAccountUpdate(ctx context.Context, args ExpenseAccountUpdateParams) (any, error) {
	if err := internal.RequireID(args.AccountID, "account_id"); err != nil {
		return nil, err
	}
	body := compactBody(map[string]any{"name": args.Name, "code": args.Code})
	return doJSON(ctx, "holded.expense_accounts.update", true, http.MethodPut, "/expensesaccounts/"+args.AccountID, url.Values{}, body, nil)
}

func expenseAccountDelete(ctx context.Context, args ExpenseAccountIDParams) (any, error) {
	if err := internal.RequireID(args.AccountID, "account_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.expense_accounts.delete", true, http.MethodDelete, "/expensesaccounts/"+args.AccountID, url.Values{}, nil, nil)
}

func AddExpenseAccountTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.expense_accounts.list", "List expense accounts.", expenseAccountsList, readOnlyOptions("List expense accounts")...).Register(m)
	mcpholded.MustTool("holded.expense_accounts.create", "Create an expense account (write).", expenseAccountCreate, writeOptions("Create expense account")...).Register(m)
	mcpholded.MustTool("holded.expense_accounts.get", "Retrieve an expense account.", expenseAccountGet, readOnlyOptions("Get expense account")...).Register(m)
	mcpholded.MustTool("holded.expense_accounts.update", "Update an expense account (write).", expenseAccountUpdate, writeOptions("Update expense account")...).Register(m)
	mcpholded.MustTool("holded.expense_accounts.delete", "Delete an expense account (write).", expenseAccountDelete, destructiveOptions("Delete expense account")...).Register(m)
}

func salesChannelsList(ctx context.Context, args ListParams) (any, error) {
	return listSimple(ctx, "holded.sales_channels.list", "/saleschannels", args)
}

func salesChannelCreate(ctx context.Context, args NameOnlyParams) (any, error) {
	return createNameOnly(ctx, "holded.sales_channels.create", "/saleschannels", args)
}

func salesChannelGet(ctx context.Context, args SalesChannelIDParams) (any, error) {
	if err := internal.RequireID(args.ChannelID, "channel_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.sales_channels.get", false, http.MethodGet, "/saleschannels/"+args.ChannelID, url.Values{}, nil, nil)
}

func salesChannelUpdate(ctx context.Context, args SalesChannelUpdateParams) (any, error) {
	if err := internal.RequireID(args.ChannelID, "channel_id"); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.sales_channels.update", true, http.MethodPut, "/saleschannels/"+args.ChannelID, url.Values{}, map[string]any{"name": args.Name}, nil)
}

func salesChannelDelete(ctx context.Context, args SalesChannelIDParams) (any, error) {
	if err := internal.RequireID(args.ChannelID, "channel_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.sales_channels.delete", true, http.MethodDelete, "/saleschannels/"+args.ChannelID, url.Values{}, nil, nil)
}

func AddSalesChannelTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.sales_channels.list", "List sales channels.", salesChannelsList, readOnlyOptions("List sales channels")...).Register(m)
	mcpholded.MustTool("holded.sales_channels.create", "Create a sales channel (write).", salesChannelCreate, writeOptions("Create sales channel")...).Register(m)
	mcpholded.MustTool("holded.sales_channels.get", "Retrieve a sales channel.", salesChannelGet, readOnlyOptions("Get sales channel")...).Register(m)
	mcpholded.MustTool("holded.sales_channels.update", "Update a sales channel (write).", salesChannelUpdate, writeOptions("Update sales channel")...).Register(m)
	mcpholded.MustTool("holded.sales_channels.delete", "Delete a sales channel (write).", salesChannelDelete, destructiveOptions("Delete sales channel")...).Register(m)
}

func taxesList(ctx context.Context, args ListParams) (any, error) {
	return listSimple(ctx, "holded.taxes.list", "/taxes", args)
}

func AddTaxTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.taxes.list", "List available taxes.", taxesList, readOnlyOptions("List taxes")...).Register(m)
}

func contactGroupsList(ctx context.Context, args ListParams) (any, error) {
	return listSimple(ctx, "holded.contact_groups.list", "/contactgroups", args)
}

func contactGroupCreate(ctx context.Context, args NameOnlyParams) (any, error) {
	return createNameOnly(ctx, "holded.contact_groups.create", "/contactgroups", args)
}

func contactGroupGet(ctx context.Context, args ContactGroupIDParams) (any, error) {
	if err := internal.RequireID(args.GroupID, "group_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contact_groups.get", false, http.MethodGet, "/contactgroups/"+args.GroupID, url.Values{}, nil, nil)
}

func contactGroupUpdate(ctx context.Context, args ContactGroupUpdateParams) (any, error) {
	if err := internal.RequireID(args.GroupID, "group_id"); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contact_groups.update", true, http.MethodPut, "/contactgroups/"+args.GroupID, url.Values{}, map[string]any{"name": args.Name}, nil)
}

func contactGroupDelete(ctx context.Context, args ContactGroupIDParams) (any, error) {
	if err := internal.RequireID(args.GroupID, "group_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.contact_groups.delete", true, http.MethodDelete, "/contactgroups/"+args.GroupID, url.Values{}, nil, nil)
}

func AddContactGroupTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.contact_groups.list", "List contact groups.", contactGroupsList, readOnlyOptions("List contact groups")...).Register(m)
	mcpholded.MustTool("holded.contact_groups.create", "Create a contact group (write).", contactGroupCreate, writeOptions("Create contact group")...).Register(m)
	mcpholded.MustTool("holded.contact_groups.get", "Retrieve a contact group.", contactGroupGet, readOnlyOptions("Get contact group")...).Register(m)
	mcpholded.MustTool("holded.contact_groups.update", "Update a contact group (write).", contactGroupUpdate, writeOptions("Update contact group")...).Register(m)
	mcpholded.MustTool("holded.contact_groups.delete", "Delete a contact group (write).", contactGroupDelete, destructiveOptions("Delete contact group")...).Register(m)
}

func remittancesList(ctx context.Context, args ListParams) (any, error) {
	return listSimple(ctx, "holded.remittances.list", "/remittances", args)
}

func remittanceGet(ctx context.Context, args RemittanceIDParams) (any, error) {
	if err := internal.RequireID(args.RemittanceID, "remittance_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.remittances.get", false, http.MethodGet, "/remittances/"+args.RemittanceID, url.Values{}, nil, nil)
}

func AddRemittanceTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.remittances.list", "List remittances.", remittancesList, readOnlyOptions("List remittances")...).Register(m)
	mcpholded.MustTool("holded.remittances.get", "Retrieve a remittance.", remittanceGet, readOnlyOptions("Get remittance")...).Register(m)
}
