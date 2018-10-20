package notebook

import (
	"encoding/json"
	"fmt"
	"github.com/HankWang95/Kanna/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/smartwalle/dbs"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
	"runtime"
	"os/exec"
)

type word struct {
	Id           int64                  `json:"id"                sql:"id"`
	Word         string                 `json:"word"              sql:"word"`
	Translations map[string]interface{} `json:"translations"      sql:"translations"`
	CreatedOn    *time.Time             `json:"created_on"        sql:"created_on"`
	AppearTime   int                    `json:"appear_time"       sql:"appear_time"`
	LastAppear   *time.Time             `json:"last_appear"       sql:"last_appear"`
}

func (w *word) FormatTranslations() {
	fmt.Println("---- ", w.Word, " ----")
	if v, ok := w.Translations["translation"]; ok {
		fmt.Println("基本翻译: ", v.([]interface{}))
	}
	if v, ok := w.Translations["basic"]; ok {
		basic := v.(map[string]interface{})
		if v, ok := basic["us-phonetic"]; ok {
			fmt.Println("美式发音: ", v.(string))
		}
		if v, ok := basic["uk-phonetic"]; ok {
			fmt.Println("英式发音: ", v.(string))
		}
		if v, ok := basic["us-speech"]; ok {
			usURL := fmt.Sprint(v.(string))
			downloadMP3(w.Word, usURL)
			fmt.Println("download")
		}
		if v, ok := basic["explains"]; ok {
			fmt.Println("其他释义: ", v.([]interface{}))
		}
	}
}

// ———————————————————————————————————————————— Loader -----------------------------------------------------

var queryWordChan = make(chan string, 10)

func queryWordEnter(word string) {
	w, err := queryWord(word)
	if err != nil {
		logger.Println(err)
	}
	w.FormatTranslations()
}

type wordLoader struct{}

func NewWordLoader() *wordLoader {
	return new(wordLoader)
}

func (this *wordLoader) LoadingFlag() (flagDict map[string]*chan string) {
	flagDict = make(map[string]*chan string)
	flagDict["w"] = &queryWordChan
	go flagHandler()
	return
}

func flagHandler() {
	for {
		select {
		case word := <-queryWordChan:
			queryWordEnter(word)
		}
	}
}

// ———————————————————————————————————————————— Service -----------------------------------------------------

var httpClient *http.Client
var logger *log.Logger

var (
	MaxIdleConns        = 100
	MaxIdleConnsPerHost = 100
)

func init() {
	httpClient = createHTTPClient()
	logFile, err := os.OpenFile("./kanna.log", os.O_WRONLY, os.ModeType)
	if err != nil {
		logger.Fatal(err)
	}
	logger = log.New(logFile, "[kanna-notebook] ", log.Ltime|log.Ldate|log.Lshortfile)
}

func createHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        MaxIdleConns,
			MaxIdleConnsPerHost: MaxIdleConnsPerHost,
			IdleConnTimeout:     time.Second * 90,
		},
	}
	return client
}

