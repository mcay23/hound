package database

import (
	"fmt"
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
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	slog.Info("Attempting DB connection", "uri", connectionString)
	databaseEngine, err = xorm.NewEngine(DriverPostgres, connectionString)
	if err != nil {
		_ = helpers.LogErrorWithMessage(err, "Failed to instantiate DB connection")
		panic(err)
	}
	slog.Info("DB Connection successful")

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
