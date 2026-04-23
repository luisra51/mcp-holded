package tools

import (
	"context"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/server"

	mcpholded "github.com/luisra51/mcp-holded"
	"github.com/luisra51/mcp-holded/internal"
)

type NumberingSeriesListParams struct {
	ListParams
	DocType string `json:"doc_type" jsonschema:"description=Document type"`
}

type NumberingSeriesCreateParams struct {
	DocType    string `json:"doc_type" jsonschema:"description=Document type"`
	Name       string `json:"name" jsonschema:"description=Series name"`
	Prefix     string `json:"prefix,omitempty" jsonschema:"description=Series prefix"`
	NextNumber int    `json:"nextNumber,omitempty" jsonschema:"description=Next number in the series"`
}

type NumberingSeriesUpdateParams struct {
	DocType    string `json:"doc_type" jsonschema:"description=Document type"`
	SerieID    string `json:"serie_id" jsonschema:"description=Numbering series ID"`
	Name       string `json:"name,omitempty" jsonschema:"description=Series name"`
	Prefix     string `json:"prefix,omitempty" jsonschema:"description=Series prefix"`
	NextNumber int    `json:"nextNumber,omitempty" jsonschema:"description=Next number in the series"`
}

type NumberingSeriesIDParams struct {
	DocType string `json:"doc_type" jsonschema:"description=Document type"`
	SerieID string `json:"serie_id" jsonschema:"description=Numbering series ID"`
}

func numberingSeriesList(ctx context.Context, args NumberingSeriesListParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	q, meta, err := addListParams(url.Values{}, args.ListParams)
	if err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.numbering_series.list", false, http.MethodGet, "/numberseries/"+args.DocType, q, nil, meta)
}

func numberingSeriesCreate(ctx context.Context, args NumberingSeriesCreateParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	body := compactBody(map[string]any{"name": args.Name, "prefix": args.Prefix, "nextNumber": args.NextNumber})
	return doJSON(ctx, "holded.numbering_series.create", true, http.MethodPost, "/numberseries/"+args.DocType, url.Values{}, body, nil)
}

func numberingSeriesUpdate(ctx context.Context, args NumberingSeriesUpdateParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.SerieID, "serie_id"); err != nil {
		return nil, err
	}
	body := compactBody(map[string]any{"name": args.Name, "prefix": args.Prefix, "nextNumber": args.NextNumber})
	return doJSON(ctx, "holded.numbering_series.update", true, http.MethodPut, "/numberseries/"+args.DocType+"/"+args.SerieID, url.Values{}, body, nil)
}

func numberingSeriesDelete(ctx context.Context, args NumberingSeriesIDParams) (any, error) {
	if err := validateDocumentType(args.DocType); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.SerieID, "serie_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.numbering_series.delete", true, http.MethodDelete, "/numberseries/"+args.DocType+"/"+args.SerieID, url.Values{}, nil, nil)
}

func AddNumberingSeriesTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.numbering_series.list", "List numbering series by document type.", numberingSeriesList, readOnlyOptions("List numbering series")...).Register(m)
	mcpholded.MustTool("holded.numbering_series.create", "Create a numbering series (write).", numberingSeriesCreate, writeOptions("Create numbering series")...).Register(m)
	mcpholded.MustTool("holded.numbering_series.update", "Update a numbering series (write).", numberingSeriesUpdate, writeOptions("Update numbering series")...).Register(m)
	mcpholded.MustTool("holded.numbering_series.delete", "Delete a numbering series (write).", numberingSeriesDelete, destructiveOptions("Delete numbering series")...).Register(m)
}

type WarehouseCreateParams struct {
	Name       string `json:"name" jsonschema:"description=Warehouse name"`
	Address    string `json:"address,omitempty" jsonschema:"description=Warehouse address"`
	City       string `json:"city,omitempty" jsonschema:"description=City"`
	PostalCode string `json:"postalCode,omitempty" jsonschema:"description=Postal code"`
	Province   string `json:"province,omitempty" jsonschema:"description=Province"`
	Country    string `json:"country,omitempty" jsonschema:"description=Country"`
}

