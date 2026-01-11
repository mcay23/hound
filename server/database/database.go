package database

import (
	"hound/helpers"
	"log/slog"
	"os"
	"time"

	"xorm.io/xorm"
)

const (
	MediaTypeTVShow = "tvshow"
	MediaTypeMovie  = "movie"
	MediaTypeGame   = "game"
	DriverPostgres  = "postgres"
)

var databaseEngine *xorm.Engine

func InstantiateDB() {
	var err error
	slog.Info("DB loaded", "driver", DriverPostgres)
	databaseEngine, err = xorm.NewEngine(DriverPostgres, os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate DB connection")
		panic(err)
	}
	// always use UTC for timestamps
	tz, _ := time.LoadLocation("UTC")
	databaseEngine.SetTZDatabase(tz)
	databaseEngine.SetTZLocation(tz)
	err = instantiateUsersTable()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate users table")
		panic(err)
	}
	err = instantiateCollectionTables()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate collection tables")
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
	err = instantiateWatchTables()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate watch tables")
		panic(err)
	}
	err = instantiateMediaFilesTable()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate media files table")
		panic(err)
	}
	err = instantiateIngestTasksTable()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate ingest tasks table")
		panic(err)
	}
	slog.Info("DB tables initialized")
	err = runMigrations()
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to migrate databases!")
		panic(err)
	}
	slog.Info("DB migrations complete")
}

func NewSession() *xorm.Session {
	return databaseEngine.NewSession()
}
