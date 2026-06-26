//usr/bin/env go run $0 $@ ; exit
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/adrg/xdg"
	"github.com/gookit/color"
)

type Config struct {
	Version int
	Name    string
	Tags    []string
}

var wd word_data

type word_data struct {
	Date         string `json:"date"`
	Word         string `json:"word"`
	Definition   string `json:"definition"`
	Description  string `json:"description"`
	PartOfSpeech string `json:"partofspeech"`
	Phonetics    string `json:"phonetics"`
	Usage string `json:"usage"`
}

// i hate how many fucking structs i have to use for this but i dont understand go well enough to fix this
type Phonetic struct {
	Text      string `json:"text"`
	Audio     string `json:"audio"`
	SourceURL string `json:"sourceUrl"`
	License   struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"license"`
}

type Definition struct {
	Definition string `json:"definition"`
	Synonyms   []any  `json:"synonyms"`
	Antonyms   []any  `json:"antonyms"`
}

type Meaning struct {
	PartOfSpeech string       `json:"partOfSpeech"`
	Definitions  []Definition `json:"definitions"`
	Synonyms     []string     `json:"synonyms"`
	Antonyms     []string     `json:"antonyms"`
}

type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type DictionaryApiResp struct {
	Word       string     `json:"word"`
	Phonetic   string     `json:"phonetic"`
	Phonetics  []Phonetic `json:"phonetics"`
	Meanings   []Meaning  `json:"meanings"`
	License    License    `json:"license"`
	SourceUrls []string   `json:"sourceUrls"`
}

type WOTDApiResp struct {
	Date         string `json:"date"`
	Word         string `json:"word"`
	PartOfSpeech string `json:"pos"`
	Phonetics    string `json:"ipa"`
	Definition   string `json:"definition"`
}

// func loadConfig() {
// 	doc, fileerr := os.ReadFile(xdg.CacheHome + "/wotd_word.json")
// 	if fileerr != nil {
// 		fmt.Println("### cfg file not exist, whatever")
// 	}

// 	var cfg Config
// 	tomlerr := toml.Unmarshal([]byte(doc), &cfg)
// 	if tomlerr != nil {
// 		panic(tomlerr)
// 	}
// 	fmt.Println("cfg:", cfg)
// }

func fetchDefinition(word string) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://api.dictionaryapi.dev/api/v2/entries/en/" + word)
	if err != nil {
		// handle error
		fmt.Println("error fetching definition:", err)
		return
	}

	// read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Println("error reading response body:", err)
		return
	}

	// close the body and check for errors
	if cerr := resp.Body.Close(); cerr != nil {
		fmt.Println("error closing response body:", cerr)
	}

	var respdata []DictionaryApiResp

	jsonerr := json.Unmarshal(body, &respdata)
	if jsonerr != nil {
		fmt.Println("could not fetch definition, likely a temporary DictionaryAPI issue. try again later?")
	}

	// Assuming the API returns a list and you want the first item
	if len(respdata) > 0 {
		data := respdata[0]

		for _, meaning := range data.Meanings {
			for _, definition := range meaning.Definitions {
				// fmt.Println("got definition:", definition.Definition)

				wd.Definition = definition.Definition
				break
			}
		}

	} else {
		fmt.Println("fetchDefinition: no data found")
	}
}

func fetchWord() {
	// get the current date
	timenow := time.Now()

	// format date to YYYY-MM-DD
	date := timenow.Format("2006-01-02")

	// initialize http client
	client := http.Client{
		Timeout: 2500 * time.Millisecond,
	}

	// send a request to WOTD API using today's date
	resp, err := client.Get("https://api.wotd.site/query?date=" + date)
	if err != nil {
		// handle error
		fmt.Println("error fetching word:", err)
		return
	}

	// read response response body into a variable
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Println("error reading response body:", err)
		return
	}

	// close the body and check for errors
	if cerr := resp.Body.Close(); cerr != nil {
		fmt.Println("error closing response body:", cerr)
		panic(cerr)
	}

	var respdata WOTDApiResp

	jsonerr := json.Unmarshal(body, &respdata)
	if jsonerr != nil {
		fmt.Println("error while parsing word data, likely a temporary WOTD API issue. try again later?")
		panic(jsonerr)
	}

	wd.Date = respdata.Date
	wd.Word = respdata.Word
	wd.PartOfSpeech = respdata.PartOfSpeech
	wd.Phonetics = respdata.Phonetics
	wd.Definition = respdata.Definition

	// try getting a better definition
	fetchDefinition(wd.Word)

	// convert wd struct to json
	content, marshalErr := json.Marshal(wd)
	if marshalErr != nil {
		fmt.Println(marshalErr.Error())
	}

	// save word data to cache
	wordWriteErr := os.WriteFile(xdg.CacheHome+"/wotd_word.json", content, 0644)
	if wordWriteErr != nil {
		fmt.Println("could not save word to cache")
		fmt.Println(marshalErr.Error())
	}
}

func readWord() {
	content, err := os.ReadFile(xdg.CacheHome + "/wotd_word.json")
	if err != nil {
		fmt.Println("### cached word file not exist, calling fetchWord()")
	}

	jsonerr := json.Unmarshal(content, &wd)
	if jsonerr != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("%s (%s • %s)\n%s\n%s\n",
		color.Bold.Sprintf(wd.Word),
		color.OpItalic.Sprintf(wd.PartOfSpeech),
		wd.Phonetics, color.Gray.Sprintf(wd.Definition),
		color.OpItalic.Sprintf(color.Gray.Sprintf(wd.Usage)))

}

func cacheIsUpdated() bool {
	// read wotd cache file
	fileContent, fileReadErr := os.ReadFile(xdg.CacheHome + "/wotd_word.json")
	if fileReadErr != nil {
		// fmt.Println("### wotd cache ts file doesnt exist")
		return false
	}

	var cacheFileJSON word_data
	// parse cache file into a JSON object
	jsonerr := json.Unmarshal(fileContent, &cacheFileJSON)
	if jsonerr != nil {
		fmt.Println("failed to read cache file")
		return false
	}

	// get the current date
	timenow := time.Now()

	// format date to YYYY-MM-DD
	date := timenow.Format("2006-01-02")

	// does the cache file date match the current date?
	return cacheFileJSON.Date == date
}

func main() {
	// check if the cache is outdated
	if cacheIsUpdated() {
		// cache is up-to-date, we can use it

		// fmt.Println("### yo mistah white we got cache yo")

		readWord()
	} else {
		// cache is outdated we gotta update it

		// fmt.Println("### jesse we need to fetch")

		fetchWord()
		readWord()
	}
}
