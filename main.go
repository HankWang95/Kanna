package main

import (
	"github.com/HankWang95/Kanna/notebook/load"
	"os"
)

func main() {
	flagDict := load.LoadingFlag()
	// todo 把 where 换为读取输入
	*flagDict[os.Args[1]] <- "where"
	select {

	}
}
