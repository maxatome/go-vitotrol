package vitotrol

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
)

// MainURL is the Viessmann Vitodata API URL.
var MainURL = `https://www.viessmann.com/app_vitodata/VIIWebService-1.16.0.0/iPhoneWebService.asmx`

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

	respBodyRaw, _ := io.ReadAll(resp.Body)

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

//
// Login
//

// LoginResponse is a response to a Login request.
type LoginResponse struct {
	LoginResult struct {
		ResultHeader
		Version   string `xml:"TechVersion"`
		Firstname string `xml:"Vorname"`
		Lastname  string `xml:"Nachname"`
	} `xml:"Body>LoginResponse>LoginResult"`
}

// ResultHeader returns the ResultHeader address in the response.
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

type getDevicesDevices struct {
	ID          uint32 `xml:"GeraetId"`
	Name        string `xml:"GeraetName"`
	HasError    bool   `xml:"HatFehler"`
	IsConnected bool   `xml:"IstVerbunden"`
}

type getDevicesLocation struct {
	ID          uint32              `xml:"AnlageId"`
	Name        string              `xml:"AnlageName"`
	Devices     []getDevicesDevices `xml:"GeraeteListe>GeraetV2"`
	HasError    bool                `xml:"HatFehler"`
	IsConnected bool                `xml:"IstVerbunden"`
}

// GetDevicesResponse is a response to a GetDevices request.
type GetDevicesResponse struct {
	GetDevicesResult struct {
		ResultHeader
		Locations []getDevicesLocation `xml:"AnlageListe>AnlageV2"`
	} `xml:"Body>GetDevicesResponse>GetDevicesResult"`
}

// ResultHeader returns the ResultHeader address in the response.
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
				LocationID:   location.ID,
				LocationName: location.Name,
				DeviceID:     device.ID,
				DeviceName:   device.Name,
				HasError:     location.HasError || device.HasError,
				IsConnected:  location.IsConnected && device.IsConnected,
				Attributes:   map[AttrID]*Value{},
				Timesheets:   map[TimesheetID]map[string]TimeslotSlice{},
			})
		}
	}

	// Make sure all devices are sorted
	sort.SliceStable(v.Devices, func(i, j int) bool {
		if v.Devices[i].LocationID == v.Devices[j].LocationID {
			return v.Devices[i].DeviceID < v.Devices[j].DeviceID
		}
		return v.Devices[i].LocationID < v.Devices[j].LocationID
	})

	return nil
}

//
// RequestRefreshStatus
//

// RequestRefreshStatusResponse is a response to a
// RequestRefreshStatus request.
type RequestRefreshStatusResponse struct {
	RequestRefreshStatusResult struct {
		ResultHeader
		Status int `xml:"Status"`
	} `xml:"Body>RequestRefreshStatusResponse>RequestRefreshStatusResult"`
}

// ResultHeader returns the ResultHeader address in the response.
func (r *RequestRefreshStatusResponse) ResultHeader() *ResultHeader {
	return &r.RequestRefreshStatusResult.ResultHeader
}

// RequestRefreshStatus launches the Vitotrol™ RequestRefreshStatus
// request to follow the status of the RefreshData request matching
// the passed refresh ID. Use RefreshDataWait instead.
func (v *Session) RequestRefreshStatus(refreshID string) (int, error) {
	var resp RequestRefreshStatusResponse
	err := v.sendRequest("RequestRefreshStatus",
		"<RequestRefreshStatus><AktualisierungsId>"+
			refreshID+
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

// RequestWriteStatusResponse is a response to a RequestWriteStatus request.
type RequestWriteStatusResponse struct {
	RequestWriteStatusResult struct {
		ResultHeader
		Status int `xml:"Status"`
	} `xml:"Body>RequestWriteStatusResponse>RequestWriteStatusResult"`
}

// ResultHeader returns the ResultHeader address in the response.
func (r *RequestWriteStatusResponse) ResultHeader() *ResultHeader {
	return &r.RequestWriteStatusResult.ResultHeader
}

// RequestWriteStatus launches the Vitotrol™ RequestWriteStatus
// request to follow the status of the WriteData request matching
// the passed refresh ID. Use WriteDataWait instead.
func (v *Session) RequestWriteStatus(refreshID string) (int, error) {
	var resp RequestWriteStatusResponse
	err := v.sendRequest("RequestWriteStatus",
		"<RequestWriteStatus><AktualisierungsId>"+
			refreshID+
			"</AktualisierungsId></RequestWriteStatus>",
		&resp)
	if err != nil {
		return 0, err
	}

	return resp.RequestWriteStatusResult.Status, nil
}
