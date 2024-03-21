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

-- name: UpdateSessionLastActivityTime :exec
UPDATE session SET last_activity_time = datetime('now') WHERE session_id = ?;

-- name: InsertSpotifyAccessToken :exec
INSERT INTO spotify_access_token (user_id, access_token, token_type, expires_at) VALUES (?, ?, ?, ?);

-- name: InsertSpotifyRefreshToken :exec
INSERT INTO spotify_refresh_token (user_id, refresh_token) VALUES (?, ?);

-- name: GetSpotifyAccessToken :one
SELECT sat.user_id, sat.access_token, sat.token_type, sat.expires_at, srt.refresh_token FROM spotify_access_token sat
JOIN main.spotify_refresh_token srt on sat.user_id = srt.user_id
WHERE sat.user_id = ?;


-- name: UpdateSpotifyAccessToken :exec
UPDATE  spotify_access_token SET access_token = ?, expires_at = ? WHERE user_id = ?;

-- name: UpdateSpotifyRefreshToken :exec
UPDATE  spotify_refresh_token SET refresh_token = ? WHERE user_id = ?;

-- name: GetSpotifyUserInfo :one
SELECT * FROM spotify_user_info WHERE user_id = ?;

-- name: InsertSpotifyUserInfo :exec
INSERT INTO spotify_user_info (user_id, spotify_id, display_name) VALUES (?, ?, ?);

-- name: InsertSpotifyUserImage :exec
INSERT INTO spotify_user_images (user_id, url, width, height) VALUES (?, ?, ?, ?);

-- name: GetUserIdsWithActiveSpotify :many
SELECT user_id from spotify_user_info;

-- name: InsertHistory :one
INSERT INTO spotify_user_history (user_id, timestamp, is_playing) VALUES (?, ?, ?) RETURNING id;

-- name: InsertHistoryContext :exec
INSERT INTO spotify_user_history_context (history_id, type, href, external_url, uri) VALUES (?, ?, ?, ?, ?);

-- name: GetLastHistoryEntryForUser :one
SELECT * FROM spotify_user_history
WHERE user_id = ?
ORDER BY timestamp DESC
LIMIT 1;

-- name: InsertHistoryItem :exec
INSERT INTO spotify_user_history_item (history_id, type, href, external_url, uri, name, artists, album, album_uri,
                                       episode_description, episode_show_name, episode_show_description,
                                       episode_show_uri)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetLastHistoryEntryComplete :one
SELECT id,
       timestamp,
       is_playing,
       ctx.type ctx_type,
       ctx.href ctx_href,
       ctx.external_url ctx_external_url,
       ctx.uri ctx_uri,
       item.type item_type,
       item.href item_href,
       item.external_url item_external_url,
       item.uri item_uri,
       item.name,
       item.artists,
       item.album,
       item.album_uri,
       item.episode_description,
       item.episode_show_name,
       item.episode_show_description,
       item.episode_show_uri
FROM spotify_user_history
         JOIN main.spotify_user_history_context ctx on spotify_user_history.id = ctx.history_id
         JOIN main.spotify_user_history_item item on spotify_user_history.id = item.history_id
WHERE user_id = ?
ORDER BY timestamp DESC
LIMIT 1;

-- name: GetHistoryEntriesBetween :many
SELECT id,
       timestamp,
       is_playing,
       ctx.type ctx_type,
       ctx.href ctx_href,
       ctx.external_url ctx_external_url,
       ctx.uri ctx_uri,
       item.type item_type,
       item.href item_href,
       item.external_url item_external_url,
       item.uri item_uri,
       item.name,
       item.artists,
       item.album,
       item.album_uri,
       item.episode_description,
       item.episode_show_name,
       item.episode_show_description,
       item.episode_show_uri
FROM spotify_user_history
         JOIN main.spotify_user_history_context ctx on spotify_user_history.id = ctx.history_id
         JOIN main.spotify_user_history_item item on spotify_user_history.id = item.history_id
WHERE
user_id = ? AND timestamp > ? AND timestamp < ?
ORDER BY timestamp;
