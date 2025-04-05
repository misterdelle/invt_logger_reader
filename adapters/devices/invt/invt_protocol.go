package invt

type field struct {
	register  int
	name      string
	valueType string
	factor    float32
	unit      string
}

type registerRange struct {
	start       int
	end         int
	replyFields []field
}

var allRegisterRanges = []registerRange{
	rrGridOutput,
	rrPVOutput,
	rrLoadInfo,
	rrEnergyTodayTotals,
	rrSystemInfo,
	rrBatOutput,
	// rrRatio,
	rrINVInfo,
}

var stationRegisterRanges = []registerRange{
	rrStationInfo,
	rrStationData,
	rrStationBatterySOC,
	rrStationBatteryPower,
	rrStationPower,
	rrStationPV,
}

var energyTodayTotalsRegisterRanges = []registerRange{
	rrEnergyTodayTotals,
}

var gridOutputRegisterRanges = []registerRange{
	rrGridOutput,
}

var inverterInfoRegisterRanges = []registerRange{
	rrINVInfo,
}

var loadInfoRegisterRanges = []registerRange{
	rrLoadInfo,
}

var batteryOutputRanges = []registerRange{
	rrBatOutput,
}

var pvOutputRanges = []registerRange{
	rrPVOutput,
}

func GetAllRegisterNames() []string {
	result := make([]string, 0)
	for _, rr := range allRegisterRanges {
		for _, f := range rr.replyFields {
			if f.name == "" || f.valueType == "" {
				// Measurements without a name or value type are ignored in replies
				continue
			}
			result = append(result, f.name)
		}
	}
	return result
}

var rrPVOutput = registerRange{
	start: 0x3130,
	end:   0x3135,
	replyFields: []field{
		{0x3130, "PV: Voltage_PV1", "U16", 0.1, "V"},
		{0x3131, "PV: Current_PV1", "U16", 0.1, "A"},
		{0x3132, "PV: Power_PV1", "U16", 1, "kW"},
		{0x3133, "PV: Voltage_PV2", "U16", 0.1, "V"},
		{0x3134, "PV: Current_PV2", "U16", 0.1, "A"},
		{0x3135, "PV: Power_PV2", "U16", 1, "kW"},
	},
}

var rrSystemInfo = registerRange{
	start: 0x3500,
	end:   0x350F,
	replyFields: []field{
		{0x3500, "RR: Year_Month", "U8", 1, ""},
		{0x3501, "RR: Day_Res", "U8", 1, ""},
		{0x3502, "RR: Hour_Minute", "U8", 1, ""},
		{0x3503, "RR: Second_DayOfWeek", "U8", 1, ""},
		{0x3504, "RR: Charge Time1 Start", "U8", 1, ""},
		{0x3505, "RR: Charge Time1 End", "U8", 1, ""},
		{0x3506, "RR: Discharge Time1 Start", "U8", 1, ""},
		{0x3507, "RR: Discharge Time1 End", "U8", 1, ""},
		{0x3508, "RR: Charge Time2 Start", "U8", 1, ""},
		{0x3509, "RR: Charge Time2 End", "U8", 1, ""},
		{0x350A, "RR: Discharge Time2 Start", "U8", 1, ""},
		{0x350B, "RR: Discharge Time2 End", "U8", 1, ""},
		{0x350C, "RR: Charge Time2 Start", "U8", 1, ""},
		{0x350D, "RR: Charge Time2 End", "U8", 1, ""},
		{0x350E, "RR: Discharge Time2 Start", "U8", 1, ""},
		{0x350F, "RR: Discharge Time2 End", "U8", 1, ""},
	},
}

