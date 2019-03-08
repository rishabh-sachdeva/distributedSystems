package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Site struct {
	Name         string
	Role         string
	URI          string
	AccessPoints *AccessPoints
}

type AccessPoints struct {
	Label string
	URL   string
}

var sites []Site //instatiating struct Person

// display all from sites var
func GetSites(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(people)
}

//display a single data
func GetSite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, item := range sites {
		if item.Name == params["Name"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Site{})
}

func main() {
	router := mux.NewRouter()
	sites = append(sites, Site{Name: "A", Role: "C", URI: "Doe", AccessPoints: &AccessPoints{Label: "City X", URL: "State X"}})
	sites = append(people, Site{Name: "B", Role: "D", URI: "Doe", AccessPoints: &AccessPoints{Label: "City Z", URL: "State Y"}})
	router.HandleFunc("/sites", GetSites).Methods("GET")
	router.HandleFunc("/sites/{Name}", GetSite).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}
