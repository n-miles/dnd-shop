package main

import (
	"html/template"
	"log"
	"net/http"
	"regexp"
)

type Shop struct {
	Name string
}

var shops []*Shop

var staticHandler = http.FileServer(http.Dir("static"))

func main() {
	var err error
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/edit/", editHandler)

	editTemplate, err = template.ParseFiles("templates/edit.html")
	if err != nil {
		log.Fatal(err)
	}

	errorTemplate, err = template.ParseFiles("templates/error.html")
	if err != nil {
		log.Fatal(err)
	}

	shopFrontRegex, err = regexp.Compile("/[1-9A-HJKMNP-Za-hjkmnp-z]{6}")
	if err != nil {
		log.Fatal(err)
	}

	shops = make([]*Shop, 0, 0)
	shops = append(shops, &Shop{"asdfg"}, &Shop{"12345"})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

var editTemplate *template.Template
var errorTemplate *template.Template

func editHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/edit/"):]
	var thisShop *Shop
	for _, shop := range shops {
		if name == shop.Name {
			thisShop = shop
			break
		}
	}

	if thisShop == nil {
		errorTemplate.Execute(w, name)
		return
	}

	editTemplate.Execute(w, thisShop)
}

var shopFrontRegex *regexp.Regexp

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if !shopFrontRegex.MatchString(r.URL.Path) {
		staticHandler.ServeHTTP(w, r)
		return
	}

	w.Write([]byte("Under construction"))
}
