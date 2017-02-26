package vitotrol

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Device represents one Vitotrol™ device (a priori a boiler)
type Device struct {
	LocationID   uint32 // Vitotrol™ ID of location (AnlageId field)
	LocationName string // location name (AnlageName field)
	DeviceID     uint32 // Vitotrol™ ID of device (GeraetId field)
	DeviceName   string // device name (GeraetName field)
	HasError     bool   // ORed HatFehler field of Device & Location
	IsConnected  bool   // IstVerbunden field of Device

	// cache of last read attributes values (filled by GetData)
	Attributes map[AttrID]*Value
	// cache of last read timesheets data (filled by GetTimesheetData)
	Timesheets map[TimesheetID]map[string]TimeslotSlice
	// cache of last read errors (filled by GetErrorHistory)
	Errors []ErrorHistoryEvent
}

// FormatAttributes displays informations about selected
// attributes. Displays information about all known attributes when a
// nil slice is passed.
func (d *Device) FormatAttributes(attrs []AttrID) string {
	buf := bytes.NewBuffer(nil)

	pConcatFun := func(attrID AttrID, pValue *Value) {
		pRef := AttributesRef[attrID]
		if pRef == nil {
			buf.WriteString(
				fmt.Sprintf("%d: %s@%s\n", attrID, pValue.Value, pValue.Time))
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

	for _, attrID := range attrs {
		pConcatFun(attrID, d.Attributes[attrID])
	}

	return buf.String()
}

func (d *Device) buildBody(soapAction string, reqBody string) string {
	return fmt.Sprintf(`<%s>
<GeraetId>%d</GeraetId>
<AnlageId>%d</AnlageId>
%s
</%[1]s>`, soapAction, d.DeviceID, d.LocationID, reqBody)
}

func (d *Device) sendRequest(v *Session, soapAction string, reqBody string, respBody HasResultHeader) error {
	return v.sendRequest(soapAction, d.buildBody(soapAction, reqBody), respBody)
}

//
// GetData
//

type GetDataValue struct {
	ID    uint16 `xml:"DatenpunktId"`
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

func makeDatenpunktIDs(attrIDs []AttrID) string {
	body := bytes.NewBufferString("<DatenpunktIds>")

	for _, id := range attrIDs {
		body.WriteString(fmt.Sprintf("<int>%d</int>", id))
	}

	body.WriteString("</DatenpunktIds>")

	return body.String()
}

// GetData launches the Vitotrol™ GetData request. Populates the
// internal cache before returning (see Attributes field).
func (d *Device) GetData(v *Session, attrIDs []AttrID) error {
	var resp GetDataResponse
	err := d.sendRequest(v, "GetData", makeDatenpunktIDs(attrIDs), &resp)
	if err != nil {
		return err
	}

	// On met en cache
	for _, respValue := range resp.GetDataResult.Values {
		d.Attributes[AttrID(respValue.ID)] = &Value{
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
	RefreshID string `xml:"AktualisierungsId"`
}

func (r *WriteDataResponse) ResultHeader() *ResultHeader {
	return &r.WriteDataResult.ResultHeader
}

// WriteData launches the Vitotrol™ WriteData request and returns the
// "refresh ID" sent back by the server. Use WriteDataWait instead.
func (d *Device) WriteData(v *Session, attrID AttrID, value string) (string, error) {
	var resp WriteDataResponse
	err := d.sendRequest(v, "WriteData",
		fmt.Sprintf("<DatapointId>%d</DatapointId><Wert>%s</Wert>", attrID, value),
		&resp)
	if err != nil {
		return "", err
	}
	return resp.WriteDataResult.RefreshID, nil
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
func (d *Device) WriteDataWait(v *Session, attrID AttrID, value string) (<-chan error, error) {
	refreshID, err := d.WriteData(v, attrID, value)
	if err != nil {
		return nil, err
	}

	ch := make(chan error)

	go waitAsyncStatus(v, refreshID, ch, (*Session).RequestWriteStatus,
		WriteDataWaitDuration, WriteDataWaitMinDuration)

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
	RefreshID string `xml:"AktualisierungsId"`
}

func (r *RefreshDataResponse) ResultHeader() *ResultHeader {
	return &r.RefreshDataResult.ResultHeader
}

// RefreshData launches the Vitotrol™ RefreshData request and returns
// the "refresh ID" sent back by the server. Use RefreshDataWait
// instead.
func (d *Device) RefreshData(v *Session, attrIDs []AttrID) (string, error) {
	var resp RefreshDataResponse
	err := d.sendRequest(v, "RefreshData", makeDatenpunktIDs(attrIDs), &resp)
	if err != nil {
		return "", err
	}
	return resp.RefreshDataResult.RefreshID, nil
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
func (d *Device) RefreshDataWait(v *Session, attrIDs []AttrID) (<-chan error, error) {
	refreshID, err := d.RefreshData(v, attrIDs)
	if err != nil {
		return nil, err
	}

	ch := make(chan error)

	go waitAsyncStatus(v, refreshID, ch, (*Session).RequestRefreshStatus,
		RefreshDataWaitDuration, RefreshDataWaitMinDuration)

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
	ID       uint16    `xml:"SchaltsatzDaten>DatenpunktID"`
	DaySlots []DaySlot `xml:"SchaltsatzDaten>Schaltzeiten>Schaltzeit"`
}

func (r *GetTimesheetDataResponse) ResultHeader() *ResultHeader {
	return &r.GetTimesheetDataResult.ResultHeader
}

// GetTimesheetData launches the Vitotrol™ GetTimesheetData
// request. Populates the internal cache before returning (see
// Timesheets field).
func (d *Device) GetTimesheetData(v *Session, id TimesheetID) error {
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

//
// WriteTimesheetData
//

type WriteTimesheetDataResponse struct {
	WriteTimesheetDataResult WriteTimesheetDataResult `xml:"Body>WriteTimesheetDataResponse>WriteTimesheetDataResult"`
}

type WriteTimesheetDataResult struct {
	ResultHeader
	RefreshID string `xml:"AktualisierungsId"`
}

func (r *WriteTimesheetDataResponse) ResultHeader() *ResultHeader {
	return &r.WriteTimesheetDataResult.ResultHeader
}

var timesheetDays = []string{
	"MON",
	"TUE",
	"WED",
	"THU",
	"FRI",
	"SAT",
	"SUN",
}
var timesheetDaysSet = func() map[string]struct{} {
	tss := make(map[string]struct{}, 7)
	for _, day := range timesheetDays {
		tss[day] = struct{}{}
	}
	return tss
}()

// WriteTimesheetData launches the Vitotrol™ WriteTimesheetData
// request and returns the "refresh ID" sent back by the server. Does
// not populate the internal cache before returning (Timesheets
// field), use WriteTimesheetDataWait instead.
func (d *Device) WriteTimesheetData(v *Session, id TimesheetID, data map[string]TimeslotSlice) (string, error) {
	buf := bytes.NewBufferString(
		`<SchaltzeitTyp>1</SchaltzeitTyp>` +
			`<DatenpunktId>`)
	buf.WriteString(strconv.Itoa(int(id)))
	buf.WriteString(`</DatenpunktId>` +
		`<Schaltzeiten>`)

	preDays := make(map[string]*bytes.Buffer, 7)
	for day, daySlots := range data {
		day = strings.ToUpper(day)
		_, ok := timesheetDaysSet[day]
		if !ok {
			return "", fmt.Errorf("Bad timesheet day `%s'", day)
		}

		sort.Sort(daySlots) // sort slots in place

		tmpBuf := bytes.NewBuffer(nil)
		preDays[day] = tmpBuf

		for idx, slot := range daySlots {
			tmpBuf.WriteString(fmt.Sprintf(
				`<Schaltzeit>`+
					`<Wochentag>%s</Wochentag>`+
					`<ZeitVon>%04d</ZeitVon>`+
					`<ZeitBis>%04d</ZeitBis>`+
					`<Wert>1</Wert>`+
					`<Position>%d</Position>`+
					`</Schaltzeit>`,
				day, slot.From, slot.To, idx))
		}
	}

	// Write sorted days
	for _, day := range timesheetDays {
		tmpBuf := preDays[day]
		if tmpBuf != nil {
			buf.Write(tmpBuf.Bytes())
		}
	}

	buf.WriteString(`</Schaltzeiten>`)

	// Oddly, WriteTimesheetData has a nested layer SchaltsatzData
	// before GeraetId and AnlageId fields, so use the
	// Session.sendRequest method instead of Device.sendRequest
	var resp WriteTimesheetDataResponse
	err := v.sendRequest("WriteTimesheetData",
		`<WriteTimesheetData>`+
			d.buildBody("SchaltsatzData", buf.String())+
			`</WriteTimesheetData>`,
		&resp)
	if err != nil {
		return "", err
	}
	return resp.WriteTimesheetDataResult.RefreshID, nil
}

var (
	// WriteTimesheetDataWaitDuration defines the duration to wait in
	// WriteTimesheetDataWait after the WriteTimesheetData call before
	// calling RequestWriteStatus for the first time. After that first
	// call, the next pause duration will be divided by 4 and so on.
	WriteTimesheetDataWaitDuration = 8 * time.Second
	// WriteTimesheetDataWaitMinDuration defines the minimal duration of pauses
	// between RequestWriteStatus calls.
	WriteTimesheetDataWaitMinDuration = 1 * time.Second
)

// WriteTimesheetDataWait launches the Vitotrol™ WriteTimesheetData
// request and returns a channel on which the final error
// (asynchronous one) will be received (nil if the data has been
// correctly written).
//
// If an error occurs during the WriteTimesheetData call (synchronous
// one), a nil channel is returned with an error.
func (d *Device) WriteTimesheetDataWait(v *Session, id TimesheetID, data map[string]TimeslotSlice) (<-chan error, error) {
	refreshID, err := d.WriteTimesheetData(v, id, data)
	if err != nil {
		return nil, err
	}

	ch := make(chan error)

	go waitAsyncStatus(v, refreshID, ch, (*Session).RequestWriteStatus,
		WriteTimesheetDataWaitDuration, WriteTimesheetDataWaitMinDuration)

	return ch, nil
}

func waitAsyncStatus(v *Session, refreshID string, ch chan error,
	requestStatus func(*Session, string) (int, error),
	waitFirstDuration, waitminDuration time.Duration) {
	start := time.Now()
	// Waiting availability of data, yes *8* seconds the first time :(
	for wait := waitFirstDuration; true; {
		time.Sleep(wait)

		status, err := requestStatus(v, refreshID)
		if err != nil {
			ch <- err
			break
		}

		if status == 4 {
			break
		}

		wait /= 4
		if wait < waitminDuration {
			wait = waitminDuration
		}

		if v.Debug {
			log.Printf("waitAsyncStatus(%s): status %d, wait %d secs...\n",
				refreshID, status, wait/time.Second)
		}
	}
	if v.Debug {
		log.Printf("waitAsyncStatus(%s) done in %s",
			refreshID, time.Now().Sub(start))
	}
	close(ch)
}
