package models

import (
	"os"

	adapter "github.com/beego/beego/v2/adapter"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	_ "github.com/lib/pq"
)

var TestnetFlg bool

func init() {
	// postgres driver register
	err := orm.RegisterDriver("postgres", orm.DRPostgres)
	if err != nil {
		adapter.Error("Postgres register driver error:", err)
	}

	//get database url config
	DATABASE_URL, err := beego.AppConfig.String("DATABASE_URL")
	if err != nil {
		adapter.Error("Postgres register driver error:", err.Error())
		os.Exit(1)
	}

	// register and connect database
	err = orm.RegisterDataBase("default", "postgres", DATABASE_URL)
	if err != nil {
		logs.Error(err.Error())
		os.Exit(1)
	}
	// register model
	orm.RegisterModel(new(User))

	// Run sync db
	orm.RunSyncdb("default", false, true)
}
