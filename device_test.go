package vitotrol

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testDeviceId   = 1234
	testLocationId = 5678
	testTimeStr    = "2016-10-30 12:13:14"
)

var (
	_ = []HasResultHeader{
		(*GetDataResponse)(nil),
		(*WriteDataResponse)(nil),
		(*RefreshDataResponse)(nil),
		(*GetErrorHistoryResponse)(nil),
		(*GetTimesheetDataResponse)(nil),
	}

	testTime = func() Time {
		tm, _ := ParseVitotrolTime(testTimeStr)
		return tm
	}()
)

func TestFormatAttributes(t *testing.T) {
	assert := assert.New(t)

	pDevice := &Device{
		Attributes: map[AttrId]*Value{
			NoAttr: {
				Value: "unknown-attr",
				Time:  testTime,
			},
			BurnerState: {
				Value: "invalid-value",
				Time:  testTime,
			},
			IndoorTemp: {
				Value: "22",
				Time:  testTime,
			},
			OutdoorTemp: nil,
		},
	}

	assert.Equal(
		fmt.Sprintf("%d: unknown-attr@%s\n", NoAttr, testTime)+
			fmt.Sprintf("BurnerState: unknown-value<invalid-value>@%s (%s)\n",
				testTime, AttributesRef[BurnerState].Doc)+
			fmt.Sprintf("IndoorTemp: 22@%s (%s)\n",
				testTime, AttributesRef[IndoorTemp].Doc)+
			fmt.Sprintf("OutdoorTemp: uninitialized (%s)\n",
				AttributesRef[OutdoorTemp].Doc),
		pDevice.FormatAttributes(
			[]AttrId{NoAttr, BurnerState, IndoorTemp, OutdoorTemp}))
}

func TestMakeDatenpunktIds(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(`<DatenpunktIds><int>11</int><int>22</int></DatenpunktIds>`,
		makeDatenpunktIds([]AttrId{11, 22}))
}

type requestDeviceCommon struct {
	DeviceId   uint32 `xml:"GeraetId"`
	LocationId uint32 `xml:"AnlageId"`
}

var deviceCommon = requestDeviceCommon{
	DeviceId:   testDeviceId,
	LocationId: testLocationId,
}

func intoDeviceResponse(action, content string) string {
	return fmt.Sprintf(
		`<%[1]sResponse xmlns="http://www.e-controlnet.de/services/vii/">
<%[1]sResult>
%[2]s
</%[1]sResult>
</%[1]sResponse>`, action, content)
}

func testSendRequestDeviceAny(assert *assert.Assertions,
	sendReq func(v *Session, d *Device) bool, soapAction string,
	expectedRequest interface{}, serverResponse string,
	testName string) bool {
	return testSendRequestAny(assert,
		func(v *Session) bool {
			v.Devices = []Device{
				{
					DeviceId:   testDeviceId,
					LocationId: testLocationId,
					Attributes: map[AttrId]*Value{},
					Timesheets: map[TimesheetId]map[string]TimeslotSlice{},
				},
			}
			return sendReq(v, &v.Devices[0])
		},
		soapAction, expectedRequest,
		intoDeviceResponse(soapAction, serverResponse),
		testName)
}

//
// GetData
//

func TestGetData(t *testing.T) {
	assert := assert.New(t)

	type requestGetData struct {
		requestDeviceCommon
		Ids []int `xml:"DatenpunktIds>int"`
	}

	type requestBody struct {
		GetData requestGetData `xml:"Body>GetData"`
	}

	expectedRequest := &requestBody{
		GetData: requestGetData{
			requestDeviceCommon: deviceCommon,
			Ids:                 []int{11, 22},
		},
	}

	// No problem
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			err := d.GetData(v, []AttrId{11, 22})
			if !assert.Nil(err) {
				return false
			}
			return assert.Equal(map[AttrId]*Value{
				11: {
					Value: "value11",
					Time:  testTime,
				},
				22: {
					Value: "value22",
					Time:  testTime,
				},
			}, d.Attributes)
		},
		// SOAP action
		"GetData",
		expectedRequest,
		// Response to reply
		`<Ergebnis>0</Ergebnis>
<ErgebnisText>Kein Fehler</ErgebnisText>
<DatenwerteListe>
  <WerteListe>
    <DatenpunktId>11</DatenpunktId>
    <Wert>value11</Wert>
    <Zeitstempel>`+testTimeStr+`</Zeitstempel>
  </WerteListe>
  <WerteListe>
    <DatenpunktId>22</DatenpunktId>
    <Wert>value22</Wert>
    <Zeitstempel>`+testTimeStr+`</Zeitstempel>
  </WerteListe>
</DatenwerteListe>
<Status>12</Status>`,
		"GetData")

	// With an error
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			return assert.NotNil(d.GetData(v, []AttrId{11, 22}))
		},
		// SOAP action
		"GetData",
		expectedRequest,
		// Response to reply
		`<bad XML>`,
		"GetData with error")
}

