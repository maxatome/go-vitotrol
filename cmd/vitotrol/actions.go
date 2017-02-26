package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/maxatome/go-vitotrol"
	"io/ioutil"
	"strings"
)

func checkAttributeAccess(attrName string, reqAccess vitotrol.AttrAccess) (vitotrol.AttrID, error) {
	attrID, ok := vitotrol.AttributesNames2IDs[attrName]
	if !ok {
		return vitotrol.NoAttr, fmt.Errorf("unknown attribute `%s'", attrName)
	}

	if (vitotrol.AttributesRef[attrID].Access & reqAccess) != reqAccess {
		return vitotrol.NoAttr, fmt.Errorf("attribute `%s' is not %s",
			attrName, vitotrol.AccessToStr[reqAccess])
	}

	return attrID, nil
}

func existTimesheetName(tsName string) (vitotrol.TimesheetID, error) {
	tID, ok := vitotrol.TimesheetsNames2IDs[tsName]
	if !ok {
		return 0, fmt.Errorf("unknown timesheet `%s'", tsName)
	}
	return tID, nil
}

// An Action can be typically called by main to do a job.
type Action interface {
	// NeedAuth tells whether this Action needs an authentication or not.
	NeedAuth() bool

	// Do executes the action.
	Do(pOptions *Options, params []string) error
}

var actions = map[string]Action{
	"list":          &listAction{},
	"get":           &getAction{},
	"rget":          &getAction{rget: true},
	"set":           &setAction{},
	"errors":        &errorsAction{},
	"timesheet":     &timesheetAction{},
	"set_timesheet": &setTimesheetAction{},
}

type authAction struct {
	v       *vitotrol.Session
	d       *vitotrol.Device
	options *Options
}

func (a *authAction) initVitotrol(pOptions *Options) error {
	v := &vitotrol.Session{
		Debug: pOptions.debug,
	}

	err := v.Login(pOptions.login, pOptions.password)
	if err != nil {
		return fmt.Errorf("Login failed: %s", err)
	}

	err = v.GetDevices()
	if err != nil {
		return fmt.Errorf("GetDevices failed: %s", err)
	}
	if len(v.Devices) == 0 {
		return errors.New("No device found")
	}

	pDevice := &v.Devices[0]
	if !pDevice.IsConnected {
		return fmt.Errorf("Device %s@%s is not connected",
			pDevice.DeviceName, pDevice.LocationName)
	}

	if pOptions.verbose {
		fmt.Printf("Working with device %s@%s\n",
			pDevice.DeviceName, pDevice.LocationName)
	}

	a.v = v
	a.d = pDevice
	a.options = pOptions

	return nil
}

func (a *authAction) NeedAuth() bool {
	return true
}

// listAction implements the "list" action.
type listAction struct{}

func (a *listAction) NeedAuth() bool {
	return false
}

func (a *listAction) Do(pOptions *Options, params []string) error {
	if len(params) == 0 || params[0] == "attrs" {
		for _, pAttrRef := range vitotrol.AttributesRef {
			fmt.Println(pAttrRef)
		}
		return nil
	}

	if params[0] == "timesheets" {
		for _, pTimesheetRef := range vitotrol.TimesheetsRef {
			fmt.Println(pTimesheetRef)
		}
		return nil
	}

	return fmt.Errorf(
		"`list' action allows `attrs' or `timesheets' params, not `%s'",
		params[0])
}

// getAction implements the "get" and "rget" actions.
type getAction struct {
	authAction
	rget bool
}

func (a *getAction) Do(pOptions *Options, params []string) error {
	if len(params) == 0 {
		return errors.New("at least one PARAM is missing")
	}

	var attrs []vitotrol.AttrID
	// Special case -> all attributes
	if len(params) == 1 && params[0] == "all" {
		attrs = vitotrol.Attributes
	} else {
		attrs = make([]vitotrol.AttrID, len(params))
		var err error
		for idx, attrName := range params {
			attrs[idx], err = checkAttributeAccess(attrName, vitotrol.ReadOnly)
			if err != nil {
				return err
			}
		}
	}

	err := a.initVitotrol(pOptions)
	if err != nil {
		return err
	}

	if a.rget {
		ch, err := a.d.RefreshDataWait(a.v, attrs)
		if err != nil {
			return fmt.Errorf("RefreshData error: %s", err)
		}

		if err = <-ch; err != nil {
			return fmt.Errorf("RefreshData failed: %s", err)
		}
	}

	err = a.d.GetData(a.v, attrs)
	if err != nil {
		return fmt.Errorf("GetData error: %s", err)
	}

	fmt.Print(a.d.FormatAttributes(attrs))
	return nil
}

