CREATE TABLE product_categories (
    id BIGSERIAL PRIMARY KEY,
    product_id bigint not null,
    category_id bigint not null,
    date_added timestamp NOT NULL,
    date_updated timestamp,
    CONSTRAINT c_constraint UNIQUE (product_id, category_id)
);