//
// WriteData
//

type requestWriteData struct {
	requestDeviceCommon
	Id    int    `xml:"DatapointId"`
	Value string `xml:"Wert"`
}

type requestWriteDataBody struct {
	WriteData requestWriteData `xml:"Body>WriteData"`
}

const (
	writeDataTestId    = 12
	writeDataTestValue = "value12"
)

var writeDataTest = testAction{
	expectedRequest: &requestWriteDataBody{
		WriteData: requestWriteData{
			requestDeviceCommon: deviceCommon,
			Id:                  writeDataTestId,
			Value:               writeDataTestValue,
		},
	},
	serverResponse: `<Ergebnis>0</Ergebnis>
<ErgebnisText>Kein Fehler</ErgebnisText>
<AktualisierungsId>123456789</AktualisierungsId>`,
}

func TestWriteData(t *testing.T) {
	assert := assert.New(t)

	// No problem
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			refreshId, err := d.WriteData(v, writeDataTestId, writeDataTestValue)
			if !assert.Nil(err) {
				return false
			}
			return assert.Equal("123456789", refreshId)
		},
		// SOAP action
		"WriteData",
		writeDataTest.expectedRequest,
		// Response to reply
		writeDataTest.serverResponse,
		"WriteData")

	// With an error
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			_, err := d.WriteData(v, writeDataTestId, writeDataTestValue)
			return assert.NotNil(err)
		},
		// SOAP action
		"WriteData",
		writeDataTest.expectedRequest,
		// Response to reply
		`<bad XML>`,
		"WriteData with error")
}

//
// RefreshData
//

type requestRefreshData struct {
	requestDeviceCommon
	Ids []AttrId `xml:"DatenpunktIds>int"`
}

type requestRefreshDataBody struct {
	RefreshData requestRefreshData `xml:"Body>RefreshData"`
}

var refreshDataTestIds = []AttrId{11, 22, 33}

var refreshDataTest = testAction{
	expectedRequest: &requestRefreshDataBody{
		RefreshData: requestRefreshData{
			requestDeviceCommon: deviceCommon,
			Ids:                 refreshDataTestIds,
		},
	},
	serverResponse: `<Ergebnis>0</Ergebnis>
<ErgebnisText>Kein Fehler</ErgebnisText>
<AktualisierungsId>123456789</AktualisierungsId>`,
}

func TestRefreshData(t *testing.T) {
	assert := assert.New(t)

	// No problem
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			refreshId, err := d.RefreshData(v, refreshDataTestIds)
			if !assert.Nil(err) {
				return false
			}
			return assert.Equal("123456789", refreshId)
		},
		// SOAP action
		"RefreshData",
		refreshDataTest.expectedRequest,
		// Response to reply
		refreshDataTest.serverResponse,
		"RefreshData")

	// With an error
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			return assert.NotNil(d.RefreshData(v, refreshDataTestIds))
		},
		// SOAP action
		"RefreshData",
		refreshDataTest.expectedRequest,
		// Response to reply
		`<bad XML>`,
		"RefreshData with error")
}

func TestErrorHistoryEvent(t *testing.T) {
	assert := assert.New(t)

	ehe := &ErrorHistoryEvent{
		Error:    "EC",
		Message:  "Error message",
		Time:     testTime,
		IsActive: false,
	}

	expectedStr := "EC@" + testTimeStr + " = Error message"

	assert.Equal(expectedStr, ehe.String())

	ehe.IsActive = true
	assert.Equal(expectedStr+" *ACTIVE*", ehe.String())
}

//
// GetErrorHistory
//

