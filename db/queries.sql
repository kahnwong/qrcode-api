-- name: UpsertQrcode :exec
INSERT OR REPLACE INTO qrcode (id, name, image)
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.arg(image));

-- name: GetQrcodeTitle :one
SELECT id, name FROM qrcode WHERE id = sqlc.arg(id);

-- name: GetQrcodeImage :one
SELECT id, image FROM qrcode WHERE id = sqlc.arg(id);
