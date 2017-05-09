package api

import "net/http"

type StatusController struct{}

func NewStatusController(server *Server) ControllerInterface {
	return &StatusController{}
}

func (controller *StatusController) serve(response http.ResponseWriter, request *http.Request, requestData *RequestData) {
	response.Write([]byte(`{ "status": "ok" }`))
}
