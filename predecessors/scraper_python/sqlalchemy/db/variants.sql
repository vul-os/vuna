CREATE TABLE variants (
    id SERIAL primary key,
    name text unique NOT NULL,
    url text unique,
    store_id int NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp NOT NULL
);
