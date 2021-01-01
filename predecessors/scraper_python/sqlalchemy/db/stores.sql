CREATE TABLE stores (
    id SERIAL primary key,
    name text NOT NULL unique,
    url text NOT NULL unique,
    date_added timestamp NOT NULL unique
);
