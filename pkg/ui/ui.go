package ui

import (
	"fmt"
	"net/http"
	"tealfs/pkg/cmds"
	"tealfs/pkg/node"
)

type Ui struct {
	node     *node.Node
	userCmds chan cmds.User
}

func NewUi(node *node.Node, userCmds chan cmds.User) Ui {
	return Ui{node, userCmds}
}

func (ui Ui) Start() {
	ui.handleUserCommands()
	handleRoot(nodePort, hostid)
	http.ListenAndServe(":0", nil)
}

func (ui Ui) handleUserCommands() {
	http.HandleFunc("/connect-to", func(w http.ResponseWriter, r *http.Request) {
		hostAndPort := r.FormValue("hostandport")
		ui.userCmds <- cmds.User{
			CmdType:  cmds.ConnectTo,
			Argument: hostAndPort,
		}
		fmt.Fprintf(w, "Connecting to: %s", hostAndPort)
	})
	http.HandleFunc("/add-storage", func(w http.ResponseWriter, r *http.Request) {
		newStorageLocation := r.FormValue("newStorageLocation")
		ui.userCmds <- cmds.User{
			CmdType:  cmds.AddStorage,
			Argument: newStorageLocation,
		}
		fmt.Fprintf(w, "Adding storage location: %s", newStorageLocation)
	})
}

func (ui Ui) handleRoot() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
			<!DOCTYPE html>
			<html>
			<head>
				<title>TealFS</title>
				<link rel="stylesheet" href="https://unpkg.com/mvp.css@1.12/mvp.css" /> 
				<script src="https://unpkg.com/htmx.org@1.9.2"></script>
			</head>
			<body>
			    <main>
					<h1>TealFS: ` + ui.node.NodeId.String() + `</h1>
					` + htmlMyhost(nodePort) + `
					<p>Input the host and port of a node to add</p>
					<form hx-put="/connect-to">
						<label for="textbox">Host and port:</label>
						<input type="text" id="hostandport" name="hostandport">
						<input type="submit" value="Connect">
					</form>
					<form hx-put="/add-storage">
						<label for="textbox">Storage location</label>
						<input type="text" id="newStorageLocation" name="newStorageLocation">
						<input type="submit" value="Add">
					</form>
				</main>
			</body>
			</html>
		`

		// Write the HTML content to the response writer
		fmt.Fprintf(w, "%s", html)
	})
}

func htmlMyhost(port int) string {
	return `
		<div id="myhost">
			<h2>My host</h2>
			<p>Host: localhost:` + fmt.Sprint(port) + `</p>
		</div>`
}
