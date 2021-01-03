CREATE TABLE attributes (
    id SERIAL primary key,
    attribute_name text NOT NULL,
    store_id int NOT NULL,
    url text unique,
    date_added timestamp NOT NULL,
    date_updated timestamp
);