func TestGetErrorHistory(t *testing.T) {
	assert := assert.New(t)

	type requestGetErrorHistory struct {
		requestDeviceCommon
		Locale string `xml:"Culture"`
	}

	type requestBody struct {
		GetErrorHistory requestGetErrorHistory `xml:"Body>GetErrorHistory"`
	}

	expectedRequest := &requestBody{
		GetErrorHistory: requestGetErrorHistory{
			requestDeviceCommon: deviceCommon,
			Locale:              "fr-fr",
		},
	}

	// No problem
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			if !assert.Nil(d.GetErrorHistory(v)) {
				return false
			}
			return assert.Equal([]ErrorHistoryEvent{
				{
					Error:    "AB",
					Message:  "First error",
					Time:     testTime,
					IsActive: true,
				},
				{
					Error:    "CD",
					Message:  "Second error",
					Time:     testTime,
					IsActive: false,
				},
			}, d.Errors)
		},
		// SOAP action
		"GetErrorHistory",
		expectedRequest,
		// Response to reply
		`<Ergebnis>0</Ergebnis>
<ErgebnisText>Kein Fehler</ErgebnisText>
<FehlerListe>
  <FehlerHistorie>
    <FehlerCode>AB</FehlerCode>
    <FehlerMeldung>First error</FehlerMeldung>
    <Zeitstempel>`+testTimeStr+`</Zeitstempel>
    <FehlerIstAktiv>1</FehlerIstAktiv>
  </FehlerHistorie>
  <FehlerHistorie>
    <FehlerCode>CD</FehlerCode>
    <FehlerMeldung>Second error</FehlerMeldung>
    <Zeitstempel>`+testTimeStr+`</Zeitstempel>
    <FehlerIstAktiv>0</FehlerIstAktiv>
  </FehlerHistorie>
</FehlerListe>`,
		"GetErrorHistory")

	// With an error
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			return assert.NotNil(d.GetErrorHistory(v))
		},
		// SOAP action
		"GetErrorHistory",
		expectedRequest,
		// Response to reply
		`<bad XML>`,
		"GetErrorHistory with error")
}

//
// GetTimesheetData
//

func TestGetTimesheetData(t *testing.T) {
	assert := assert.New(t)

	type requestGetTimesheetData struct {
		requestDeviceCommon
		Id int `xml:"DatenpunktId"`
	}

	type requestBody struct {
		GetTimesheetData requestGetTimesheetData `xml:"Body>GetTimesheetData"`
	}

	expectedRequest := &requestBody{
		GetTimesheetData: requestGetTimesheetData{
			requestDeviceCommon: deviceCommon,
			Id:                  23,
		},
	}

	// No problem
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			if !assert.Nil(d.GetTimesheetData(v, 23)) {
				return false
			}
			return assert.Equal(map[string]TimeslotSlice{
				"mon": {
					{From: 900, To: 1011},
					{From: 1015, To: 1222},
					{From: 1230, To: 1345},
				},
				"wed": {
					{From: 1900, To: 2011},
					{From: 2015, To: 2222},
					{From: 2230, To: 2345},
				},
			}, d.Timesheets[23])
		},
		// SOAP action
		"GetTimesheetData",
		expectedRequest,
		// Response to reply
		`<Ergebnis>0</Ergebnis>
<ErgebnisText>Kein Fehler</ErgebnisText>
<SchaltsatzDaten>
  <DatenpunktId>23</DatenpunktId>
  <Schaltzeiten>
    <Schaltzeit>
      <Wochentag>Mon</Wochentag>
      <ZeitVon>1230</ZeitVon>
      <ZeitBis>1345</ZeitBis>
    </Schaltzeit>
    <Schaltzeit>
      <Wochentag>Wed</Wochentag>
      <ZeitVon>2015</ZeitVon>
      <ZeitBis>2222</ZeitBis>
    </Schaltzeit>
    <Schaltzeit>
      <Wochentag>Mon</Wochentag>
      <ZeitVon>900</ZeitVon>
      <ZeitBis>1011</ZeitBis>
    </Schaltzeit>
    <Schaltzeit>
      <Wochentag>Wed</Wochentag>
      <ZeitVon>2230</ZeitVon>
      <ZeitBis>2345</ZeitBis>
    </Schaltzeit>
    <Schaltzeit>
      <Wochentag>Mon</Wochentag>
      <ZeitVon>1015</ZeitVon>
      <ZeitBis>1222</ZeitBis>
    </Schaltzeit>
    <Schaltzeit>
      <Wochentag>Wed</Wochentag>
      <ZeitVon>1900</ZeitVon>
      <ZeitBis>2011</ZeitBis>
    </Schaltzeit>
  </Schaltzeiten>
</SchaltsatzDaten>`,
		"GetTimesheetData")

	// With an error
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			return assert.NotNil(d.GetTimesheetData(v, 23))
		},
		// SOAP action
		"GetTimesheetData",
		expectedRequest,
		// Response to reply
		`<bad XML>`,
		"GetTimesheetData with error")
}
