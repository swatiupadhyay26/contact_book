package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a App

const tableCreationQuery = `
CREATE TABLE IF NOT EXISTS contact_details
(
    emailID VARCHAR(50) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
	city VARCHAR(50) ,
	state VARCHAR(50) 
)`

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize("root", "2627", "Contacts")
	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

// Below method is to test the Empty Table
func TestEmptyTable(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/contacts", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentUser(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/contacts/swati@gmail.com", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "contact not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'contact not found'. Got '%s'", m["error"])
	}
}

// Create Contact
func TestCreateContact(t *testing.T) {
	clearTable()
	payload := []byte(`{"emailid":"swati2@gmail.com","name":"swati2","city":"BLR2","state":"KA2"}`)
	req, _ := http.NewRequest("POST", "/contacts", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)
}

// Get Contact details
func TestGetContactDetail(t *testing.T) {
	clearTable()
	addContact(3)
	req, _ := http.NewRequest("GET", "/contacts/User3@gmail.com", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

// Delete Contact
func TestDeleteContact(t *testing.T) {
	clearTable()
	addContact(1)
	req, _ := http.NewRequest("GET", "/contacts/User1@gmail.com", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	req, _ = http.NewRequest("DELETE", "/contacts/User1@gmail.com", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	req, _ = http.NewRequest("GET", "/contacts/User1@gmail.com", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

// Update Contact
func TestUpdateContact(t *testing.T) {
	clearTable()
	addContact(3)
	req, _ := http.NewRequest("GET", "/contacts/User1@gmail.com", nil)
	response := executeRequest(req)
	var originalUser map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalUser)
	payload := []byte(`{"name":"UpdUser1","city":"UpdCity1"}`)
	req, _ = http.NewRequest("PUT", "/contacts/User1@gmail.com", bytes.NewBuffer(payload))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] == originalUser["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalUser["name"], m["name"], m["name"])
	}
	if m["city"] == originalUser["city"] {
		t.Errorf("Expected the city to change from '%v' to '%v'. Got '%v'", originalUser["city"], m["city"], m["city"])
	}
}

// List the contacts based on limit(pagination)
func TestGetContactList(t *testing.T) {
	clearTable()
	addContact(11)
	req, _ := http.NewRequest("GET", "/contacts", nil)
	response := executeRequest(req)

	if 10 != response.Code {
		t.Errorf("Expected count of rows %d. retrieved %d\n", 10, response.Code)
	}

}

func TestContactByNameOrEmail(t *testing.T) {
	clearTable()
	addContact(3)
	payload := []byte(`{"emailid":"swati5@gmail.com","name":"swati2"}`)
	req, _ := http.NewRequest("POST", "/contacts/getContactByNameOrEmail", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func addContact(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		statement := fmt.Sprintf("INSERT INTO contact_details(emailId, name, city, state) VALUES('%s', '%s','%s', '%s')", ("User" + strconv.Itoa(i+1) + "@gmail.com"), ("Swati" + strconv.Itoa(i+1)), ("City" + strconv.Itoa(i+1)), ("State" + strconv.Itoa(i+1)))
		a.DB.Exec(statement)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}
func clearTable() {
	a.DB.Exec("DELETE FROM contact_details")
}
