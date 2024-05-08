package invt

import "github.com/misterdelle/invt_logger_reader/ports"

type Logger struct {
	serialNumber uint
	connPort     ports.CommunicationPort
}

type Station struct {
	lastUpdateTimeUnix int
	lastUpdateTime     string
	generationTotal    int
	generationPower    int
	chargePower        int
	dischargePower     int
	batteryPower       int
	batterySOC         int
	usePower           int
}

type EnergyTodayTotals struct {
	S_BUS_Voltage              int
	N_BUS_Voltage              int
	DCDC_Temperature           int
	PV_Day_Energy              int
	Grid_Day_Energy            int
	Load_Day_Energy            int
	PV_Month_Energy            int
	Grid_Month_Energy          int
	Load_Month_Energy          int
	PV_Year_Energy             int
	Grid_Year_Energy           int
	Load_Year_Energy           int
	PV_Total_Energy            int
	Grid_Total_Energy          int
	Load_Total_Energy          int
	Purchasing_Day_Energy      int
	Bat_Charge_Day_Energy      int
	Bat_Discharge_Day_Energy   int
	Purchasing_Month_Energy    int
	Bat_Charge_Month_Energy    int
	Bat_Discharge_Month_Energy int
	Purchasing_Year_Energy     int
	Bat_Charge_Year_Energy     int
	Bat_Discharge_Year_Energy  int
	Purchasing_Total_Energy    int
	Bat_Charge_Total_Energy    int
	Bat_Discharge_Total_Energy int
}

type GridOutput struct {
	Grid_A_Voltage   int
	Grid_A_Current   int
	Grid_A_Power     int
	Grid_B_Voltage   int
	Grid_B_Current   int
	Grid_B_Power     int
	Grid_C_Voltage   int
	Grid_C_Current   int
	Grid_C_Power     int
	Grid_Freq        int
	INV1_Temperature int
	INV2_Temperature int
}

type InverterInfo struct {
	INV_A_Voltage int
	INV_A_Current int
	INV_A_Power   int
	INV_B_Voltage int
	INV_B_Current int
	INV_B_Power   int
	INV_C_Voltage int
	INV_C_Current int
	INV_C_Power   int
	INV_A_Freq    int
	INV_B_Freq    int
	INV_C_Freq    int
	Leak_Current  int
}

type LoadInfo struct {
	Load_A_Voltage           int
	Load_A_Current           int
	Load_A_Power             int
	Load_A_Rate              int
	Load_B_Voltage           int
	Load_B_Current           int
	Load_B_Power             int
	Load_B_Rate              int
	Load_C_Voltage           int
	Load_C_Current           int
	Load_C_Power             int
	Load_C_Rate              int
	Generator_Port_Voltage_A int
	Generator_Port_Voltage_B int
	Generator_Port_Voltage_C int
}

type BatteryOutput struct {
	BAT_Voltage                  int
	BAT_Current                  int
	BAT_1_Current                int
	BAT_2_Current                int
	BAT_3_Current                int
	BAT_SOC                      int
	BAT_Temperature              int
	BAT_Charge_Voltage           int
	BAT_Charge_Current_Limit     int
	BAT_Discharge_Current_Limit  int
	BAT_Power                    int
	BMS_BAT_Voltage              int
	BMS_BAT_Current              int
	BMS_BAT_Cell_Max_Voltage     int
	BMS_BAT_Cell_Min_Voltage     int
	BMS_BAT_Cell_Max_Temperature int
	BMS_BAT_Cell_Min_Temperature int
}

type PVOutput struct {
	Voltage_PV1 int
	Current_PV1 int
	Power_PV1   int
	Voltage_PV2 int
	Current_PV2 int
	Power_PV2   int
}

func NewInvtLogger(serialNumber uint, connPort ports.CommunicationPort) *Logger {
	return &Logger{
		serialNumber: serialNumber,
		connPort:     connPort,
	}
}

func (s *Logger) Query() (map[string]interface{}, error) {
	return readData(s.connPort, s.serialNumber)
}

func (s *Logger) Name() string {
	return "sofar"
}

func (s *Logger) QueryStation() (map[string]interface{}, error) {
	return readStationData(s.connPort, s.serialNumber)
}

func (s *Logger) QueryEnergyTodayTotals() (map[string]interface{}, error) {
	return readEnergyTodayTotalsData(s.connPort, s.serialNumber)
}

func (s *Logger) QueryGridOutput() (map[string]interface{}, error) {
	return readGridOutput(s.connPort, s.serialNumber)
}

func (s *Logger) QueryInverterInfo() (map[string]interface{}, error) {
	return readInverterInfo(s.connPort, s.serialNumber)
}

func (s *Logger) QueryLoadInfo() (map[string]interface{}, error) {
	return readLoadInfo(s.connPort, s.serialNumber)
}

func (s *Logger) QueryBatteryOutput() (map[string]interface{}, error) {
	return readBatteryOutput(s.connPort, s.serialNumber)
}

func (s *Logger) QueryPVOutput() (map[string]interface{}, error) {
	return readPVOutput(s.connPort, s.serialNumber)
}

func NewStation(lastUpdateTime string, lastUpdateTimeUnix, generationTotal, generationPower, chargePower, dischargePower, batteryPower, batterySOC, usePower int) *Station {
	return &Station{
		lastUpdateTime:     lastUpdateTime,
		lastUpdateTimeUnix: lastUpdateTimeUnix,
		generationTotal:    generationTotal,
		generationPower:    generationPower,
		chargePower:        chargePower,
		dischargePower:     dischargePower,
		batteryPower:       batteryPower,
		batterySOC:         batterySOC,
		usePower:           usePower,
	}
}

// func NewInverter(deviceSN string, deviceId int, deviceType string, deviceState, collectionTime int) *Inverter {
// 	return &Inverter{
// 		deviceSN:       deviceSN,
// 		deviceId:       deviceId,
// 		deviceType:     deviceType,
// 		deviceState:    deviceState,
// 		collectionTime: collectionTime,
// 	}
// }

// func NewDataLogger(deviceSN string, deviceId int, deviceType string, deviceState, collectionTime int) *DataLogger {
// 	return &DataLogger{
// 		deviceSN:       deviceSN,
// 		deviceId:       deviceId,
// 		deviceType:     deviceType,
// 		deviceState:    deviceState,
// 		collectionTime: collectionTime,
// 	}
// }
