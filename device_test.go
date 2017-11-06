package vitotrol

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testDeviceID   = 1234
	testLocationID = 5678
	testTimeStr    = "2016-10-30 12:13:14"
)

var (
	_ = []HasResultHeader{
		(*GetDataResponse)(nil),
		(*WriteDataResponse)(nil),
		(*RefreshDataResponse)(nil),
		(*GetErrorHistoryResponse)(nil),
		(*GetTimesheetDataResponse)(nil),
		(*WriteTimesheetDataResponse)(nil),
		(*GetTypeInfoResponse)(nil),
	}

	testTime = func() Time {
		tm, _ := ParseVitotrolTime(testTimeStr)
		return tm
	}()
)

func TestFormatAttributes(t *testing.T) {
	assert := assert.New(t)

	pDevice := &Device{
		Attributes: map[AttrID]*Value{
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
			[]AttrID{NoAttr, BurnerState, IndoorTemp, OutdoorTemp}))
}

func TestMakeDatenpunktIDs(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(`<DatenpunktIds><int>11</int><int>22</int></DatenpunktIds>`,
		makeDatenpunktIDs([]AttrID{11, 22}))
}

type requestDeviceCommon struct {
	DeviceID   uint32 `xml:"GeraetId"`
	LocationID uint32 `xml:"AnlageId"`
}

