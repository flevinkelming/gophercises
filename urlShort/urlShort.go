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

func main() {
	fname := flag.String("fname", "paths.yaml", "Specify the redirect YAML or JSON encoded file.")
	ftype := flag.String("ftype", "yaml", "Specify the file encoding type.")
	flag.Parse()

	if strings.Split(*fname, ".")[1] != *ftype {
		flag.Usage()
		os.Exit(1)
	}

	bs, err := ioutil.ReadFile(*fname) // "Because ReadFile reads the whole file,
					   // it does not treat an EOF from Read as an error to be reported."
	if err != nil {
		log.Fatalf("[!] Error while reading file: %s", err) // "Fatalf is equivalent to Printf()
								    // followed by a call to os.Exit(1)."
	}

	var paths pathURLs

	var enc encoding = *ftype
	if err := enc.Unmarshal(bs, &paths); err != nil {
		log.Fatalf("[!] Error while decoding .%s file: %s", *ftype, err)
	}

	pathMap := newMap(paths)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)

	rmw := redirectMiddleware(pathMap, mux)

	log.Fatal(http.ListenAndServe(":8080", rmw))
}

type encoding string

func (e *encoding) Unmarshal(data []byte, v interface{}) error {
	if e == "yaml" {
		return yaml.Unmarshal(data, v)
	}

	return json.Unmarshal(data, v)
}

type pathURL struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url" json:"url"`
}

type pathURLs []pathURL

// Factory function (constructor) w/ identifier as type, prepended by `new` (`New` if exported).
func newMap(paths pathURLs) map[string]string {
	pathMap := make(map[string]string, len(paths))
	for _, p := range paths {
		pathMap[p.Path] = p.URL
	}

	return pathMap
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}

func redirectMiddleware(pathMap map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if dest, ok := pathMap[r.URL.Path]; ok {
			http.Redirect(w, r, dest, http.StatusFound)
			return
		}

		fallback.ServeHTTP(w, r)
	}
}
