package invt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/misterdelle/invt_logger_reader/ports"
	"github.com/sigurn/crc16"
)

type LSWRequest struct {
	serialNumber  uint
	startRegister int
	endRegister   int
}

func NewLSWRequest(serialNumber uint, startRegister int, endRegister int) LSWRequest {
	return LSWRequest{
		serialNumber:  serialNumber,
		startRegister: startRegister,
		endRegister:   endRegister,
	}
}

func (l LSWRequest) ToBytes() []byte {
	buf := make([]byte, 36)

	// preamble
	buf[0] = 0xa5
	binary.BigEndian.PutUint16(buf[1:], 0x1700)
	binary.BigEndian.PutUint16(buf[3:], 0x1045)
	buf[5] = 0x00
	buf[6] = 0x00

	binary.LittleEndian.PutUint32(buf[7:], uint32(l.serialNumber))

	buf[11] = 0x02

	binary.BigEndian.PutUint16(buf[26:], 0x0103)

	binary.BigEndian.PutUint16(buf[28:], uint16(l.startRegister))
	binary.BigEndian.PutUint16(buf[30:], uint16(l.endRegister-l.startRegister+1))

	// compute crc
	table := crc16.MakeTable(crc16.CRC16_MODBUS)
	modbusCRC := crc16.Checksum(buf[26:32], table)

	// append crc
	binary.LittleEndian.PutUint16(buf[32:], modbusCRC)

	// compute & append frame crc
	buf[34] = l.checksum(buf)

	// end of frame
	buf[35] = 0x15

	return buf

}

func (l LSWRequest) String() string {
	return fmt.Sprintf("% 0X", l.ToBytes())
}

func (l LSWRequest) checksum(buf []byte) uint8 {
	var checksum uint8
	for _, b := range buf[1 : len(buf)-2] {
		checksum += b
	}
	return checksum
}

func readData(connPort ports.CommunicationPort, serialNumber uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, rr := range allRegisterRanges {
		reply, err := readRegisterRange(rr, connPort, serialNumber)
		if err != nil {
			return nil, err
		}

		for k, v := range reply {
			result[k] = v
		}
	}
	return result, nil
}

func readRegisterRange(rr registerRange, connPort ports.CommunicationPort, serialNumber uint) (map[string]interface{}, error) {
	lswRequest := NewLSWRequest(serialNumber, rr.start, rr.end)

	commandBytes := lswRequest.ToBytes()

	err := connPort.Open()
	if err != nil {
		return nil, err
	}

	defer func(connPort ports.CommunicationPort) {
		if err := connPort.Close(); err != nil {
			log.Printf("error during connection close: %s", err)
		}
	}(connPort)

	// send the command
	_, err = connPort.Write(commandBytes)
	if err != nil {
		return nil, err
	}

	// read the result
	buf := make([]byte, 2048)
	n, err := connPort.Read(buf)
	if err != nil {
		return nil, err
	}

	// truncate the buffer
	buf = buf[:n]
	if len(buf) < 27 {
		// short reply
		return nil, fmt.Errorf("short reply: %d bytes", n)
	}

	replyBytesCount := buf[27]

	modbusReply := buf[28 : 28+replyBytesCount]

	// shove the data into the reply
	reply := make(map[string]interface{})

	for _, f := range rr.replyFields {
		fieldOffset := (f.register - rr.start) * 2

		if fieldOffset > len(modbusReply)-2 {
			// skip invalid offset
			continue
		}

		switch f.valueType {
		case "U8":
			mr := modbusReply[fieldOffset : fieldOffset+2]
			v1 := int(mr[0])
			v2 := int(mr[1])
			reply[f.name] = fmt.Sprintf("%v-%v", v1, v2)
		case "U16":
			mr := modbusReply[fieldOffset : fieldOffset+2]
			be := binary.BigEndian.Uint16(mr)
			v := float64(be) * float64(f.factor)
			reply[f.name] = strconv.FormatFloat(v, 'f', 2, 64)
		case "U32":
			mr1 := modbusReply[fieldOffset : fieldOffset+2]
			// mr2 := modbusReply[fieldOffset+2 : fieldOffset+4]
			be1 := binary.BigEndian.Uint16(mr1)
			// be2 := binary.BigEndian.Uint16(mr2)

			v := float64(be1) * float64(f.factor)
			reply[f.name] = strconv.FormatFloat(v, 'f', 2, 64)
		case "S16":
			mr := modbusReply[fieldOffset : fieldOffset+2]
			be := TwoComplement(mr)
			v := float64(be) * float64(f.factor)
			reply[f.name] = strconv.FormatFloat(v, 'f', 2, 64)
		default:
		}
	}

	return reply, nil
}