var deviceCommon = requestDeviceCommon{
	DeviceID:   testDeviceID,
	LocationID: testLocationID,
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
					DeviceID:   testDeviceID,
					LocationID: testLocationID,
					Attributes: map[AttrID]*Value{},
					Timesheets: map[TimesheetID]map[string]TimeslotSlice{},
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
		IDs []int `xml:"DatenpunktIds>int"`
	}

	type requestBody struct {
		GetData requestGetData `xml:"Body>GetData"`
	}

	expectedRequest := &requestBody{
		GetData: requestGetData{
			requestDeviceCommon: deviceCommon,
			IDs:                 []int{11, 22},
		},
	}

	// No problem
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			err := d.GetData(v, []AttrID{11, 22})
			if !assert.Nil(err) {
				return false
			}
			return assert.Equal(map[AttrID]*Value{
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
			return assert.NotNil(d.GetData(v, []AttrID{11, 22}))
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
	ID    int    `xml:"DatapointId"`
	Value string `xml:"Wert"`
}

type requestWriteDataBody struct {
	WriteData requestWriteData `xml:"Body>WriteData"`
}

const (
	writeDataTestID    = 12
	writeDataTestValue = "value12"
)

var writeDataTest = testAction{
	expectedRequest: &requestWriteDataBody{
		WriteData: requestWriteData{
			requestDeviceCommon: deviceCommon,
			ID:                  writeDataTestID,
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
			refreshID, err := d.WriteData(v, writeDataTestID, writeDataTestValue)
			if !assert.Nil(err) {
				return false
			}
			return assert.Equal("123456789", refreshID)
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
			_, err := d.WriteData(v, writeDataTestID, writeDataTestValue)
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
	IDs []AttrID `xml:"DatenpunktIds>int"`
}

type requestRefreshDataBody struct {
	RefreshData requestRefreshData `xml:"Body>RefreshData"`
}

var refreshDataTestIDs = []AttrID{11, 22, 33}

var refreshDataTest = testAction{
	expectedRequest: &requestRefreshDataBody{
		RefreshData: requestRefreshData{
			requestDeviceCommon: deviceCommon,
			IDs:                 refreshDataTestIDs,
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
			refreshID, err := d.RefreshData(v, refreshDataTestIDs)
			if !assert.Nil(err) {
				return false
			}
			return assert.Equal("123456789", refreshID)
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
			return assert.NotNil(d.RefreshData(v, refreshDataTestIDs))
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
		ID int `xml:"DatenpunktId"`
	}

	type requestBody struct {
		GetTimesheetData requestGetTimesheetData `xml:"Body>GetTimesheetData"`
	}

	expectedRequest := &requestBody{
		GetTimesheetData: requestGetTimesheetData{
			requestDeviceCommon: deviceCommon,
			ID:                  23,
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

//
// WriteTimesheetData
//

func TestWriteTimesheetData(t *testing.T) {
	assert := assert.New(t)

	type requestDaySlot struct {
		Day      string `xml:"Wochentag"`
		From     string `xml:"ZeitVon"`
		To       string `xml:"ZeitBis"`
		Value    int    `xml:"Wert"`
		Position int    `xml:"Position"`
	}

	type requestWriteTimesheetData struct {
		requestDeviceCommon
		ID       int              `xml:"DatenpunktId"`
		Type     int              `xml:"SchaltzeitTyp"`
		DaySlots []requestDaySlot `xml:"Schaltzeiten>Schaltzeit"`
	}

	// Be careful to SchaltsatzData nested layer
	type requestBody struct {
		WriteTimesheetData requestWriteTimesheetData `xml:"Body>WriteTimesheetData>SchaltsatzData"`
	}

	expectedRequest := &requestBody{
		WriteTimesheetData: requestWriteTimesheetData{
			requestDeviceCommon: deviceCommon,
			ID:                  23,
			Type:                1,
			DaySlots: []requestDaySlot{
				{Day: "MON", From: "0610", To: "0820", Value: 1, Position: 0},
				{Day: "MON", From: "1610", To: "1820", Value: 1, Position: 1},
				{Day: "TUE", From: "0610", To: "0820", Value: 1, Position: 0},
				{Day: "WED", From: "0610", To: "0820", Value: 1, Position: 0},
				{Day: "THU", From: "0610", To: "0820", Value: 1, Position: 0},
				{Day: "FRI", From: "0610", To: "0820", Value: 1, Position: 0},
				{Day: "SAT", From: "0610", To: "0820", Value: 1, Position: 0},
				{Day: "SAT", From: "1610", To: "1820", Value: 1, Position: 1},
				{Day: "SUN", From: "0610", To: "0820", Value: 1, Position: 0},
			},
		},
	}

	inputOK := map[string]TimeslotSlice{
		"mon":     {{From: 1610, To: 1820}, {From: 610, To: 820}},
		"Tue":     {{From: 610, To: 820}},
		"weD-FRI": {{From: 610, To: 820}},
		"sat":     {{From: 1610, To: 1820}, {From: 610, To: 820}},
		"sun":     {{From: 610, To: 820}},
	}

	// No problem
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			id, err := d.WriteTimesheetData(v, 23, inputOK)
			return assert.Equal("123456789", id) && assert.Nil(err)
		},
		// SOAP action
		"WriteTimesheetData",
		expectedRequest,
		// Response to reply
		`<Ergebnis>0</Ergebnis>
<ErgebnisText>Kein Fehler</ErgebnisText>
<AktualisierungsId>123456789</AktualisierungsId>`,
		"WriteTimesheetData")

	// Bad dayslot
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			id, err := d.WriteTimesheetData(v, 23, map[string]TimeslotSlice{
				"foo": {{From: 1610, To: 1820}},
			})
			return assert.Empty(id) &&
				assert.Error(err) &&
				assert.Equal("Bad timesheet day `FOO'", err.Error())
		},
		"", nil, "", "WriteTimesheetData with bad day")

	// Bad dayslot (day range)
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			id, err := d.WriteTimesheetData(v, 23, map[string]TimeslotSlice{
				"foo-bar": {{From: 1610, To: 1820}},
			})
			return assert.Empty(id) &&
				assert.Error(err) &&
				assert.Equal("Bad timesheet range of days `FOO-BAR'", err.Error())
		},
		"", nil, "", "WriteTimesheetData with bad day range")

	// Bad dayslot (duplicate)
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			id, err := d.WriteTimesheetData(v, 23, map[string]TimeslotSlice{
				"mon":     {{From: 1610, To: 1820}},
				"sun-tue": {{From: 1610, To: 1820}},
			})
			return assert.Empty(id) &&
				assert.Error(err) &&
				assert.Equal("Duplicate day `MON'", err.Error())
		},
		"", nil, "", "WriteTimesheetData with bad day")

	// Async error
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			id, err := d.WriteTimesheetData(v, 23, inputOK)
			return assert.Empty(id) && assert.Error(err)
		},
		// SOAPAction
		"WriteTimesheetData",
		expectedRequest,
		// Response to reply
		`<bad XML>`,
		"WriteTimesheetData with async error")
}

