package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type Sites struct {
	Sites []Site `json:"Sites"`
}

type Site struct {
	Name         string        `json:"Name,omitempty"`github
	Role         string        `json:"Role,omitempty"`
	URI          string        `json:"URI,omitempty"`
	Accesspoints []AccessPoint `json:"Accesspoints,omitempty"`
}

type AccessPoint struct {
	Label string `json:"Label,omitempty"`
	URL   string `json:"URL,omitempty"`
}

var apts []AccessPoint
var siteArr []Site
var sites Sites

func main() {
	// apts = append(apts, AccessPoint{Label: "label-1", URL: "url-1"})
	// apts = append(apts, AccessPoint{Label: "label-2", URL: "url-2"})

	// siteArr = append(siteArr, Site{Name: "site-1", Role: "role-1", URI: "uri-1", Accesspoints: apts})
	// siteArr = append(siteArr, Site{Name: "site-2", Role: "role-2", URI: "uri-2", Accesspoints: apts})

	// Open our jsonFile
	jsonFile, err := os.Open("Sites.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &sites)
	siteArr = sites.Sites[:]

	router := mux.NewRouter()
	router.HandleFunc("/", projectIndex).Methods("GET")
	router.HandleFunc("/sites", GetSites).Methods("GET")
	router.HandleFunc("/sites/{name}", GetSite).Methods("GET")
	router.HandleFunc("/sites/{name}", CreateSite).Methods("POST")
	router.HandleFunc("/sites/{name}", DeleteSite).Methods("DELETE")
	router.HandleFunc("/sites/{name}", UpdateSite).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8001", router))
}

func projectIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Project 1 \n Group Submission \n Group members: \n\tAniruddha\n\tArshitha\n\tNipun Ramagiri\n\tRishab")
}

// GetSites - Returns all the Sites
func GetSites(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(siteArr)
}

// GetSite - Returns single Site of Name: {name}
func GetSite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, site := range siteArr {
		if site.Name == params["name"] {
			json.NewEncoder(w).Encode(site)
			return
		}
	}
}

// CreateSite - Creates a new Site [Original]
func CreateSite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fmt.Println("POST: ", r.Body)

	var site Site
	_ = json.NewDecoder(r.Body).Decode(&site)

	site.Name = params["name"]
	siteArr = append(siteArr, site)
	json.NewEncoder(w).Encode(site)
}

// DeleteSite -
func DeleteSite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for idx, site := range siteArr {
		if site.Name == params["name"] {
			siteArr = append(siteArr[:idx], siteArr[idx+1:]...)
			break
		}
		//json.NewEncoder(w).Encode(siteArr)
	}
}

// UpdateSite -
func UpdateSite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var updatedSite Site
	_ = json.NewDecoder(r.Body).Decode(&updatedSite)
	fmt.Println(&updatedSite)
	for idx, site := range siteArr {
		if site.Name == params["name"] {
			updatedSite.Name = params["name"]
			siteArr[idx] = updatedSite
			break
		}
	}
	json.NewEncoder(w).Encode(updatedSite)
}
