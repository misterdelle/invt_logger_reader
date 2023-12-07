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

	// fmt.Printf("serial number: %0X\n", uint32SerialNumber)
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
		case "STRING12":
			mr := modbusReply[fieldOffset : fieldOffset+12]
			v1 := int(mr[0])
			v2 := int(mr[1])
			v3 := int(mr[2])
			v4 := int(mr[3])
			v5 := int(mr[4])
			v6 := int(mr[5])
			v7 := int(mr[6])
			v8 := int(mr[7])
			v9 := int(mr[8])
			v10 := int(mr[9])
			v11 := int(mr[10])
			v12 := int(mr[11])
			reply[f.name] = fmt.Sprintf("%v%v%v%v%v%v%v%v%v%v%v%v", v1, v2, v3, v4, v5, v6, v7, v8, v9, v10, v11, v12)
		case "U8":
			mr := modbusReply[fieldOffset : fieldOffset+2]
			v1 := int(mr[0])
			v2 := int(mr[1])
			reply[f.name] = fmt.Sprintf("%v-%v", v1, v2)
		case "U16":
			mr := modbusReply[fieldOffset : fieldOffset+2]
			be := binary.BigEndian.Uint16(mr)
			v := float64(be) * float64(f.factor)
			// v := float64(binary.BigEndian.Uint16(modbusReply[fieldOffset:fieldOffset+2])) * float64(f.factor)
			// reply[f.name] = float32(binary.BigEndian.Uint16(modbusReply[fieldOffset:fieldOffset+2])) * f.factor
			reply[f.name] = strconv.FormatFloat(v, 'f', 2, 64)
		case "U32":
			// mr := modbusReply[fieldOffset : fieldOffset+4]
			// be := binary.BigEndian.Uint32(mr)
			// v := float64(be) * float64(f.factor)

			mr1 := modbusReply[fieldOffset : fieldOffset+2]
			mr2 := modbusReply[fieldOffset+2 : fieldOffset+4]
			be1 := binary.BigEndian.Uint16(mr1)
			be2 := binary.BigEndian.Uint16(mr2)
			_ = be1
			_ = be2

			v := float64(be1) * float64(f.factor)
			// v := float64(binary.BigEndian.Uint32(modbusReply[fieldOffset:fieldOffset+4])) * float64(f.factor)
			// reply[f.name] = float32(binary.BigEndian.Uint32(modbusReply[fieldOffset:fieldOffset+4])) * f.factor
			reply[f.name] = strconv.FormatFloat(v, 'f', 2, 64)
		case "S16":
			mr := modbusReply[fieldOffset : fieldOffset+2]
			// be := binary.BigEndian.Uint16(mr)
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
	year := "20" + yearMonth[:sep]
	month := yearMonth[sep+1:]
	mm, _ := strconv.Atoi(month)
	if mm <= 9 {
		month = "0" + month
	}

	sep = strings.Index(dayRes, "-")
	day := dayRes[:sep]
	dd, _ := strconv.Atoi(day)
	if dd <= 9 {
		day = "0" + day
	}

	sep = strings.Index(hourMinute, "-")
	hour := hourMinute[:sep]
	hh, _ := strconv.Atoi(hour)
	if hh <= 9 {
		hour = "0" + hour
	}

	minute := hourMinute[sep+1:]
	min, _ := strconv.Atoi(minute)
	if min <= 9 {
		minute = "0" + minute
	}

	sep = strings.Index(secondDayOfWeek, "-")
	second := secondDayOfWeek[:sep]
	sec, _ := strconv.Atoi(second)
	if sec <= 9 {
		second = "0" + second
	}
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

	result = make(map[string]interface{})

	lastUpdateTime := fmt.Sprintf("%s-%s-%s %s:%s:%s", year, month, day, hour, minute, second)
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

	// generationTotal := 0
	// generationPower := 0
	// chargePower := 0
	// dischargePower := 0
	// batteryPower := 0
	// usePower := 0

	// station := NewStation(lastUpdateTime, lastUpdateTimeUnix, generationTotal, generationPower, chargePower, dischargePower, batteryPower, batterySOC, usePower)

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
