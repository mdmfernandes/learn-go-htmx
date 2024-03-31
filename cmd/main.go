package main

import (
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// HTML Templates
type Templates struct {
	templates *template.Template
}

// Render renders a template
func (t *Templates) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

// Contact
var id = 0

type Contact struct {
	Id    int
	Name  string
	Email string
}

func newContact(name, email string) Contact {
	id++
	return Contact{
		Id:    id,
		Name:  name,
		Email: email,
	}
}

// Contacts is a list of Contacts
type Contacts []Contact

// Data to put on the page
type Data struct {
	Contacts Contacts
}

func newData() Data {
	return Data{
		Contacts: Contacts{
			newContact("Alice", "alice@example.com"),
			newContact("Bob", "bob@example.com"),
		},
	}
}

// hasEmail returns true if there is a contact with the provided email
func (d *Data) hasEmail(email string) bool {
	for _, c := range d.Contacts {
		if c.Email == email {
			return true
		}
	}
	return false
}

// indexOf returns the index of the contact (in the Contacts list) with the provided ID
func (d *Data) indexOf(id int) int {
	for i, c := range d.Contacts {
		if c.Id == id {
			return i
		}
	}
	return -1
}

// FormData is the data of a HTML form
type FormData struct {
	Values map[string]string
	Errors map[string]string
}

// newFormData creates a new FormData
func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

// Page is an HTML page
type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())

	// Renderer
	e.Renderer = newTemplate()

	// Serve static pages
	e.Static("/images", "images")
	e.Static("/css", "css")

	page := newPage()

	// Routes
	e.GET("/", func(c echo.Context) error {
		// Render the "index" block
		return c.Render(200, "index", page)
	})
	e.POST("/contacts", contactsPostHandler(page))
	e.DELETE("/contacts/:id", contactsDeleteHandler(page))

	// Start server
	e.Logger.Fatal(e.Start(":1337"))
}

// Handler: POST /contacts
func contactsPostHandler(page Page) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"
			return c.Render(422, "createcontact", formData)
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)

		// Render a form
		c.Render(200, "createcontact", newFormData())
		// Render the "oob-contact" block (so we just send the contact that is created)
		// The less data the server sends, the better
		return c.Render(200, "oob-contact", contact)
	}
}

// Handler: DELETE /contacts/:id
func contactsDeleteHandler(page Page) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Simulate a slow server
		time.Sleep(1 * time.Second)
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(400, "Invalid ID")
		}

		index := page.Data.indexOf(id)
		if index == -1 {
			return c.String(404, "Contact not found")
		}

		// Delete the contact from the list
		// This is a simple way to delete an element from a slice
		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

		return c.NoContent(200)
	}
}