func readStationData(connPort ports.CommunicationPort, serialNumber uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, rr := range stationRegisterRanges {
		if rr.start == 0x3195 {
			fmt.Println(rr.start)
		}

		reply, err := readRegisterRange(rr, connPort, serialNumber)
		if err != nil {
			return nil, err
		}

		for k, v := range reply {
			result[k] = v
		}
	}

	yearMonth := result["Year_Month"].(string)
	dayRes := result["Day_Res"].(string)
	hourMinute := result["Hour_Minute"].(string)
	secondDayOfWeek := result["Second_DayOfWeek"].(string)

	sep := strings.Index(yearMonth, "-")
	year, _ := strconv.Atoi("20" + yearMonth[:sep])
	month, _ := strconv.Atoi(yearMonth[sep+1:])

	sep = strings.Index(dayRes, "-")
	day, _ := strconv.Atoi(dayRes[:sep])

	sep = strings.Index(hourMinute, "-")
	hour, _ := strconv.Atoi(hourMinute[:sep])
	minute, _ := strconv.Atoi(hourMinute[sep+1:])

	sep = strings.Index(secondDayOfWeek, "-")
	second, _ := strconv.Atoi(secondDayOfWeek[:sep])
	dayOfWeek := secondDayOfWeek[sep+1:]
	_ = dayOfWeek

	batterySOC := result["batterySOC"].(string)
	batteryPower := result["batteryPower"].(string)
	currentConsumptionPower := result["currentConsumptionPower"].(string)
	batteryChargeDayEnergy := result["Bat Charge Day Energy"].(string)
	batteryDischargeDayEnergy := result["Bat Discharge Day Energy"].(string)
	batteryChargeTotalEnergy := result["Bat Charge Total Energy"].(string)
	batteryDischargeTotalEnergy := result["Bat Discharge Total Energy"].(string)
	pvDayEnergy := result["PV Day Energy"].(string)
	gridDayEnergy := result["Grid Day Energy"].(string)
	loadDayEnergy := result["Load Day Energy"].(string)
	pvTotalEnergy := result["PV Total Energy"].(string)
	gridTotalEnergy := result["Grid Total Energy"].(string)
	loadTotalEnergy := result["Load Total Energy"].(string)
	purchasingDayEnergy := result["Purchasing Day Energy"].(string)
	purchasingTotalEnergy := result["Purchasing Total Energy"].(string)
	powerPV1 := result["Power PV1"].(string)
	powerPV2 := result["Power PV2"].(string)
	powPV1, _ := strconv.ParseFloat(powerPV1, 64)
	powPV2, _ := strconv.ParseFloat(powerPV2, 64)
	totalPowerFromPV := fmt.Sprintf("%v", powPV1+powPV2)

	result = make(map[string]interface{})

	lastUpdateTime := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, minute, second)
	t, err := time.Parse("2006-01-02 15:04:05", lastUpdateTime)
	if err != nil {
		fmt.Println(err)
	}
	lastUpdateTimeUnix := t.Unix()

	result["lastUpdateTime"] = lastUpdateTime
	result["lastUpdateTimeUnix"] = lastUpdateTimeUnix
	result["batterySOC"] = batterySOC
	result["batteryPower"] = batteryPower
	result["currentConsumptionPower"] = currentConsumptionPower
	result["batteryChargeDayEnergy"] = batteryChargeDayEnergy
	result["batteryDischargeDayEnergy"] = batteryDischargeDayEnergy
	result["batteryChargeTotalEnergy"] = batteryChargeTotalEnergy
	result["batteryDischargeTotalEnergy"] = batteryDischargeTotalEnergy
	result["pvDayEnergy"] = pvDayEnergy
	result["gridDayEnergy"] = gridDayEnergy
	result["loadDayEnergy"] = loadDayEnergy
	result["pvTotalEnergy"] = pvTotalEnergy
	result["gridTotalEnergy"] = gridTotalEnergy
	result["loadTotalEnergy"] = loadTotalEnergy
	result["purchasingDayEnergy"] = purchasingDayEnergy
	result["purchasingTotalEnergy"] = purchasingTotalEnergy
	result["powerFromPV1"] = powerPV1
	result["powerFromPV2"] = powerPV2
	result["totalPowerFromPV"] = totalPowerFromPV

	return result, nil
}

