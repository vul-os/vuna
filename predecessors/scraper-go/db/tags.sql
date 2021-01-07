CREATE TABLE tags (
    id BIGSERIAL PRIMARY KEY,
    tag_name text NOT NULL,
    url text unique,
    store int NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp
);
