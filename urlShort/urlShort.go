package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func exitOnFail(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

type encoding struct {
	Type string
}

func (e *encoding) Unmarshal(data []byte, out interface{}) error {
	if e.Type == "yaml" {
		return yaml.Unmarshal(data, out)
	}

	return json.Unmarshal(data, out)
}

type urlPath struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url" json:"url"`
}

type urlPaths []urlPath

func convertToMap(paths urlPaths) map[string]string {
	urlPathsMap := make(map[string]string)
	for _, p := range paths {
		urlPathsMap[p.Path] = p.URL
	}

	return urlPathsMap
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}

func newServeMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	return mux
}

func encodingHandler(urlPathsMap map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if dest, ok := urlPathsMap[r.URL.Path]; ok {
			http.Redirect(w, r, dest, http.StatusFound)
			return
		}

		fallback.ServeHTTP(w, r)
	}
}

func main() {
	pathEncodedFile := flag.String("enc", "paths.yaml", "Specify a YAML or JSON file containing 'path: url' information.")
	flag.Parse()

	raw, err := ioutil.ReadFile(*pathEncodedFile)
	if err != nil {
		exitOnFail(fmt.Sprintf("Unable to parse file: %s", *pathEncodedFile))
	}

	enc := encoding{Type: strings.Split(*pathEncodedFile, ".")[1]}
	var paths urlPaths
	if err := enc.Unmarshal(raw, &paths); err != nil {
		exitOnFail(fmt.Sprintf("An error ocurred: %s", err.Error()))
	}
	urlPathsMap := convertToMap(paths)

	mux := newServeMux()
	redirectHandler := encodingHandler(urlPathsMap, mux)

	log.Fatal(http.ListenAndServe(":8080", redirectHandler))
}
