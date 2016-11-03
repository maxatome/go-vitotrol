package vitotrol

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// Device represents one Vitotrol™ device (a priori a boiler)
type Device struct {
	LocationId   uint32 // Vitotrol™ ID of location (AnlageId field)
	LocationName string // location name (AnlageName field)
	DeviceId     uint32 // Vitotrol™ ID of device (GeraetId field)
	DeviceName   string // device name (GeraetName field)
	HasError     bool   // ORed HatFehler field of Device & Location
	IsConnected  bool   // IstVerbunden field of Device

	// cache of last read attributes values (filled by GetData)
	Attributes map[AttrId]*Value
	// cache of last read timesheets data (filled by GetTimesheetData)
	Timesheets map[TimesheetId]map[string]TimeslotSlice
	// cache of last read errors (filled by GetErrorHistory)
	Errors []ErrorHistoryEvent
}

// FormatAttributes displays informations about selected
// attributes. Displays information about all known attributes when a
// nil slice is passed.
func (d *Device) FormatAttributes(attrs []AttrId) string {
	buf := bytes.NewBuffer(nil)

	pConcatFun := func(attrId AttrId, pValue *Value) {
		pRef := AttributesRef[attrId]
		if pRef == nil {
			buf.WriteString(
				fmt.Sprintf("%d: %s@%s\n", attrId, pValue.Value, pValue.Time))
		} else if pValue != nil {
			humanValue, err := pRef.Type.Vitodata2HumanValue(pValue.Value)
			if err != nil {
				humanValue = fmt.Sprintf("unknown-value<%s>", pValue.Value)
			}
			buf.WriteString(
				fmt.Sprintf("%s: %s@%s (%s)\n",
					pRef.Name, humanValue, pValue.Time, pRef.Doc))
		} else {
			buf.WriteString(
				fmt.Sprintf("%s: uninitialized (%s)\n", pRef.Name, pRef.Doc))
		}
	}

	for _, attrId := range attrs {
		pConcatFun(attrId, d.Attributes[attrId])
	}

	return buf.String()
}

func (d *Device) sendRequest(v *Session, soapAction string, reqBody string, respBody HasResultHeader) error {
	return v.sendRequest(soapAction, fmt.Sprintf(`<%s>
<GeraetId>%d</GeraetId>
<AnlageId>%d</AnlageId>
%s
</%[1]s>`, soapAction, d.DeviceId, d.LocationId, reqBody), respBody)
}

//
// GetData
//

type GetDataValue struct {
	Id    uint16 `xml:"DatenpunktId"`
	Value string `xml:"Wert"`
	Time  Time   `xml:"Zeitstempel"`
}

type GetDataResponse struct {
	GetDataResult GetDataResult `xml:"Body>GetDataResponse>GetDataResult"`
}

type GetDataResult struct {
	ResultHeader
	Values []GetDataValue `xml:"DatenwerteListe>WerteListe"`
}

func (r *GetDataResponse) ResultHeader() *ResultHeader {
	return &r.GetDataResult.ResultHeader
}

func makeDatenpunktIds(attrIds []AttrId) string {
	body := bytes.NewBufferString("<DatenpunktIds>")

	for _, id := range attrIds {
		body.WriteString(fmt.Sprintf("<int>%d</int>", id))
	}

	body.WriteString("</DatenpunktIds>")

	return body.String()
}

// GetData launches the Vitotrol™ GetData request. Populates the
// internal cache before returning (see Attributes field).
func (d *Device) GetData(v *Session, attrIds []AttrId) error {
	var resp GetDataResponse
	err := d.sendRequest(v, "GetData", makeDatenpunktIds(attrIds), &resp)
	if err != nil {
		return err
	}

	// On met en cache
	for _, respValue := range resp.GetDataResult.Values {
		d.Attributes[AttrId(respValue.Id)] = &Value{
			Time:  respValue.Time,
			Value: respValue.Value,
		}
	}

	return nil
}

//
// WriteData
//

type WriteDataResponse struct {
	WriteDataResult WriteDataResult `xml:"Body>WriteDataResponse>WriteDataResult"`
}

type WriteDataResult struct {
	ResultHeader
	RefreshId string `xml:"AktualisierungsId"`
}

