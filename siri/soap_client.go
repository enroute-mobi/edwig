package siri

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/enroute-mobi/edwig/version"
	"github.com/jbowtie/gokogiri/xml"
	"golang.org/x/text/encoding/charmap"
)

type requestType int

const (
	DEFAULT requestType = iota
	SUBSCRIPTION
	NOTIFICATION
	CHECK_STATUS
)

type Request interface {
	BuildXML() (string, error)
}

type SOAPClient struct {
	SOAPClientUrls
}

type SOAPClientUrls struct {
	Url              string
	SubscriptionsUrl string
	NotificationsUrl string
}

type soapClientArguments struct {
	request          Request
	requestType      requestType
	expectedResponse string
	acceptGzip       bool
}

func NewSOAPClient(urls SOAPClientUrls) *SOAPClient {
	return &SOAPClient{SOAPClientUrls: urls}
}

func (client *SOAPClient) responseFromFormat(body io.Reader, contentType string) io.Reader {
	r, _ := regexp.Compile("^text/xml;charset=([ -~]+)")
	s := r.FindStringSubmatch(contentType)
	if len(s) == 0 {
		return body
	}
	if s[1] == "ISO-8859-1" {
		return charmap.ISO8859_1.NewDecoder().Reader(body)
	}
	return body
}

func (client *SOAPClient) prepareAndSendRequest(args soapClientArguments) (xml.Node, error) {
	// Wrap the request XML
	soapEnvelope := NewSOAPEnvelopeBuffer()
	xml, err := args.request.BuildXML()
	if err != nil {
		return nil, err
	}

	soapEnvelope.WriteXML(xml)

	// For tests
	// logger.Log.Debugf("%v", soapEnvelope.String())

	// Create http request
	httpRequest, err := http.NewRequest("POST", client.getURL(args.requestType), soapEnvelope)
	if err != nil {
		return nil, err
	}
	if args.acceptGzip {
		httpRequest.Header.Set("Accept-Encoding", "gzip, deflate")
	}
	httpRequest.Header.Set("Content-Type", "text/xml; charset=utf-8")
	httpRequest.Header.Set("User-Agent", version.ApplicationName())
	httpRequest.ContentLength = soapEnvelope.Length()

	// Send http request
	httpClient := &http.Client{Timeout: getTimeOut(args.requestType)}
	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Do nothing if request is a notification
	if args.requestType == NOTIFICATION {
		return nil, nil
	}

	// Check response status
	if response.StatusCode != http.StatusOK {
		return nil, NewSiriError(strings.Join([]string{"SIRI CRITICAL: HTTP status ", strconv.Itoa(response.StatusCode)}, ""))
	}

	if !strings.Contains(response.Header.Get("Content-Type"), "text/xml") {
		return nil, NewSiriError(fmt.Sprintf("SIRI CRITICAL: HTTP Content-Type %v", response.Header.Get("Content-Type")))
	}

	// Check if response is gzip
	var responseReader io.Reader
	if args.acceptGzip && response.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()
		responseReader = gzipReader
	} else {
		responseReader = client.responseFromFormat(response.Body, response.Header.Get("Content-Type"))
	}

	// Create SOAPEnvelope and check body type
	envelope, err := NewSOAPEnvelope(responseReader)
	if err != nil {
		return nil, err
	}
	if envelope.BodyType() != args.expectedResponse {
		return nil, NewSiriError(fmt.Sprintf("SIRI CRITICAL: Wrong Soap from server: %v", envelope.BodyType()))
	}
	return envelope.Body(), nil
}

func (client *SOAPClient) getURL(requestType requestType) string {
	switch requestType {
	case SUBSCRIPTION:
		if client.SubscriptionsUrl != "" {
			return client.SubscriptionsUrl
		}
	case NOTIFICATION:
		if client.NotificationsUrl != "" {
			return client.NotificationsUrl
		}
	}
	return client.Url
}

