package vitotrol

import (
	"fmt"
	"strconv"
)

// An AttrID defines an attribute ID.
type AttrID uint16

// Attribute IDs currently supported by the library. For each, the
// Vitotrol™ name.
const (
	IndoorTemp             AttrID = 5367   // temp_rts_r
	OutdoorTemp            AttrID = 5373   // temp_ats_r
	SmokeTemp              AttrID = 5372   // temp_agt_r
	BoilerTemp             AttrID = 5374   // temp_kts_r
	HotWaterTemp           AttrID = 5381   // temp_ww_r
	HotWaterOutTemp        AttrID = 5382   // temp_auslauf_r
	HeatWaterOutTemp       AttrID = 6052   // temp_vts_r
	HeatNormalTemp         AttrID = 82     // konf_raumsolltemp_rw
	PartyModeTemp          AttrID = 79     // konf_partysolltemp_rw
	HeatReducedTemp        AttrID = 85     // konf_raumsolltemp_reduziert_rw
	HotWaterSetpointTemp   AttrID = 51     // konf_ww_solltemp_rw
	BurnerHoursRun         AttrID = 104    // anzahl_brennerstunden_r
	BurnerHoursRunReset    AttrID = 106    // anzahl_brennerstunden_w
	BurnerState            AttrID = 600    // zustand_brenner_r
	BurnerStarts           AttrID = 111    // anzahl_brennerstart_r
	InternalPumpStatus     AttrID = 245    // zustand_interne_pumpe_r
	HeatingPumpStatus      AttrID = 729    // zustand_heizkreispumpe_r
	CirculationPumpState   AttrID = 7181   // zustand_zirkulationspumpe_r
	PartyMode              AttrID = 7855   // konf_partybetrieb_rw
	EnergySavingMode       AttrID = 7852   // konf_sparbetrieb_rw
	DateTime               AttrID = 5385   // konf_uhrzeit_rw
	CurrentError           AttrID = 7184   // aktuelle_fehler_r
	HolidaysStart          AttrID = 306    // konf_ferien_start_rw
	HolidaysEnd            AttrID = 309    // konf_ferien_ende_rw
	HolidaysStatus         AttrID = 714    // zustand_ferienprogramm_r
	Way3ValveStatus        AttrID = 5389   // info_status_umschaltventil_r
	OperatingModeRequested AttrID = 92     // konf_betriebsart_rw
	OperatingModeCurrent   AttrID = 708    // aktuelle_betriebsart_r
	FrostProtectionStatus  AttrID = 717    // zustand_frostgefahr_r
	NoAttr                 AttrID = 0xffff // Used in error cases
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
	Custom bool
}

// String returns all information contained in this attribute reference.
func (r *AttrRef) String() string {
	return fmt.Sprintf("%s: %s (%s - %s)",
		r.Name, r.Doc, r.Type.Type(), AccessToStr[r.Access])
}

// AttributesRef lists the reference for each attribute ID.
var AttributesRef = map[AttrID]*AttrRef{
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
	SmokeTemp: {
		Type:   TypeDouble,
		Access: ReadOnly,
		Doc:    "Smoke temperature",
		Name:   "SmokeTemp",
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

// AddAttributeRef adds a new attribute to the "official" list. This
// new attribute will only differ from others by its Custom field set
// to true.
//
// No check is done to avoid overriding existing attributes.
func AddAttributeRef(attrID AttrID, ref AttrRef) {
	ref.Custom = true
	AttributesRef[attrID] = &ref

	AttributesNames2IDs = computeNames2IDs()
	Attributes = computeAttributes()
}

func computeNames2IDs() map[string]AttrID {
	ret := make(map[string]AttrID, len(AttributesRef))
	for attrID, pAttrRef := range AttributesRef {
		ret[pAttrRef.Name] = attrID
	}
	return ret
}

func computeAttributes() []AttrID {
	ret := make([]AttrID, 0, len(AttributesRef))
	for attrID := range AttributesRef {
		ret = append(ret, attrID)
	}
	return ret
}

// AttributesNames2IDs maps the attributes names to their AttrID
// counterpart.
var AttributesNames2IDs = computeNames2IDs()

// Attributes lists the AttrIDs for all available attributes.
var Attributes = computeAttributes()

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
