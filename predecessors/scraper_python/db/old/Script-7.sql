CREATE TABLE variants (
    id SERIAL primary key,
    variant_name text NOT NULL,
    product_id INT,
    date_added timestamp NOT NULL
);
