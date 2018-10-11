package service

import (
	"fmt"
	"github.com/HankWang95/Kanna/notebook/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/HankWang95/Kanna/notebook"
)

// e.g. http://fanyi.youdao.com/openapi.do?keyfrom=YouDaoCV&key=659600698&type=data&doctype=json&version=1.2&q=search

var (
	// todo 写配置文件中
	KeyFrom = "YouDaoCV"
	Key     = "659600698"
)

func translateWord(word string) (result *notebook.WordFromYouDao,err error) {
	requestUrl := fmt.Sprintf("http://fanyi.youdao.com/openapi.do?keyfrom=%s&key=%s&type=data&doctype=json&version=1.2&q=%s", KeyFrom, Key, word)
	resp, err := http.Get(requestUrl)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.Status != "200 OK" {
		log.Print("调用翻译api发生错误")
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var wordTranslation = &notebook.WordFromYouDao{}
	wordTranslation.Word = word
	wordTranslation.CleanedTranslation(string(body))

	return wordTranslation, nil
}

func packQueryWordAndSave(word *notebook.WordFromYouDao) (result *notebook.Word, err error) {
	var wordStruct = &notebook.Word{}
	wordStruct.Word = word.Word
	wordStruct.Translations = fmt.Sprint(word.Translation.Translation, word.Translation.Basic.Explains)
	err = mysql.AddWord(wordStruct)
	if err != nil{
		return nil, err
	}
	return wordStruct, nil
}

func youDaoTranslate(word string) (result *notebook.Word, err error) {
	youDaoWord, err := translateWord(word)
	if err != nil {
		return nil, err
	}
	wordStruct, err := packQueryWordAndSave(youDaoWord)
	if err != nil {
		return nil, err
	}
	return wordStruct, nil
}

func QueryWord(word string) (result *notebook.Word, err error) {
	ws, err := mysql.GetWord(word)
	if err != nil {
		return nil, err
	}
	if ws != nil{
		err = mysql.UpdateWord(ws.Id)
		log.Print(err)
		return ws, nil
	}
	result, err = youDaoTranslate(word)
	if err != nil {
		return nil, err
	}
	return
}