var rrEnergyTodayTotals = registerRange{
	start: 0x3150,
	end:   0x3181,
	replyFields: []field{
		{0x3150, "ETT: S BUS Voltage", "U16", 1, "V"},
		{0x3151, "ETT: N BUS Voltage", "S16", 1, "V"},
		{0x3152, "ETT: DCDC Temperature", "S16", 1, "°C"},
		{0x3153, "ETT: PV Day Energy", "U32", 0.001, "kWh"},
		{0x3155, "ETT: Grid Day Energy", "U32", 0.001, "kWh"},
		{0x3157, "ETT: Load Day Energy", "U32", 0.001, "kWh"},
		{0x3159, "ETT: PV Month Energy", "U32", 0.01, "kWh"},
		{0x315B, "ETT: Grid Month Energy", "U32", 0.001, "kWh"},
		{0x315D, "ETT: Load Month Energy", "U32", 0.001, "kWh"},
		{0x315F, "ETT: PV Year Energy", "U32", 0.01, "kWh"},
		{0x3161, "ETT: Grid Year Energy", "U32", 0.001, "kWh"},
		{0x3163, "ETT: Load Year Energy", "U32", 0.001, "kWh"},
		{0x3165, "ETT: PV Total Energy", "U32", 0.001, "kWh"},
		{0x3167, "ETT: Grid Total Energy", "U32", 0.001, "kWh"},
		{0x3169, "ETT: Load Total Energy", "U32", 0.001, "kWh"},
		{0x316B, "ETT: Purchasing Day Energy", "U32", 0.001, "kWh"},
		{0x316D, "ETT: Bat Charge Day Energy", "U32", 0.001, "kWh"},
		{0x316F, "ETT: Bat Discharge Day Energy", "U32", 0.001, "kWh"},
		{0x3171, "ETT: Purchasing Month Energy", "U32", 0.001, "kWh"},
		{0x3173, "ETT: Bat Charge Month Energy", "U32", 0.001, "kWh"},
		{0x3175, "ETT: Bat Discharge Month Energy", "U32", 0.001, "kWh"},
		{0x3177, "ETT: Purchasing Year Energy", "U32", 0.001, "kWh"},
		{0x3179, "ETT: Bat Charge Year Energy", "U32", 0.001, "kWh"},
		{0x317B, "ETT: Bat Discharge Year Energy", "U32", 0.001, "kWh"},
		{0x317D, "ETT: Purchasing Total Energy", "U32", 0.001, "kWh"},
		{0x317F, "ETT: Bat Charge Total Energy", "U32", 0.001, "kWh"},
		{0x3181, "ETT: Bat Discharge Total Energy", "U32", 0.001, "kWh"},
	},
}

var rrGridOutput = registerRange{
	start: 0x3110,
	end:   0x311B,
	replyFields: []field{
		{0x3110, "GO: Grid A Voltage", "U16", 0.1, "V"},
		{0x3111, "GO: Grid A Current", "S16", 0.1, "A"},
		{0x3112, "GO: Grid A Power", "S16", 1, "W"},
		{0x3113, "GO: Grid B Voltage", "U16", 0.1, "V"},
		{0x3114, "GO: Grid B Current", "S16", 0.1, "A"},
		{0x3115, "GO: Grid B Power", "S16", 1, "W"},
		{0x3116, "GO: Grid C Voltage", "U16", 0.1, "V"},
		{0x3117, "GO: Grid C Current", "S16", 0.1, "A"},
		{0x3118, "GO: Grid C Power", "S16", 1, "W"},
		{0x3119, "GO: Grid Freq", "U16", 0.01, "Hz"},
		{0x311A, "GO: INV1 Temperature", "U16", 1, "°C"},
		{0x311B, "GO: INV2 Temperature", "U16", 1, "°C"},
	},
}