func (r *WriteDataResponse) ResultHeader() *ResultHeader {
	return &r.WriteDataResult.ResultHeader
}

// WriteData launches the Vitotrol™ WriteData request and returns the
// "refresh ID" sent back by the server. Use WriteDataWait instead.
func (d *Device) WriteData(v *Session, attrId AttrId, value string) (string, error) {
	var resp WriteDataResponse
	err := d.sendRequest(v, "WriteData",
		fmt.Sprintf("<DatapointId>%d</DatapointId><Wert>%s</Wert>", attrId, value),
		&resp)
	if err != nil {
		return "", err
	}
	return resp.WriteDataResult.RefreshId, nil
}

var (
	// WriteDataWaitDuration defines the duration to wait in
	// WriteDataWait after the WriteData call before calling
	// RequestWriteStatus for the first time. After that first call, the
	// next pause duration will be divided by 4 and so on.
	WriteDataWaitDuration = 4 * time.Second
	// WriteDataWaitMinDuration defines the minimal duration of pauses
	// between RequestWriteStatus calls.
	WriteDataWaitMinDuration = 1 * time.Second
)

// WriteDataWait launches the Vitotrol™ WriteData request and returns
// a channel on which the final error (asynchronous one) will be
// received (nil if the data has been correctly written).
//
// If an error occurs during the WriteData call (synchronous one), a
// nil channel is returned with an error.
func (d *Device) WriteDataWait(v *Session, attrId AttrId, value string) (<-chan error, error) {
	refreshId, err := d.WriteData(v, attrId, value)
	if err != nil {
		return nil, err
	}

	ch := make(chan error)

	go func() {
		start := time.Now()
		// Waiting for update to be done
		for wait := WriteDataWaitDuration; true; {
			time.Sleep(wait)

			status, err := v.RequestWriteStatus(refreshId)
			if err != nil {
				ch <- err
				break
			}

			if status == 4 {
				break
			}

			wait /= 4
			if wait < WriteDataWaitMinDuration {
				wait = WriteDataWaitMinDuration
			}

			if v.Debug {
				log.Printf("WriteDataWait: status %d, wait %d secs...\n",
					status, wait/time.Second)
			}
		}
		if v.Debug {
			log.Println("WriteDataWait done in", time.Now().Sub(start))
		}
		close(ch)
	}()

	return ch, nil
}

//
// RefreshData
//

type RefreshDataResponse struct {
	RefreshDataResult RefreshDataResult `xml:"Body>RefreshDataResponse>RefreshDataResult"`
}
type RefreshDataResult struct {
	ResultHeader
	RefreshId string `xml:"AktualisierungsId"`
}

func (r *RefreshDataResponse) ResultHeader() *ResultHeader {
	return &r.RefreshDataResult.ResultHeader
}

// RefreshData launches the Vitotrol™ RefreshData request and returns
// the "refresh ID" sent back by the server. Use RefreshDataWait
// instead.
func (d *Device) RefreshData(v *Session, attrIds []AttrId) (string, error) {
	var resp RefreshDataResponse
	err := d.sendRequest(v, "RefreshData", makeDatenpunktIds(attrIds), &resp)
	if err != nil {
		return "", err
	}
	return resp.RefreshDataResult.RefreshId, nil
}

var (
	// RefreshDataWaitDuration defines the duration to wait in
	// RefreshDataWait after the RefreshData call before calling
	// RequestRefreshStatus for the first time. After that first call, the
	// next pause duration will be divided by 4 and so on.
	RefreshDataWaitDuration = 8 * time.Second
	// RefreshDataWaitMinDuration defines the minimal duration of pauses
	// between RequestRefreshStatus calls.
	RefreshDataWaitMinDuration = 1 * time.Second
)

