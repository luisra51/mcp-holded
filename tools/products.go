package tools

import (
	"context"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/server"

	mcpholded "github.com/luisra51/mcp-holded"
	"github.com/luisra51/mcp-holded/internal"
)

type ProductsListParams struct {
	ListParams
}

type ProductCreateParams struct {
	Name        string  `json:"name" jsonschema:"description=Product name"`
	SKU         string  `json:"sku,omitempty" jsonschema:"description=Product SKU"`
	Barcode     string  `json:"barcode,omitempty" jsonschema:"description=Product barcode"`
	Price       float64 `json:"price,omitempty" jsonschema:"description=Product price"`
	Cost        float64 `json:"cost,omitempty" jsonschema:"description=Product cost"`
	CostPrice   float64 `json:"costPrice,omitempty" jsonschema:"description=Product cost price"`
	Tax         float64 `json:"tax,omitempty" jsonschema:"description=Tax percentage"`
	Description string  `json:"description,omitempty" jsonschema:"description=Product description"`
	Unit        string  `json:"unit,omitempty" jsonschema:"description=Unit name"`
	Stock       float64 `json:"stock,omitempty" jsonschema:"description=Initial stock"`
	Kind        string  `json:"kind,omitempty" jsonschema:"description=Product kind: product|service"`
}

type ProductUpdateParams struct {
	ProductID string `json:"product_id" jsonschema:"description=Product ID"`
	ProductCreateParams
}

type ProductIDParams struct {
	ProductID string `json:"product_id" jsonschema:"description=Product ID"`
}

type ProductImageParams struct {
	ProductID string `json:"product_id" jsonschema:"description=Product ID"`
	ImageID   string `json:"image_id" jsonschema:"description=Image ID"`
}

type ProductStockUpdateParams struct {
	ProductID   string `json:"product_id" jsonschema:"description=Product ID"`
	WarehouseID string `json:"warehouse_id,omitempty" jsonschema:"description=Warehouse ID"`
	Units       int    `json:"units" jsonschema:"description=Units to add or subtract"`
}

func productsList(ctx context.Context, args ProductsListParams) (any, error) {
	q, meta, err := addListParams(url.Values{}, args.ListParams)
	if err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.products.list", false, http.MethodGet, "/products", q, nil, meta)
}

func productCreate(ctx context.Context, args ProductCreateParams) (any, error) {
	if err := internal.RequireID(args.Name, "name"); err != nil {
		return nil, err
	}
	if args.Kind != "" {
		if err := internal.RequireOneOf(args.Kind, "kind", "product", "service"); err != nil {
			return nil, err
		}
	}
	return doJSON(ctx, "holded.products.create", true, http.MethodPost, "/products", url.Values{}, args, nil)
}

func productGet(ctx context.Context, args ProductIDParams) (any, error) {
	if err := internal.RequireID(args.ProductID, "product_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.products.get", false, http.MethodGet, "/products/"+args.ProductID, url.Values{}, nil, nil)
}

func productUpdate(ctx context.Context, args ProductUpdateParams) (any, error) {
	if err := internal.RequireID(args.ProductID, "product_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.products.update", true, http.MethodPut, "/products/"+args.ProductID, url.Values{}, args.ProductCreateParams, nil)
}

func productDelete(ctx context.Context, args ProductIDParams) (any, error) {
	if err := internal.RequireID(args.ProductID, "product_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.products.delete", true, http.MethodDelete, "/products/"+args.ProductID, url.Values{}, nil, nil)
}

func productMainImageGet(ctx context.Context, args ProductIDParams) (any, error) {
	if err := internal.RequireID(args.ProductID, "product_id"); err != nil {
		return nil, err
	}
	return doRawBase64(ctx, "holded.products.image.main.get", "/products/"+args.ProductID+"/image")
}

func productImagesList(ctx context.Context, args ProductIDParams) (any, error) {
	if err := internal.RequireID(args.ProductID, "product_id"); err != nil {
		return nil, err
	}
	return doJSON(ctx, "holded.products.images.list", false, http.MethodGet, "/products/"+args.ProductID+"/images", url.Values{}, nil, nil)
}

func productImageGet(ctx context.Context, args ProductImageParams) (any, error) {
	if err := internal.RequireID(args.ProductID, "product_id"); err != nil {
		return nil, err
	}
	if err := internal.RequireID(args.ImageID, "image_id"); err != nil {
		return nil, err
	}
	return doRawBase64(ctx, "holded.products.images.get", "/products/"+args.ProductID+"/images/"+args.ImageID)
}

func productStockUpdate(ctx context.Context, args ProductStockUpdateParams) (any, error) {
	if err := internal.RequireID(args.ProductID, "product_id"); err != nil {
		return nil, err
	}
	body := compactBody(map[string]any{"warehouseId": args.WarehouseID, "units": args.Units})
	return doJSON(ctx, "holded.products.stock.update", true, http.MethodPut, "/products/"+args.ProductID+"/stock", url.Values{}, body, nil)
}

var (
	ProductsList        = mcpholded.MustTool("holded.products.list", "List products.", productsList, readOnlyOptions("List products")...)
	ProductCreate       = mcpholded.MustTool("holded.products.create", "Create a product (write).", productCreate, writeOptions("Create product")...)
	ProductGet          = mcpholded.MustTool("holded.products.get", "Retrieve a product by ID.", productGet, readOnlyOptions("Get product")...)
	ProductUpdate       = mcpholded.MustTool("holded.products.update", "Update a product (write).", productUpdate, writeOptions("Update product")...)
	ProductDelete       = mcpholded.MustTool("holded.products.delete", "Delete a product (write).", productDelete, destructiveOptions("Delete product")...)
	ProductMainImageGet = mcpholded.MustTool("holded.products.image.main.get", "Retrieve product main image as base64.", productMainImageGet, readOnlyOptions("Get main product image")...)
	ProductImagesList   = mcpholded.MustTool("holded.products.images.list", "List product images.", productImagesList, readOnlyOptions("List product images")...)
	ProductImageGet     = mcpholded.MustTool("holded.products.images.get", "Retrieve product secondary image as base64.", productImageGet, readOnlyOptions("Get product image")...)
	ProductStockUpdate  = mcpholded.MustTool("holded.products.stock.update", "Update product stock (write).", productStockUpdate, writeOptions("Update product stock")...)
)

func AddProductTools(m *server.MCPServer) {
	ProductsList.Register(m)
	ProductCreate.Register(m)
	ProductGet.Register(m)
	ProductUpdate.Register(m)
	ProductDelete.Register(m)
	ProductMainImageGet.Register(m)
	ProductImagesList.Register(m)
	ProductImageGet.Register(m)
	ProductStockUpdate.Register(m)
}
