package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"regexp"

	"encoding/json"

	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Shop represents a store front
type Shop struct {
	Name        string
	ShopFrontID string
	DMID        string
}

var session *mgo.Session

func main() {
	var err error

	r := mux.NewRouter()
	r.HandleFunc("/{shopfrontID:[2-9A-HJKMNP-Za-hjkmnp-z]{6}}", shopFrontHandler).Methods("GET")
	r.HandleFunc("/edit/{shopID:[2-9A-HJKMNP-Za-hjkmnp-z]{8}}", editHandler).Methods("GET")
	r.HandleFunc("/shops", newShopHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	http.Handle("/", r)

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
	err = c.Find(bson.M{"dmid": "asdfasdf"}).One(result)

	if err != nil {
		log.Fatal("Error retrieving document:\n", err)
	}
	log.Print("Got ", result.DMID, ": ", result.Name)

	log.Fatal(http.ListenAndServe(":8081", nil))
}

var editTemplate *template.Template
var errorTemplate *template.Template

func editHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["shopID"]
	thisShop := Shop{}
	err := session.DB("test").C("shops").Find(bson.M{"dmid": id}).One(&thisShop)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	editTemplate.Execute(w, thisShop)
}

var shopFrontRegex *regexp.Regexp

func shopFrontHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Shopfront under construction"))
}

type newShopBody struct {
	Name string
}

func newShopHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var body newShopBody

	err := decoder.Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error reading POST body:\n%s", err)
		w.Write([]byte("Bad POST request body."))
		return
	}

	newShop := Shop{Name: body.Name, DMID: ""}
	// loop until we add it
	log.Printf("Adding shop \"%s\"", newShop.Name)
	for {
		newShop.DMID = getNewDMID()
		newShop.ShopFrontID = getNewShopfrontID()
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
	json.NewEncoder(w).Encode(bson.M{"newId": newShop.DMID})
}

var idCharacters = []byte("abcdefghjkmnpqrstuvwxyz23456789")

func getNewDMID() string {
	var newID [8]byte
	for i := range newID {
		newID[i] = idCharacters[rand.Intn(len(idCharacters))]
	}
	return string(newID[:])
}

func getNewShopfrontID() string {
	var newID [6]byte
	for i := range newID {
		newID[i] = idCharacters[rand.Intn(len(idCharacters))]
	}
	return string(newID[:])
}