// setAction implements the "set" action.
type setAction struct {
	authAction
}

func (a *setAction) Do(pOptions *Options, params []string) error {
	if len(params) == 0 || (len(params)&1) != 0 {
		return errors.New("PARAMS must be a list of pairs: ATTR_NAME, VALUE")
	}

	attrsValues := make(map[vitotrol.AttrID]string, len(params)/2)
	for idx := 0; idx < len(params); idx += 2 {
		attrID, err := checkAttributeAccess(params[idx], vitotrol.WriteOnly)
		if err != nil {
			return err
		}

		value, err :=
			vitotrol.AttributesRef[attrID].Type.Human2VitodataValue(params[idx+1])
		if err != nil {
			return fmt.Errorf("value `%s' of attribute %s is invalid: %s",
				params[idx+1], params[idx], err)
		}

		attrsValues[attrID] = value
	}

	err := a.initVitotrol(pOptions)
	if err != nil {
		return err
	}

	// Set them all
	for attrID, value := range attrsValues {
		ch, err := a.d.WriteDataWait(a.v, attrID, value)
		if err != nil {
			return fmt.Errorf("WriteData error: %s", err)
		}

		if err = <-ch; err != nil {
			return fmt.Errorf("WriteData failed: %s", err)
		}

		if pOptions.verbose {
			fmt.Printf("%s attribute successfully set to `%s'\n",
				vitotrol.AttributesRef[attrID].Name, value)
		}
	}

	return nil
}

// errorsAction implements the "errors" action.
type errorsAction struct {
	authAction
}

func (a *errorsAction) Do(pOptions *Options, params []string) error {
	err := a.initVitotrol(pOptions)
	if err != nil {
		return err
	}

	err = a.d.GetErrorHistory(a.v)
	if err != nil {
		return fmt.Errorf("GetErrorHistory error: %s", err)
	}

	if len(a.d.Errors) == 0 {
		fmt.Println("No errors")
	} else {
		fmt.Printf("%d error(s):\n", len(a.d.Errors))
		for _, error := range a.d.Errors {
			fmt.Println("-", &error)
		}
	}

	return nil
}

// setTimesheetAction implements the "set_timesheet" action.
type setTimesheetAction struct {
	authAction
}

func (a *setTimesheetAction) Do(pOptions *Options, params []string) error {
	if len(params) == 0 {
		return errors.New("timesheet name is missing")
	}

	var err error

	tID, err := existTimesheetName(params[0])
	if err != nil {
		return err
	}

	if len(params) == 1 {
		return errors.New("JSON definition of timesheet is missing")
	}
	tss := make(map[string]vitotrol.TimeslotSlice)
	var data []byte
	if strings.HasPrefix(params[1], "@") && len(params[1]) > 1 {
		data, err = ioutil.ReadFile(params[1][1:])
		if err != nil {
			return fmt.Errorf("Cannot read file %s: %s", params[1][1:], err)
		}
	} else {
		data = []byte(params[1])
	}
	err = json.Unmarshal(data, &tss)
	if err != nil {
		return fmt.Errorf("JSON definition of timesheet is invalid: %s", err)
	}

	err = a.initVitotrol(pOptions)
	if err != nil {
		return err
	}

	ch, err := a.d.WriteTimesheetDataWait(a.v, tID, tss)
	if err != nil {
		return fmt.Errorf("WriteTimesheetData error: %s", err)
	}

	if err = <-ch; err != nil {
		return fmt.Errorf("WriteTimesheetData failed: %s", err)
	}

	return nil
}

// timesheetAction implements the "timesheet" action.
type timesheetAction struct {
	authAction
}

func (a *timesheetAction) Do(pOptions *Options, params []string) error {
	if len(params) == 0 {
		return errors.New("timesheet name is missing")
	}

	var err error

	timesheetIDs := make([]vitotrol.TimesheetID, len(params))
	for idx, name := range params {
		timesheetIDs[idx], err = existTimesheetName(name)
		if err != nil {
			return err
		}
	}

	err = a.initVitotrol(pOptions)
	if err != nil {
		return err
	}

	for _, tID := range timesheetIDs {
		err := a.d.GetTimesheetData(a.v, tID)
		if err != nil {
			return fmt.Errorf("GetTimesheetData error: %s", err)
		}

		ts := a.d.Timesheets[tID]
		if a.options.jsonOutput {
			buf, _ := json.Marshal(ts)
			fmt.Println(string(buf))
		} else {
			fmt.Println(vitotrol.TimesheetsRef[tID])
			for _, day := range []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"} {
				fmt.Printf("- %s:\n", day)
				for _, slot := range ts[day] {
					fmt.Printf("  %s\n", &slot)
				}
			}
		}
	}

	return nil
}
