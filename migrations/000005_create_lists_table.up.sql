CREATE TABLE IF NOT EXISTS lists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(225) NOT NULL,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    book_list_id INT REFERENCES book_list(id) ON DELETE SET NULL,
    status TEXT NOT NULL
);