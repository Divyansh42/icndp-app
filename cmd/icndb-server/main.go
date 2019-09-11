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
	port string
	name string
	version = "2.0"
)
func main() {
	flag.StringVar(&port, "port", "8000", "server listening port")
	flag.StringVar(&name, "name", "Chuck Norris", "name in jokes")
	flag.Parse()

	addr := ":" + port

	http.HandleFunc("/", crackJoke)
	log.Printf("Listening on: %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func crackJoke(w http.ResponseWriter, r *http.Request)  {
	tmplPath := filepath.Join("templates", "index.gohtml")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	joke, err := fetchJoke()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data := struct {
		Joke string
		Name string
		Version string
	}{
		Joke: joke,
		Name: name,
		Version: version,
	}
	log.Printf("joke: %s", joke)
	t.Execute(w, data)
}

type icndbRes struct {
	Type string `json:"type"`
	Value icndbPayload `json:"value"`
}

type icndbPayload struct {
	ID float32 `json:"id"`
	Joke string `json:"joke"`
	Categories []string `json:"categories"`
}

func fetchJoke() (string, error)  {
	firstLast := strings.Split(name, " ")
	url := "http://api.icndb.com/jokes/random?limitTo=nerdy"
	if len(firstLast) > 0 {
		url += "&firstName=" + firstLast[0] + "&lastName="
	}
	if len(firstLast) > 1 {
		url += firstLast[1]
	}
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
	return data.Value.Joke, nil
 }