var rrBatOutput = registerRange{
	start: 0x313E,
	end:   0x314E,
	replyFields: []field{
		{0x313E, "BO: BMS BAT Voltage", "U16", 0.1, "V"},
		{0x313F, "BO: BMS BAT Current", "S16", 0.1, "A"},
		{0x3140, "BO: BAT Voltage", "U16", 0.1, "V"},
		{0x3141, "BO: BAT Current", "S16", 0.1, "A"},
		{0x3142, "BO: BAT 1 Current", "S16", 0.1, "A"},
		{0x3143, "BO: BAT 2 Current", "S16", 0.1, "A"}, //eccolo
		{0x3144, "BO: BAT 3 Current", "S16", 0.1, "A"},
		{0x3145, "BO: BAT SOC", "U16", 0.1, "%"},
		{0x3146, "BO: BAT Temperature", "U16", 0.1, "℃"},
		{0x3147, "BO: BAT Charge Voltage", "U16", 0.1, "V"},
		{0x3148, "BO: BAT Charge Current Limit", "U16", 0.1, "A"},
		{0x3149, "BO: BAT Discharge Current Limit", "U16", 0.1, "A"},
		{0x314A, "BO: BAT Power", "S16", 1, "W"},
		{0x314B, "BO: BMS BAT Cell Max Voltage", "U16", 1, "mV"},
		{0x314C, "BO: BMS BAT Cell Min Voltage", "U16", 1, "mV"},
		{0x314D, "BO: BMS BAT Cell Max Temperature", "S16", 1, "°C"},
		{0x314E, "BO: BMS BAT Cell Min Temperature", "S16", 1, "°C"},
	},
}

var rrLoadInfo = registerRange{
	start: 0x3120,
	end:   0x313F,
	replyFields: []field{
		{0x3120, "LI: Load A Voltage", "U16", 0.1, "V"},
		{0x3121, "LI: Load A Current", "U16", 0.1, "A"},
		{0x3122, "LI: Load A Power", "U16", 1, "W"},
		{0x3123, "LI: Load A Rate", "U16", 0.1, "%"},
		{0x3124, "LI: Load B Voltage", "U16", 0.1, "V"},
		{0x3125, "LI: Load B Current", "U16", 0.1, "A"},
		{0x3126, "LI: Load B Power", "U16", 1, "W"},
		{0x3127, "LI: Load B Rate", "U16", 0.1, "%"},
		{0x3128, "LI: Load C Voltage", "U16", 0.1, "V"},
		{0x3129, "LI: Load C Current", "U16", 0.1, "A"},
		{0x313A, "LI: Load C Power", "U16", 1, "W"},
		{0x313B, "LI: Load C Rate", "U16", 0.1, "%"},
		{0x313C, "", "", 1, ""},
		{0x313D, "LI: Generator Port Voltage A", "U16", 0.1, "V"},
		{0x313E, "LI: Generator Port Voltage B", "U16", 0.1, "V"},
		{0x313F, "LI: Generator Port Voltage C", "U16", 0.1, "V"},
	},
}

var rrINVInfo = registerRange{
	start: 0x3190,
	end:   0x319C,
	replyFields: []field{
		{0x3190, "II: INV A Voltage", "U16", 0.1, "V"},
		{0x3191, "II: INV A Current", "U16", 0.1, "A"},
		{0x3192, "II: INV A Power", "U16", 1, "W"},
		{0x3193, "II: INV B Voltage", "U16", 0.1, "V"},
		{0x3194, "II: INV B Current", "U16", 0.1, "A"},
		{0x3195, "II: INV B Power", "U16", 1, "W"}, // currentConsumptionPower
		{0x3196, "II: INV C Voltage", "U16", 0.1, "V"},
		{0x3197, "II: INV C Current", "U16", 0.1, "A"},
		{0x3198, "II: INV C Power", "U16", 1, "W"},
		{0x3199, "II: INV A Freq", "U16", 0.01, "Hz"},
		{0x319A, "II: INV B Freq", "U16", 0.01, "Hz"},
		{0x319B, "II: INV C Freq", "U16", 0.01, "Hz"},
		{0x319C, "II: Leak Current", "U16", 1, "mA"},
	},
}

