package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/maxatome/go-vitotrol"
	"io/ioutil"
	"strconv"
	"strings"
)

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
	"devices":       &devicesAction{authAction: authAction{noDefaultDev: true}},
	"list":          &listAction{},
	"get":           &getAction{},
	"rget":          &getAction{rget: true},
	"bget":          &getAction{bget: true},
	"set":           &setAction{},
	"errors":        &errorsAction{},
	"timesheet":     &timesheetAction{},
	"set_timesheet": &setTimesheetAction{},
	"remote_attrs":  &remoteAttrsAction{},
}

type authAction struct {
	v            *vitotrol.Session
	d            *vitotrol.Device
	noDefaultDev bool
	options      *Options
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

	if !a.noDefaultDev {
		var pDevice *vitotrol.Device
		if pOptions.device == "" {
			pDevice = &v.Devices[0]
		} else if idx, err := strconv.Atoi(pOptions.device); err == nil {
			// Check if a device exists with this ID
			for _, device := range v.Devices {
				if uint32(idx) == device.DeviceID {
					pDevice = &device
					break
				}
			}

			// Else, take it as an index in devices array
			if pDevice == nil {
				if idx >= len(v.Devices) {
					return fmt.Errorf(
						"%d is not a device ID and too big to be an index "+
							"(>= %d available devices).",
						idx, len(v.Devices))
				}
				pDevice = &v.Devices[idx]
			}
		} else {
			checkDevLoc := strings.ContainsRune(pOptions.device, '@')

			// Check if a device exists with this name
			for _, device := range v.Devices {
				if pOptions.device == device.DeviceName {
					pDevice = &device
					break
				}

				// More checks: DeviceId@LocationID & DeviceName@LocationName
				if checkDevLoc {
					// DeviceId@LocationID
					if pOptions.device == fmt.Sprintf("%d@%d",
						device.DeviceID, device.LocationID) {
						pDevice = &device
						break
					}

					// DeviceName@LocationName
					if pOptions.device == device.DeviceName+"@"+device.LocationName {
						pDevice = &device
						break
					}
				}
			}

			if pDevice == nil {
				return fmt.Errorf("Cannot find device named `%s'", pOptions.device)
			}
		}

		if pOptions.verbose {
			fmt.Printf("Working with device %s@%s\n",
				pDevice.DeviceName, pDevice.LocationName)
		}

		a.d = pDevice
	}

	a.v = v
	a.options = pOptions

	return nil
}

func (a *authAction) NeedAuth() bool {
	return true
}

type foreignAttrs struct {
	authAction
	cachePopulated bool
}

func (f *foreignAttrs) checkAttributeAccess(attrName string, reqAccess vitotrol.AttrAccess) (vitotrol.AttrID, error) {
	var attrID vitotrol.AttrID
	var ok bool

	for {
		id, err := strconv.ParseUint(attrName, 0, 16)
		if err == nil {
			attrID = vitotrol.AttrID(id)
			_, ok = vitotrol.AttributesRef[attrID]
		} else {
			attrID, ok = vitotrol.AttributesNames2IDs[attrName]
		}

		if ok || f.cachePopulated {
			break
		}
		f.populateCache()
	}

	if !ok {
		return vitotrol.NoAttr, fmt.Errorf("unknown attribute `%s'", attrName)
	}

	if (vitotrol.AttributesRef[attrID].Access & reqAccess) != reqAccess {
		return vitotrol.NoAttr, fmt.Errorf("attribute `%s' is not %s",
			attrName, vitotrol.AccessToStr[reqAccess])
	}

	return attrID, nil
}

