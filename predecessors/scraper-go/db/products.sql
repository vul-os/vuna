CREATE TABLE products (
    id SERIAL primary key,
    product_name text NOT NULL,
    url text not null unique,
    store integer NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp
);
