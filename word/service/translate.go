package service

import (
	"fmt"
	"net/http"
	"encoding/json"
	"log"
	"io/ioutil"
)

// e.g. http://fanyi.youdao.com/openapi.do?keyfrom=YouDaoCV&key=659600698&type=data&doctype=json&version=1.2&q=search

var (
	KeyFrom = "YouDaoCV"
	Key = "659600698"
)

type WordTranslation struct {
	Id int64
	Word string
	Translation *Translation
}

type Translation struct {
	WordId int64 `json:"word_id"`
	Translation []string `json:"translation"`
	Basic *Basic `json:"basic"`
}

type Basic struct {
	UkSpeech string `json:"uk-speech"`
	Explains []string `json:"explains"`
}



func (w *WordTranslation) CleanedTranslation(p string) () {
	var result = &Translation{}
	err := json.Unmarshal([]byte(p), &result)
	if err != nil{
		log.Println(err)
		return
	}
	w.Translation = result
	return
}

func TranslateWord(word string) error{
	requestUrl := fmt.Sprintf("http://fanyi.youdao.com/openapi.do?keyfrom=%s&key=%s&type=data&doctype=json&version=1.2&q=%s", KeyFrom, Key, word)
	resp,err := http.Get(requestUrl)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.Status != "200 OK" {
		log.Print("调用翻译api发生错误")
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var wordTranslation = &WordTranslation{}
	wordTranslation.Word = word
	wordTranslation.CleanedTranslation(string(body))

	return nil
}

func main() {
	TranslateWord("what")
}

