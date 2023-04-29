package products

type Vairation struct {
	VarID      int               `json:"variation_id"`
	Sku        string            `json:"sku"`
	MaxQty     int               `json:"max_qty"`
	Price      float32           `json:"display_price"`
	Attributes map[string]string `json:"attributes"`
}

/* 
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    url text not null unique,
    store bigint NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp
);

CREATE TABLE product_variations (
    id BIGSERIAL PRIMARY KEY,
    product_id bigint not null,
    variation_id bigint not null,
    date_added timestamp NOT NULL,
    date_updated timestamp,
    CONSTRAINT u_constraint UNIQUE (product_id, variation_id)
);

CREATE TABLE variations (
    id serial PRIMARY KEY,
    variation_id BIGSERIAL NOT NULL DEFAULT -1,
    sku text NOT NULL DEFAULT '',
    date_added timestamp NOT NULL,
    date_updated timestamp,
    CONSTRAINT p_constraint UNIQUE (attribute_id, variation_id_raw, sku)
);
*/