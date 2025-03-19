package ui

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*.html
var templateFS embed.FS

func initTemplates() *template.Template {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(err)
	}
	return tmpl
}

func (ui *Ui) index(w http.ResponseWriter, tmpl *template.Template) {
	status := []struct {
		Status  string
		Address string
	}
	err := tmpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
