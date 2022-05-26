package vitotrol

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	td "github.com/maxatome/go-testdeep"
)

var _ = []HasResultHeader{
	(*LoginResponse)(nil),
	(*GetDevicesResponse)(nil),
	(*RequestRefreshStatusResponse)(nil),
	(*RequestWriteStatusResponse)(nil),
}

const (
	respHeader = `<?xml version="1.0" encoding="utf-8"?><soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema"><soap:Body>`
	respFooter = `</soap:Body></soap:Envelope>`
)

func extractRequestBody(t *td.T, r *http.Request, reqBody interface{}, testName string) bool {
	t.Helper()

	bodyRaw, err := ioutil.ReadAll(r.Body)
	if !t.CmpNoError(err, "%s: request body ReadAll OK", testName) {
		return false
	}

	err = xml.Unmarshal(bodyRaw, reqBody)
	return t.CmpNoError(err, "%s: request body Unmarshal OK", testName)
}

func virginInstance(pOrig interface{}) interface{} {
	return reflect.New(
		reflect.Indirect(reflect.ValueOf(pOrig)).Type()).
		Interface()
}

func testSendRequestAny(t *td.T,
	sendReq func(v *Session) bool, soapAction string,
	expectedRequest interface{}, serverResponse string,
	testName string) bool {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Check header
			t.CmpDeeply(r.Header.Get("SOAPAction"), soapURL+soapAction,
				"%s: SOAPAction header matches", testName)
			t.CmpDeeply(r.Header.Get("Content-Type"), "text/xml; charset=utf-8",
				"%s: Content-Type header matches", testName)

			if cookie := r.Header.Get("Cookie"); cookie != "" {
				w.Header().Add("Set-Cookie", cookie)
			}

			// Extract request body in the same struct type as the expectedRequest
			recvReq := virginInstance(expectedRequest)
			if !extractRequestBody(t, r, recvReq, testName) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			t.CmpDeeply(recvReq, expectedRequest, "%s: request OK", testName)

			// Send response
			fmt.Fprintln(w, respHeader+serverResponse+respFooter)
		}))
	defer ts.Close()

	MainURL = ts.URL

	return sendReq(&Session{})
}

//
// sendRequest
//

type TestResponse struct {
	TestResult TestResult `xml:"Body>TestResponse>TestResult"`
}

type TestResult struct {
	ResultHeader
	Pipo string `xml:"Pipo"`
}

func (r *TestResponse) ResultHeader() *ResultHeader {
	return &r.TestResult.ResultHeader
}

func TestSendRequestErrors(tt *testing.T) {
	t := td.NewT(tt)

	v := &Session{}

	// bad URL -> parse URL will fail
	MainURL = ":"
	var resp TestResponse
	err := v.sendRequest("bad", `<xxx></xxx>`, &resp)
	t.CmpError(err)

	// bad scheme -> Do request will fail
	MainURL = "bad-scheme:..."
	err = v.sendRequest("bad", `<xxx></xxx>`, &resp)
	t.CmpError(err)

	// HTTP status error
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
	defer ts.Close()

	MainURL = ts.URL
	err = v.sendRequest("bad", `<xxx></xxx>`, &resp)
	t.CmpError(err)
}