func getTimeOut(rt requestType) time.Duration {
	switch rt {
	case SUBSCRIPTION:
		return 30 * time.Second
	case CHECK_STATUS:
		return 9 * time.Second
	default:
		return 5 * time.Second
	}
}

func (client *SOAPClient) CheckStatus(request *SIRICheckStatusRequest) (*XMLCheckStatusResponse, error) {
	node, err := client.prepareAndSendRequest(soapClientArguments{
		request:          request,
		expectedResponse: "CheckStatusResponse",
		requestType:      CHECK_STATUS,
		acceptGzip:       true,
	})
	if err != nil {
		return nil, err
	}

	checkStatus := NewXMLCheckStatusResponse(node)
	return checkStatus, nil
}

func (client *SOAPClient) StopMonitoring(request *SIRIGetStopMonitoringRequest) (*XMLStopMonitoringResponse, error) {
	node, err := client.prepareAndSendRequest(soapClientArguments{
		request:          request,
		expectedResponse: "GetStopMonitoringResponse",
	})
	if err != nil {
		return nil, err
	}

	stopMonitoring := NewXMLStopMonitoringResponse(node)
	return stopMonitoring, nil
}

func (client *SOAPClient) SituationMonitoring(request *SIRIGetGeneralMessageRequest) (*XMLGeneralMessageResponse, error) {
	node, err := client.prepareAndSendRequest(soapClientArguments{
		request:          request,
		expectedResponse: "GetGeneralMessageResponse",
	})
	if err != nil {
		return nil, err
	}

	generalMessage := NewXMLGeneralMessageResponse(node)
	return generalMessage, nil
}

func (client *SOAPClient) StopMonitoringSubscription(request *SIRIStopMonitoringSubscriptionRequest) (*XMLSubscriptionResponse, error) {
	node, err := client.prepareAndSendRequest(soapClientArguments{
		request:          request,
		requestType:      SUBSCRIPTION,
		expectedResponse: "SubscribeResponse",
	})
	if err != nil {
		return nil, err
	}
	response := NewXMLSubscriptionResponse(node)
	return response, nil
}

func (client *SOAPClient) GeneralMessageSubscription(request *SIRIGeneralMessageSubscriptionRequest) (*XMLSubscriptionResponse, error) {
	node, err := client.prepareAndSendRequest(soapClientArguments{
		request:          request,
		requestType:      SUBSCRIPTION,
		expectedResponse: "SubscribeResponse",
	})
	if err != nil {
		return nil, err
	}
	response := NewXMLSubscriptionResponse(node)
	return response, nil
}

func (client *SOAPClient) DeleteSubscription(request *SIRIDeleteSubscriptionRequest) (*XMLDeleteSubscriptionResponse, error) {
	node, err := client.prepareAndSendRequest(soapClientArguments{
		request:          request,
		requestType:      SUBSCRIPTION,
		expectedResponse: "DeleteSubscriptionResponse",
	})
	if err != nil {
		return nil, err
	}

	terminatedSub := NewXMLDeleteSubscriptionResponse(node)
	return terminatedSub, nil
}

func (client *SOAPClient) NotifyStopMonitoring(request *SIRINotifyStopMonitoring) error {
	_, err := client.prepareAndSendRequest(soapClientArguments{
		request:     request,
		requestType: NOTIFICATION,
	})
	if err != nil {
		return err
	}
	return nil
}

func (client *SOAPClient) NotifyGeneralMessage(request *SIRINotifyGeneralMessage) error {
	_, err := client.prepareAndSendRequest(soapClientArguments{
		request:     request,
		requestType: NOTIFICATION,
	})
	if err != nil {
		return err
	}
	return nil
}

func (client *SOAPClient) NotifyEstimatedTimeTable(request *SIRINotifyEstimatedTimeTable) error {
	_, err := client.prepareAndSendRequest(soapClientArguments{
		request:     request,
		requestType: NOTIFICATION,
	})
	if err != nil {
		return err
	}
	return nil
}