// RefreshDataWait launches the Vitotrol™ RefreshData request and
// returns a channel on which the final error (asynchronous one) will
// be received (nil if the data has been correctly written).
//
// If an error occurs during the RefreshData call (synchronous one), a
// nil channel is returned with an error.
func (d *Device) RefreshDataWait(v *Session, attrIds []AttrId) (<-chan error, error) {
	refreshId, err := d.RefreshData(v, attrIds)
	if err != nil {
		return nil, err
	}

	ch := make(chan error)

	go func() {
		start := time.Now()
		// Waiting availability of data, yes *8* seconds the first time :(
		for wait := RefreshDataWaitDuration; true; {
			time.Sleep(wait)

			status, err := v.RequestRefreshStatus(refreshId)
			if err != nil {
				ch <- err
				break
			}

			if status == 4 {
				break
			}

			wait /= 4
			if wait < RefreshDataWaitMinDuration {
				wait = RefreshDataWaitMinDuration
			}

			if v.Debug {
				log.Printf("RefreshDataWait: status %d, wait %d secs...\n",
					status, wait/time.Second)
			}
		}
		if v.Debug {
			log.Println("RefreshDataWait done in", time.Now().Sub(start))
		}
		close(ch)
	}()

	return ch, nil
}

//
// GetErrorHistory
//

type ErrorHistoryEvent struct {
	Error    string `xml:"FehlerCode"`
	Message  string `xml:"FehlerMeldung"`
	Time     Time   `xml:"Zeitstempel"`
	IsActive bool   `xml:"FehlerIstAktiv"`
}

func (e *ErrorHistoryEvent) String() string {
	var isActive string
	if e.IsActive {
		isActive = " *ACTIVE*"
	}
	return fmt.Sprintf("%s@%s = %s%s", e.Error, e.Time, e.Message, isActive)
}

type GetErrorHistoryResponse struct {
	GetErrorHistoryResult GetErrorHistoryResult `xml:"Body>GetErrorHistoryResponse>GetErrorHistoryResult"`
}
type GetErrorHistoryResult struct {
	ResultHeader
	Events []ErrorHistoryEvent `xml:"FehlerListe>FehlerHistorie"`
}

func (r *GetErrorHistoryResponse) ResultHeader() *ResultHeader {
	return &r.GetErrorHistoryResult.ResultHeader
}

// GetErrorHistory launches the Vitotrol™ GetErrorHistory
// request. Populates the internal cache before returning (see Errors
// field).
func (d *Device) GetErrorHistory(v *Session) error {
	var resp GetErrorHistoryResponse
	err := d.sendRequest(v, "GetErrorHistory", "<Culture>fr-fr</Culture>", &resp)
	if err != nil {
		return err
	}

	d.Errors = resp.GetErrorHistoryResult.Events
	return nil
}

//
// GetTimesheetData
//

type DaySlot struct {
	Day  string `xml:"Wochentag"`
	From uint16 `xml:"ZeitVon"`
	To   uint16 `xml:"ZeitBis"`
}

type GetTimesheetDataResponse struct {
	GetTimesheetDataResult GetTimesheetDataResult `xml:"Body>GetTimesheetDataResponse>GetTimesheetDataResult"`
}

type GetTimesheetDataResult struct {
	ResultHeader
	Id       uint16    `xml:"SchaltsatzDaten>DatenpunktId"`
	DaySlots []DaySlot `xml:"SchaltsatzDaten>Schaltzeiten>Schaltzeit"`
}

func (r *GetTimesheetDataResponse) ResultHeader() *ResultHeader {
	return &r.GetTimesheetDataResult.ResultHeader
}

// GetTimesheetData launches the Vitotrol™ GetTimesheetData
// request. Populates the internal cache before returning (see
// Timesheets field).
func (d *Device) GetTimesheetData(v *Session, id TimesheetId) error {
	var resp GetTimesheetDataResponse
	err := d.sendRequest(v, "GetTimesheetData",
		fmt.Sprintf("<DatenpunktId>%d</DatenpunktId>", id), &resp)
	if err != nil {
		return err
	}

	timesheet := make(map[string]TimeslotSlice)

	for _, slot := range resp.GetTimesheetDataResult.DaySlots {
		day := strings.ToLower(slot.Day)
		timesheet[day] = append(timesheet[day], Timeslot{
			From: slot.From,
			To:   slot.To,
		})
	}

	// To be clean, sort slots for each day
	for _, daySlots := range timesheet {
		sort.Sort(daySlots)
	}

	d.Timesheets[id] = timesheet

	return nil
}
