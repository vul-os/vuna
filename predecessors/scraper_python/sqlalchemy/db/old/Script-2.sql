CREATE TABLE datapoints (
    id SERIAL primary key,
    product integer NOT NULL,
    stock integer NOT NULL,
    date_scraped timestamp NOT null,
    date_added timestamp NOT NULL
);
