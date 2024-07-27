package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/adrg/xdg"
	"github.com/gocolly/colly"
	"github.com/gookit/color"
)

var wd word_data

type word_data struct {
	Word         string `json:"word"`
	Definition   string `json:"definition"`
	Description  string `json:"description"`
	PartofSpeech string `json:"partofspeech"`
	Syllables    string `json:"syllables"`
	Usage        string `json:"usage"`
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

func fetchDefinition(word string) {
	resp, err := http.Get("https://api.dictionaryapi.dev/api/v2/entries/en/" + word)
	if err != nil {
		// handle error
		fmt.Println("Error fetching definition:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Println("Error reading response body:", err)
		return
	}

	var respdata []DictionaryApiResp

	jsonerr := json.Unmarshal(body, &respdata)
	if jsonerr != nil {
		panic(jsonerr)
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
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		// fmt.Println("visiting", r.URL)
	})

	c.OnHTML("h2.word-header-txt", func(e *colly.HTMLElement) {
		// fmt.Println("got word:", e.Text)

		wd.Word = e.Text
	})

	c.OnHTML(".word-attributes span.main-attr", func(e *colly.HTMLElement) {
		// fmt.Println("got word_partofspeech:", e.Text)

		wd.PartofSpeech = e.Text
	})

	c.OnHTML(".word-attributes span.word-syllables", func(e *colly.HTMLElement) {
		// fmt.Println("got word_syllables:", e.Text)

		wd.Syllables = e.Text
	})

	c.OnHTML(".wod-definition-container > p:nth-child(2)", func(e *colly.HTMLElement) {
		// fmt.Println("got word_description:", e.Text)

		wd.Description = e.Text
	})

	c.OnHTML(".wod-definition-container > p:nth-child(3)", func(e *colly.HTMLElement) {
		// fmt.Println("got word_usage:", e.Text)

		wd.Usage = e.Text
	})

	// do the scrapey
	c.Visit("https://www.merriam-webster.com/word-of-the-day")

	// get definition
	fetchDefinition(wd.Word)

	// convert wd struct to json
	content, err := json.Marshal(wd)
	if err != nil {
		fmt.Println(err.Error())
	}

	// save word data to cache
	os.WriteFile(xdg.CacheHome+"/wotd_word.json", content, 0644)

	// save word date to cache
	timenow := time.Now()

	date := fmt.Sprintf("%d-%d-%d\n",
		timenow.Year(),
		timenow.Month(),
		timenow.Day())
	os.WriteFile(xdg.CacheHome+"/wotd_cache_ts.txt", []byte(date), 0644)
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
		color.OpItalic.Sprintf(wd.PartofSpeech),
		wd.Syllables, color.Gray.Sprintf(wd.Definition),
		color.OpItalic.Sprintf(color.Gray.Sprintf(wd.Usage)))

}

func cacheIsUpdated() bool {
	content, err := os.ReadFile(xdg.CacheHome + "/wotd_cache_ts.txt")
	if err != nil {
		// fmt.Println("### wotd cache ts file doesnt exist")
		return false
	}

	timenow := time.Now()
	date := fmt.Sprintf("%d-%d-%d\n",
		timenow.Year(),
		timenow.Month(),
		timenow.Day())

	return date == string(content)
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
