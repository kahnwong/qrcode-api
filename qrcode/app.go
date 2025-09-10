package qrcode

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	sqliteBase "github.com/kahnwong/sqlite-base"
)

const dbName = "qrcode"

var Qrcode *Application

type Application struct {
	DB *sqlx.DB
}

type QrcodeItem struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Image []byte `db:"image"`
}

func (Qrcode *Application) Add(qrcode QrcodeItem) error {
	query := `INSERT OR REPLACE INTO qrcode (id, name, image) VALUES (?, ?, ?)`
	_, err := Qrcode.DB.Exec(query, qrcode.ID, qrcode.Name, qrcode.Image)
	if err != nil {
		return fmt.Errorf("error inserting activity for qrcode: '%s' - %w", qrcode.Name, err)
	}

	return nil
}

func (Qrcode *Application) GetTitle(id int) (*QrcodeItem, error) {
	query := `SELECT id, name FROM qrcode WHERE id = ?`
	var qrcodeItem QrcodeItem
	err := Qrcode.DB.Get(&qrcodeItem, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("qrcode with ID '%d' not found", id)
		}
		return nil, fmt.Errorf("error getting qrcode by ID '%d': %w", id, err)
	}

	return &qrcodeItem, nil
}
func (Qrcode *Application) GetImage(id int) (*QrcodeItem, error) {
	query := `SELECT id, image FROM qrcode WHERE id = ?`
	var qrcodeItem QrcodeItem
	err := Qrcode.DB.Get(&qrcodeItem, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("qrcode with ID '%d' not found", id)
		}
		return nil, fmt.Errorf("error getting qrcode by ID '%d': %w", id, err)
	}

	return &qrcodeItem, nil
}

func init() {
	var dbFileName string
	if os.Getenv("MODE") != "DEVELOPMENT" {
		dbFileName = fmt.Sprintf("/data/%s.sqlite", dbName)
	} else {
		dbFileName = fmt.Sprintf("./%s.sqlite", dbName)
	}

	// init app
	dbExists := sqliteBase.IsDBExists(dbFileName)
	Qrcode = &Application{
		DB: sqliteBase.InitDB(dbFileName),
	}
	sqliteBase.InitSchema(dbFileName, Qrcode.DB, tableSchemas, allExpectedColumns, dbExists)
}