var rrStationInfo = registerRange{
	start: 0x3500,
	end:   0x3503,
	replyFields: []field{
		{0x3500, "Year_Month", "U8", 1, ""},
		{0x3501, "Day_Res", "U8", 1, ""},
		{0x3502, "Hour_Minute", "U8", 1, ""},
		{0x3503, "Second_DayOfWeek", "U8", 1, ""},
	},
}

var rrStationBatterySOC = registerRange{
	start: 0x3145,
	end:   0x3145,
	replyFields: []field{
		{0x3145, "batterySOC", "U16", 0.1, "%"},
	},
}

var rrStationBatteryPower = registerRange{
	start: 0x314A,
	end:   0x314A,
	replyFields: []field{
		{0x314A, "batteryPower", "S16", 0.001, "kW"},
	},
}

var rrStationPower = registerRange{
	start: 0x3195,
	end:   0x3195,
	replyFields: []field{
		//{0x3126, "currentConsumptionPower", "U16", 1, "W"},
		{0x3195, "currentConsumptionPower", "U16", 1, "W"},
	},
}

var rrStationData = registerRange{
	start: 0x3153,
	end:   0x3181,
	replyFields: []field{
		{0x3153, "PV Day Energy", "U32", 0.001, "kWh"},
		{0x3155, "Grid Day Energy", "U32", 0.001, "kWh"},
		{0x3157, "Load Day Energy", "U32", 0.001, "kWh"},
		{0x3159, "PV Month Energy", "U32", 0.001, "kWh"},
		{0x315B, "Grid Month Energy", "U32", 0.001, "kWh"},
		{0x315D, "Load Month Energy", "U32", 0.001, "kWh"},
		{0x315F, "PV Year Energy", "U32", 0.001, "kWh"},
		{0x3161, "Grid Year Energy", "U32", 0.001, "kWh"},
		{0x3163, "Load Year Energy", "U32", 0.001, "kWh"},
		{0x3165, "PV Total Energy", "U32", 0.001, "kWh"},
		{0x3167, "Grid Total Energy", "U32", 0.001, "kWh"},
		{0x3169, "Load Total Energy", "U32", 0.001, "kWh"},
		{0x316B, "Purchasing Day Energy", "U32", 0.001, "kWh"},
		{0x316D, "Bat Charge Day Energy", "U32", 0.001, "kWh"},
		{0x316F, "Bat Discharge Day Energy", "U32", 0.001, "kWh"},
		{0x3171, "Purchasing Month Energy", "U32", 0.001, "kWh"},
		{0x3173, "Bat Charge Month Energy", "U32", 0.001, "kWh"},
		{0x3175, "Bat Discharge Month Energy", "U32", 0.001, "kWh"},
		{0x3177, "Purchasing Year Energy", "U32", 0.001, "kWh"},
		{0x3179, "Bat Charge Year Energy", "U32", 0.001, "kWh"},
		{0x317B, "Bat Discharge Year Energy", "U32", 0.001, "kWh"},
		{0x317D, "Purchasing Total Energy", "U32", 0.001, "kWh"},
		{0x317F, "Bat Charge Total Energy", "U32", 0.001, "kWh"},
		{0x3181, "Bat Discharge Total Energy", "U32", 0.001, "kWh"},
	},
}

var rrStationPV = registerRange{
	start: 0x3132,
	end:   0x3135,
	replyFields: []field{
		{0x3132, "Power PV1", "U16", 1, "kW"},
		{0x3133, "", "", 1, ""},
		{0x3134, "", "", 1, ""},
		{0x3135, "Power PV2", "U16", 1, "kW"},
	},
}

func GetStationRegisterNames() []string {

	result := make([]string, 0)
	for _, rr := range allRegisterRanges {
		for _, f := range rr.replyFields {
			if f.name == "" || f.valueType == "" {
				// Measurements without a name or value type are ignored in replies
				continue
			}
			result = append(result, f.name)
		}
	}
	return result
}
