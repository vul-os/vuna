package products

type Site struct {
    Id      string               `bigquery:"id"`
	Url         string            `bigquery:"url"`
    Name        string          `bigquery:"name"`

	DateAdded   bigquery.NullDate `bigquery:"date_added"`
	DateUpdated bigquery.NullDate `bigquery:"date_updated"`
}

type Product struct {
    Id      string               `bigquery:"id"`
	Url         string            `bigquery:"name"`
	SiteId     string            `bigquery:"site_id"`

	DateAdded   bigquery.NullDate `bigquery:"date_added"`
	DateUpdated bigquery.NullDate `bigquery:"date_updated"`
}

type Variations struct {
    Id      string               `bigquery:"id"`
	Sku     string               `bigquery:"sku"`

    DateAdded   bigquery.NullDate `bigquery:"date_added"`
	DateUpdated bigquery.NullDate `bigquery:"date_updated"`
}

type DataPoint struct {
    Id      string               `bigquery:"id"`
	VarID      string               `bigquery:"var_id"`

	MaxQty     int               `bigquery:"max_qty"`
	Price      float32           `bigquery:"price"`

    DateAdded   bigquery.NullDate `bigquery:"date_added"`
	DateUpdated bigquery.NullDate `bigquery:"date_updated"`
}

// CREATE TABLE product_variations (
//     id BIGSERIAL PRIMARY KEY,
//     product_id bigint not null,
//     variation_id bigint not null,
//     CONSTRAINT u_constraint UNIQUE (product_id, variation_id)
// );

/* 
CREATE TABLE variations (
    id serial PRIMARY KEY,
    variation_id BIGSERIAL NOT NULL DEFAULT -1,
    sku text NOT NULL DEFAULT '',
    date_added timestamp NOT NULL,
    date_updated timestamp,
    CONSTRAINT p_constraint UNIQUE (attribute_id, variation_id_raw, sku)
);

CREATE TABLE product_variations (
    id BIGSERIAL PRIMARY KEY,
    product_id bigint not null,
    variation_id bigint not null,
    date_added timestamp NOT NULL,
    date_updated timestamp,
    CONSTRAINT u_constraint UNIQUE (product_id, variation_id)
);
*/