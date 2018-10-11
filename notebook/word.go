package notebook

import (
	"encoding/json"
	"log"
	"time"
)

type WordFromYouDao struct {
	Word        string `json:"word" sql:"word"`
	Translation *Translation
}

func (w *WordFromYouDao) CleanedTranslation(p string) {
	var result = &Translation{}
	err := json.Unmarshal([]byte(p), &result)
	if err != nil {
		log.Println(err)
		return
	}
	w.Translation = result
	return
}

type Translation struct {
	Translation []string `json:"translation"`
	Basic       *Basic   `json:"basic"`
}

type Basic struct {
	UkSpeech string   `json:"uk-speech"`
	Explains []string `json:"explains"`
}

type Word struct {
	Id           int64      `json:"id"                sql:"id"`
	Word         string     `json:"word"              sql:"word"`
	Translations string     `json:"translations"      sql:"translations"`
	CreatedOn    *time.Time `json:"created_on"        sql:"created_on"`
	AppearTime   int        `json:"appear_time"       sql:"appear_time"`
	LastAppear   *time.Time `json:"last_appear"       sql:"last_appear"`
}
