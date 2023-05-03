package shopify


type ProductJSON struct {
	Product Product `json:"product"`
}

type Product struct {
	ID             int          `json:"id"`
	Title          string       `json:"title"`
	BodyHTML       string       `json:"body_html"`
	Vendor         string       `json:"vendor"`
	ProductType    string       `json:"product_type"`
	CreatedAt      string       `json:"created_at"`
	Handle         string       `json:"handle"`
	UpdatedAt      string       `json:"updated_at"`
	PublishedAt    string       `json:"published_at"`
	TemplateSuffix interface{}  `json:"template_suffix"`
	PublishedScope string       `json:"published_scope"`
	Tags           string       `json:"tags"`
	Variants       []Variant    `json:"variants"`
	Options        []Option     `json:"options"`
	Images         []Image      `json:"images"`
	Image          Image        `json:"image"`
}

type Variant struct {
	ID                int              `json:"id"`
	ProductID         int              `json:"product_id"`
	Title             string           `json:"title"`
	Price             string           `json:"price"`
	SKU               string           `json:"sku"`
	Position          int              `json:"position"`
	InventoryPolicy   string           `json:"inventory_policy"`
	CompareAtPrice    string           `json:"compare_at_price"`
	FulfillmentService string           `json:"fulfillment_service"`
	InventoryManagement string         `json:"inventory_management"`
	Option1           string           `json:"option1"`
	Option2           interface{}      `json:"option2"`
	Option3           interface{}      `json:"option3"`
	CreatedAt         string           `json:"created_at"`
	UpdatedAt         string           `json:"updated_at"`
	Taxable           bool             `json:"taxable"`
	Barcode           interface{}      `json:"barcode"`
	Grams             int              `json:"grams"`
	ImageID           interface{}      `json:"image_id"`
	Weight            float64          `json:"weight"`
	WeightUnit        string           `json:"weight_unit"`
	InventoryQuantity int              `json:"inventory_quantity"`
	OldInventoryQuantity int          `json:"old_inventory_quantity"`
	RequiresShipping  bool             `json:"requires_shipping"`
	QuantityRule      VariantQuantityRule `json:"quantity_rule"`
}

type VariantQuantityRule struct {
	Min       int `json:"min"`
	Max       int `json:"max"`
	Increment int `json:"increment"`
}

type Option struct {
	ID        int      `json:"id"`
	ProductID int      `json:"product_id"`
	Name      string   `json:"name"`
	Position  int      `json:"position"`
	Values    []string `json:"values"`
}

type Image struct {
	ID          int           `json:"id"`
	ProductID   int           `json:"product_id"`
	Position    int           `json:"position"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
	Alt         interface{}   `json:"alt"`
	Width       int           `json:"width"`
	Height      int           `json:"height"`
	Src         string        `json:"src"`
	VariantIDs  []interface{} `json:"variant_ids"`
}