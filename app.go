package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {

	connectionString := fmt.Sprintf("%s:%s@/%s", user, password, dbname)
	var err error
	a.DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/contacts", a.GetContactList).Methods("GET")
	a.Router.HandleFunc("/contacts/{emailid}", a.GetContactDetail).Methods("GET")
	a.Router.HandleFunc("/contacts", a.createContact).Methods("POST")
	a.Router.HandleFunc("/contacts/getContactByNameOrEmail", a.getContactByNameOrEmail).Methods("POST")
	a.Router.HandleFunc("/contacts/{emailid}", a.deleteContact).Methods("DELETE")
	a.Router.HandleFunc("/contacts/{emailid}", a.updateContact).Methods("PUT")

}

// Get Contact details
func (a *App) GetContactDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailid := vars["emailid"]
	name := vars["name"]

	c := Contact{EmailID: emailid, Name: name}

	if err := c.GetContactDetail(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "contact not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, c)
}

// Create contact
func (a *App) createContact(w http.ResponseWriter, r *http.Request) {
	var c Contact
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := c.createContact(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, c)
}

func (a *App) getContactByNameOrEmail(w http.ResponseWriter, r *http.Request) {
	var c Contact
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := c.GetContactDetail(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, c)
}

// Delete contact
func (a *App) deleteContact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailid := vars["emailid"]
	/*if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}*/

	c := Contact{EmailID: emailid}
	if err := c.deleteContact(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// Update contact details
func (a *App) updateContact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailid := vars["emailid"]
	/*if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}*/

	var c Contact
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&c); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	c.EmailID = emailid

	if err := c.updateContactDetails(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, c)
}

// List contacts based on the page limit
func (a *App) GetContactList(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	contacts, err := getContactsList(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(contacts) == count {
		respondWithJSON(w, len(contacts), contacts)
	}
	respondWithJSON(w, http.StatusOK, contacts)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
