package load

import (
	"fmt"
	"github.com/HankWang95/Kanna/notebook/service"
)

var queryWordChan = make(chan string)

// todo 统一装载 接口
func queryWord(word string) {
	fmt.Println(word, "ojbk")
	w, err := service.QueryWord(word)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(w.Translations)
}

func LoadingFlag() (flagDict map[string]*chan string) {
	flagDict = make(map[string]*chan string)
	flagDict["word"] = &queryWordChan
	go flagHandler()
	return
}

func flagHandler()  {
	select {
	case word :=<- queryWordChan:
		queryWord(word)
	}
}