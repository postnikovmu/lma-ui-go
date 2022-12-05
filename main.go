package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

type vacancies []struct {
	URL          string `json:"strUrl"`
	Name         string `json:"strJobTitle"`
	AreaName     string `json:"strArea"`
	EmployerName string `json:"strCompany"`
	Description  string `json:"strBodyFull"`
	KeySkills    []struct {
		Name string `json:"name"`
	} `json:"strArrKeySkills"`
}

type RespData struct {
	Title    string
	Response string
	List     PairList
	Text     string
	Area     string
	ItemsNum string
}

func rankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func handler(w http.ResponseWriter, r *http.Request) {

	//Create a variable of the same type as our model
	var ltVacancies vacancies

	lmSkills := make(map[string]int)

	var strText, strArea, itemsNum string
	if r.Method == "POST" {
		strText = r.FormValue("strText")
		strArea = r.FormValue("strArea")
		fmt.Println(strText, strArea)
	}

	if strText != "" && strArea != "" {
		lvText := url.QueryEscape(strText)
		lvArea := url.QueryEscape(strArea)
		//Build The URL string
		URL := "https://lma-extractor-hh.cfapps.us10.hana.ondemand.com/hh4?text=" + lvText + "&" + "area=" + lvArea
		//We make HTTP request using the Get function
		resp, err := http.Get(URL)
		if err != nil {
			log.Fatal("Sorry, an error occurred, please try again")
		}
		defer resp.Body.Close()

		//Decode the data
		if err := json.NewDecoder(resp.Body).Decode(&ltVacancies); err != nil {
			log.Fatal("Sorry, an error occurred, please try again")
		}

		for _, line := range ltVacancies {
			for _, lineSkill := range line.KeySkills {
				lmSkills[lineSkill.Name] += 1
			}
		}

		itemsNum = strconv.Itoa(len(lmSkills)) + " skills are found"
	}

	lmSortedSkills := rankByWordCount(lmSkills)

	respData := RespData{
		Title:    "Skills analyzer",
		Response: "Welcome to the skills analyzer",
		List:     lmSortedSkills,
		Text:     strText,
		Area:     strArea,
		ItemsNum: itemsNum,
	}

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Add("Content-Type", "text/html")
	t.Execute(w, respData)
}

func handler2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from Go")
}

func main() {
	http.HandleFunc("/hh4/", handler)
	http.HandleFunc("/", handler2)
	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./templates"))))
	//http.ListenAndServe("localhost:8080", nil) //locally
	http.ListenAndServe(":8080", nil) //SAP BTP CloudFounry
}
