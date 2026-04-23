package tools

import "github.com/mark3labs/mcp-go/server"

// AddAllTools registers every tool group on the MCP server.
func AddAllTools(m *server.MCPServer) {
	AddContactTools(m)
	AddDocumentTools(m)
	AddProductTools(m)
	AddTreasuryTools(m)
	AddExpenseAccountTools(m)
	AddNumberingSeriesTools(m)
	AddSalesChannelTools(m)
	AddWarehouseTools(m)
	AddPaymentTools(m)
	AddTaxTools(m)
	AddContactGroupTools(m)
	AddRemittanceTools(m)
	AddServiceTools(m)
}
