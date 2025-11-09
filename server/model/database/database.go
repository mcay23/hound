package database

import (
	"hound/helpers"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
	"xorm.io/xorm"
)

const (
	MediaTypeTVShow = "tvshow"
	MediaTypeMovie  = "movie"
	MediaTypeGame   = "game"
)

var databaseEngine *xorm.Engine

func InstantiateDB() {
	var err error
	slog.Info("DB loaded:", "driver", os.Getenv("DB_DRIVER"))
	databaseEngine, err = xorm.NewEngine(os.Getenv("DB_DRIVER"), os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate DB connection")
		panic(err)
	}
	err = instantiateUsersTable()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate users table")
		panic(err)
	}
	err = instantiateMediaTables()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate media tables")
		panic(err)
	}
	err = instantiateCommentTable()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate comment table")
		panic(err)
	}
	err = runMigrations()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to migrate databases!")
		panic(err)
	}
	slog.Info("DB tables initialized")
}
