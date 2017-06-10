package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"

	"encoding/json"

	"os"

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

var dbName string

func main() {

	if len(os.Args) != 4 {
		log.Fatal("Usage: go run main.go [port] [db address] [db name]")
	}

	port := os.Args[1]
	dbAddress := os.Args[2]
	dbName = os.Args[3]

	r := mux.NewRouter()
	r.HandleFunc("/{shopfrontID:[2-9A-HJKMNP-Za-hjkmnp-z]{6}}", shopFrontHandler).Methods("GET")
	r.HandleFunc("/edit/{shopID:[2-9A-HJKMNP-Za-hjkmnp-z]{8}}", editHandler).Methods("GET")
	r.HandleFunc("/shops", newShopHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	http.Handle("/", r)

	var err error
	editTemplate, err = template.ParseFiles("templates/edit.html")
	if err != nil {
		log.Fatal(err)
	}

	errorTemplate, err = template.ParseFiles("templates/error.html")
	if err != nil {
		log.Fatal(err)
	}

	session, err = mgo.Dial(dbAddress)
	if err != nil {
		log.Fatal("Error creating databse session:\n", err)
	}

	c := session.DB(dbName).C("shops")

	result := &Shop{}
	err = c.Find(bson.M{"dmid": "asdfasdf"}).One(result)

	if err != nil {
		log.Fatal("Error retrieving document:\n", err)
	}
	log.Print("Got ", result.DMID, ": ", result.Name)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

var editTemplate *template.Template
var errorTemplate *template.Template

func editHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["shopID"]
	thisShop := Shop{}
	err := session.DB(dbName).C("shops").Find(bson.M{"dmid": id}).One(&thisShop)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	editTemplate.Execute(w, thisShop)
}

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
		err = session.DB(dbName).C("shops").Insert(newShop)
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
