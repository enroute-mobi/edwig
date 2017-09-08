Feature: Support SIRI StopLinesDiscovery

  Background:
    Given a Referential "test" is created

@wip
  Scenario: 4397 - Handle a SIRI StopLinesDiscovery request
    Given a Partner "test" exists with connectors [siri-stop-lines-discovery-request-broadcaster] and the following settings:
      | local_credential     | test     |
      | remote_objectid_kind | internal |
      | local_url            | address  |
    And a Line exists with the following attributes:
      | Name      | Line 1                          |
      | ObjectIDs | "internal":"STIF:Line::C00272:" |
    And a Line exists with the following attributes:
      | Name      | Line 2                          |
      | ObjectIDs | "internal":"STIF:Line::C00273:" |
    And a Line exists with the following attributes:
      | Name      | Line 3                          |
      | ObjectIDs | "internal":"STIF:Line::C00274:" |
    When I send this SIRI request
      """
      <S:Envelope xmlns:S="http://schemas.xmlsoap.org/soap/envelope/"
            xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
        <S:Body>
          <ns7:StopLinesDiscovery xmlns:ns2="http://www.siri.org.uk/siri"
                           xmlns:ns3="http://www.ifopt.org.uk/acsb"
                           xmlns:ns4="http://www.ifopt.org.uk/ifopt"
                           xmlns:ns5="http://datex2.eu/schema/2_0RC1/2_0"
                           xmlns:ns6="http://scma/siri" xmlns:ns7="http://wsdl.siri.org.uk">
            <Request>
              <ns2:RequestTimestamp>2017-03-03T11:28:00.359Z</ns2:RequestTimestamp>
              <ns2:RequestorRef>test</ns2:RequestorRef>
              <ns2:MessageIdentifier>STIF:Message::2345Fsdfrg35df:LOC</ns2:MessageIdentifier>
            </Request>
            <RequestExtension />
          </ns7:StopLinesDiscovery>
        </S:Body>
        </S:Envelope>
        """
    Then I should receive this SIRI response
      """
      <?xml version="1.0" encoding="UTF-8"?>
      <S:Envelope xmlns:S="http://schemas.xmlsoap.org/soap/envelope/">
        <S:Body>
          <ns8:StopLinesDiscoveryResponse xmlns:ns8="http://wsdl.siri.org.uk" xmlns:ns3="http://www.siri.org.uk/siri" xmlns:ns4="http://www.ifopt.org.uk/acsb" xmlns:ns5="http://www.ifopt.org.uk/ifopt" xmlns:ns6="http://datex2.eu/schema/2_0RC1/2_0" xmlns:ns7="http://scma/siri" xmlns:ns9="http://wsdl.siri.org.uk/siri">
            <Answer version="2.0">
            <ns3:ResponseTimestamp>2017-01-01T12:00:00.000Z</ns3:ResponseTimestamp>
            <ns3:Address>address</ns3:Address>
            <ns3:ProducerRef>Edwig</ns3:ProducerRef>
            <ns3:RequestMessageRef>STIF:Message::2345Fsdfrg35df:LOC</ns3:RequestMessageRef>
            <ns3:ResponseMessageIdentifier>RATPDev:ResponseMessage::6ba7b814-9dad-11d1-8-00c04fd430c8:LOC</ns3:ResponseMessageIdentifier>
            <ns3:Status>true</ns3:Status>
              <ns3:AnnotatedLineStructure>
                <ns3:LineRef>STIF:Line::C00272:</ns3:LineRef>
                <ns3:LineName>Line 1</ns3:LineName>
              </ns3:AnnotatedLineStructure>
              <ns3:AnnotatedLineStructure>
                <ns3:LineRef>STIF:Line::C00273:</ns3:LineRef>
                <ns3:LineName>Line 2</ns3:LineName>
              </ns3:AnnotatedLineStructure>
              <ns3:AnnotatedLineStructure>
                <ns3:LineRef>STIF:Line::C00274:</ns3:LineRef>
                <ns3:LineName>Line 3</ns3:LineName>
              </ns3:AnnotatedLineStructure>
            </Answer>
            <AnswerExtension />
          </ns8:StopLinesDiscoveryResponse>
        </S:Body>
      </S:Envelope>
      """