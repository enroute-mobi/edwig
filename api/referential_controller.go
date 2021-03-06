package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"bitbucket.org/enroute-mobi/ara/core"
	"bitbucket.org/enroute-mobi/ara/logger"
	"bitbucket.org/enroute-mobi/ara/monitoring"
)

type ReferentialController struct {
	server *Server
}

func NewReferentialController(server *Server) ControllerInterface {
	return &Controller{
		restfulResource: &ReferentialController{
			server: server,
		},
	}
}

func (controller *ReferentialController) Action(response http.ResponseWriter, requestData *RequestData) {
	if requestData.Action == "reload" {
		controller.reload(requestData.Id, response)
		return
	}
	http.Error(response, fmt.Sprintf("Action not supported: %s", requestData.Action), http.StatusInternalServerError)
}

func (controller *ReferentialController) findReferential(identifier string) *core.Referential {
	referential := controller.server.CurrentReferentials().FindBySlug(core.ReferentialSlug(identifier))
	if referential != nil {
		return referential
	}
	return controller.server.CurrentReferentials().Find(core.ReferentialId(identifier))
}

func (controller *ReferentialController) Index(response http.ResponseWriter, filters url.Values) {
	logger.Log.Debugf("Referentials Index")

	jsonBytes, _ := json.Marshal(controller.server.CurrentReferentials().FindAll())
	response.Write(jsonBytes)
}

func (controller *ReferentialController) Show(response http.ResponseWriter, identifier string) {
	referential := controller.findReferential(identifier)
	if referential == nil {
		http.Error(response, fmt.Sprintf("Referential not found: %s", identifier), http.StatusNotFound)
		return
	}
	logger.Log.Debugf("Get referential %s", identifier)

	jsonBytes, _ := referential.MarshalJSON()
	response.Write(jsonBytes)
}

func (controller *ReferentialController) Delete(response http.ResponseWriter, identifier string) {
	referential := controller.findReferential(identifier)
	if referential == nil {
		http.Error(response, fmt.Sprintf("Referential not found: %s", identifier), http.StatusNotFound)
		return
	}
	logger.Log.Debugf("Delete referential %s", identifier)

	jsonBytes, _ := referential.MarshalJSON()
	referential.Stop()
	controller.server.CurrentReferentials().Delete(referential)
	response.Write(jsonBytes)
}

func (controller *ReferentialController) Update(response http.ResponseWriter, identifier string, body []byte) {
	referential := controller.findReferential(identifier)
	if referential == nil {
		http.Error(response, fmt.Sprintf("Referential not found: %s", identifier), http.StatusNotFound)
		return
	}

	logger.Log.Debugf("Update referential %s: %s", identifier, string(body))

	apiReferential := referential.Definition()
	err := json.Unmarshal(body, apiReferential)
	if err != nil {
		http.Error(response, fmt.Sprintf("Invalid request: can't parse request body: %v", err), http.StatusBadRequest)
		return
	}

	if !apiReferential.Validate() {
		jsonBytes, _ := json.Marshal(apiReferential)
		response.WriteHeader(http.StatusBadRequest)
		response.Write(jsonBytes)
		return
	}

	referential.Stop()
	referential.SetDefinition(apiReferential)
	referential.Save()
	referential.Start()

	jsonBytes, _ := referential.MarshalJSON()
	response.Write(jsonBytes)
}

func (controller *ReferentialController) Create(response http.ResponseWriter, body []byte) {
	logger.Log.Debugf("Create referential: %s", string(body))

	referential := controller.server.CurrentReferentials().New("")
	apiReferential := referential.Definition()
	err := json.Unmarshal(body, apiReferential)
	if err != nil {
		http.Error(response, fmt.Sprintf("Invalid request: can't parse request body: %v", err), http.StatusBadRequest)
		return
	}

	if !apiReferential.Validate() {
		jsonBytes, _ := json.Marshal(apiReferential)
		response.WriteHeader(http.StatusBadRequest)
		response.Write(jsonBytes)
		return
	}

	referential.SetDefinition(apiReferential)
	referential.Save()
	referential.Start()
	jsonBytes, _ := referential.MarshalJSON()
	response.Write(jsonBytes)
}

func (controller *ReferentialController) Save(response http.ResponseWriter) {
	logger.Log.Debugf("Saving referentials to database")

	status, err := controller.server.CurrentReferentials().SaveToDatabase()

	if err != nil {
		monitoring.ReportError(err)

		response.WriteHeader(status)
		jsonBytes, _ := json.Marshal(map[string]string{"error": err.Error()})
		response.Write(jsonBytes)
	}
}

func (controller *ReferentialController) reload(identifier string, response http.ResponseWriter) {
	referential := controller.findReferential(identifier)
	if referential == nil {
		http.Error(response, fmt.Sprintf("Referential not found: %s", identifier), http.StatusNotFound)
		return
	}

	logger.Log.Debugf("Reload referential %v from API request", referential.Slug())

	referential.ReloadModel()

	response.WriteHeader(http.StatusOK)
}
