package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	hashids "github.com/speps/go-hashids"
)

// URL holds all data for a URL
type URL struct {
	gorm.Model
	Code     string `gorm:"unique",json:"code,omitempty"`
	ShortURL string `json:"shortURL,omitempty"`
	LongURL  string `json:"longURL,omitempty"`
}

func main() {
	db, _ := gorm.Open("sqlite3", "./db/shortner.db")
	defer db.Close()

	db.AutoMigrate(&URL{})
	db.Model(&URL{}).AddIndex("idx_url_code", "code")

	router := mux.NewRouter()
	router.HandleFunc("/create", CreateEndpoint).Methods("POST")
	//router.HandleFunc("/expand", ExpandEndpoint).Methods("GET")
	router.HandleFunc("/{id}", RedirectEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(":1337", router))
}

// CreateEndpoint handles creating and storing a new URL
func CreateEndpoint(w http.ResponseWriter, req *http.Request) {
	db, _ := gorm.Open("sqlite3", "./db/shortner.db")
	defer db.Close()

	var resURL URL
	err := json.NewDecoder(req.Body).Decode(&resURL)
	if err != nil {
		panic(err)
	}

	var newURL URL
	db.Where(&URL{LongURL: resURL.LongURL}).First(&newURL)

	if newURL == (URL{}) {
		hd := hashids.NewData()
		h := hashids.NewWithData(hd)
		now := time.Now()
		newURL.Code, _ = h.Encode([]int{int(now.Unix())})
		newURL.ShortURL = "http://localhost:1337/" + newURL.Code
		newURL.LongURL = resURL.LongURL
		db.Create(&newURL)
	}

	json.NewEncoder(w).Encode(newURL)
}

// RedirectEndpoint handles redirecting to long URL by ID
func RedirectEndpoint(w http.ResponseWriter, req *http.Request) {
	db, _ := gorm.Open("sqlite3", "./db/shortner.db")
	defer db.Close()

	params := mux.Vars(req)
	var url URL
	db.Where(&URL{Code: params["id"]}).First(&url)
	http.Redirect(w, req, url.LongURL, 301)
}