func readEnergyTodayTotalsData(connPort ports.CommunicationPort, serialNumber uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, rr := range energyTodayTotalsRegisterRanges {
		reply, err := readRegisterRange(rr, connPort, serialNumber)
		if err != nil {
			return nil, err
		}

		for k, v := range reply {
			result[k] = v
		}
	}

	sBUSVoltage := result["ETT: S BUS Voltage"].(string)
	nBUSVoltage := result["ETT: N BUS Voltage"].(string)
	dcdcTemperature := result["ETT: DCDC Temperature"].(string)
	pvDayEnergy := result["ETT: PV Day Energy"].(string)
	gridDayEnergy := result["ETT: Grid Day Energy"].(string)
	loadDayEnergy := result["ETT: Load Day Energy"].(string)
	pvMonthEnergy := result["ETT: PV Month Energy"].(string)
	gridMonthEnergy := result["ETT: Grid Month Energy"].(string)
	loadMonthEnergy := result["ETT: Load Month Energy"].(string)
	pvYearEnergy := result["ETT: PV Year Energy"].(string)
	gridYearEnergy := result["ETT: Grid Year Energy"].(string)
	loadYearEnergy := result["ETT: Load Year Energy"].(string)
	pvTotalEnergy := result["ETT: PV Total Energy"].(string)
	gridTotalEnergy := result["ETT: Grid Total Energy"].(string)
	loadTotalEnergy := result["ETT: Load Total Energy"].(string)
	purchasingDayEnergy := result["ETT: Purchasing Day Energy"].(string)
	batChargeDayEnergy := result["ETT: Bat Charge Day Energy"].(string)
	batDischargeDayEnergy := result["ETT: Bat Discharge Day Energy"].(string)
	purchasingMonthEnergy := result["ETT: Purchasing Month Energy"].(string)
	batChargeMonthEnergy := result["ETT: Bat Charge Month Energy"].(string)
	batDischargeMonthEnergy := result["ETT: Bat Discharge Month Energy"].(string)
	purchasingYearEnergy := result["ETT: Purchasing Year Energy"].(string)
	batChargeYearEnergy := result["ETT: Bat Charge Year Energy"].(string)
	batDischargeYearEnergy := result["ETT: Bat Discharge Year Energy"].(string)
	purchasingTotalEnergy := result["ETT: Purchasing Total Energy"].(string)
	batChargeTotalEnergy := result["ETT: Bat Charge Total Energy"].(string)
	batDischargeTotalEnergy := result["ETT: Bat Discharge Total Energy"].(string)

	result = make(map[string]interface{})

	result["S BUS Voltage"] = sBUSVoltage
	result["N BUS Voltage"] = nBUSVoltage
	result["DC DC Temperature"] = dcdcTemperature
	result["PV Day Energy"] = pvDayEnergy
	result["Grid Day Energy"] = gridDayEnergy
	result["Load Day Energy"] = loadDayEnergy
	result["PV Month Energy"] = pvMonthEnergy
	result["Grid Month Energy"] = gridMonthEnergy
	result["Load Month Energy"] = loadMonthEnergy
	result["PV Year Energy"] = pvYearEnergy
	result["Grid Year Energy"] = gridYearEnergy
	result["Load Year Energy"] = loadYearEnergy
	result["PV Total Energy"] = pvTotalEnergy
	result["Grid Total Energy"] = gridTotalEnergy
	result["Load Total Energy"] = loadTotalEnergy
	result["Purchasing Day Energy"] = purchasingDayEnergy
	result["BAT Charge Day Energy"] = batChargeDayEnergy
	result["BAT Discharge Day Energy"] = batDischargeDayEnergy
	result["Purchasing Month Energy"] = purchasingMonthEnergy
	result["BAT Charge Month Energy"] = batChargeMonthEnergy
	result["BAT Discharge Month Energy"] = batDischargeMonthEnergy
	result["Purchasing Year Energy"] = purchasingYearEnergy
	result["BAT Charge Year Energy"] = batChargeYearEnergy
	result["BAT Discharge Year Energy"] = batDischargeYearEnergy
	result["Purchasing Total Energy"] = purchasingTotalEnergy
	result["BAT Charge Total Energy"] = batChargeTotalEnergy
	result["BAT Discharge Total Energy"] = batDischargeTotalEnergy

	return result, nil
}

