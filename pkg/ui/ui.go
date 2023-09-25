package ui

import (
	"tealfs/pkg/cmds"
	"fmt"
	"net/http"
	"strconv"
)

func StartUi(uiPort int, nodePort int, userCmds chan cmds.User, hostid uint32) {
	handleUserCommands(userCmds)
	handleRoot(nodePort, hostid)
	http.ListenAndServe(":"+fmt.Sprint(uiPort), nil)
}

func handleUserCommands(userCmds chan cmds.User) {
	http.HandleFunc("/connect-to", func(w http.ResponseWriter, r *http.Request) {
		hostAndPort := r.FormValue("hostandport")
		userCmds <- cmds.User{
			CmdType:  cmds.ConnectTo,
			Argument: hostAndPort,
		}
		fmt.Fprintf(w, "Connecting to: %s", hostAndPort)
	})
	http.HandleFunc("/add-storage", func(w http.ResponseWriter, r *http.Request) {
		newStorageLocation := r.FormValue("newStorageLocation")
		userCmds <- cmds.User{
			CmdType:  cmds.AddStorage,
			Argument: newStorageLocation,
		}
		fmt.Fprintf(w, "Adding storage location: %s", newStorageLocation)
	})
}

func handleRoot(nodePort int, hostId uint32) {
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
					<h1>TealFS: ` + strconv.Itoa(int(hostId)) + `</h1>
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