//
// GetTypeInfo
//

func TestGetTypeInfo(t *testing.T) {
	assert := assert.New(t)

	type requestGetTypeInfo struct {
		requestDeviceCommon
	}

	type requestBody struct {
		GetTypeInfo requestGetTypeInfo `xml:"Body>GetTypeInfo"`
	}

	expectedRequest := &requestBody{
		GetTypeInfo: requestGetTypeInfo{
			requestDeviceCommon: deviceCommon,
		},
	}

	// No problem
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			list, err := d.GetTypeInfo(v)
			if !assert.Nil(err) || !assert.NotEmpty(list) {
				return false
			}
			return assert.Equal([]*AttributeInfo{
				&AttributeInfo{
					AttributeInfoBase: AttributeInfoBase{
						AttributeName:      "anzahl_brennerstunden_r",
						AttributeType:      "Double",
						AttributeTypeValue: 0,
						MinValue:           "",
						MaxValue:           "",
						DataPointGroup:     "ecnsysEventTypeGroupHC~VScotHO1_72",
						HeatingCircuitID:   19178,
						DefaultValue:       "",
						Readable:           true,
						Writable:           false,
					},
					AttributeID: 104,
				},
				&AttributeInfo{
					AttributeInfoBase: AttributeInfoBase{
						AttributeName:      "konf_ww_solltemp_rw",
						AttributeType:      "Integer",
						AttributeTypeValue: 0,
						MinValue:           "10",
						MaxValue:           "95",
						DataPointGroup:     "viessmann.eventtypegroupHC.name.VScotHO1_72~HC1",
						HeatingCircuitID:   19179,
						DefaultValue:       "50",
						Readable:           true,
						Writable:           true,
					},
					AttributeID: 51,
				},
				&AttributeInfo{
					AttributeInfoBase: AttributeInfoBase{
						AttributeName:      "zustand_interne_pumpe_r",
						AttributeType:      "ENUM",
						AttributeTypeValue: 0,
						MinValue:           "",
						MaxValue:           "",
						DataPointGroup:     "ecnsysEventTypeGroupHC~VScotHO1_72",
						HeatingCircuitID:   19178,
						DefaultValue:       "",
						Readable:           true,
						Writable:           false,
					},
					AttributeID: 245,
					EnumValues: map[uint32]string{
						0: "Aus",
						1: "Ein",
					},
				},
			}, list)
		},
		// SOAP action
		"GetTypeInfo",
		expectedRequest,
		// Response to reply
		`<Ergebnis>0</Ergebnis>
<ErgebnisText>Kein Fehler</ErgebnisText>
<TypeInfoListe>
  <DatenpunktTypInfo>
    <AnlageId>88888</AnlageId>
    <GeraetId>77777</GeraetId>
    <DatenpunktId>104</DatenpunktId>
    <DatenpunktName>anzahl_brennerstunden_r</DatenpunktName>
    <DatenpunktTyp>Double</DatenpunktTyp>
    <DatenpunktTypWert>0</DatenpunktTypWert>
    <MinimalWert />
    <MaximalWert />
    <DatenpunktGruppe>ecnsysEventTypeGroupHC~VScotHO1_72</DatenpunktGruppe>
    <HeizkreisId>19178</HeizkreisId>
    <Auslieferungswert />
    <IstLesbar>true</IstLesbar>
    <IstSchreibbar>false</IstSchreibbar>
  </DatenpunktTypInfo>
  <DatenpunktTypInfo>
    <AnlageId>88888</AnlageId>
    <GeraetId>77777</GeraetId>
    <DatenpunktId>51</DatenpunktId>
    <DatenpunktName>konf_ww_solltemp_rw</DatenpunktName>
    <DatenpunktTyp>Integer</DatenpunktTyp>
    <DatenpunktTypWert>0</DatenpunktTypWert>
    <MinimalWert>10</MinimalWert>
    <MaximalWert>95</MaximalWert>
    <DatenpunktGruppe>viessmann.eventtypegroupHC.name.VScotHO1_72~HC1</DatenpunktGruppe>
    <HeizkreisId>19179</HeizkreisId>
    <Auslieferungswert>50</Auslieferungswert>
    <IstLesbar>true</IstLesbar>
    <IstSchreibbar>true</IstSchreibbar>
  </DatenpunktTypInfo>
  <DatenpunktTypInfo>
    <AnlageId>88888</AnlageId>
    <GeraetId>77777</GeraetId>
    <DatenpunktId>245</DatenpunktId>
    <DatenpunktName>zustand_interne_pumpe_r</DatenpunktName>
    <DatenpunktTyp>ENUM</DatenpunktTyp>
    <DatenpunktTypWert>0</DatenpunktTypWert>
    <MinimalWert />
    <MaximalWert />
    <DatenpunktGruppe>ecnsysEventTypeGroupHC~VScotHO1_72</DatenpunktGruppe>
    <HeizkreisId>19178</HeizkreisId>
    <Auslieferungswert />
    <IstLesbar>true</IstLesbar>
    <IstSchreibbar>false</IstSchreibbar>
  </DatenpunktTypInfo>
  <DatenpunktTypInfo>
    <AnlageId>88888</AnlageId>
    <GeraetId>77777</GeraetId>
    <DatenpunktId>245-0</DatenpunktId>
    <DatenpunktName>zustand_interne_pumpe_r</DatenpunktName>
    <DatenpunktTyp>ENUM</DatenpunktTyp>
    <DatenpunktTypWert>0</DatenpunktTypWert>
    <MinimalWert>Aus</MinimalWert>
    <MaximalWert />
    <DatenpunktGruppe>ecnsysEventTypeGroupHC~VScotHO1_72</DatenpunktGruppe>
    <HeizkreisId>19178</HeizkreisId>
    <Auslieferungswert />
    <IstLesbar>true</IstLesbar>
    <IstSchreibbar>false</IstSchreibbar>
  </DatenpunktTypInfo>
  <DatenpunktTypInfo>
    <AnlageId>88888</AnlageId>
    <GeraetId>77777</GeraetId>
    <DatenpunktId>245-1</DatenpunktId>
    <DatenpunktName>zustand_interne_pumpe_r</DatenpunktName>
    <DatenpunktTyp>ENUM</DatenpunktTyp>
    <DatenpunktTypWert>0</DatenpunktTypWert>
    <MinimalWert>Ein</MinimalWert>
    <MaximalWert />
    <DatenpunktGruppe>ecnsysEventTypeGroupHC~VScotHO1_72</DatenpunktGruppe>
    <HeizkreisId>19178</HeizkreisId>
    <Auslieferungswert />
    <IstLesbar>true</IstLesbar>
    <IstSchreibbar>false</IstSchreibbar>
  </DatenpunktTypInfo>
</TypeInfoListe>`,
		"GetTypeInfo")

	// With an error
	testSendRequestDeviceAny(assert,
		// Send request and check result
		func(v *Session, d *Device) bool {
			_, err := d.GetTypeInfo(v)
			return assert.NotNil(err)
		},
		// SOAP action
		"GetTypeInfo",
		expectedRequest,
		// Response to reply
		`<bad XML>`,
		"GetTypeInfo with error")
}
