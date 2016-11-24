package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/maxatome/go-vitotrol"
	"os"
	"path"
)

var allowedActions = map[string]bool{
	"errors":    true,
	"get":       true,
	"list":      true,
	"rget":      true,
	"set":       true,
	"timesheet": true,
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS] ACTION [PARAMS]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
ACTION & PARAMS can be:
- list [attrs|timesheets]  list attribute (default) or timesheet names
- get ATTR_NAME ...    get the value of attributes ATTR_NAME, ... on vitodata
                         server
- get all              get all known attributes on vitodata server
- rget ATTR_NAME ...   refresh then get the value of attributes ATTR_NAME, ...
                         on vitodata server
- rget all             refresh than get all known attributes on vitodata server
- set ATTR_NAME VALUE  set the value of attribute ATTR_NAME to VALUE
- timesheet TIMESHEET  get the timesheet TIMESHEET data
- errors               get the error history`)
	}

	var login, password, config string
	var verbose, debug, jsonOutput bool
	flag.StringVar(&login, "login", "", "login on vitotrol API")
	flag.StringVar(&password, "password", "", "password on vitotrol API")
	flag.StringVar(&config, "config", "", "login+password config file")
	flag.BoolVar(&verbose, "verbose", false, "print verbose information")
	flag.BoolVar(&debug, "debug", false, "print debug information")
	flag.BoolVar(&jsonOutput, "json", false,
		"used by `timesheet' action to display timesheets using JSON format")

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	action, params := flag.Args()[0], flag.Args()[1:]

	if !allowedActions[action] {
		fmt.Fprintf(os.Stderr, "*** bad action `%s'\n", action)
		flag.Usage()
		os.Exit(1)
	}

	// This action dos not need auth
	if action == "list" {
		if len(params) == 0 || params[0] == "attrs" {
			for _, pAttrRef := range vitotrol.AttributesRef {
				fmt.Println(pAttrRef)
			}
		} else if params[0] == "timesheets" {
			for _, pTimesheetRef := range vitotrol.TimesheetsRef {
				fmt.Println(pTimesheetRef)
			}
		} else {
			fmt.Fprintf(os.Stderr,
				"`list' action allows `attrs' or `timesheets' params, not `%s'\n",
				params[0])
			flag.Usage()
			os.Exit(1)
		}
		return
	}

	// Load config if login OR password is missing
	if login == "" || password == "" {
		if config == "" {
			config = path.Join(os.Getenv("HOME"), ".vitotrol-api")
		}

		file, err := os.Open(config)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"--login & --password are mandatory "+
					"EXCEPT if `%s' file exists and is readdable\n",
				config)
			os.Exit(1)
		}

		info, err := file.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot stat `%s' file: %s\n", config, err)
			os.Exit(1)
		}

		if (info.Mode() & 06) != 0 {
			fmt.Fprintf(os.Stderr,
				"`%s' file readdable and/or writtable by others. Abort!\n", config)
			os.Exit(1)
		}

		rd := bufio.NewReader(file)
		confLogin, _ := rd.ReadString('\n')
		confPassword, err := rd.ReadString('\n')
		file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid config file `%s' contents, must contain "+
				"login and password on two separate lines\n", config)
			os.Exit(1)
		}

		if login == "" {
			login = confLogin[:len(confLogin)-1]
		}
		if password == "" {
			password = confPassword[:len(confPassword)-1]
		}
	}

	if action != "errors" && len(params) == 0 {
		fmt.Fprintln(os.Stderr, "*** PARAMS is missing!")
		flag.Usage()
		os.Exit(1)
	}

	switch action {
	case "get", "rget":
		var attrs []vitotrol.AttrId
		// Special case -> all attributes
		if len(params) == 1 && params[0] == "all" {
			attrs = vitotrol.Attributes
		} else {
			attrs = make([]vitotrol.AttrId, len(params))
			for idx, attrName := range params {
				attrs[idx] = mustCheckAttributeAccess(attrName, vitotrol.ReadOnly)
			}
		}

		v, pDevice := mustInitVitotrol(login, password, verbose, debug)

		if action == "rget" {
			ch, err := pDevice.RefreshDataWait(v, attrs)
			if err != nil {
				fmt.Fprintln(os.Stderr, "RefreshData error:", err)
				os.Exit(1)
			}

			if err = <-ch; err != nil {
				fmt.Fprintln(os.Stderr, "RefreshData failed:", err)
				os.Exit(1)
			}
		}

		err := pDevice.GetData(v, attrs)
		if err != nil {
			fmt.Fprintln(os.Stderr, "GetData error:", err)
			os.Exit(1)
		}

		fmt.Print(pDevice.FormatAttributes(attrs))

	case "set":
		if (len(params) & 1) != 0 {
			fmt.Fprintln(os.Stderr,
				"*** PARAMS must be a list of pairs: ATTR_NAME VALUE ...")
			flag.Usage()
			os.Exit(1)
		}
		attrsValues := make(map[vitotrol.AttrId]string, len(params)/2)
		for idx := 0; idx < len(params); idx += 2 {
			attrId := mustCheckAttributeAccess(params[idx], vitotrol.WriteOnly)

			value, err :=
				vitotrol.AttributesRef[attrId].Type.Human2VitodataValue(params[idx+1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "*** value `%s' of attribute %s is invalid: %s\n",
					params[idx+1], params[idx], err)
				os.Exit(1)
			}

			attrsValues[attrId] = value
		}

		v, pDevice := mustInitVitotrol(login, password, verbose, debug)

		// Set them all
		for attrId, value := range attrsValues {
			ch, err := pDevice.WriteDataWait(v, attrId, value)
			if err != nil {
				fmt.Fprintln(os.Stderr, "WriteData error:", err)
				os.Exit(1)
			}

			if err = <-ch; err != nil {
				fmt.Fprintln(os.Stderr, "WriteData failed:", err)
				os.Exit(1)
			}

			if verbose {
				fmt.Printf("%s attribute successfully set to `%s'\n",
					vitotrol.AttributesRef[attrId].Name, value)
			}
		}

	case "errors":
		v, pDevice := mustInitVitotrol(login, password, verbose, debug)

		err := pDevice.GetErrorHistory(v)
		if err != nil {
			fmt.Fprintln(os.Stderr, "GetErrorHistory error:", err)
			os.Exit(1)
		}

		if len(pDevice.Errors) == 0 {
			fmt.Println("No errors")
		} else {
			fmt.Printf("%d error(s):\n", len(pDevice.Errors))
			for _, error := range pDevice.Errors {
				fmt.Println("-", &error)
			}
		}

	case "timesheet":
		timesheetIds := make([]vitotrol.TimesheetId, len(params))
		for idx, name := range params {
			tId, ok := vitotrol.TimesheetsNames2Ids[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "*** unknown timesheet `%s'\n", name)
				flag.Usage()
				os.Exit(1)
			}
			timesheetIds[idx] = tId
		}

		v, pDevice := mustInitVitotrol(login, password, verbose, debug)

		for _, tId := range timesheetIds {
			err := pDevice.GetTimesheetData(v, tId)
			if err != nil {
				fmt.Fprintln(os.Stderr, "GetTimesheetData error:", err)
				os.Exit(1)
			}

			ts := pDevice.Timesheets[tId]
			if jsonOutput {
				buf, _ := json.Marshal(ts)
				fmt.Println(string(buf))
			} else {
				fmt.Println(vitotrol.TimesheetsRef[tId])
				for _, day := range []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"} {
					fmt.Printf("- %s:\n", day)
					for _, slot := range ts[day] {
						fmt.Printf("  %s\n", &slot)
					}
				}
			}
		}

	default:
		panic("Unhandled action `" + action + "'")
	}
}

