CREATE TABLE IF NOT EXISTS forums (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    tags TEXT NOT NULL,
    user_id INTEGER NOT NULL, 
    created DATETIME NOT NULL,
    expires DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NUll UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    hashed_password CHAR(60) NOT NULL,
    created DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    token CHAR(43) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expiry TEXT NOT NUll
);

CREATE TABLE IF NOT EXISTS forum_likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    forum_id INTEGER,
    user_id INTEGER NOT NULL,
    comment_id INTEGER,
    like_status INTEGER NOT NULL,
    FOREIGN KEY (forum_id) REFERENCES forums (id),
    FOREIGN KEY (comment_id) REFERENCES forum_comments (id),
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS forum_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    forum_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    comment TEXT NOT NULL,
    FOREIGN KEY (forum_id) REFERENCES forums (id),
    FOREIGN KEY (user_id) REFERENCES users (id)
);
