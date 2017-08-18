package core

import (
	"fmt"
	"strconv"

	"github.com/af83/edwig/audit"
	"github.com/af83/edwig/model"
	"github.com/af83/edwig/siri"
)

type EstimatedTimetableBroadcaster interface {
	RequestLine(request *siri.XMLGetEstimatedTimetable) *siri.SIRIEstimatedTimeTableResponse
}

type SIRIEstimatedTimetableBroadcaster struct {
	model.ClockConsumer

	siriConnector
}

type SIRIEstimatedTimetableBroadcasterFactory struct{}

func NewSIRIEstimatedTimetableBroadcaster(partner *Partner) *SIRIEstimatedTimetableBroadcaster {
	broadcaster := &SIRIEstimatedTimetableBroadcaster{}
	broadcaster.partner = partner
	return broadcaster
}

func (connector *SIRIEstimatedTimetableBroadcaster) remoteObjectIDKind() string {
	if connector.partner.Setting("siri-estimated-timetable-request-broadcaster.remote_objectid_kind") != "" {
		return connector.partner.Setting("siri-estimated-timetable-request-broadcaster.remote_objectid_kind")
	}
	return connector.partner.Setting("remote_objectid_kind")
}

func (connector *SIRIEstimatedTimetableBroadcaster) RequestLine(request *siri.XMLGetEstimatedTimetable) *siri.SIRIEstimatedTimeTableResponse {
	tx := connector.Partner().Referential().NewTransaction()
	defer tx.Close()

	logStashEvent := connector.newLogStashEvent()
	defer audit.CurrentLogStash().WriteEvent(logStashEvent)

	logXMLEstimatedTimetableRequest(logStashEvent, request)

	response := &siri.SIRIEstimatedTimeTableResponse{
		Address:                   connector.Partner().Setting("local_url"),
		ProducerRef:               connector.Partner().Setting("remote_credential"),
		ResponseMessageIdentifier: connector.SIRIPartner().IdentifierGenerator("response_message_identifier").NewMessageIdentifier(),
	}
	if response.ProducerRef == "" {
		response.ProducerRef = "Edwig"
	}

	response.SIRIEstimatedTimetableDelivery = connector.getEstimatedTimetableDelivery(tx, &request.XMLEstimatedTimetableRequest)

	logSIRIEstimatedTimetableResponse(logStashEvent, response)

	return response
}

func (connector *SIRIEstimatedTimetableBroadcaster) getEstimatedTimetableDelivery(tx *model.Transaction, request *siri.XMLEstimatedTimetableRequest) siri.SIRIEstimatedTimetableDelivery {
	currentTime := connector.Clock().Now()

	delivery := siri.SIRIEstimatedTimetableDelivery{
		RequestMessageRef: request.MessageIdentifier(),
		ResponseTimestamp: currentTime,
		Status:            true,
	}

	// SIRIEstimatedJourneyVersionFrame
	for _, lineId := range request.Lines() {
		lineObjectId := model.NewObjectID(connector.remoteObjectIDKind(), lineId)
		line, ok := tx.Model().Lines().FindByObjectId(lineObjectId)
		if !ok {
			continue
		}
		journeyFrame := &siri.SIRIEstimatedJourneyVersionFrame{
			RecordedAtTime: currentTime,
		}

		// SIRIEstimatedVehicleJourney
		for _, vehicleJourney := range tx.Model().VehicleJourneys().FindByLineId(line.Id()) {
			// Handle vehicleJourney Objectid
			vehicleJourneyId, ok := vehicleJourney.ObjectID(connector.remoteObjectIDKind())
			var datedVehicleJourneyRef string
			if ok {
				datedVehicleJourneyRef = vehicleJourneyId.Value()
			} else {
				defaultObjectID, ok := vehicleJourney.ObjectID("_default")
				if !ok {
					continue
				}
				referenceGenerator := connector.SIRIPartner().IdentifierGenerator("reference_identifier")
				datedVehicleJourneyRef = referenceGenerator.NewIdentifier(IdentifierAttributes{Type: "VehicleJourney", Default: defaultObjectID.HashValue()})
			}

			estimatedVehicleJourney := &siri.SIRIEstimatedVehicleJourney{
				LineRef:                lineObjectId.Value(),
				DatedVehicleJourneyRef: datedVehicleJourneyRef,
				Attributes:             make(map[string]string),
				References:             make(map[string]model.Reference),
			}
			estimatedVehicleJourney.References = connector.getEstimatedVehicleJourneyReferences(vehicleJourney, tx)
			estimatedVehicleJourney.Attributes = vehicleJourney.Attributes

			// SIRIEstimatedCall
			for _, stopVisit := range tx.Model().StopVisits().FindFollowingByVehicleJourneyId(vehicleJourney.Id()) {
				// Handle StopPointRef
				stopArea, ok := tx.Model().StopAreas().Find(stopVisit.StopAreaId)
				if !ok {
					continue
				}
				stopAreaId, ok := stopArea.ObjectID(connector.remoteObjectIDKind())
				if !ok {
					continue
				}

				estimatedCall := &siri.SIRIEstimatedCall{
					ArrivalStatus:         string(stopVisit.ArrivalStatus),
					DepartureStatus:       string(stopVisit.DepartureStatus),
					AimedArrivalTime:      stopVisit.Schedules.Schedule("aimed").ArrivalTime(),
					ExpectedArrivalTime:   stopVisit.Schedules.Schedule("expected").ArrivalTime(),
					AimedDepartureTime:    stopVisit.Schedules.Schedule("aimed").DepartureTime(),
					ExpectedDepartureTime: stopVisit.Schedules.Schedule("expected").DepartureTime(),
					Order:              stopVisit.PassageOrder,
					StopPointRef:       stopAreaId.Value(),
					StopPointName:      stopArea.Name,
					DestinationDisplay: stopVisit.Attributes["DestinationDisplay"],
					VehicleAtStop:      stopVisit.VehicleAtStop,
				}

				estimatedVehicleJourney.EstimatedCalls = append(estimatedVehicleJourney.EstimatedCalls, estimatedCall)
			}

			journeyFrame.EstimatedVehicleJourneys = append(journeyFrame.EstimatedVehicleJourneys, estimatedVehicleJourney)
		}
		delivery.EstimatedJourneyVersionFrames = append(delivery.EstimatedJourneyVersionFrames, journeyFrame)
	}

	return delivery
}

