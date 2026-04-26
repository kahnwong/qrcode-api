-- +goose Up
CREATE TABLE qrcode (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    image BLOB
);

-- +goose Down
DROP TABLE qrcode;
