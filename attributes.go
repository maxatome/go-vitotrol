package vitotrol

import (
	"fmt"
	"strconv"
)

// An AttrId defines an attribute ID
type AttrId uint16

// Attribute IDs currently supported by the library. For each, the
// Vitotrol™ name.
const (
	IndoorTemp             AttrId = 5367   // temp_rts_r
	OutdoorTemp            AttrId = 5373   // temp_ats_r
	BoilerTemp             AttrId = 5374   // temp_kts_r
	HotWaterTemp           AttrId = 5381   // temp_ww_r
	HotWaterOutTemp        AttrId = 5382   // temp_auslauf_r
	HeatWaterOutTemp       AttrId = 6052   // temp_vts_r
	HeatNormalTemp         AttrId = 82     // konf_raumsolltemp_rw
	PartyModeTemp          AttrId = 79     // konf_partysolltemp_rw
	HeatReducedTemp        AttrId = 85     // konf_raumsolltemp_reduziert_rw
	HotWaterSetpointTemp   AttrId = 51     // konf_ww_solltemp_rw
	BurnerHoursRun         AttrId = 104    // anzahl_brennerstunden_r
	BurnerHoursRunReset    AttrId = 106    // anzahl_brennerstunden_w
	BurnerState            AttrId = 600    // zustand_brenner_r
	BurnerStarts           AttrId = 111    // anzahl_brennerstart_r
	InternalPumpStatus     AttrId = 245    // zustand_interne_pumpe_r
	HeatingPumpStatus      AttrId = 729    // zustand_heizkreispumpe_r
	CirculationPumpState   AttrId = 7181   // zustand_zirkulationspumpe_r
	PartyMode              AttrId = 7855   // konf_partybetrieb_rw
	EnergySavingMode       AttrId = 7852   // konf_sparbetrieb_rw
	DateTime               AttrId = 5385   // konf_uhrzeit_rw
	CurrentError           AttrId = 7184   // aktuelle_fehler_r
	HolidaysStart          AttrId = 306    // konf_ferien_start_rw
	HolidaysEnd            AttrId = 309    // konf_ferien_ende_rw
	HolidaysStatus         AttrId = 714    // zustand_ferienprogramm_r
	Way3ValveStatus        AttrId = 5389   // info_status_umschaltventil_r
	OperatingModeRequested AttrId = 92     // konf_betriebsart_rw
	OperatingModeCurrent   AttrId = 708    // aktuelle_betriebsart_r
	FrostProtectionStatus  AttrId = 717    // zustand_frostgefahr_r
	NoAttr                 AttrId = 0xffff // Used in error cases
)

// An AttrAccess defines attributes access rights.
type AttrAccess uint8

// Availables access rights.
const (
	ReadOnly AttrAccess = 1 << iota
	WriteOnly
	ReadWrite AttrAccess = ReadOnly | WriteOnly
)

// AccessToStr map allows to translate AttrAccess values to strings.
var AccessToStr = map[AttrAccess]string{
	ReadOnly:  "read-only",
	WriteOnly: "write-only",
	ReadWrite: "read/write",
}

// An AttrRef describes an attribute reference: its type, access and name.
type AttrRef struct {
	Type   VitodataType
	Access AttrAccess
	Name   string
	Doc    string
}

// String returns all information contained in this attribute reference.
func (r *AttrRef) String() string {
	return fmt.Sprintf("%s: %s (%s - %s)",
		r.Name, r.Doc, r.Type.Type(), AccessToStr[r.Access])
}