func (connector *SIRIEstimatedTimetableBroadcaster) getEstimatedVehicleJourneyReferences(vehicleJourney model.VehicleJourney, tx *model.Transaction) (references map[string]model.Reference) {
	for _, refType := range []string{"OriginRef", "DestinationRef"} {
		ref, ok := vehicleJourney.Reference(refType)
		if !ok || ref == (model.Reference{}) {
			continue
		}
		if foundStopArea, ok := tx.Model().StopAreas().Find(model.StopAreaId(ref.Id)); ok {
			obj, ok := foundStopArea.ObjectID(connector.remoteObjectIDKind())
			if ok {
				references[refType] = *model.NewReference(obj)
				continue
			}
		}
		generator := connector.SIRIPartner().IdentifierGenerator("reference_stop_area_identifier")
		defaultObjectID := model.NewObjectID(connector.remoteObjectIDKind(), generator.NewIdentifier(IdentifierAttributes{Default: ref.GetSha1()}))
		references[refType] = *model.NewReference(defaultObjectID)
	}
	return
}

func (connector *SIRIEstimatedTimetableBroadcaster) newLogStashEvent() audit.LogStashEvent {
	event := connector.partner.NewLogStashEvent()
	event["connector"] = "EstimatedTimetableRequestBroadcaster"
	return event
}

func (factory *SIRIEstimatedTimetableBroadcasterFactory) Validate(apiPartner *APIPartner) bool {
	ok := apiPartner.ValidatePresenceOfSetting("remote_objectid_kind")
	ok = ok && apiPartner.ValidatePresenceOfSetting("local_credential")
	return ok
}

func (factory *SIRIEstimatedTimetableBroadcasterFactory) CreateConnector(partner *Partner) Connector {
	return NewSIRIEstimatedTimetableBroadcaster(partner)
}

func logXMLEstimatedTimetableRequest(logStashEvent audit.LogStashEvent, request *siri.XMLGetEstimatedTimetable) {
	logStashEvent["messageIdentifier"] = request.MessageIdentifier()
	logStashEvent["requestorRef"] = request.RequestorRef()
	logStashEvent["requestTimestamp"] = request.RequestTimestamp().String()
	logStashEvent["requestXML"] = request.RawXML()
}

func logSIRIEstimatedTimetableResponse(logStashEvent audit.LogStashEvent, response *siri.SIRIEstimatedTimeTableResponse) {
	logStashEvent["address"] = response.Address
	logStashEvent["producerRef"] = response.ProducerRef
	logStashEvent["requestMessageRef"] = response.RequestMessageRef
	logStashEvent["responseMessageIdentifier"] = response.ResponseMessageIdentifier
	logStashEvent["responseTimestamp"] = response.ResponseTimestamp.String()
	logStashEvent["status"] = strconv.FormatBool(response.Status)
	xml, err := response.BuildXML()
	if err != nil {
		logStashEvent["responseXML"] = fmt.Sprintf("%v", err)
		return
	}
	logStashEvent["responseXML"] = xml
}
