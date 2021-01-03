CREATE TABLE price_options (
    id SERIAL primary key,
    option_name text NOT NULL,
    product_id INT,
    date_added timestamp NOT NULL
);
