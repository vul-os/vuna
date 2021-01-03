CREATE TABLE variations (
    id SERIAL primary key,
    variation_id text unique NOT NULL,
    sku text unique,
    date_added timestamp NOT NULL,
    date_updated timestamp
);