CREATE TABLE skus (
    id SERIAL primary key,
    variation_id int NOT NULL,
    product_id int NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp
);