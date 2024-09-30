CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS lists (
    id serial PRIMARY KEY,
    title varchar(255) NOT NULL,
    description varchar(255)
);

CREATE TABLE IF NOT EXISTS items (
    id serial PRIMARY KEY,
    title varchar(255) NOT NULL,
    description varchar(255),
    done boolean NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS users_lists (
    id serial PRIMARY KEY,
    user_id int REFERENCES users (id) ON DELETE CASCADE NOT NULL,
    list_id int REFERENCES lists (id) ON DELETE CASCADE NOT NULL
);

CREATE TABLE IF NOT EXISTS lists_items (
    id serial PRIMARY KEY,
    item_id int REFERENCES items (id) ON DELETE CASCADE NOT NULL,
    list_id int REFERENCES lists (id) ON DELETE CASCADE NOT NULL
);

CREATE INDEX idx_users_lists_user_id ON users_lists(user_id);
CREATE INDEX idx_users_lists_list_id ON users_lists(list_id);

CREATE INDEX idx_lists_items_item_id ON lists_items(item_id);
CREATE INDEX idx_lists_items_list_id ON lists_items(list_id);
