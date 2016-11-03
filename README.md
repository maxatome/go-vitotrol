Package `go-vitotrol` provides access to the Viessmann™
Vitotrol™ cloud API for controlloing/monitoring boilers.

[![Build Status](https://travis-ci.org/maxatome/go-vitotrol.svg)](https://travis-ci.org/maxatome/go-vitotrol)
[![Coverage Status](https://coveralls.io/repos/github/maxatome/go-vitotrol/badge.svg?branch=master)](https://coveralls.io/github/maxatome/go-vitotrol?branch=master)
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

Any pull-request is welcome.

## Example

See `cmd/vitotrol/main.go` for an example.

`cmd/vitotrol/vitotrol` usage follows:

```
usage: ./cmd/vitotrol/vitotrol [OPTIONS] ACTION [PARAMS]
  -config string
    	login+password config file
  -debug
    	print debug information
  -login string
    	login on vitotrol API
  -password string
    	password on vitotrol API
  -verbose
    	print verbose information

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
