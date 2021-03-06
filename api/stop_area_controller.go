package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"bitbucket.org/enroute-mobi/ara/core"
	"bitbucket.org/enroute-mobi/ara/logger"
	"bitbucket.org/enroute-mobi/ara/model"
)

type StopAreaController struct {
	referential *core.Referential
}

func NewStopAreaController(referential *core.Referential) ControllerInterface {
	return &Controller{
		restfulResource: &StopAreaController{
			referential: referential,
		},
	}
}

func (controller *StopAreaController) findStopArea(tx *model.Transaction, identifier string) (model.StopArea, bool) {
	idRegexp := "([0-9a-zA-Z-]+):([0-9a-zA-Z-:]+)"
	pattern := regexp.MustCompile(idRegexp)
	foundStrings := pattern.FindStringSubmatch(identifier)
	if foundStrings != nil {
		objectid := model.NewObjectID(foundStrings[1], foundStrings[2])
		return tx.Model().StopAreas().FindByObjectId(objectid)
	}
	return tx.Model().StopAreas().Find(model.StopAreaId(identifier))
}

func (controller *StopAreaController) Index(response http.ResponseWriter, filters url.Values) {
	tx := controller.referential.NewTransaction()
	defer tx.Close()

	logger.Log.Debugf("StopAreas Index")

	stime := controller.referential.Clock().Now()
	sas := tx.Model().StopAreas().FindAll()
	logger.Log.Debugf("StopAreaController FindAll time : %v", controller.referential.Clock().Since(stime))
	stime = controller.referential.Clock().Now()
	jsonBytes, _ := json.Marshal(sas)
	logger.Log.Debugf("StopAreaController Json Marshal time : %v ", controller.referential.Clock().Since(stime))
	response.Write(jsonBytes)
}

func (controller *StopAreaController) Show(response http.ResponseWriter, identifier string) {
	tx := controller.referential.NewTransaction()
	defer tx.Close()

	stopArea, ok := controller.findStopArea(tx, identifier)
	if !ok {
		http.Error(response, fmt.Sprintf("Stop area not found: %s", identifier), http.StatusNotFound)
		return
	}
	logger.Log.Debugf("Get stopArea %s", identifier)

	jsonBytes, _ := stopArea.MarshalJSON()
	response.Write(jsonBytes)
}

func (controller *StopAreaController) Delete(response http.ResponseWriter, identifier string) {
	// New transaction
	tx := controller.referential.NewTransaction()
	defer tx.Close()

	stopArea, ok := controller.findStopArea(tx, identifier)
	if !ok {
		http.Error(response, fmt.Sprintf("Stop area not found: %s", identifier), http.StatusNotFound)
		return
	}
	logger.Log.Debugf("Delete stopArea %s", identifier)

	jsonBytes, _ := stopArea.MarshalJSON()
	tx.Model().StopAreas().Delete(&stopArea)
	err := tx.Commit()
	if err != nil {
		logger.Log.Debugf("Transaction error: %v", err)
		http.Error(response, "Internal error", http.StatusInternalServerError)
		return
	}
	response.Write(jsonBytes)
}

func (controller *StopAreaController) Update(response http.ResponseWriter, identifier string, body []byte) {
	// New transaction
	tx := controller.referential.NewTransaction()
	defer tx.Close()

	stopArea, ok := controller.findStopArea(tx, identifier)
	if !ok {
		http.Error(response, fmt.Sprintf("Stop area not found: %s", identifier), http.StatusNotFound)
		return
	}

	logger.Log.Debugf("Update stopArea %s: %s", identifier, string(body))

	err := json.Unmarshal(body, &stopArea)
	if err != nil {
		http.Error(response, fmt.Sprintf("Invalid request: can't parse request body: %v", err), http.StatusBadRequest)
		return
	}

	for _, obj := range stopArea.ObjectIDs() {
		sa, ok := tx.Model().StopAreas().FindByObjectId(obj)
		if ok && sa.Id() != stopArea.Id() {
			http.Error(response, fmt.Sprintf("Invalid request: stopArea %v already have an objectid %v", sa.Id(), obj.String()), http.StatusBadRequest)
			return
		}
	}

	tx.Model().StopAreas().Save(&stopArea)
	err = tx.Commit()
	if err != nil {
		logger.Log.Debugf("Transaction error: %v", err)
		http.Error(response, "Internal error", http.StatusInternalServerError)
		return
	}
	jsonBytes, _ := stopArea.MarshalJSON()
	response.Write(jsonBytes)
}

func (controller *StopAreaController) Create(response http.ResponseWriter, body []byte) {
	// New transaction
	tx := controller.referential.NewTransaction()
	defer tx.Close()

	logger.Log.Debugf("Create stopArea: %s", string(body))

	stopArea := tx.Model().StopAreas().New()

	err := json.Unmarshal(body, &stopArea)
	if err != nil {
		http.Error(response, fmt.Sprintf("Invalid request: can't parse request body: %v", err), http.StatusBadRequest)
		return
	}

	if stopArea.Id() != "" {
		http.Error(response, "Invalid request", http.StatusBadRequest)
		return
	}

	for _, obj := range stopArea.ObjectIDs() {
		sa, ok := tx.Model().StopAreas().FindByObjectId(obj)
		if ok {
			http.Error(response, fmt.Sprintf("Invalid request: stopArea %v already have an objectid %v", sa.Id(), obj.String()), http.StatusBadRequest)
			return
		}
	}

	tx.Model().StopAreas().Save(&stopArea)
	err = tx.Commit()
	if err != nil {
		logger.Log.Debugf("Transaction error: %v", err)
		http.Error(response, "Internal error", http.StatusInternalServerError)
		return
	}
	jsonBytes, _ := stopArea.MarshalJSON()
	response.Write(jsonBytes)
}
