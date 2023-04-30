CREATE TABLE stores (
    id BIGSERIAL PRIMARY KEY,
    store_name text NOT NULL unique,
    url text NOT NULL unique,
    date_added timestamp NOT NULL,
    date_updated timestamp
);
