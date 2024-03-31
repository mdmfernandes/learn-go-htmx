package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// HTML Templates
type Templates struct {
	templates *template.Template
}

// Render renders a template
func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Count struct {
	Count int
}

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())

	// Renderer
	e.Renderer = newTemplate()

	// Start counter
	count := Count{0}

	// Routes
	e.GET("/", rootHandler(count))
	e.POST("/count", countHandler(count))

	// Start server
	e.Logger.Fatal(e.Start(":1337"))
}

// Handler: root
func rootHandler(count Count) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Render the "index" block
		return c.Render(200, "index", count)
	}
}

// Handler: count
func countHandler(count Count) echo.HandlerFunc {
	return func(c echo.Context) error {
		count.Count++
		// Render the "count" block. The less amount of data you pass to the template, the better (faster).
		return c.Render(200, "count", count)
	}
}
