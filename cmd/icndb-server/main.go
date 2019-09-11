package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	port  string
	names string
	version = "exp"
)

func main() {
	flag.StringVar(&port, "port", "8000", "server listening port")
	flag.StringVar(&names, "names", "Chuck Norris", "name(s) in jokes 'FirstName LastName,FirstName LastName...")
	flag.Parse()

	addr := ":" + port

	http.HandleFunc("/", crackJoke)
	log.Printf("Listening on: %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Joke struct {
	Name string
	Joke string
}

func collectJokes() ([]Joke, error) {
	jokes := []Joke{}
	firstlastPairs := strings.Split(names, ",")
	for _, fullName := range firstlastPairs {
		url := buildUrl(fullName)
		jokeText, err := fetchJoke(url)
		if err != nil {
			return []Joke{}, err
		}
		joke := Joke{
			Name: fullName,
			Joke: jokeText,
		}
		jokes = append(jokes, joke)
	}
	return jokes, nil
}

func crackJoke(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "index.gohtml")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jokes, err := collectJokes()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := struct{
		Jokes []Joke
		Version string
	}{
		Jokes:   jokes,
		Version: version,
	}
	t.Execute(w, data)
}

type icndbRes struct {
	Type  string       `json:"type"`
	Value icndbPayload `json:"value"`
}

type icndbPayload struct {
	ID         float32  `json:"id"`
	Joke       string   `json:"joke"`
	Categories []string `json:"categories"`
}

func buildUrl(fullName string) string {
	url := "http://api.icndb.com/jokes/random?limitTo=nerdy"

	firstLast := strings.Split(fullName, " ")
	if len(firstLast) > 0 {
		url += "&firstName=" + firstLast[0] + "&lastName="
	}
	if len(firstLast) > 1 {
		url += firstLast[1]
	}
	return url
}

func fetchJoke(url string) (string, error) {
	log.Printf("url: %s", url)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	buff, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return "", err
	}

	data := icndbRes{}
	err = json.Unmarshal(buff, &data)
	if err != nil {
		return "", err
	}
	joke := data.Value.Joke
	log.Printf("joke: %s", joke)
	return joke, nil
}