func mustInitVitotrol(login, password string, verbose, debug bool) (*vitotrol.Session, *vitotrol.Device) {
	v, pDevice, err := initVitotrol(login, password, verbose, debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Fatal error:", err)
		os.Exit(1)
	}
	return v, pDevice
}

func initVitotrol(login, password string, verbose, debug bool) (*vitotrol.Session, *vitotrol.Device, error) {
	v := &vitotrol.Session{
		Debug: debug,
	}

	err := v.Login(login, password)
	if err != nil {
		return nil, nil, fmt.Errorf("Login failed: %s", err)
	}

	err = v.GetDevices()
	if err != nil {
		return nil, nil, fmt.Errorf("GetDevices failed: %s", err)
	}
	if len(v.Devices) == 0 {
		return nil, nil, errors.New("No device found!")
	}

	pDevice := &v.Devices[0]
	if !pDevice.IsConnected {
		return nil, nil, fmt.Errorf("Device %s@%s is not connected",
			pDevice.DeviceName, pDevice.LocationName)
	}

	if verbose {
		fmt.Printf("Working with device %s@%s\n",
			pDevice.DeviceName, pDevice.LocationName)
	}

	return v, pDevice, nil
}

func checkAttributeAccess(attrName string, reqAccess vitotrol.AttrAccess) (vitotrol.AttrId, error) {
	attrId, ok := vitotrol.AttributesNames2Ids[attrName]
	if !ok {
		return vitotrol.NoAttr, fmt.Errorf("unknown attribute `%s'", attrName)
	}

	if (vitotrol.AttributesRef[attrId].Access & reqAccess) != reqAccess {
		return vitotrol.NoAttr, fmt.Errorf("attribute `%s' is not %s",
			attrName, vitotrol.AccessToStr[reqAccess])
	}

	return attrId, nil
}

func mustCheckAttributeAccess(attrName string, reqAccess vitotrol.AttrAccess) vitotrol.AttrId {
	attrId, err := checkAttributeAccess(attrName, reqAccess)
	if err != nil {
		fmt.Fprintln(os.Stderr, "***", err)
		flag.Usage()
		os.Exit(1)
	}
	return attrId
}
