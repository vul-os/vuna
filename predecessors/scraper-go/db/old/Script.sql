CREATE TABLE products (
    id SERIAL primary key,
    name text NOT NULL,
    url text not null unique,
    store integer NOT NULL,
    price real not null,
    date_added timestamp NOT NULL
);
