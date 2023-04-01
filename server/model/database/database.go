package database

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"xorm.io/xorm"
)

var databaseEngine *xorm.Engine

func InstantiateDB() {
	var err error
	databaseEngine, err = xorm.NewEngine(viper.GetString("DB_DRIVER"), viper.GetString("DB_CONNECTION_STRING"))
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
}
