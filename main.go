package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

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
	r.HandleFunc("/shops/{shopfrontID:[2-9A-HJKMNP-Za-hjkmnp-z]{6}}/{playerName}", getPlayerView).Methods("GET")
	r.HandleFunc("/shops/{shopID:[2-9A-HJKMNP-Za-hjkmnp-z]{8}}", getDMView).Methods("GET")
	r.HandleFunc("/shops/{shopID:[2-9A-HJKMNP-Za-hjkmnp-z]{8}}", deleteShop).Methods("DELETE")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/"))) // Don't add handlers after this. The order is important. Add before.

	http.Handle("/", r)

	var err error
	session, err = mgo.Dial(dbAddress)
	if err != nil {
		log.Fatal("Error creating database session:\n", err)
	}

	c := session.DB(dbName).C("shops")

	result := &Shop{}
	err = c.Find(bson.M{"dmid": "asdfasdf"}).One(result)

	if err != nil {
		log.Fatal("Error retrieving asdfasdf during startup:\n", err)
	}
	log.Print("Setup complete. Starting server on port ", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "dynamic/edit.html")
}

func shopFrontHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "dynamic/shopfront.html")
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
		w.Write([]byte("Bad POST request body."))
		return
	}

	newShop := Shop{Name: body.Name}
	// loop until we add it
	log.Printf("Adding shop \"%s\" with DMID %s", newShop.Name, newShop.DMID)
	for {
		newShop.DMID = getNewDMID()
		newShop.ShopFrontID = getNewShopfrontID()
		err = session.DB(dbName).C("shops").Insert(newShop)
		if err == nil {
			// if no error, we're done
			break
		}
		if mgo.IsDup(err) {
			// if it was a duplicate ID, try again
			continue
		}

		// we had an error and it wasn't a duplicate ID, so something else is wrong
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error while creating new shop\"%s\":\n%s", newShop.Name, err)
		w.Write([]byte("Error creating shop"))
		return

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

func getPlayerView(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"status": "Under Construction"})
}

func getDMView(w http.ResponseWriter, r *http.Request) {
	id := strings.ToLower(mux.Vars(r)["shopID"])
	thisShop := Shop{}
	err := session.DB(dbName).C("shops").Find(bson.M{"dmid": id}).One(&thisShop)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thisShop)
}

func deleteShop(w http.ResponseWriter, r *http.Request) {
	id := strings.ToLower(mux.Vars(r)["shopID"])
	err := session.DB(dbName).C("shops").Remove(bson.M{"dmid": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/text")
	w.Write([]byte("Successfully deleted the shop"))
}
