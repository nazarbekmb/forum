CREATE TABLE IF NOT EXISTS users (
    user_id  INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE CONSTRAINT users_uc_username,
    email TEXT NOT NULL UNIQUE CONSTRAINT users_uc_email,
    hash_password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS posts (
    post_id  INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_date DATE NOT NULL,
    user_id INTEGER,
    author TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES sessions (user_id)
);

CREATE TABLE IF NOT EXISTS categories (
    category_id  INTEGER PRIMARY KEY,
    category TEXT
);

INSERT OR IGNORE INTO categories (category_id, category) VALUES 
(1, "Counter-Strike 2"), 
(2, "Dota 2"),
(3, "Valorant"),
(4, "Overwatch 2"),
(5, "Other");

CREATE TABLE IF NOT EXISTS posts_categories (
    post_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts (post_id),
    FOREIGN KEY (category_id) REFERENCES categories (category_id)
);

CREATE TABLE IF NOT EXISTS sessions (
    user_id INTEGER,
    session_token TEXT,
    expires_at TIME
);

CREATE TABLE IF NOT EXISTS comments (
    comment_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    comment TEXT NOT NULL,
    created_at DATE NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (post_id) REFERENCES posts(post_id)
);


CREATE TABLE IF NOT EXISTS reactions (
    user_id INTEGER NOT NULL,
    post_id INTEGER NOT NULL,
    like_status INTEGER DEFAULT 0 NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (post_id) REFERENCES posts(post_id)
    PRIMARY KEY (user_id, post_id)
)
