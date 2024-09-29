package main

import (
	"fmt"
	"go_collector_server/database"
	"go_collector_server/server"
	"go_collector_server/util"
)

var logLevel = "debug"

func main() {
	fmt.Printf("Set log level:%s\n", logLevel)
	util.BuildLogger(logLevel)

	go func() {
		database.DBChannel <- database.CreateDatabase()
		database.SetupSqliteDatabase() // 调用创建方法
	}()

	fmt.Println("unix server start!")
	server.StartUnix()
}
