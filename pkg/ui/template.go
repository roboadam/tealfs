// Copyright (C) 2026 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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

func (ui *Ui) connectToGet(w http.ResponseWriter, tmpl *template.Template) {
	err := tmpl.ExecuteTemplate(w, "connect-to.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ui *Ui) addDiskGet(
	w http.ResponseWriter,
	tmpl *template.Template,
	local model.NodeId,
) {
	nodeData := []struct {
		Id   string
		Name string
	}{}
	nodeData = append(nodeData, struct {
		Id   string
		Name string
	}{Id: string(local), Name: "local"})
	remotes := ui.NodeConnMap.Nodes()
	for _, remote := range remotes.GetValues() {
		nodeData = append(nodeData, struct {
			Id   string
			Name string
		}{Id: string(remote), Name: string(remote)})
	}
	err := tmpl.ExecuteTemplate(w, "add-disk.html", nodeData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (ui *Ui) connectionStatus(w http.ResponseWriter, tmpl *template.Template) {
	ui.sMux.Lock()
	defer ui.sMux.Unlock()
	status := []struct {
		Status  string
		Address string
	}{}
	connections := ui.NodeConnMap.ConnectionTriples()
	for _, value := range connections {
		status = append(status, struct {
			Status  string
			Address string
		}{Status: statusToString(value), Address: value.Address})
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

func (ui *Ui) diskStatus(w http.ResponseWriter, tmpl *template.Template) {

	local := []struct {
		Status string
		Path   string
	}{}
	remote := []struct {
		Status string
		Path   string
		Node   string
	}{}
	for _, value := range ui.diskStatuses {
		status := availablenessToString(value.Availableness)
		switch value.Localness {
		case model.Local:
			val := struct {
				Status string
				Path   string
			}{Status: status, Path: value.Path}
			local = append(local, val)
		case model.Remote:
			val := struct {
				Status string
				Path   string
				Node   string
			}{Status: status, Path: value.Path, Node: string(value.Node)}
			remote = append(remote, val)
		default:
			panic("Unknown Localness")
		}
	}
	sort.Slice(local, func(i, j int) bool {
		if local[i].Status == local[j].Status {
			return local[i].Path < local[j].Path
		}
		return local[i].Status < local[j].Status
	})
	sort.Slice(remote, func(i, j int) bool {
		if remote[i].Status == remote[j].Status {
			if remote[i].Node == remote[j].Node {
				return remote[i].Path < remote[j].Path
			} else {
				return remote[i].Node < remote[j].Node
			}
		}
		return local[i].Status < local[j].Status
	})
	data := struct {
		Local  any
		Remote any
	}{Local: local, Remote: remote}
	err := tmpl.ExecuteTemplate(w, "disk-status.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func statusToString(status model.ConnectionTriple) string {
	if status.IsConnected {
		return "Connected"
	}
	return "Disconnected"
}

func availablenessToString(status model.DiskAvailableness) string {
	switch status {
	case model.Available:
		return "Available"
	case model.Unavailable:
		return "Unavailable"
	default:
		panic("Unknown availableness")
	}
}
