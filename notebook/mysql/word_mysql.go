package mysql

import (
	"github.com/HankWang95/Kanna/notebook"
	"github.com/smartwalle/dbs"
	_ "github.com/go-sql-driver/mysql"
	"time"
)
// todo 数据库连接选项移到配置文件
var db, _ = dbs.NewSQL("mysql", "root:whc5608698@tcp(127.0.0.1)/kanna?parseTime=true", 30, 5)

const(
	K_DB_NOTEBOOK_WORD = "notebook_word"
)

func AddWord(word *notebook.Word) (err error) {
	var now = time.Now()
	var ib = dbs.NewInsertBuilder()
	ib.Columns("word", "translations", "created_on", "appear_time", "last_appear")
	ib.Values(word.Word, word.Translations, now, 1, now)
	ib.Table(K_DB_NOTEBOOK_WORD)
	_, err = ib.Exec(db)
	if err != nil{
		return err
	}
	return nil
}

func GetWord(word string) (result *notebook.Word, err error) {
	var sb = dbs.NewSelectBuilder()
	sb.Selects("word", "translations", "created_on", "appear_time", "last_appear")
	sb.From(K_DB_NOTEBOOK_WORD)
	sb.Where("word = ?", word)
	err = sb.Scan(db, &result)
	if err != nil {
		return nil, err
	}
	return
}

func UpdateWord(id int64) (err error) {
	var ib = dbs.NewInsertBuilder()
	ib.SET("appear_time", "appear_time + 1")
	ib.SET("last_appear", time.Now())
	_, err = ib.Exec(db)
	return
}