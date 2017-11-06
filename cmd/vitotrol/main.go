package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
)

// Options gathers user parameters together.
type Options struct {
	login      string
	password   string
	verbose    bool
	debug      bool
	jsonOutput bool
	device     string
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS] ACTION [PARAMS]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
ACTION & PARAMS can be:
- devices              list all available devices
- list [attrs|timesheets]  list attribute (default) or timesheet names
- get ATTR_NAME ...    get the value of attributes ATTR_NAME, ... on vitodata
                         server
- get all              get all known attributes on vitodata server
- rget ATTR_NAME ...   refresh then get the value of attributes ATTR_NAME, ...
                         on vitodata server
- rget all             refresh than get all known attributes on vitodata server
- set ATTR_NAME VALUE  set the value of attribute ATTR_NAME to VALUE
- timesheet TIMESHEET ...
                       get the timesheet TIMESHEET data
- set_timesheet TIMESHEET '{"wday":[{"from":630,"to":2200},...],...}'
                       replace the whole timesheet TIMESHEET
                       wday is either a day (eg. mon) or a range of days
                       (eg. mon-wed or sat-mon)
                       The JSON content can be in a file with the syntax @file
- errors               get the error history
- remote_attrs         list server available attributes
                         (for developing purpose)`)
	}

	var options Options
	var config string
	flag.StringVar(&options.login, "login", "", "login on vitotrol API")
	flag.StringVar(&options.password, "password", "", "password on vitotrol API")
	flag.StringVar(&config, "config", "", "login+password config file")
	flag.StringVar(&options.device, "device", "0",
		"DeviceID, index, DeviceName, "+
			"DeviceId@LocationID, DeviceName@LocationName (see `devices' action)")
	flag.BoolVar(&options.verbose, "verbose", false, "print verbose information")
	flag.BoolVar(&options.debug, "debug", false, "print debug information")
	flag.BoolVar(&options.jsonOutput, "json", false,
		"used by `timesheet' action to display timesheets using JSON format")

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	actionName, params := flag.Args()[0], flag.Args()[1:]

	action := actions[actionName]
	if action == nil {
		fmt.Fprintf(os.Stderr, "*** bad action `%s'\n", action)
		flag.Usage()
		os.Exit(1)
	}

	var err error
	if action.NeedAuth() {
		// Load config if login OR password is missing
		if options.login == "" || options.password == "" {
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

			if options.login == "" {
				options.login = confLogin[:len(confLogin)-1]
			}
			if options.password == "" {
				options.password = confPassword[:len(confPassword)-1]
			}
		}
	}

	err = action.Do(&options, params)
	if err != nil {
		fmt.Fprintln(os.Stderr, "***", err)
		os.Exit(1)
	}
}
