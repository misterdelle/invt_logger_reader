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

type Inverter struct {
	deviceSN       string
	deviceId       int
	deviceType     string
	deviceState    int
	collectionTime int
}

type DataLogger struct {
	deviceSN       string
	deviceId       int
	deviceType     string
	deviceState    int
	collectionTime int
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

func NewInverter(deviceSN string, deviceId int, deviceType string, deviceState, collectionTime int) *Inverter {
	return &Inverter{
		deviceSN:       deviceSN,
		deviceId:       deviceId,
		deviceType:     deviceType,
		deviceState:    deviceState,
		collectionTime: collectionTime,
	}
}

func NewDataLogger(deviceSN string, deviceId int, deviceType string, deviceState, collectionTime int) *DataLogger {
	return &DataLogger{
		deviceSN:       deviceSN,
		deviceId:       deviceId,
		deviceType:     deviceType,
		deviceState:    deviceState,
		collectionTime: collectionTime,
	}
}
