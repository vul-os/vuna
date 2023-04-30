CREATE TABLE attributes (
     id BIGSERIAL PRIMARY KEY,
     attribute_name text NOT NULL,
     attribute_value text NOT NULL,
     store_id bigint NOT NULL,
     date_added timestamp NOT NULL,
     date_updated timestamp,
     CONSTRAINT att_constraint UNIQUE (attribute_name, attribute_value, store_id)
);
