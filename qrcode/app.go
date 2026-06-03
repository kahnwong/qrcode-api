package qrcode

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/kahnwong/qrcode-api/qrcode/store"
	sqliteBase "github.com/kahnwong/sqlite-base"
)

const dbName = "qrcode"

var Qrcode *Application

type Application struct {
	DB      *sql.DB
	Queries *store.Queries
}

type QrcodeItem struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Image []byte `db:"image"`
}

func (Qrcode *Application) Add(ctx context.Context, qrcode QrcodeItem) error {
	err := Qrcode.Queries.UpsertQrcode(ctx, store.UpsertQrcodeParams{
		ID: int64(qrcode.ID),
		Name: sql.NullString{
			String: qrcode.Name,
			Valid:  true,
		},
		Image: qrcode.Image,
	})
	if err != nil {
		return fmt.Errorf("error inserting activity for qrcode: '%s' - %w", qrcode.Name, err)
	}

	return nil
}

func (Qrcode *Application) GetTitle(ctx context.Context, id int) (*QrcodeItem, error) {
	row, err := Qrcode.Queries.GetQrcodeTitle(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("qrcode with ID '%d' not found", id)
		}
		return nil, fmt.Errorf("error getting qrcode by ID '%d': %w", id, err)
	}

	return &QrcodeItem{
		ID:   int(row.ID),
		Name: row.Name.String,
	}, nil
}
func (Qrcode *Application) GetImage(ctx context.Context, id int) (*QrcodeItem, error) {
	row, err := Qrcode.Queries.GetQrcodeImage(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("qrcode with ID '%d' not found", id)
		}
		return nil, fmt.Errorf("error getting qrcode by ID '%d': %w", id, err)
	}

	return &QrcodeItem{
		ID:    int(row.ID),
		Image: row.Image,
	}, nil
}

func initializeApp(dbFileName string) (*Application, error) {
	config := sqliteBase.Config{
		Path:         dbFileName,
		MigrationDir: "migrations",
	}

	db, err := sqliteBase.Open(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	app := &Application{
		DB:      db,
		Queries: store.New(db),
	}

	return app, nil
}

func init() {
	var dbFileName string
	if os.Getenv("MODE") != "DEVELOPMENT" {
		dbFileName = fmt.Sprintf("/data/%s.sqlite", dbName)
	} else {
		dbFileName = fmt.Sprintf("./%s.sqlite", dbName)
	}

	app, err := initializeApp(dbFileName)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize application: %v", err))
	}

	Qrcode = app
}
