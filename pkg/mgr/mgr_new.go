package mgr

type MgrNew struct {
	uiCommands  <-chan UiCommand
	connections <-chan IncomingConnection
}

type UiCommand struct {
}

type IncomingConnection struct {
}

type MutationRequest struct {
}

func (m *MgrNew) Start() {
	for {
		var mutationRequest MutationRequest

		select {
		case cmd := <-m.uiCommands:
			mutationRequest = m.uiCommandToMutationRequest(cmd)
		case conn := <-m.connections:
			mutationRequest = m.incomingConnectionToMutationRequest(conn)
		}
	}
}

func (m *MgrNew) uiCommandToMutationRequest(_ UiCommand) MutationRequest {
	return MutationRequest{}
}

func (m *MgrNew) incomingConnectionToMutationRequest(_ IncomingConnection) MutationRequest {
	return MutationRequest{}
}