func (f *foreignAttrs) populateCache() {
	f.cachePopulated = true

	attrs, err := f.d.GetTypeInfo(f.v)
	if err != nil {
		fmt.Printf("GetTypeInfo failed: %s", err)
		return
	}

	for _, pAttrInfo := range attrs {
		attrID := pAttrInfo.AttributeID
		if vitotrol.AttributesRef[attrID] == nil {
			// Unknown attribute
			pType := vitotrol.TypeNames[pAttrInfo.AttributeType]
			if pType == nil {
				if pAttrInfo.AttributeType != "ENUM" {
					// No warning for type used for timesheets...
					if pAttrInfo.AttributeType != "CircuitTime" {
						fmt.Printf("populateCache: unrecognized type %s for attribute "+
							"%s-0x%04x. Discard it.\n",
							pAttrInfo.AttributeType, pAttrInfo.AttributeName, attrID)
					}
					continue
				}

				var maxIdx uint32
				for idx := range pAttrInfo.EnumValues {
					if idx > maxIdx {
						maxIdx = idx
					}
				}
				enumValues := make([]string, maxIdx+1)
				for idx, value := range pAttrInfo.EnumValues {
					enumValues[idx] = value
				}
				pType = vitotrol.NewEnum(enumValues)
			}

			ref := vitotrol.AttrRef{
				Type: pType,
				Name: fmt.Sprintf("%s-0x%04x", pAttrInfo.AttributeName, attrID),
				Doc:  pAttrInfo.AttributeName,
			}

			if pAttrInfo.Readable {
				ref.Access = vitotrol.ReadOnly
			}
			if pAttrInfo.Writable {
				ref.Access |= vitotrol.WriteOnly
			}

			vitotrol.AttributesRef[attrID] = &ref
			vitotrol.AttributesNames2IDs[ref.Name] = attrID
			vitotrol.Attributes = append(vitotrol.Attributes, attrID)
		}
	}
}

// devicesAction implements the "devices" action.
type devicesAction struct {
	authAction
}

func (a *devicesAction) Do(pOptions *Options, params []string) error {
	err := a.initVitotrol(pOptions)
	if err != nil {
		return err
	}

	for idx, device := range a.v.Devices {
		fmt.Printf(`Index %d
  LocationName (LocationID): %s (%d)
      DeviceName (DeviceID): %s (%d)
                   HasError: %v
                IsConnected: %v
`, idx,
			device.LocationName, device.LocationID,
			device.DeviceName, device.DeviceID,
			device.HasError,
			device.IsConnected)
	}
	return nil
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
	foreignAttrs
	rget bool
	bget bool
}

func (a *getAction) Do(pOptions *Options, params []string) error {
	if len(params) == 0 {
		return errors.New("at least one PARAM is missing")
	}

	err := a.initVitotrol(pOptions)
	if err != nil {
		return err
	}

	var attrs []vitotrol.AttrID
	// Special case -> all attributes
	if len(params) == 1 && params[0] == "all" {
		a.populateCache()
		attrs = vitotrol.Attributes
	} else {
		var err error
		attrs = make([]vitotrol.AttrID, len(params))

		if a.bget {
			a.populateCache()

			var id uint64
			for idx, attrName := range params {
				id, err = strconv.ParseUint(attrName, 0, 16)
				if err != nil {
					return err
				}
				attrs[idx] = vitotrol.AttrID(id)

				// Create a fake String entry for this attribute
				if _, ok := vitotrol.AttributesRef[vitotrol.AttrID(id)]; !ok {
					vitotrol.AttributesRef[vitotrol.AttrID(id)] = &vitotrol.AttrRef{
						Type:   vitotrol.TypeString,
						Access: vitotrol.ReadOnly,
						Name:   fmt.Sprintf("0x%04x", id),
					}
				}
			}
		} else {
			for idx, attrName := range params {
				attrs[idx], err = a.checkAttributeAccess(attrName, vitotrol.ReadOnly)
				if err != nil {
					return err
				}
			}
		}
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
	foreignAttrs
}

func (a *setAction) Do(pOptions *Options, params []string) error {
	if len(params) == 0 || (len(params)&1) != 0 {
		return errors.New("PARAMS must be a list of pairs: ATTR_NAME, VALUE")
	}

	err := a.initVitotrol(pOptions)
	if err != nil {
		return err
	}

	attrsValues := make(map[vitotrol.AttrID]string, len(params)/2)
	for idx := 0; idx < len(params); idx += 2 {
		attrID, err := a.checkAttributeAccess(params[idx], vitotrol.WriteOnly)
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

// remoteAttrsAction implements the "remote_attrs" action.
type remoteAttrsAction struct {
	authAction
}

func (a *remoteAttrsAction) Do(pOptions *Options, params []string) error {
	err := a.initVitotrol(pOptions)
	if err != nil {
		return err
	}

	list, err := a.d.GetTypeInfo(a.v)
	if err != nil {
		return err
	}

	for _, pAttrInfo := range list {
		fmt.Printf("- %#v\n", *pAttrInfo)
	}
	return nil
}
