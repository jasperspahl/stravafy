CREATE TABLE IF NOT EXISTS user
(
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    strava_id      INT          NOT NULL UNIQUE,
    first_name     VARCHAR(255) NOT NULL,
    last_name      VARCHAR(255) NOT NULL,
    profile        VARCHAR(255) NOT NULL,
    profile_medium VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS spotify_user_info
(
    user_id      INTEGER PRIMARY KEY NOT NULL,
    spotify_id   VARCHAR(255)        NOT NULL,
    display_name VARCHAR(255)        NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user (id)
);

CREATE TABLE IF NOT EXISTS spotify_user_images
(
    user_id INTEGER NOT NULL,
    url VARCHAR(255) NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    PRIMARY KEY (user_id, width, height),
    FOREIGN KEY (user_id) REFERENCES user (id)
);

CREATE TABLE IF NOT EXISTS strava_access_token
(
    user_id      INT           NOT NULL PRIMARY KEY,
    access_token VARCHAR(1000) NOT NULL,
    expires_at   INT           NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user (id)
);

CREATE INDEX IF NOT EXISTS strava_access_token_expires_at_idx ON strava_access_token (expires_at);

CREATE TABLE IF NOT EXISTS strava_refresh_token
(
    user_id       INT           NOT NULL PRIMARY KEY,
    refresh_token VARCHAR(1000) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user (id)
);

CREATE TABLE IF NOT EXISTS session
(
    session_id         VARCHAR(50) PRIMARY KEY,
    user_id            INT,
    start_time         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address         VARCHAR(45),
    user_agent         VARCHAR(255),
    last_activity_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user (id)
);

CREATE TABLE IF NOT EXISTS spotify_access_token
(
    user_id      INT           NOT NULL PRIMARY KEY,
    access_token VARCHAR(1000) NOT NULL,
    token_type   VARCHAR(255)  NOT NULL,
    expires_at   INT           NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user (id)
);

CREATE INDEX IF NOT EXISTS spotify_access_token_expires_at_idx ON spotify_access_token (expires_at);

CREATE TABLE IF NOT EXISTS spotify_refresh_token
(
    user_id       INT           NOT NULL PRIMARY KEY,
    refresh_token VARCHAR(1000) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user (id)
);