func TestSendRequest(tt *testing.T) {
	t := td.NewT(tt)

	type testRequest struct {
		Foo string `xml:"Body>Test>Foo"`
		Bar string `xml:"Body>Test>Bar"`
	}

	// No problem
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			v.Cookies = []string{"foo=123", "bar=456"}

			var resp TestResponse
			err := v.sendRequest("foobar", `
<Test>
  <Foo>foo</Foo>
  <Bar>bar</Bar>
</Test>`, &resp)
			if !t.CmpNoError(err) {
				return false
			}
			return t.CmpDeeply(&resp,
				&TestResponse{
					TestResult: TestResult{
						ResultHeader: ResultHeader{
							ErrorNum: 0,
							ErrorStr: "Kein Fehler",
						},
						Pipo: "hello",
					},
				})
		},
		// SOAP action
		"foobar",
		// Expected request
		&testRequest{
			Foo: "foo",
			Bar: "bar",
		},
		// Response to reply
		`<TestResponse xmlns="http://www/">
  <TestResult>
   <Ergebnis>0</Ergebnis>
   <ErgebnisText>Kein Fehler</ErgebnisText>
   <Pipo>hello</Pipo>
  </TestResult>
</TestResponse>`,
		"sendRequest")

	// XML decoding error
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			v.Debug = true

			var resp TestResponse
			err := v.sendRequest("foobar", `
<Test>
  <Foo>foo</Foo>
  <Bar>bar</Bar>
</Test>`, &resp)
			return t.CmpError(err)
		},
		// SOAP action
		"foobar",
		// Expected request
		&testRequest{
			Foo: "foo",
			Bar: "bar",
		},
		// Response to reply
		"<bad XML>",
		"sendRequest XML error")

	// Applicative error
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			var resp TestResponse
			err := v.sendRequest("foobar", `
<Test>
  <Foo>foo</Foo>
  <Bar>bar</Bar>
</Test>`, &resp)
			if !t.CmpError(err) || !t.Isa(err, &ResultHeader{}) {
				return false
			}
			return t.CmpDeeply(err.(*ResultHeader),
				td.Struct(&ResultHeader{
					ErrorNum: 42,
					ErrorStr: "ERROR!!!",
				}, nil))
		},
		// SOAP action
		"foobar",
		// Expected request
		&testRequest{
			Foo: "foo",
			Bar: "bar",
		},
		// Response to reply
		`<TestResponse xmlns="http://www/">
  <TestResult>
   <Ergebnis>42</Ergebnis>
   <ErgebnisText>ERROR!!!</ErgebnisText>
   <Pipo>hello</Pipo>
  </TestResult>
</TestResponse>`,
		"sendRequest app error")
}

//
// Login.
//
func TestLogin(tt *testing.T) {
	t := td.NewT(tt)

	type loginRequest struct {
		AppID      string `xml:"Body>Login>AppId"`
		AppVersion string `xml:"Body>Login>AppVersion"`
		Password   string `xml:"Body>Login>Passwort"`
		System     string `xml:"Body>Login>Betriebssystem"`
		Login      string `xml:"Body>Login>Benutzer"`
	}

	expectedRequest := &loginRequest{
		AppID:      "prod",
		AppVersion: "4.3.1",
		Password:   "bingo",
		System:     "Android",
		Login:      "pipo",
	}

	// No problem
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			return t.CmpNoError(v.Login("pipo", "bingo"))
		},
		// SOAP action
		"Login",
		expectedRequest,
		// Response to reply
		`<LoginResponse xmlns="http://www.e-controlnet.de/services/vii/">
  <LoginResult>
    <Ergebnis>0</Ergebnis>
    <ErgebnisText>Kein Fehler</ErgebnisText>
    <TechVersion>2.5.6.0</TechVersion>
    <Anrede>1</Anrede>
    <Vorname>Maxime</Vorname>
    <Nachname>Soulé</Nachname>
  </LoginResult>
</LoginResponse>`,
		"Login")

	// With an error
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			return t.CmpError(v.Login("pipo", "bingo"))
		},
		// SOAP action
		"Login",
		expectedRequest,
		// Response to reply
		`<bad XML>`,
		"Login with error")
}

//
// GetDevices.
//
func TestGetDevices(tt *testing.T) {
	t := td.NewT(tt)

	type getDevicesRequest struct {
		Dummy string `xml:"Body>GetDevices,omitempty"`
	}

	expectedRequest := &getDevicesRequest{}

	// No problem
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			err := v.GetDevices()
			if !t.CmpNoError(err) {
				return false
			}
			return t.CmpDeeply(v.Devices,
				[]Device{
					{
						LocationID:   31456,
						LocationName: "Paris",
						DeviceID:     40213,
						DeviceName:   "VT 200 (HO1C)",
						HasError:     true,
						IsConnected:  true,
						Attributes:   map[AttrID]*Value{},
						Timesheets:   map[TimesheetID]map[string]TimeslotSlice{},
					},
				})
		},
		// SOAP action
		"GetDevices",
		expectedRequest,
		// Response to reply
		`<GetDevicesResponse xmlns="http://www.e-controlnet.de/services/vii/GetDevices">
  <GetDevicesResult>
    <Ergebnis>0</Ergebnis>
    <ErgebnisText>Kein Fehler</ErgebnisText>
    <AnlageListe>
      <AnlageV2>
        <AnlageId>31456</AnlageId>
        <AnlageName>Paris</AnlageName>
        <AnlageStandort>Paris</AnlageStandort>
        <AnlageTyp />
        <GeraeteListe>
          <GeraetV2>
            <GeraetId>40213</GeraetId>
            <GeraetName>VT 200 (HO1C)</GeraetName>
            <GeraetTyp>350</GeraetTyp>
            <Heizkreise>
              <BenutzerHeizkreis>
                <HeizkreisId>19179</HeizkreisId>
                <HeizkreisBezeichnung>viessmann.eventtypegroupHC.name.VScotHO1_72~HC1</HeizkreisBezeichnung>
                <Benutzerfreigabe>true</Benutzerfreigabe>
              </BenutzerHeizkreis>
            </Heizkreise>
            <ViaFreigabe>true</ViaFreigabe>
            <Regelungstype>GWG</Regelungstype>
            <Regelungsadresse>VScotHO1_72</Regelungsadresse>
            <HatFehler>true</HatFehler>
            <IstVerbunden>true</IstVerbunden>
          </GeraetV2>
        </GeraeteListe>
        <VerbindungsTyp />
        <HatFehler>false</HatFehler>
        <IstVerbunden>true</IstVerbunden>
      </AnlageV2>
    </AnlageListe>
  </GetDevicesResult>
</GetDevicesResponse>`,
		"GetDevices")

	// With an error
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			return t.CmpError(v.GetDevices())
		},
		// SOAP action
		"GetDevices",
		expectedRequest,
		// Response to reply
		`<bad XML>`,
		"GetDevices with error")
}

