package qrcode

import (
	sqliteBase "github.com/kahnwong/sqlite-base"
)

var tableDefinitions = []sqliteBase.TableDefinition{
	{
		Name: "qrcode",
		CreateSQL: `
	CREATE TABLE qrcode (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT,
    image BLOB
);`,
	},
}
