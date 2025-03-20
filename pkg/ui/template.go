package ui

import (
	"embed"
	"html/template"
	"net/http"
	"sort"
	"tealfs/pkg/model"
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
	err := tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ui *Ui) connectionStatus(w http.ResponseWriter, tmpl *template.Template) {

	status := []struct {
		Status  string
		Address string
	}{}
	for _, value := range ui.statuses {
		status = append(status, struct {
			Status  string
			Address string
		}{Status: statusToString(value.Type), Address: value.RemoteAddress})
	}
	sort.Slice(status, func(i, j int) bool {
		if status[i].Status == status[j].Status {
			return status[i].Address < status[j].Address
		}
		return status[i].Status < status[j].Status
	})
	err := tmpl.ExecuteTemplate(w, "connection-status.html", status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func statusToString(status model.ConnectedStatus) string {
	switch status {
	case model.Connected:
		return "Connected"
	case model.NotConnected:
		return "Disconnected"
	default:
		return "Unknown"
	}
}
