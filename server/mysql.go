package server

import (
	"github.com/smartwalle/dbs"
	"log"
)

var db dbs.DB

func InitMySQL() {
	var err error
	// todo 配置文件
	db, err = dbs.NewSQL("mysql", "root:hankwang@tcp(127.0.0.1)/kanna?parseTime=true", 30, 5)
	if err != nil{
		log.Fatal(err)
	}
}

func GetMySQLSession() dbs.DB {
	return db
}