// CMSC 621, Advanced Operating Systems. Spring 2019
// Project 1
// Group Submission
// Group Members: Aniruddha, Arshita, Nipun, Rishab

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
)

// Site struct
type Site struct {
	Name         string        `json:"Name,omitempty"`
	Role         string        `json:"Role"`
	URI          string        `json:"URI"`
	Accesspoints []AccessPoint `json:"Accesspoints"`
}

// AccessPoint struct
type AccessPoint struct {
	Label string `json:"Label"`
	URL   string `json:"URL"`
}

var siteArr []Site

// To maintain persistent file storage, Mutex is used.
var lock sync.Mutex

func readPersistentFile() {
	lock.Lock()
	defer lock.Unlock()

	jsonFile, err := os.Open("sites.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	b, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(b, &siteArr)
}

func updatePersistentFile() {
	lock.Lock()
	defer lock.Unlock()

	fileWriter, err := os.Create("sites.json")
	if err != nil {
		panic(err)
	}
	json.NewEncoder(fileWriter).Encode(siteArr)
	defer fileWriter.Close()
}

func main() {
	// Reading Persistent storage file
	readPersistentFile()

	router := mux.NewRouter()
	router.HandleFunc("/", projectIndex).Methods("GET")

	// Sites APIs
	// Create API
	router.HandleFunc("/sites/{name}", CreateSite).Methods("PUT")
	router.HandleFunc("/sites", CreateSite).Methods("POST")
	// Read API
	router.HandleFunc("/sites", GetSites).Methods("GET")
	router.HandleFunc("/sites/{name}", GetSite).Methods("GET")
	// Update API
	router.HandleFunc("/sites/{name}", UpdateSite).Methods("POST")
	// Delete API
	router.HandleFunc("/sites/{name}", DeleteSite).Methods("DELETE")

	// Access Points APIs
	// Create API
	router.HandleFunc("/sites/{name}/accesspoints/{label}", CreateAP).Methods("PUT")
	router.HandleFunc("/sites/{name}/accesspoints", CreateAP).Methods("POST")
	// Read API
	router.HandleFunc("/sites/{name}/accesspoints", GetAPs).Methods("GET")
	router.HandleFunc("/sites/{name}/accesspoints/{label}", GetAP).Methods("GET")
	// Update API
	router.HandleFunc("/sites/{name}/accesspoints/{label}", UpdateAP).Methods("POST")
	// Delete API
	router.HandleFunc("/sites/{name}/accesspoints/{label}", DeleteAP).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8001", router))
}

func projectIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<h1>Project 1</h1> 
					<h2>Group Submission</h2> <br>
					Group members: <br>
					<br>Aniruddha<br>Arshitha<br>Nipun<br>Rishab`)
}

// Site API Handlers

// CreateSite - Creates a new Site
func CreateSite(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	params := mux.Vars(r)
	var tempSite Site
	_ = json.NewDecoder(r.Body).Decode(&tempSite)

	if r.Method == "PUT" {
		var newSite Site
		newSite = tempSite
		newSite.Name = params["name"]
		// Checking if {name} already present
		for idx, site := range siteArr {
			if site.Name == params["name"] {
				newSite.Name = site.Name
				newSite.Accesspoints = site.Accesspoints
				siteArr[idx] = newSite
				fmt.Fprint(w, "Site - PUT Update: ")
				json.NewEncoder(w).Encode(newSite)
				updatePersistentFile()
				return
			}
		}

		newSite.Accesspoints = nil
		siteArr = append(siteArr, newSite)
		fmt.Fprint(w, "Site - PUT Create: ")
		updatePersistentFile()
		json.NewEncoder(w).Encode(newSite)
		return
	}

	if tempSite.Name == "" {
		fmt.Fprintln(w, "No Name in Body.")
		return
	}
	// Checking if Name in Body already present or not
	for _, site := range siteArr {
		if site.Name == tempSite.Name {
			fmt.Fprint(w, "Name already exists")
			return
		}
	}
	tempSite.Accesspoints = nil
	siteArr = append(siteArr, tempSite)
	updatePersistentFile()
	fmt.Fprint(w, "Site - POST Create: ")
	json.NewEncoder(w).Encode(tempSite)
	return
}

// GetSites - Returns all the Sites
func GetSites(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	json.NewEncoder(w).Encode(siteArr)
	return
}

// GetSite - Returns single Site of Name: {name}
func GetSite(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	params := mux.Vars(r)
	for _, site := range siteArr {
		if site.Name == params["name"] {
			json.NewEncoder(w).Encode(site)
			return
		}
	}

	fmt.Fprintln(w, "Name not found.")
	return
}

// UpdateSite - Updates a single Site of Name: {name}
func UpdateSite(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	params := mux.Vars(r)
	var updatedSite Site
	_ = json.NewDecoder(r.Body).Decode(&updatedSite)
	updatedSite.Accesspoints = nil
	for idx, site := range siteArr {
		if site.Name == params["name"] {
			updatedSite.Name = params["name"]
			updatedSite.Accesspoints = site.Accesspoints
			siteArr[idx] = updatedSite
			updatePersistentFile()
			fmt.Fprint(w, "Site - Updated as: ")
			json.NewEncoder(w).Encode(updatedSite)
			return
		}
	}

	fmt.Fprintln(w, "Name not found. No update done.")
	return
}

// DeleteSite - Deletes a single Site of Name: {name}
func DeleteSite(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	params := mux.Vars(r)
	for idx, site := range siteArr {
		if site.Name == params["name"] {
			siteArr = append(siteArr[:idx], siteArr[idx+1:]...)
			updatePersistentFile()
			fmt.Fprint(w, "Site Deleted: ")
			json.NewEncoder(w).Encode(site)
			return
		}
	}

	fmt.Fprintln(w, "Name not found. Nothing deleted.")
	return
}

// Access Point API Handlers

// CreateAP - Creates a new AP
func CreateAP(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	params := mux.Vars(r)
	var AP AccessPoint
	_ = json.NewDecoder(r.Body).Decode(&AP)

	// Finding the corresponding site
	var corrSite Site
	var corrSiteidx int
	for idx, site := range siteArr {
		if site.Name == params["name"] {
			corrSite = site
			corrSiteidx = idx
		}
	}

	// If Site absent, no insertion
	if len(corrSite.Name) == 0 {
		fmt.Fprint(w, "Name not found.")
		return
	}

	if r.Method == "PUT" {
		var newAP AccessPoint
		newAP = AP
		newAP.Label = params["label"]
		// Checking if {label} already present
		for idx, ap := range corrSite.Accesspoints {
			if ap.Label == params["label"] {
				corrSite.Accesspoints[idx] = newAP
				siteArr[corrSiteidx] = corrSite
				updatePersistentFile()
				fmt.Fprint(w, "AP - PUT Update: ")
				json.NewEncoder(w).Encode(siteArr[corrSiteidx])
				return
			}
		}

		corrSite.Accesspoints = append(corrSite.Accesspoints, newAP)
		siteArr[corrSiteidx] = corrSite
		updatePersistentFile()
		fmt.Fprint(w, "AP - PUT Create: ")
		json.NewEncoder(w).Encode(newAP)
		return
	}

	if AP.Label == "" {
		fmt.Fprintln(w, "No Label in Body.")
		return
	}
	// Checking if Label in Body already present or not
	for _, ap := range corrSite.Accesspoints {
		if ap.Label == AP.Label {
			fmt.Fprint(w, "Label already exists.")
			return
		}
	}
	siteArr[corrSiteidx].Accesspoints = append(siteArr[corrSiteidx].Accesspoints, AP)
	updatePersistentFile()
	fmt.Fprint(w, "AP - POST Create: ")
	json.NewEncoder(w).Encode(siteArr[corrSiteidx])
	return
}

// GetAPs - Returns all the AP
func GetAPs(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	params := mux.Vars(r)
	for _, site := range siteArr {
		if site.Name == params["name"] {
			json.NewEncoder(w).Encode(site.Accesspoints)
			return
		}
	}

	fmt.Fprintln(w, "Name not found.")
	return
}

// GetAP - Returns single AP of Label: {label}
func GetAP(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	params := mux.Vars(r)
	for _, site := range siteArr {
		if site.Name == params["name"] {
			for _, ap := range site.Accesspoints {
				if ap.Label == params["label"] {
					json.NewEncoder(w).Encode(ap)
					return
				}
			}
			fmt.Fprintln(w, "Label not found.")
			return
		}
	}
	fmt.Fprintln(w, "Name not found.")
	return
}

// UpdateAP - Updates a single AP of Label: {label}
func UpdateAP(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	params := mux.Vars(r)

	// Finding the corresponding site
	var corrSite Site
	var corrSiteidx int
	for idx, site := range siteArr {
		if site.Name == params["name"] {
			corrSite = site
			corrSiteidx = idx
			break
		}
	}
	// If Site absent, no update
	if len(corrSite.Name) == 0 {
		fmt.Fprintln(w, "Name not found. No update done.")
		return
	}

	var AP AccessPoint
	_ = json.NewDecoder(r.Body).Decode(&AP)
	AP.Label = params["label"]

	for idx, ap := range corrSite.Accesspoints {
		if ap.Label == params["label"] {
			corrSite.Accesspoints[idx] = AP
			siteArr[corrSiteidx] = corrSite
			updatePersistentFile()
			fmt.Fprint(w, "Site AP - Updated as: ")
			json.NewEncoder(w).Encode(corrSite)
			return
		}
	}

	fmt.Fprintln(w, "Label not found. No update done.")
	return
}

// DeleteAP - Deletes a single AP of Label: {label}
func DeleteAP(w http.ResponseWriter, r *http.Request) {
	readPersistentFile()
	params := mux.Vars(r)
	var AP AccessPoint
	_ = json.NewDecoder(r.Body).Decode(&AP)

	// Finding the corresponding site
	var corrSite Site
	var corrSiteidx int
	for idx, site := range siteArr {
		if site.Name == params["name"] {
			corrSite = site
			corrSiteidx = idx
			break
		}
	}

	// If Site absent, no insertion
	if len(corrSite.Name) == 0 {
		fmt.Fprintln(w, "Name not found. No deletion done.")
		return
	}

	for idx, ap := range corrSite.Accesspoints {
		if ap.Label == params["label"] {
			corrSite.Accesspoints = append(corrSite.Accesspoints[:idx], corrSite.Accesspoints[idx+1:]...)
			siteArr[corrSiteidx] = corrSite
			updatePersistentFile()
			fmt.Fprint(w, "AP Deleted: ")
			json.NewEncoder(w).Encode(ap)
			return
		}
	}

	fmt.Fprintln(w, "Label not found. Nothing deleted.")
	return
}
