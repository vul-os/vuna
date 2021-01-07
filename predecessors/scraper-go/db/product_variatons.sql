CREATE TABLE product_variations (
    id BIGSERIAL PRIMARY KEY,
    product_id bigint not null,
    variation_id bigint not null,
    date_added timestamp NOT NULL,
    date_updated timestamp,
    CONSTRAINT u_constraint UNIQUE (product_id, variation_id)
);