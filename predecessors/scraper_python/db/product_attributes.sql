CREATE TABLE product_attributes (
    id SERIAL primary key,
    product_id int NOT NULL,
    attribute_id int NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp NOT NULL,
);