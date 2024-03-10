-- name: GetUserIdByStravaId :one
SELECT id FROM user where strava_id = ?;

-- name: GetUserByStravaId :one
SELECT * FROM user where strava_id = ?;

-- name: GetUserById :one
SELECT * FROM user WHERE id = ?;

-- name: GetTokenByUserId :one
SELECT sat.user_id, sat.access_token, sat.expires_at, srt.refresh_token FROM strava_access_token sat
JOIN main.strava_refresh_token srt on sat.user_id = srt.user_id
WHERE sat.user_id = ?;

-- name: InsertUser :one
INSERT INTO user (strava_id, first_name, last_name, profile, profile_medium) VALUES (?, ?, ?, ?, ?) RETURNING id;

-- name: InsertStravaAccessToken :exec
INSERT INTO strava_access_token (user_id, access_token, expires_at) VALUES (?, ?, ?);

-- name: InsertStravaRefreshToken :exec
INSERT INTO strava_refresh_token (user_id, refresh_token) VALUES (?, ?);

-- name: UpdateStravaAccessToken :exec
UPDATE strava_access_token SET access_token = ?, expires_at = ? WHERE user_id = ?;

-- name: UpdateStravaRefreshToken :exec
UPDATE strava_refresh_token SET refresh_token = ? where user_id = ?;

-- name: InsertSession :exec
INSERT INTO session (
    session_id ,
    ip_address,
    user_agent
) VALUES (?, ?, ?);

-- name: UpdateSessionUserId :exec
UPDATE session SET user_id = ? WHERE  session_id = ?;

-- name: GetUserIdFromSession :one
SELECT user_id FROM session WHERE session_id = ? AND last_activity_time > datetime('now', '-1 day');

-- name: GetSession :one
SELECT session_id FROM session WHERE session_id = ? AND last_activity_time > datetime('now', '-1 day');

-- name: DeleteSession :exec
DELETE FROM session
WHERE session_id = ?;
