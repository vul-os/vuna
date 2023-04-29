CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    url text not null unique,
    store bigint NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp
);
