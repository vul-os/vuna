CREATE TABLE datapoints (
    id SERIAL primary key,
    sku_id integer NOT NULL,
    stock integer NOT NULL,
    price integer NOT NULL,
    date_added timestamp NOT null unique
);