func youDaoTranslate(searchWord string) (result *word, err error) {
	//requestUrl := fmt.Sprintf("http://fanyi.youdao.com/openapi.do?keyfrom=%s&key=%s&type=data&doctype=json&version=1.2&q=%s", KeyFrom, Key, searchWord)
	requestUrl := fmt.Sprintf("http://fanyi.youdao.com/openapi.do?keyfrom=YouDaoCV&key=659600698&type=data&doctype=json&version=1.2&q=%s", searchWord)
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)

	// ---------------- set cookie ---------------
	//cookie1 := http.Cookie{
	//	Name: "OUTFOX_SEARCH_USER_ID", Value: "1573694584@171.223.98.5",
	//}
	//req.AddCookie(&cookie1)
	//cookie2 := http.Cookie{
	//	Name: "OUTFOX_SEARCH_USER_ID_NCOO", Value: "OUTFOX_SEARCH_USER_ID_NCOO",
	//}
	//req.AddCookie(&cookie2)
	//cookie3 := http.Cookie{
	//	Name: "SESSION_FROM_COOKIE", Value: "YouDaoCV",
	//}
	//req.AddCookie(&cookie3)
	//cookie4 := http.Cookie{
	//	Name: "UM_distinctid", Value: "164b7c2827886c-070e1ce5101c3c-163e6952-13c680-164b7c28279868",
	//}
	//req.AddCookie(&cookie4)
	//cookie5 := http.Cookie{
	//	Name: "_ntes_nnid", Value: "eeb80d4edb02b6dd360641d2eae0debd,1533630352086",
	//}
	//req.AddCookie(&cookie5)

	resp, err := httpClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		logger.Fatal("获取api出错", err)
	}

	if resp.Status != "200 OK" {
		logger.Print("调用翻译api发生错误")
		return nil, err
	}
	//body, err := ioutil.ReadAll(resp.Body)
	var wordStruct = new(word)
	wordStruct.Word = searchWord
	var f interface{}
	err = json.NewDecoder(resp.Body).Decode(&f)
	if err != nil {
		logger.Println(" decode err: ", err)
		return
	}
	wordStruct.Translations = f.(map[string]interface{})

	err = sqlAddWord(wordStruct)
	if err != nil {
		return nil, err
	}
	return wordStruct, nil
}

func downloadMP3(name, url string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Println(" download MP3 err :", err)
		return
	}

	// 查看是否有存放语音的文件夹
	_, err = os.Stat("./speech/")
	if err != nil{
		os.Mkdir("./speech/", os.ModeDir|0777)
	}

	f, err := os.OpenFile(fmt.Sprint("./speech/", name, ".mp3"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer f.Close()
	if err != nil {
		logger.Fatal(err)
		return
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Println(" download MP3 err :", err)
		return
	}
	io.Copy(f, resp.Body)
	resp.Body.Close()
	// todo afplay only in mac
	if runtime.GOOS == "darwin"{
		cmd := exec.Command("afplay", f.Name())
		cmd.Start()
	}


}

func queryWord(searchWord string) (result *word, err error) {
	//ws, err := mysql.GetWord(searchWord)
	//if err != nil {
	//	return nil, err
	//}
	//if ws != nil{
	//	err = mysql.UpdateWord(ws.Id)
	//	fmt.Println(ws.Id)
	//	log.Print(err)
	//	return ws, nil
	//}
	result, err = youDaoTranslate(searchWord)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ———————————————————————————————————————————— MySQL -----------------------------------------------------

const (
	K_DB_NOTEBOOK_WORD = "notebook_word"
)

func sqlAddWord(word *word) (err error) {
	//var now = time.Now()
	//var ib = dbs.NewInsertBuilder()
	//ib.Columns("word", "translations", "created_on", "appear_time", "last_appear")
	//ib.Values(word.word, word.Translations, now, 1, now)
	//ib.Table(K_DB_NOTEBOOK_WORD)
	//_, err = ib.Exec(server.GetMySQLSession())
	//if err != nil{
	//	return err
	//}
	return nil
}

func sqlGetWord(word string) (result *word, err error) {
	var sb = dbs.NewSelectBuilder()
	sb.Selects("id", "word", "translations", "created_on", "appear_time", "last_appear")
	sb.From(K_DB_NOTEBOOK_WORD)
	sb.Where("word = ?", word)
	err = sb.Scan(server.GetMySQLSession(), &result)
	if err != nil {
		return nil, err
	}
	return
}

func sqlUpdateWord(id int64) (err error) {
	var ub = dbs.NewUpdateBuilder()
	ub.SET("appear_time", "appear_time + 1")
	ub.SET("last_appear", time.Now())
	ub.Table(K_DB_NOTEBOOK_WORD)
	ub.Where("id = ?", id)
	_, err = ub.Exec(server.GetMySQLSession())
	return
}
