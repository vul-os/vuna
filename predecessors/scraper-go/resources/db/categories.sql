CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    category_name text NOT NULL,
    url text unique,
    store bigint NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp
);