//
// RequestRefreshStatus
//

type requestRefreshStatusRequest struct {
	AktualisierungsID string `xml:"Body>RequestRefreshStatus>AktualisierungsId"`
}

var requestRefreshStatusTest = testAction{
	expectedRequest: &requestRefreshStatusRequest{
		AktualisierungsID: "123456789",
	},
	serverResponse: `<RequestRefreshStatusResponse xmlns="http://www.e-controlnet.de/services/vii/">
  <RequestRefreshStatusResult>
    <Ergebnis>0</Ergebnis>
    <ErgebnisText>Kein Fehler</ErgebnisText>
    <Status>4</Status>
  </RequestRefreshStatusResult>
</RequestRefreshStatusResponse>`,
}

func TestRequestRefreshStatus(tt *testing.T) {
	t := td.NewT(tt)

	// No problem
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			status, err := v.RequestRefreshStatus("123456789")
			return t.CmpNoError(err) && t.CmpDeeply(status, 4)
		},
		// SOAP action
		"RequestRefreshStatus",
		requestRefreshStatusTest.expectedRequest,
		// Response to reply
		requestRefreshStatusTest.serverResponse,
		"RequestRefreshStatus")

	// With an error
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			_, err := v.RequestRefreshStatus("123456789")
			return t.CmpError(err)
		},
		// SOAP action
		"RequestRefreshStatus",
		requestRefreshStatusTest.expectedRequest,
		// Response to reply
		`<bad XML>`,
		"RequestRefreshStatus with error")
}

//
// RequestWriteStatus
//

type requestWriteStatusRequest struct {
	AktualisierungsID string `xml:"Body>RequestWriteStatus>AktualisierungsId"`
}

var requestWriteStatusTest = testAction{
	expectedRequest: &requestWriteStatusRequest{
		AktualisierungsID: "123456789",
	},
	serverResponse: `<RequestWriteStatusResponse xmlns="http://www.e-controlnet.de/services/vii/">
  <RequestWriteStatusResult>
    <Ergebnis>0</Ergebnis>
    <ErgebnisText>Kein Fehler</ErgebnisText>
    <Status>4</Status>
  </RequestWriteStatusResult>
</RequestWriteStatusResponse>`,
}

func TestRequestWriteStatus(tt *testing.T) {
	t := td.NewT(tt)

	// No problem
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			status, err := v.RequestWriteStatus("123456789")
			return t.CmpNoError(err) && t.CmpDeeply(status, 4)
		},
		// SOAP action
		"RequestWriteStatus",
		requestWriteStatusTest.expectedRequest,
		// Response to reply
		requestWriteStatusTest.serverResponse,
		"RequestWriteStatus")

	// With an error
	testSendRequestAny(t,
		// Send request and check result
		func(v *Session) bool {
			_, err := v.RequestWriteStatus("123456789")
			return t.CmpError(err)
		},
		// SOAP action
		"RequestWriteStatus",
		requestWriteStatusTest.expectedRequest,
		// Response to reply
		`<bad XML>`,
		"RequestWriteStatus with error")
}
