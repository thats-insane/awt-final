CREATE TABLE IF NOT EXISTS reviews (
    id bigserial PRIMARY KEY,
    book_id INT REFERENCES books(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    rating INT CHECK(rating BETWEEN 1 AND 5),
    description VARCHAR(225) NOT NULL,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);