Package `go-vitotrol` provides access to the Viessmann™
Vitotrol™ cloud API for controlling/monitoring boilers.

[![Build Status](https://travis-ci.org/maxatome/go-vitotrol.svg)](https://travis-ci.org/maxatome/go-vitotrol)
[![Coverage Status](https://coveralls.io/repos/github/maxatome/go-vitotrol/badge.svg?branch=master)](https://coveralls.io/github/maxatome/go-vitotrol?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/maxatome/go-vitotrol)](https://goreportcard.com/report/github.com/maxatome/go-vitotrol)
[![GoDoc](https://godoc.org/github.com/maxatome/go-vitotrol?status.svg)](https://godoc.org/github.com/maxatome/go-vitotrol)

See https://www.viessmann.com/app_vitodata/VIIWebService-1.16.0.0/iPhoneWebService.asmx

Only requests I really need are currently implemented:
- Login
- GetDevices
- RequestRefreshStatus
- RequestWriteStatus
- GetData
- WriteData
- RefreshData
- GetErrorHistory
- GetTimesheet
- WriteTimesheetData
- GetTypeInfo

Any pull-request is welcome.

## Install

The library:
```
go get -u github.com/maxatome/go-vitotrol
```

The vitotrol command:
```
go get -u github.com/maxatome/go-vitotrol/cmd/vitotrol
```

resulting a `vitotrol` executable in `$GOPATH/bin/` directory.

## Example

See `cmd/vitotrol/*.go` for an example of use.

Executable `vitotrol` usage follows:

```
usage: ./vitotrol [OPTIONS] ACTION [PARAMS]
  -config string
        login+password config file
  -debug
        print debug information
  -device string
        DeviceID, index, DeviceName, DeviceId@LocationID, DeviceName@LocationName (see `devices' action)
  -json
        used by `timesheet' action to display timesheets using JSON format
  -login string
        login on vitotrol API
  -password string
        password on vitotrol API
  -verbose
        print verbose information

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
```

The config file is a two lines file containing the LOGIN on the first
line and the PASSWORD on the second, and is named
`$HOME/.vitotrol-api` by default (when all `--config`, `--login` and
`--password` options are missing or empty):

```
LOGIN
PASSWORD
```

## License

go-vitotrol is released under the MIT License.
