CREATE TABLE skus (
    id SERIAL primary key,
    sku_id text NOT NULL,
    product_id int NOT NULL,
    variant_id int,
    attribute_id int,
    date_scraped timestamp NOT null,
    date_added timestamp NOT NULL
);
