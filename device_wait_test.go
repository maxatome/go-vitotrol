package vitotrol

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type testAction struct {
	expectedRequest interface{}
	serverResponse  string
}

func testSendRequestAnyMulti(assert *assert.Assertions,
	sendReqs func(v *Session, d *Device) bool,
	actions map[string]*testAction, testName string) bool {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			soapActionUrl := r.Header.Get("SOAPAction")
			if !assert.NotEmpty(soapActionUrl,
				"%s: SOAPAction header found", testName) {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}
			soapAction := soapActionUrl[strings.LastIndex(soapActionUrl, "/")+1:]
			pAction := actions[soapAction]
			if !assert.NotNil(pAction,
				"%s: SOAPAction header `%s' matches one expected action",
				testName, soapAction) {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}

			assert.Equal("text/xml; charset=utf-8", r.Header.Get("Content-Type"),
				"%s: Content-Type header matches", testName)

			if cookie := r.Header.Get("Cookie"); cookie != "" {
				w.Header().Add("Set-Cookie", cookie)
			}

			// Extract request body in the same struct type as the expectedRequest
			recvReq := virginInstance(pAction.expectedRequest)
			if !extractRequestBody(assert, r, recvReq, testName) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			assert.Equal(pAction.expectedRequest, recvReq, "%s: request OK", testName)

			// Send response
			fmt.Fprintln(w, respHeader+pAction.serverResponse+respFooter)
		}))
	defer ts.Close()

	MainURL = ts.URL
	v := &Session{
		Devices: []Device{
			{
				DeviceId:   testDeviceId,
				LocationId: testLocationId,
				Attributes: map[AttrId]*Value{},
				Timesheets: map[TimesheetId]map[string]TimeslotSlice{},
			},
		},
	}
	return sendReqs(v, &v.Devices[0])
}

//
// WriteDataWait
//

func TestWriteDataWait(t *testing.T) {
	assert := assert.New(t)

	// No problem
	testSendRequestAnyMulti(assert,
		func(v *Session, d *Device) bool {
			WriteDataWaitDuration = 0
			WriteDataWaitMinDuration = 0
			ch, err := d.WriteDataWait(v, writeDataTestId, writeDataTestValue)
			if !assert.Nil(err) {
				return false
			}
			timeoutTicker := time.NewTicker(100 * time.Millisecond)
			defer timeoutTicker.Stop()

			select {
			case err = <-ch:
				return assert.Nil(err)
			case <-timeoutTicker.C:
				return assert.Fail("TIMEOUT!")
			}
		},
		map[string]*testAction{
			"WriteData": {
				expectedRequest: writeDataTest.expectedRequest,
				serverResponse: intoDeviceResponse(
					"WriteData", writeDataTest.serverResponse),
			},
			"RequestWriteStatus": &requestWriteStatusTest,
		},
		"WriteDataWait")

	// Error during WriteData
	testSendRequestAnyMulti(assert,
		func(v *Session, d *Device) bool {
			WriteDataWaitDuration = 0
			ch, err := d.WriteDataWait(v, writeDataTestId, writeDataTestValue)
			assert.NotNil(err)
			return assert.Nil(ch)
		},
		map[string]*testAction{
			"WriteData": {
				expectedRequest: writeDataTest.expectedRequest,
				serverResponse:  `<bad XML>`,
			},
			"RequestWriteStatus": &requestWriteStatusTest,
		},
		"WriteDataWait, error during WriteData")

	// Error during RequestWriteStatus
	testSendRequestAnyMulti(assert,
		func(v *Session, d *Device) bool {
			ch, err := d.WriteDataWait(v, writeDataTestId, writeDataTestValue)
			if !assert.Nil(err) {
				return false
			}
			timeoutTicker := time.NewTicker(100 * time.Millisecond)
			defer timeoutTicker.Stop()

			select {
			case err = <-ch:
				return assert.NotNil(err)
			case <-timeoutTicker.C:
				return assert.Fail("TIMEOUT!")
			}
		},
		map[string]*testAction{
			"WriteData": {
				expectedRequest: writeDataTest.expectedRequest,
				serverResponse: intoDeviceResponse(
					"WriteData", writeDataTest.serverResponse),
			},
			"RequestWriteStatus": {
				expectedRequest: requestWriteStatusTest.expectedRequest,
				serverResponse:  `<bad XML>`,
			},
		},
		"WriteDataWait, error during RequestWriteStatus")
}

//
// RefreshDataWait
//

func TestRefreshDataWait(t *testing.T) {
	assert := assert.New(t)

	// No problem
	testSendRequestAnyMulti(assert,
		func(v *Session, d *Device) bool {
			RefreshDataWaitDuration = 0
			RefreshDataWaitMinDuration = 0
			ch, err := d.RefreshDataWait(v, refreshDataTestIds)
			if !assert.Nil(err) {
				return false
			}
			timeoutTicker := time.NewTicker(100 * time.Millisecond)
			defer timeoutTicker.Stop()

			select {
			case err = <-ch:
				return assert.Nil(err)
			case <-timeoutTicker.C:
				return assert.Fail("TIMEOUT!")
			}
		},
		map[string]*testAction{
			"RefreshData": {
				expectedRequest: refreshDataTest.expectedRequest,
				serverResponse: intoDeviceResponse(
					"RefreshData", refreshDataTest.serverResponse),
			},
			"RequestRefreshStatus": &requestRefreshStatusTest,
		},
		"RefreshDataWait")

	// Error during RefreshData
	testSendRequestAnyMulti(assert,
		func(v *Session, d *Device) bool {
			RefreshDataWaitDuration = 0
			ch, err := d.RefreshDataWait(v, refreshDataTestIds)
			assert.NotNil(err)
			return assert.Nil(ch)
		},
		map[string]*testAction{
			"RefreshData": {
				expectedRequest: refreshDataTest.expectedRequest,
				serverResponse:  `<bad XML>`,
			},
			"RequestRefreshStatus": &requestRefreshStatusTest,
		},
		"RefreshDataWait, error during RefreshData")

	// Error during RequestRefreshStatus
	testSendRequestAnyMulti(assert,
		func(v *Session, d *Device) bool {
			ch, err := d.RefreshDataWait(v, refreshDataTestIds)
			if !assert.Nil(err) {
				return false
			}
			timeoutTicker := time.NewTicker(100 * time.Millisecond)
			defer timeoutTicker.Stop()

			select {
			case err = <-ch:
				return assert.NotNil(err)
			case <-timeoutTicker.C:
				return assert.Fail("TIMEOUT!")
			}
		},
		map[string]*testAction{
			"RefreshData": {
				expectedRequest: refreshDataTest.expectedRequest,
				serverResponse: intoDeviceResponse(
					"RefreshData", refreshDataTest.serverResponse),
			},
			"RequestRefreshStatus": {
				expectedRequest: requestRefreshStatusTest.expectedRequest,
				serverResponse:  `<bad XML>`,
			},
		},
		"RefreshDataWait, error during RequestRefreshStatus")
}