// AttributesRef lists the reference for each attribute ID.
var AttributesRef = map[AttrId]*AttrRef{
	IndoorTemp: {
		Type:   TypeDouble,
		Access: ReadOnly,
		Doc:    "Indoor temperature",
		Name:   "IndoorTemp",
	},
	OutdoorTemp: {
		Type:   TypeDouble,
		Access: ReadOnly,
		Doc:    "Outdoor temperature",
		Name:   "OutdoorTemp",
	},
	BoilerTemp: {
		Type:   TypeDouble,
		Access: ReadOnly,
		Doc:    "Boiler temperature",
		Name:   "BoilerTemp",
	},
	HotWaterTemp: {
		Type:   TypeDouble,
		Access: ReadOnly,
		Doc:    "Hot water temperature",
		Name:   "HotWaterTemp",
	},
	HotWaterOutTemp: {
		Type:   TypeDouble,
		Access: ReadOnly,
		Doc:    "Hot water outlet temperature",
		Name:   "HotWaterOutTemp",
	},
	HeatWaterOutTemp: {
		Type:   TypeDouble,
		Access: ReadOnly,
		Doc:    "Heating water outlet temperature",
		Name:   "HeatWaterOutTemp",
	},
	HeatNormalTemp: {
		Type:   TypeDouble,
		Access: ReadWrite,
		Doc:    "Setpoint of the normal room temperature",
		Name:   "HeatNormalTemp",
	},
	PartyModeTemp: {
		Type:   TypeDouble,
		Access: ReadWrite,
		Doc:    "Party mode temperature",
		Name:   "PartyModeTemp",
	},
	HeatReducedTemp: {
		Type:   TypeDouble,
		Access: ReadWrite,
		Doc:    "Setpoint of the reduced room temperature",
		Name:   "HeatReducedTemp",
	},
	HotWaterSetpointTemp: {
		Type:   TypeDouble,
		Access: ReadWrite,
		Doc:    "Setpoint of the domestic hot water temperature",
		Name:   "HotWaterSetpointTemp",
	},
	BurnerHoursRun: {
		Type:   TypeDouble,
		Access: ReadOnly,
		Doc:    "Burner hours run",
		Name:   "BurnerHoursRun",
	},
	BurnerHoursRunReset: {
		Type:   TypeDouble,
		Access: WriteOnly,
		Doc:    "Reset the burner hours run",
		Name:   "BurnerHoursRunReset",
	},
	BurnerState: {
		Type:   TypeOnOffEnum,
		Access: ReadOnly,
		Doc:    "Burner status",
		Name:   "BurnerState",
	},
	BurnerStarts: {
		Type:   TypeDouble,
		Access: ReadWrite,
		Doc:    "Burner starts",
		Name:   "BurnerStarts",
	},
	InternalPumpStatus: {
		Type: NewEnum([]string{ // 0 -> 3
			"off",
			"on",
			"off2",
			"on2",
		}),
		Access: ReadOnly,
		Doc:    "Internal pump status",
		Name:   "InternalPumpStatus",
	},
	HeatingPumpStatus: {
		Type:   TypeOnOffEnum,
		Access: ReadOnly,
		Doc:    "Heating pump status",
		Name:   "HeatingPumpStatus",
	},
	CirculationPumpState: {
		Type:   TypeOnOffEnum,
		Access: ReadOnly,
		Doc:    "Statut pompe circulation",
		Name:   "CirculationPumpState",
	},
	PartyMode: {
		Type:   TypeEnabledEnum,
		Access: ReadWrite,
		Doc:    "Party mode status",
		Name:   "PartyMode",
	},
	EnergySavingMode: {
		Type:   TypeEnabledEnum,
		Access: ReadWrite,
		Doc:    "Energy saving mode status",
		Name:   "EnergySavingMode",
	},
	DateTime: {
		Type:   TypeDate,
		Access: ReadWrite,
		Doc:    "Current date and time",
		Name:   "DateTime",
	},
	CurrentError: {
		Type:   TypeString,
		Access: ReadOnly,
		Doc:    "Current error",
		Name:   "CurrentError",
	},
	HolidaysStart: {
		Type:   TypeDate,
		Access: ReadWrite,
		Doc:    "Holidays begin date",
		Name:   "HolidaysStart",
	},
	HolidaysEnd: {
		Type:   TypeDate,
		Access: ReadWrite,
		Doc:    "Holidays end date",
		Name:   "HolidaysEnd",
	},
	HolidaysStatus: {
		Type:   TypeEnabledEnum,
		Access: ReadOnly,
		Doc:    "Holidays program status",
		Name:   "HolidaysStatus",
	},
	Way3ValveStatus: {
		Type: NewEnum([]string{ // 0 -> 3
			"undefined",
			"heating",
			"middle position",
			"hot water",
		}),
		Access: ReadOnly,
		Doc:    "3-way valve status",
		Name:   "3WayValveStatus",
	},
	OperatingModeRequested: {
		Type: NewEnum([]string{ // 0 -> 4
			"off",
			"DHW only",
			"heating+DHW",
			"continuous reduced",
			"continuous normal",
		}),
		Access: ReadWrite,
		Doc:    "Operating mode requested",
		Name:   "OperatingModeRequested",
	},
	OperatingModeCurrent: {
		Type: NewEnum([]string{ // 0 -> 3
			"stand-by",
			"reduced",
			"normal",
			"continuous normal",
		}),
		Access: ReadOnly,
		Doc:    "Operating mode",
		Name:   "OperatingModeCurrent",
	},
	FrostProtectionStatus: {
		Type:   TypeEnabledEnum,
		Access: ReadOnly,
		Doc:    "Frost protection status",
		Name:   "FrostProtectionStatus",
	},
}

// AttributesNames2Ids maps the attributes names to their AttrId
// counterpart.
var AttributesNames2Ids = func() map[string]AttrId {
	ret := make(map[string]AttrId, len(AttributesRef))
	for attrId, pAttrRef := range AttributesRef {
		ret[pAttrRef.Name] = attrId
	}
	return ret
}()

// Attributes lists the AttrIds for all available attributes.
var Attributes = func() []AttrId {
	ret := make([]AttrId, 0, len(AttributesRef))
	for attrId := range AttributesRef {
		ret = append(ret, attrId)
	}
	return ret
}()

// Value is the timestamped value of an attribute.
type Value struct {
	Value string
	Time  Time
}

// Num returns the numerical value of this value. If the value is not
// a numerical one, 0 is returned.
func (v *Value) Num() (ret float64) {
	ret, _ = strconv.ParseFloat(v.Value, 64)
	return
}