func readGridOutput(connPort ports.CommunicationPort, serialNumber uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, rr := range gridOutputRegisterRanges {
		reply, err := readRegisterRange(rr, connPort, serialNumber)
		if err != nil {
			return nil, err
		}

		for k, v := range reply {
			result[k] = v
		}
	}

	gridAVoltage := result["GO: Grid A Voltage"].(string)
	gridACurrent := result["GO: Grid A Current"].(string)
	gridAPower := result["GO: Grid A Power"].(string)
	gridBVoltage := result["GO: Grid B Voltage"].(string)
	gridBCurrent := result["GO: Grid B Current"].(string)
	gridBPower := result["GO: Grid B Power"].(string)
	gridCVoltage := result["GO: Grid C Voltage"].(string)
	gridCCurrent := result["GO: Grid C Current"].(string)
	gridCPower := result["GO: Grid C Power"].(string)
	gridFreq := result["GO: Grid Freq"].(string)
	inv1Temperature := result["GO: INV1 Temperature"].(string)
	inv2Temperature := result["GO: INV2 Temperature"].(string)

	result = make(map[string]interface{})

	result["Grid A Voltage"] = gridAVoltage
	result["Grid A Current"] = gridACurrent
	result["Grid A Power"] = gridAPower
	result["Grid B Voltage"] = gridBVoltage
	result["Grid B Current"] = gridBCurrent
	result["Grid B Power"] = gridBPower
	result["Grid C Voltage"] = gridCVoltage
	result["Grid C Current"] = gridCCurrent
	result["Grid C Power"] = gridCPower
	result["Grid Freq"] = gridFreq
	result["Inv 1 Temperature"] = inv1Temperature
	result["Inv 2 Temperature"] = inv2Temperature

	return result, nil
}

func readInverterInfo(connPort ports.CommunicationPort, serialNumber uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, rr := range inverterInfoRegisterRanges {
		reply, err := readRegisterRange(rr, connPort, serialNumber)
		if err != nil {
			return nil, err
		}

		for k, v := range reply {
			result[k] = v
		}
	}

	invAVoltage := result["II: INV A Voltage"].(string)
	invACurrent := result["II: INV A Current"].(string)
	invAPower := result["II: INV A Power"].(string)
	invBVoltage := result["II: INV B Voltage"].(string)
	invBCurrent := result["II: INV B Current"].(string)
	invBPower := result["II: INV B Power"].(string)
	invCVoltage := result["II: INV C Voltage"].(string)
	invCCurrent := result["II: INV C Current"].(string)
	invCPower := result["II: INV C Power"].(string)
	invAFreq := result["II: INV A Freq"].(string)
	invBFreq := result["II: INV B Freq"].(string)
	invCFreq := result["II: INV C Freq"].(string)
	leakCurrent := result["II: Leak Current"].(string)

	result = make(map[string]interface{})

	result["Inv A Voltage"] = invAVoltage
	result["Inv A Current"] = invACurrent
	result["Inv A Power"] = invAPower
	result["Inv B Voltage"] = invBVoltage
	result["Inv B Current"] = invBCurrent
	result["Inv B Power"] = invBPower
	result["Inv C Voltage"] = invCVoltage
	result["Inv C Current"] = invCCurrent
	result["Inv C Power"] = invCPower
	result["Inv A Freq"] = invAFreq
	result["Inv B Freq"] = invBFreq
	result["Inv C Freq"] = invCFreq
	result["Leak Current"] = leakCurrent

	return result, nil
}

func readLoadInfo(connPort ports.CommunicationPort, serialNumber uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, rr := range loadInfoRegisterRanges {
		reply, err := readRegisterRange(rr, connPort, serialNumber)
		if err != nil {
			return nil, err
		}

		for k, v := range reply {
			result[k] = v
		}
	}

	loadAVoltage := result["LI: Load A Voltage"].(string)
	loadACurrent := result["LI: Load A Current"].(string)
	loadAPower := result["LI: Load A Power"].(string)
	loadARate := result["LI: Load A Rate"].(string)
	loadBVoltage := result["LI: Load B Voltage"].(string)
	loadBCurrent := result["LI: Load B Current"].(string)
	loadBPower := result["LI: Load B Power"].(string)
	loadBRate := result["LI: Load B Rate"].(string)
	loadCVoltage := result["LI: Load C Voltage"].(string)
	loadCCurrent := result["LI: Load C Current"].(string)
	loadCPower := result["LI: Load C Power"].(string)
	loadCRate := result["LI: Load C Rate"].(string)
	generatorPortVoltageA := result["LI: Generator Port Voltage A"].(string)
	generatorPortVoltageB := result["LI: Generator Port Voltage B"].(string)
	generatorPortVoltageC := result["LI: Generator Port Voltage C"].(string)

	result = make(map[string]interface{})

	result["Load A Voltage"] = loadAVoltage
	result["Load A Current"] = loadACurrent
	result["Load A Power"] = loadAPower
	result["Load A Rate"] = loadARate
	result["Load B Voltage"] = loadBVoltage
	result["Load B Current"] = loadBCurrent
	result["Load B Power"] = loadBPower
	result["Load B Rate"] = loadBRate
	result["Load C Voltage"] = loadCVoltage
	result["Load C Current"] = loadCCurrent
	result["Load C Power"] = loadCPower
	result["Load C Rate"] = loadCRate
	result["Generator Port Voltage A"] = generatorPortVoltageA
	result["Generator Port Voltage B"] = generatorPortVoltageB
	result["Generator Port Voltage C"] = generatorPortVoltageC

	return result, nil
}

