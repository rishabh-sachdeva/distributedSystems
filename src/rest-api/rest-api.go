package main

import (
    "encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "fmt"

)

type Site struct {
	Name         string `json:"Name,omitempty"`
	Role         string `json:"Role,omitempty"`
	URI          string `json:"URI,omitempty"`
	Accesspoints []AccessPoint `json:"Accesspoints,omitempty"`
}

type AccessPoint struct{
	Label string;
	URL string;
}
	var apts []AccessPoint
	var siteArr []Site;

// our main function
func main() {
	apts = append(apts, AccessPoint{Label:"label-1",URL:"url-1"})
	apts = append(apts, AccessPoint{Label:"label-2",URL:"url-2"})
	
	siteArr = append(siteArr, Site{Name:"site-1",Role:"role-1",URI:"uri-1",Accesspoints:apts})
	siteArr = append(siteArr, Site{Name:"site-2",Role:"role-2",URI:"uri-2",Accesspoints:apts})

    router := mux.NewRouter()
    router.HandleFunc("/sites", GetSites).Methods("GET")
    router.HandleFunc("/sites/{name}", GetSite).Methods("GET")
    router.HandleFunc("/sites/{name}", CreateSite).Methods("POST")
    router.HandleFunc("/sites/{name}", DeleteSite).Methods("DELETE")
    router.HandleFunc("/sites/{name}", UpdateSite).Methods("PUT")


    log.Fatal(http.ListenAndServe(":8001", router))
}

//Return all the sites
func GetSites(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(siteArr)
}

//return single site of Name: {name}
func GetSite(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    for _, site := range siteArr {
    if site.Name == params["name"]{
    	json.NewEncoder(w).Encode(site)
        return
    }
}
}   
 // create a new site
func CreateSite(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    var site Site
    _ = json.NewDecoder(r.Body).Decode(&site)
    fmt.Println("post",site)

    site.Name = params["name"]
    siteArr = append(siteArr, site)
    json.NewEncoder(w).Encode(site)
}

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
func UpdateSite(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
     var updated_site Site
    _ = json.NewDecoder(r.Body).Decode(&updated_site)
    fmt.Println(&updated_site)
    for idx, site := range siteArr {
	    if site.Name == params["name"]{
	    	updated_site.Name = params["name"]
	    	siteArr[idx] =updated_site
	    	break
	    }
    }
     json.NewEncoder(w).Encode(siteArr)
}