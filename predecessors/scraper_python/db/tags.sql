CREATE TABLE tags (
    id SERIAL primary key,
    tag_name text NOT NULL,
    url text unique,
    store int NOT NULL,
    date_added timestamp NOT NULL,
    date_updated timestamp
);
