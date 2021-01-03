CREATE TABLE datapoints (
    id SERIAL primary key,
    product_variations_id integer NOT NULL,
    stock integer NOT NULL,
    price integer NOT NULL,
    date_added timestamp NOT null unique
);
