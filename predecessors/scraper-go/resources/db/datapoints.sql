CREATE TABLE datapoints (
    id BIGSERIAL PRIMARY KEY,
    variation_id bigint NOT NULL,
    stock integer NOT NULL,
    price float NOT NULL,
    date_added timestamp NOT null
);
