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

type Site struct {
	Name         string        `json:"Name,omitempty"`
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

func main() {
	// Opening Persistent storage file
	jsonFile, err := os.Open("sites.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	b, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(b, &siteArr)

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
					<br>Aniruddha<br>Arshitha<br>Nipun Ramagiri<br>Rishab`)
}

// Site API Handlers
// CreateSite - Creates a new Site
func CreateSite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fmt.Println("POST: ", r.Body)
	var site Site
	_ = json.NewDecoder(r.Body).Decode(&site)

	if r.Method == "PUT" {
		var newSite Site
		newSite = site
		newSite.Name = params["name"]
		fmt.Println(&newSite)
		var isChanged bool
		for idx, site := range siteArr {
			if site.Name == params["name"] {
				newSite.Name = params["name"]
				newSite.Accesspoints = site.Accesspoints
				siteArr[idx] = newSite
				isChanged = true
				break
			}
		}
		if !isChanged {
			siteArr = append(siteArr, newSite)
		}
		fileWriter, _ := os.Create("sites.json")
		json.NewEncoder(fileWriter).Encode(siteArr)
		defer fileWriter.Close()
		json.NewEncoder(w).Encode(newSite)
		return
	}

	site.Accesspoints = nil
	siteArr = append(siteArr, site)
	fileWriter, _ := os.Create("sites.json")
	json.NewEncoder(fileWriter).Encode(siteArr)
	json.NewEncoder(w).Encode(site)
	defer fileWriter.Close()

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

// UpdateSite - Updates a single Site of Name: {name}
func UpdateSite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var updatedSite Site
	_ = json.NewDecoder(r.Body).Decode(&updatedSite)
	fmt.Println(&updatedSite)
	for idx, site := range siteArr {
		if site.Name == params["name"] {
			updatedSite.Name = params["name"]
			updatedSite.Accesspoints = site.Accesspoints
			siteArr[idx] = updatedSite
			break
		}
	}
	fileWriter, _ := os.Create("sites.json")
	json.NewEncoder(fileWriter).Encode(siteArr)
	defer fileWriter.Close()
	json.NewEncoder(w).Encode(updatedSite)
}

// DeleteSite - Deletes a single Site of Name: {name}
func DeleteSite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for idx, site := range siteArr {
		if site.Name == params["name"] {
			siteArr = append(siteArr[:idx], siteArr[idx+1:]...)
			break
		}
	}
	fileWriter, _ := os.Create("sites.json")
	json.NewEncoder(fileWriter).Encode(siteArr)
	defer fileWriter.Close()
}

// Access Point API Handlers
// CreateAP - Creates a new AP
func CreateAP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fmt.Println("POST: ", r.Body)
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
		return
	}

	if r.Method == "PUT" {
		var newAP AccessPoint
		newAP = AP
		newAP.Label = params["label"]
		fmt.Println(&newAP)
		var isChanged bool
		for idx, ap := range corrSite.Accesspoints {
			if ap.Label == params["label"] {
				corrSite.Accesspoints[idx] = newAP
				isChanged = true
				break
			}
		}
		if !isChanged {
			corrSite.Accesspoints = append(corrSite.Accesspoints, newAP)
		}

		siteArr[corrSiteidx] = corrSite
		fileWriter, _ := os.Create("sites.json")
		json.NewEncoder(fileWriter).Encode(siteArr)
		defer fileWriter.Close()
		json.NewEncoder(w).Encode(newAP)
		return
	}

	siteArr[corrSiteidx].Accesspoints = append(siteArr[corrSiteidx].Accesspoints, AP)
	fileWriter, _ := os.Create("sites.json")
	json.NewEncoder(fileWriter).Encode(siteArr)
	json.NewEncoder(w).Encode(corrSite)
	defer fileWriter.Close()

}

// GetAPs - Returns all the AP
func GetAPs(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, site := range siteArr {
		if site.Name == params["name"] {
			json.NewEncoder(w).Encode(site.Accesspoints)
			return
		}
	}
}

// GetAP - Returns single AP of Label: {label}
func GetAP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, site := range siteArr {
		if site.Name == params["name"] {
			for _, ap := range site.Accesspoints {
				if ap.Label == params["label"] {
					json.NewEncoder(w).Encode(ap)
					return
				}
			}
		}
	}
}

// UpdateAP - Updates a single AP of Label: {label}
func UpdateAP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fmt.Println("POST: ", r.Body)
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
		return
	}

	for idx, ap := range corrSite.Accesspoints {
		if ap.Label == params["label"] {
			corrSite.Accesspoints[idx] = AP
		}
	}

	siteArr[corrSiteidx] = corrSite
	fileWriter, _ := os.Create("sites.json")
	json.NewEncoder(fileWriter).Encode(siteArr)
	json.NewEncoder(w).Encode(corrSite)
	defer fileWriter.Close()
}

// DeleteAP - Deletes a single AP of Label: {label}
func DeleteAP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	fmt.Println("POST: ", r.Body)
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
		return
	}

	for idx, ap := range corrSite.Accesspoints {
		if ap.Label == params["label"] {
			corrSite.Accesspoints = append(corrSite.Accesspoints[:idx], corrSite.Accesspoints[idx+1:]...)
			break
		}
	}

	siteArr[corrSiteidx] = corrSite
	fileWriter, _ := os.Create("sites.json")
	json.NewEncoder(fileWriter).Encode(siteArr)
	json.NewEncoder(w).Encode(corrSite)
	defer fileWriter.Close()
}
