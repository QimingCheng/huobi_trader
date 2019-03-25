package main

import (
	_ "github.com/xfrzrcj/huobi_trader/routers"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego"
)

const(
	DATA_SOURCE_STR = "data_source"
	DEFAULT_STR = "default"
	MYSQL_STR = "mysql"
	MAX_IDLE_CONN = 10
	MAX_OPEN_CONN = 10
	RUN_MODE = "dev"
)

func init(){
	dataSource := beego.AppConfig.String(DATA_SOURCE_STR)
	orm.RegisterDataBase(DEFAULT_STR,MYSQL_STR,dataSource,MAX_IDLE_CONN,MAX_OPEN_CONN)

}

func main() {
	logs.SetLogger("console")
	logs.EnableFuncCallDepth(true)
	if beego.BConfig.RunMode == RUN_MODE {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}
