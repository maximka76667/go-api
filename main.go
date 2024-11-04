package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"database/sql"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	DBHost  = "127.0.0.1"
	DBPort  = "5432"
	DBUser  = "postgres"
	DBPass  = "Max135135"
	DBDbase = "postgres"
	PORT    = ":8081"
)

// Define color constants
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

var database *sql.DB

type Message struct {
	Text string `json:"text"`
}

type Page struct {
	Title      string
	Content    template.HTML
	RawContent string
	Date       string
	GUID       string
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := Message{Text: "Hello, Go API!"}
	json.NewEncoder(w).Encode(response)
}

// func getUsers(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	response :=
// }

func ServeDynamic(w http.ResponseWriter, r *http.Request) {
	response := "The time is now " + time.Now().String()
	fmt.Fprintln(w, response)
}

func ServeStatic(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static.html")
}

func ServeError(w http.ResponseWriter, r *http.Request) {
	response := "No way it works! Error1"
	fmt.Fprintln(w, response)
}

func ServePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["id"]

	query := "SELECT page_title, page_content, page_date FROM pages WHERE page_guid = $1"

	thisPage := Page{}
	err := database.QueryRow(query, pageGUID).Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date)

	thisPage.Content = template.HTML(thisPage.RawContent)

	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println(Red + "Couldn't get page " + pageGUID + Reset)
		log.Println(Red + err.Error() + Reset)
		return
	}

	fmt.Println(White + "Page_guid " + Green + pageGUID + White + " accessed" + Reset)

	t, _ := template.ParseFiles("templates/blog.html")
	t.Execute(w, thisPage)
}

func RedirIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", http.StatusMovedPermanently)
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	var Pages = []Page{}

	query := "SELECT page_title, page_guid, page_content, page_date FROM pages ORDER BY $1 DESC"

	pages, err := database.Query(query, "page_date")
	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
	defer pages.Close()

	for pages.Next() {
		thisPage := Page{}
		pages.Scan(&thisPage.Title, &thisPage.GUID, &thisPage.RawContent, &thisPage.Date)
		thisPage.Content = template.HTML(thisPage.RawContent)
		Pages = append(Pages, thisPage)
	}

	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, Pages)
}

func (p Page) TruncatedText() template.HTML {
	if len(p.Content) > 150 {
		return p.Content[:150] + `...`
	}

	return p.Content
}

func main() {
	fmt.Println(Magenta + "Starting server..." + Reset)
	http.HandleFunc("/api/hello", helloHandler)

	// // Static and dynamic rotues
	http.HandleFunc("/api/static", ServeStatic)
	http.HandleFunc("/api/", ServeDynamic)
	http.HandleFunc("/api/error", ServeError)

	// Router
	routes := mux.NewRouter()
	routes.HandleFunc("/pages/{id:[0-9a-zA\\-]+}", ServePage)
	routes.HandleFunc("/", RedirIndex)
	routes.HandleFunc("/home", ServeIndex)
	http.Handle("/", routes)

	fmt.Println(Magenta + "Connecting to database..." + Reset)

	// Connect to database
	dbConn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		DBHost, DBPort, DBUser, DBPass, DBDbase)
	db, err := sql.Open("postgres", dbConn)
	if err != nil {
		log.Println(Red + "Couldn't connect to database" + Reset)
		log.Println(Red + err.Error() + Reset)
	}

	// Assign connected db to global database variable
	database = db
	fmt.Println(Cyan + "Successfully connected to database" + Reset)

	// Run server
	fmt.Println("Server is running on " + Green + "http://localhost" + PORT + Reset)
	log.Fatal(http.ListenAndServe(PORT, nil))
}
