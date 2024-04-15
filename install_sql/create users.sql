CREATE TABLE users (
    id SERIAL PRIMARY KEY,              -- Automatically incrementing integer, used as a primary key
    username VARCHAR(255) NOT NULL,     -- Variable character field for usernames, cannot be null
    password VARCHAR(255) NOT NULL      -- Variable character field for passwords, cannot be null
);
