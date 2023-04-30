CREATE TABLE variations (
    id serial PRIMARY KEY,
    variation_id BIGSERIAL NOT NULL DEFAULT -1,
    sku text NOT NULL DEFAULT '',
    date_added timestamp NOT NULL,
    date_updated timestamp,
    CONSTRAINT p_constraint UNIQUE (attribute_id, variation_id_raw, sku)
);