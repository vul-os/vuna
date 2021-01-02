CREATE TABLE categories (
    id SERIAL primary key,
    name text NOT NULL,
    url text,
    parent_id INT,
    date_added timestamp NOT NULL
);
