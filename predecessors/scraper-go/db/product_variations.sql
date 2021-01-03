CREATE TABLE product_variations (
    id SERIAL primary key,
    product_id int NOT NULL,
    variation_id int,
    date_added timestamp NOT NULL,
    date_updated timestamp NOT NULL
);