type WarehouseUpdateParams struct {
	WarehouseID string `json:"warehouse_id" jsonschema:"description=Warehouse ID"`
	WarehouseCreateParams
}

type WarehouseIDParams struct {
	WarehouseID string `json:"warehouse_id" jsonschema:"description=Warehouse ID"`
}

type WarehouseStockListParams struct {
	ListParams
	WarehouseID string `json:"warehouse_id" jsonschema:"description=Warehouse ID"`
}

func warehousesList(ctx context.Context, args ListParams) (any, error) {
	return listSimple(ctx, "holded.warehouses.list", "/warehouses", args)
}

func warehouseCreate(ctx context.Context, args WarehouseCreateParams) (any, error) {
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.warehouses.create", true, http.MethodPost, "/warehouses", url.Values{}, args, nil)
}

func warehouseStockList(ctx context.Context, args WarehouseStockListParams) (any, error) {
	if err := internal.RequireID(args.WarehouseID, "warehouse_id"); err != nil {
		return nil, err
	}
	q, meta, err := addListParams(url.Values{}, args.ListParams)
	if err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.warehouses.stock.list", false, http.MethodGet, "/warehouses/"+args.WarehouseID+"/stock", q, nil, meta)
}

func warehouseGet(ctx context.Context, args WarehouseIDParams) (any, error) {
	if err := internal.RequireID(args.WarehouseID, "warehouse_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.warehouses.get", false, http.MethodGet, "/warehouses/"+args.WarehouseID, url.Values{}, nil, nil)
}

func warehouseUpdate(ctx context.Context, args WarehouseUpdateParams) (any, error) {
	if err := internal.RequireID(args.WarehouseID, "warehouse_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.warehouses.update", true, http.MethodPut, "/warehouses/"+args.WarehouseID, url.Values{}, args.WarehouseCreateParams, nil)
}

func warehouseDelete(ctx context.Context, args WarehouseIDParams) (any, error) {
	if err := internal.RequireID(args.WarehouseID, "warehouse_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.warehouses.delete", true, http.MethodDelete, "/warehouses/"+args.WarehouseID, url.Values{}, nil, nil)
}

func AddWarehouseTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.warehouses.list", "List warehouses.", warehousesList, readOnlyOptions("List warehouses")...).Register(m)
	mcpholded.MustTool("holded.warehouses.create", "Create a warehouse (write).", warehouseCreate, writeOptions("Create warehouse")...).Register(m)
	mcpholded.MustTool("holded.warehouses.stock.list", "List products stock in a warehouse.", warehouseStockList, readOnlyOptions("List warehouse stock")...).Register(m)
	mcpholded.MustTool("holded.warehouses.get", "Retrieve a warehouse.", warehouseGet, readOnlyOptions("Get warehouse")...).Register(m)
	mcpholded.MustTool("holded.warehouses.update", "Update a warehouse (write).", warehouseUpdate, writeOptions("Update warehouse")...).Register(m)
	mcpholded.MustTool("holded.warehouses.delete", "Delete a warehouse (write).", warehouseDelete, destructiveOptions("Delete warehouse")...).Register(m)
}

type PaymentCreateParams struct {
	Name string `json:"name" jsonschema:"description=Payment method name"`
	Days int    `json:"days,omitempty" jsonschema:"description=Days until due"`
}

type PaymentUpdateParams struct {
	PaymentID string `json:"payment_id" jsonschema:"description=Payment ID"`
	PaymentCreateParams
}

type PaymentIDParams struct {
	PaymentID string `json:"payment_id" jsonschema:"description=Payment ID"`
}

func paymentsList(ctx context.Context, args ListParams) (any, error) {
	return listSimple(ctx, "holded.payments.list", "/payments", args)
}

func paymentCreate(ctx context.Context, args PaymentCreateParams) (any, error) {
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.payments.create", true, http.MethodPost, "/payments", url.Values{}, args, nil)
}

