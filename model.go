package main

import (
	"database/sql"
	"fmt"
)

type Contact struct {
	EmailID string `json:"emailid, omitempty"`
	Name    string `json:"name,omitempty"`
	City    string `json:"city, omitempty"`
	State   string `json:"state, omitempty"`
}

// Get Contact details
func (c *Contact) GetContactDetail(db *sql.DB) error {
	return db.QueryRow("SELECT name, city, state FROM contact_details WHERE emailId = ? OR name = ?", c.EmailID, c.Name).Scan(&c.Name, &c.City, &c.State)
}

// Update contact details
func (c *Contact) updateContactDetails(db *sql.DB) error {
	statement := fmt.Sprintf("UPDATE contact_details SET name='%s', city='%s' WHERE emailId='%s'", c.Name, c.City, c.EmailID)
	_, err := db.Exec(statement)
	return err
}

// Delete contact
func (c *Contact) deleteContact(db *sql.DB) error {
	statement := fmt.Sprintf("DELETE FROM contact_details WHERE emailId='%s'", c.EmailID)
	_, err := db.Exec(statement)
	return err
}

// Create contact
func (c *Contact) createContact(db *sql.DB) error {
	statement := fmt.Sprintf("INSERT INTO contact_details(emailId, name, city, state) VALUES('%s', '%s','%s', '%s')", c.EmailID, c.Name, c.City, c.State)
	//_, err := db.Exec("INSERT INTO contact_details (emailId, name) VALUES($1, $2)", c.EmailID, c.Name)
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&c.EmailID)
	if err != nil {
		return err
	}
	return nil
}

// List contacts based on the page limit
func getContactsList(db *sql.DB, start, count int) ([]Contact, error) {
	statement := fmt.Sprintf("SELECT emailId, name, city,state FROM contact_details LIMIT %d OFFSET %d", count, start)
	rows, err := db.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	contacts := []Contact{}
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.EmailID, &c.Name, &c.City, &c.State); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}
