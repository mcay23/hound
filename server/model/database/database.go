package database

import (
	_ "github.com/go-sql-driver/mysql"
	"os"
	"xorm.io/xorm"
)

const (
	MediaTypeTVShow          = "tvshow"
	MediaTypeMovie           = "movie"
	MediaTypeGame            = "game"
)

var databaseEngine *xorm.Engine

func InstantiateDB() {
	var err error
	databaseEngine, err = xorm.NewEngine(os.Getenv("DB_DRIVER"), os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}
	err = instantiateUsersTable()
	if err != nil {
		panic(err)
	}
	err = instantiateMediaTables()
	if err != nil {
		panic(err)
	}
	err = instantiateCommentTable()
	if err != nil {
		panic(err)
	}
}