func paymentGet(ctx context.Context, args PaymentIDParams) (any, error) {
	if err := internal.RequireID(args.PaymentID, "payment_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.payments.get", false, http.MethodGet, "/payments/"+args.PaymentID, url.Values{}, nil, nil)
}

func paymentUpdate(ctx context.Context, args PaymentUpdateParams) (any, error) {
	if err := internal.RequireID(args.PaymentID, "payment_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.payments.update", true, http.MethodPut, "/payments/"+args.PaymentID, url.Values{}, args.PaymentCreateParams, nil)
}

func paymentDelete(ctx context.Context, args PaymentIDParams) (any, error) {
	if err := internal.RequireID(args.PaymentID, "payment_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.payments.delete", true, http.MethodDelete, "/payments/"+args.PaymentID, url.Values{}, nil, nil)
}

func AddPaymentTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.payments.list", "List payments.", paymentsList, readOnlyOptions("List payments")...).Register(m)
	mcpholded.MustTool("holded.payments.create", "Create a payment method (write).", paymentCreate, writeOptions("Create payment")...).Register(m)
	mcpholded.MustTool("holded.payments.get", "Retrieve a payment method.", paymentGet, readOnlyOptions("Get payment")...).Register(m)
	mcpholded.MustTool("holded.payments.update", "Update a payment method (write).", paymentUpdate, writeOptions("Update payment")...).Register(m)
	mcpholded.MustTool("holded.payments.delete", "Delete a payment method (write).", paymentDelete, destructiveOptions("Delete payment")...).Register(m)
}

type ServiceCreateParams struct {
	Name        string  `json:"name" jsonschema:"description=Service name"`
	SKU         string  `json:"sku,omitempty" jsonschema:"description=Service SKU"`
	Price       float64 `json:"price,omitempty" jsonschema:"description=Service price"`
	Tax         float64 `json:"tax,omitempty" jsonschema:"description=Tax percentage"`
	Description string  `json:"description,omitempty" jsonschema:"description=Service description"`
}

type ServiceUpdateParams struct {
	ServiceID string `json:"service_id" jsonschema:"description=Service ID"`
	ServiceCreateParams
}

type ServiceIDParams struct {
	ServiceID string `json:"service_id" jsonschema:"description=Service ID"`
}

func servicesList(ctx context.Context, args ListParams) (any, error) {
	return listSimple(ctx, "holded.services.list", "/services", args)
}

func serviceCreate(ctx context.Context, args ServiceCreateParams) (any, error) {
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.services.create", true, http.MethodPost, "/services", url.Values{}, args, nil)
}

func serviceGet(ctx context.Context, args ServiceIDParams) (any, error) {
	if err := internal.RequireID(args.ServiceID, "service_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.services.get", false, http.MethodGet, "/services/"+args.ServiceID, url.Values{}, nil, nil)
}

func serviceUpdate(ctx context.Context, args ServiceUpdateParams) (any, error) {
	if err := internal.RequireID(args.ServiceID, "service_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.services.update", true, http.MethodPut, "/services/"+args.ServiceID, url.Values{}, args.ServiceCreateParams, nil)
}

func serviceDelete(ctx context.Context, args ServiceIDParams) (any, error) {
	if err := internal.RequireID(args.ServiceID, "service_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.services.delete", true, http.MethodDelete, "/services/"+args.ServiceID, url.Values{}, nil, nil)
}

func AddServiceTools(m *server.MCPServer) {
	mcpholded.MustTool("holded.services.list", "List services.", servicesList, readOnlyOptions("List services")...).Register(m)
	mcpholded.MustTool("holded.services.create", "Create a service (write).", serviceCreate, writeOptions("Create service")...).Register(m)
	mcpholded.MustTool("holded.services.get", "Retrieve a service.", serviceGet, readOnlyOptions("Get service")...).Register(m)
	mcpholded.MustTool("holded.services.update", "Update a service (write).", serviceUpdate, writeOptions("Update service")...).Register(m)
	mcpholded.MustTool("holded.services.delete", "Delete a service (write).", serviceDelete, destructiveOptions("Delete service")...).Register(m)
}
