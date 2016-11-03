package vitotrol

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var MainURL = `http://www.viessmann.com/app_vitodata/VIIWebService-1.16.0.0/iPhoneWebService.asmx`

const (
	soapURL = `http://www.e-controlnet.de/services/vii/`

	reqHeader = `<?xml version="1.0" encoding="UTF-8"?><soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns="` + soapURL + `">
<soap:Body>
`
	reqFooter = `
</soap:Body>
</soap:Envelope>`
)

// Session keep a cache of all informations downloaded from the
// Vitotrol™ server. See Login method as entry point.
type Session struct {
	Cookies []string

	Devices []Device

	Debug bool
}

func (v *Session) sendRequest(soapAction string, reqBody string, respBody HasResultHeader) error {
	client := &http.Client{}

	req, err := http.NewRequest("POST", MainURL,
		bytes.NewBuffer([]byte(reqHeader+reqBody+reqFooter)))
	if err != nil {
		return err
	}

	//req.Header.Set("User-Agent", userAgent)
	req.Header.Set("SOAPAction", soapURL+soapAction)
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	for _, cookie := range v.Cookies {
		req.Header.Add("Cookie", cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBodyRaw, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		cookies := resp.Header[http.CanonicalHeaderKey("Set-Cookie")]
		if cookies != nil {
			v.Cookies = cookies
		}

		if v.Debug {
			log.Println(string(respBodyRaw))
		}

		err = xml.Unmarshal(respBodyRaw, respBody)
		if err != nil {
			return err
		}

		// Applicative error
		if respBody.ResultHeader().IsError() {
			return respBody.ResultHeader()
		}
		return nil
	}

	return fmt.Errorf("HTTP error: [status=%d] %v (%+v)",
		resp.StatusCode, respBodyRaw, resp.Header)
}

/* Login
POST /app_vitodata/VIIWebService-1.16.0.0/iPhoneWebService.asmx HTTP/1.1
Content-Type: text/xml; charset=utf-8
Connection: Keep-Alive
Accept-Encoding: gzip

<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns="http://www.e-controlnet.de/services/vii/">
  <soap:Body>
    <Login>
      <AppId>prod</AppId>
      <AppVersion>4.3.1</AppVersion>
      <Passwort>PASSWORD</Passwort>
      <Betriebssystem>Android</Betriebssystem>
      <Benutzer>LOGIN</Benutzer>
    </Login>
  </soap:Body>
</soap:Envelope>

<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <soap:Body>
    <LoginResponse xmlns="http://www.e-controlnet.de/services/vii/">
      <LoginResult>
        <Ergebnis>0</Ergebnis>
        <ErgebnisText>Kein Fehler</ErgebnisText>
        <TechVersion>2.5.6.0</TechVersion>
        <Anrede>1</Anrede>
        <Vorname>Maxime</Vorname>
        <Nachname>Soulé</Nachname>
      </LoginResult>
    </LoginResponse>
  </soap:Body>
</soap:Envelope>
*/

type LoginResponse struct {
	LoginResult LoginResult `xml:"Body>LoginResponse>LoginResult"`
}

type LoginResult struct {
	ResultHeader
	Version   string `xml:"TechVersion"`
	Firstname string `xml:"Vorname"`
	Lastname  string `xml:"Nachname"`
}

func (r *LoginResponse) ResultHeader() *ResultHeader {
	return &r.LoginResult.ResultHeader
}

// Login authenticates the session on the Vitotrol™ server using the
// Login request.
func (v *Session) Login(login, password string) error {
	body := `<Login>
<AppId>prod</AppId>
<AppVersion>4.3.1</AppVersion>
<Passwort>` + password + `</Passwort>
<Betriebssystem>Android</Betriebssystem>
<Benutzer>` + login + `</Benutzer>
</Login>`

	v.Cookies = nil

	var resp LoginResponse
	err := v.sendRequest("Login", body, &resp)
	if err != nil {
		return err
	}

	return nil
}

//
// GetDevices
//

/*
<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope>
  <soap:Body>
    <GetDevicesResponse xmlns="http://www.e-controlnet.de/services/vii/GetDevices">
      <GetDevicesResult>
        <Ergebnis>0</Ergebnis>
        <ErgebnisText>Kein Fehler</ErgebnisText>
        <AnlageListe>
          <AnlageV2>
            <AnlageId>15527</AnlageId>
            <AnlageName>Houilles</AnlageName>
            <AnlageStandort>Houilles</AnlageStandort>
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
                <HatFehler>false</HatFehler>
                <IstVerbunden>true</IstVerbunden>
              </GeraetV2>
            </GeraeteListe>
            <VerbindungsTyp />
            <HatFehler>false</HatFehler>
            <IstVerbunden>true</IstVerbunden>
          </AnlageV2>
        </AnlageListe>
      </GetDevicesResult>
    </GetDevicesResponse>
  </soap:Body>
</soap:Envelope>
*/

type GetDevicesDevices struct {
	Id          uint32 `xml:"GeraetId"`
	Name        string `xml:"GeraetName"`
	HasError    bool   `xml:"HatFehler"`
	IsConnected bool   `xml:"IstVerbunden"`
}

type GetDevicesLocation struct {
	Id          uint32              `xml:"AnlageId"`
	Name        string              `xml:"AnlageName"`
	Devices     []GetDevicesDevices `xml:"GeraeteListe>GeraetV2"`
	HasError    bool                `xml:"HatFehler"`
	IsConnected bool                `xml:"IstVerbunden"`
}

type GetDevicesResponse struct {
	GetDevicesResult GetDevicesResult `xml:"Body>GetDevicesResponse>GetDevicesResult"`
}

type GetDevicesResult struct {
	ResultHeader
	Locations []GetDevicesLocation `xml:"AnlageListe>AnlageV2"`
}

func (r *GetDevicesResponse) ResultHeader() *ResultHeader {
	return &r.GetDevicesResult.ResultHeader
}

// GetDevices launches the Vitotrol™ GetDevices request. Populates the
// internal cache before returning (see Devices field).
func (v *Session) GetDevices() error {
	var resp GetDevicesResponse
	err := v.sendRequest("GetDevices", "<GetDevices/>", &resp)
	if err != nil {
		return err
	}

	// 0 or 1 Location
	for _, location := range resp.GetDevicesResult.Locations {
		for _, device := range location.Devices {
			v.Devices = append(v.Devices, Device{
				LocationId:   location.Id,
				LocationName: location.Name,
				DeviceId:     device.Id,
				DeviceName:   device.Name,
				HasError:     location.HasError || device.HasError,
				IsConnected:  location.IsConnected && device.IsConnected,
				Attributes:   map[AttrId]*Value{},
				Timesheets:   map[TimesheetId]map[string]TimeslotSlice{},
			})
		}
	}

	return nil
}

//
// RequestRefreshStatus
//

type RequestRefreshStatusResponse struct {
	RequestRefreshStatusResult RequestRefreshStatusResult `xml:"Body>RequestRefreshStatusResponse>RequestRefreshStatusResult"`
}
type RequestRefreshStatusResult struct {
	ResultHeader
	Status int `xml:"Status"`
}

func (r *RequestRefreshStatusResponse) ResultHeader() *ResultHeader {
	return &r.RequestRefreshStatusResult.ResultHeader
}

// RequestRefreshStatus launches the Vitotrol™ RequestRefreshStatus
// request to follow the status of the RefreshData request matching
// the passed refresh ID. Use RefreshDataWait instead.
func (v *Session) RequestRefreshStatus(refreshId string) (int, error) {
	var resp RequestRefreshStatusResponse
	err := v.sendRequest("RequestRefreshStatus",
		"<RequestRefreshStatus><AktualisierungsId>"+
			refreshId+
			"</AktualisierungsId></RequestRefreshStatus>",
		&resp)
	if err != nil {
		return 0, err
	}

	return resp.RequestRefreshStatusResult.Status, nil
}

//
// RequestWriteStatus
//

type RequestWriteStatusResponse struct {
	RequestWriteStatusResult RequestWriteStatusResult `xml:"Body>RequestWriteStatusResponse>RequestWriteStatusResult"`
}

type RequestWriteStatusResult struct {
	ResultHeader
	Status int `xml:"Status"`
}

func (r *RequestWriteStatusResponse) ResultHeader() *ResultHeader {
	return &r.RequestWriteStatusResult.ResultHeader
}

// RequestWriteStatus launches the Vitotrol™ RequestWriteStatus
// request to follow the status of the WriteData request matching
// the passed refresh ID. Use WriteDataWait instead.
func (v *Session) RequestWriteStatus(refreshId string) (int, error) {
	var resp RequestWriteStatusResponse
	err := v.sendRequest("RequestWriteStatus",
		"<RequestWriteStatus><AktualisierungsId>"+
			refreshId+
			"</AktualisierungsId></RequestWriteStatus>",
		&resp)
	if err != nil {
		return 0, err
	}

	return resp.RequestWriteStatusResult.Status, nil
}
