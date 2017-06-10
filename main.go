package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"regexp"

	"encoding/json"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Shop represents a store front
type Shop struct {
	Name string
	ID   string
}

var staticHandler = http.FileServer(http.Dir("static"))

var session *mgo.Session

func main() {
	var err error
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/shop", newShopHandler)

	editTemplate, err = template.ParseFiles("templates/edit.html")
	if err != nil {
		log.Fatal(err)
	}

	errorTemplate, err = template.ParseFiles("templates/error.html")
	if err != nil {
		log.Fatal(err)
	}

	shopFrontRegex, err = regexp.Compile("/[2-9A-HJKMNP-Za-hjkmnp-z]{6}")
	if err != nil {
		log.Fatal(err)
	}

	session, err = mgo.Dial("localhost")
	if err != nil {
		log.Fatal("Error creating session:\n", err)
	}

	c := session.DB("test").C("shops")

	result := &Shop{}
	err = c.Find(bson.M{"id": "asdfa"}).One(result)

	if err != nil {
		log.Fatal("Error retrieving document:\n", err)
	}
	log.Print("Got ", result.ID, ": ", result.Name)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

var editTemplate *template.Template
var errorTemplate *template.Template

func editHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/edit/"):]
	thisShop := Shop{}
	err := session.DB("test").C("shops").Find(bson.M{"id": id}).One(&thisShop)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	editTemplate.Execute(w, thisShop)
}

var shopFrontRegex *regexp.Regexp

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if shopFrontRegex.MatchString(r.URL.Path) {
		w.Write([]byte("Under construction"))
	} else {
		staticHandler.ServeHTTP(w, r)
	}
}

type newShopBody struct {
	Name string
}

func newShopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(""))
		return
	}

	decoder := json.NewDecoder(r.Body)
	var body newShopBody

	err := decoder.Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error reading POST body:\n%s", err)
		w.Write([]byte(""))
		return
	}

	newShop := Shop{Name: body.Name, ID: ""}
	// loop until we add it
	log.Printf("Adding shop \"%s\"", newShop.Name)
	for {
		newShop.ID = getNewID()
		err = session.DB("test").C("shops").Insert(newShop)
		if err == nil {
			break
		} else {
			if mgo.IsDup(err) {
				continue
			}
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Error while creating new shop\"%s\":\n%s", newShop.Name, err)
			w.Write([]byte(""))
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bson.M{"newId": newShop.ID})
}

var idCharacters = []byte("abcdefghjkmnpqrstuvwxyz23456789")

func getNewID() string {
	var newID [5]byte
	for i := range newID {
		newID[i] = idCharacters[rand.Intn(len(idCharacters))]
	}
	return string(newID[:])
}
