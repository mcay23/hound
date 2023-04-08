package database

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"hound/helpers"
	"os"
	"time"
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
	fmt.Println("driver", os.Getenv("DB_DRIVER"))
	retries := 10
	sum := 0
	// in case db container does not spin up in time
	for sum < retries {
		sum += sum
		databaseEngine, err = xorm.NewEngine(os.Getenv("DB_DRIVER"), os.Getenv("DB_CONNECTION_STRING"))
		if err != nil {
			_ = helpers.LogErrorWithMessage(errors.New(helpers.InternalServerError), "Error connecting to hound db, retrying..")
			sum += 1
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
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
