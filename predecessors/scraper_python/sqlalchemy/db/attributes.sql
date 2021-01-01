CREATE TABLE attributes (
    id SERIAL primary key,
    name text NOT NULL,
    url text unique,
    store int NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp NOT NULL
);