func readBatteryOutput(connPort ports.CommunicationPort, serialNumber uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, rr := range batteryOutputRanges {
		reply, err := readRegisterRange(rr, connPort, serialNumber)
		if err != nil {
			return nil, err
		}

		for k, v := range reply {
			result[k] = v
		}
	}

	batVoltage := result["BO: BAT Voltage"].(string)
	batCurrent := result["BO: BAT Current"].(string)
	bat1Current := result["BO: BAT 1 Current"].(string)
	bat2Current := result["BO: BAT 2 Current"].(string)
	bat3Current := result["BO: BAT 3 Current"].(string)
	batSOC := result["BO: BAT SOC"].(string)
	batTemperature := result["BO: BAT Temperature"].(string)
	batChargeVoltage := result["BO: BAT Charge Voltage"].(string)
	batChargeCurrentLimit := result["BO: BAT Charge Current Limit"].(string)
	batDischargeCurrentLimit := result["BO: BAT Discharge Current Limit"].(string)
	batPower := result["BO: BAT Power"].(string)
	bmsBatVoltage := result["BO: BMS BAT Voltage"].(string)
	bmsBatCurrent := result["BO: BMS BAT Current"].(string)
	bmsBatCellMaxVoltage := result["BO: BMS BAT Cell Max Voltage"].(string)
	bmsBatCellMinVoltage := result["BO: BMS BAT Cell Min Voltage"].(string)
	bmsBatCellMaxTemperature := result["BO: BMS BAT Cell Max Temperature"].(string)
	bmsBatCellMinTemperature := result["BO: BMS BAT Cell Min Temperature"].(string)

	result = make(map[string]interface{})

	result["BAT Voltage"] = batVoltage
	result["BAT Current"] = batCurrent
	result["BAT 1 Current"] = bat1Current
	result["BAT 2 Current"] = bat2Current
	result["BAT 3 Current"] = bat3Current
	result["BAT SOC"] = batSOC
	result["BAT Temperature"] = batTemperature
	result["BAT Charge Voltage"] = batChargeVoltage
	result["BAT Charge Current Limit"] = batChargeCurrentLimit
	result["BAT Discharge Current Limit"] = batDischargeCurrentLimit
	result["BAT Power"] = batPower
	result["BMS BAT Voltage"] = bmsBatVoltage
	result["BMS BAT Current"] = bmsBatCurrent
	result["BMS BAT Cell Max Voltage"] = bmsBatCellMaxVoltage
	result["BMS BAT Cell Min Voltage"] = bmsBatCellMinVoltage
	result["BMS BAT Cell Max Temperature"] = bmsBatCellMaxTemperature
	result["BMS BAT Cell Min Temperature"] = bmsBatCellMinTemperature

	return result, nil
}

func readPVOutput(connPort ports.CommunicationPort, serialNumber uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, rr := range pvOutputRanges {
		reply, err := readRegisterRange(rr, connPort, serialNumber)
		if err != nil {
			return nil, err
		}

		for k, v := range reply {
			result[k] = v
		}
	}

	voltagePV1 := result["PV: Voltage_PV1"].(string)
	currentPV1 := result["PV: Current_PV1"].(string)
	powerPV1 := result["PV: Power_PV1"].(string)
	voltagePV2 := result["PV: Voltage_PV2"].(string)
	currentPV2 := result["PV: Current_PV2"].(string)
	powerPV2 := result["PV: Power_PV2"].(string)

	result = make(map[string]interface{})

	result["Voltage PV 1"] = voltagePV1
	result["Current PV 1"] = currentPV1
	result["Power PV 1"] = powerPV1
	result["Voltage PV 2"] = voltagePV2
	result["Current PV 2"] = currentPV2
	result["Power PV 2"] = powerPV2

	return result, nil
}

func TwoComplement(b []byte) int16 {
	var v int16
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &v)
	if err != nil {
		fmt.Println(err)
	}

	return v
}
