CREATE TABLE product_tags (
    id BIGSERIAL PRIMARY KEY,
    product_id bigint not null,
    tag_id bigint not null,
    date_added timestamp NOT NULL,
    date_updated timestamp,
    CONSTRAINT t_constraint UNIQUE (product_id, tag_id)
);