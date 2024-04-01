package main

import (
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

// HTML Templates
type Templates struct {
	templates *template.Template
}

// Render renders a template
func (t *Templates) Render(w io.Writer, name string, data any) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// newTemplate creates an HTML template object
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
	// Router
	router := http.NewServeMux()

	// Middleware

	// Template
	template := newTemplate()

	// Serve static content
	fsi := http.FileServer(http.Dir("./images"))
	router.Handle("/images/", http.StripPrefix("/images/", fsi))
	fsc := http.FileServer(http.Dir("./css"))
	router.Handle("/css/", http.StripPrefix("/css/", fsc))

	// Components
	page := newPage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Routes
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		template.Render(w, "index", page)
	})
	router.HandleFunc("POST /contacts", contactsPostHandler(template, &page))
	router.HandleFunc("DELETE /contacts/{id}", contactsDeleteHandler(&page))

	// Server
	server := http.Server{
		Addr:    ":1337",
		Handler: logging(router, logger),
		// Use our logger to log errors from the HTTP server. We can do this because our
		// logger implements the io.Write interface. Any logs generated here will be shown
		// as an ERROR log.
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("Server is running on port :1337")
	err := server.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

// Handler: POST /contacts
func contactsPostHandler(template *Templates, page *Page) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		email := r.FormValue("email")

		if page.Data.hasEmail(email) {
			formData := newFormData()
			formData.Values["name"] = name
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists"

			// We need to write	the header before rendering the template
			// Otherwise, the status code will be 200 OK
			w.WriteHeader(http.StatusUnprocessableEntity)
			template.Render(w, "createcontact", formData)
			return
		}

		contact := newContact(name, email)
		page.Data.Contacts = append(page.Data.Contacts, contact)

		w.WriteHeader(http.StatusCreated)
		// Render a form
		template.Render(w, "createcontact", newFormData())
		// Render the "oob-contact" block (so we just send the contact that is created)
		// The less data the server sends, the better
		template.Render(w, "oob-contact", contact)
	}
}

// Handler: DELETE /contacts/:id
func contactsDeleteHandler(page *Page) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow server
		time.Sleep(1 * time.Second)
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Invalid ID"))
			return
		}

		index := page.Data.indexOf(id)
		if index == -1 {
			w.WriteHeader(404)
			w.Write([]byte("Contact not found"))
			return
		}

		// Delete the contact from the list
		// This is a simple way to delete an element from a slice
		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

		w.WriteHeader(http.StatusOK)
	}
}
