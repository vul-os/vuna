package product

import "time"

type Product struct {
	ID     string `gorm:"id"`
	Url    string `gorm:"name"`
	SiteId string `gorm:"site_id"`

	DateAdded   time.Time `gorm:"date_added"`
	DateUpdated time.Time `gorm:"date_updated"`
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
