package main

import (
	"flag"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/misterdelle/invt_logger_reader/adapters/comms/tcpip"
	"github.com/misterdelle/invt_logger_reader/adapters/devices/invt"
	"github.com/misterdelle/invt_logger_reader/adapters/export/mosquitto"
	"github.com/misterdelle/invt_logger_reader/ports"
)

// maximumFailedConnections maximum number failed logger connection, after this number will be exceeded reconnect
// interval will be extended from 5s to readInterval defined in config file
const maximumFailedConnections = 3

type Application struct {
	Env                  string
	InverterPort         string
	InverterLoggerSerial uint
	InverterReadInterval int
	MQTTURL              string
	MQTTUser             string
	MQTTPassword         string
	MQTTTopicName        string
}

var (
	config *Config
	port   ports.CommunicationPort
	mqtt   ports.DatabaseWithListener
	device ports.Device

	hasMQTT bool
)

// Set up an app config
var app = Application{}

func init() {
	flag.Parse()

	if app.Env != "" {
		fmt.Printf("app.Env        : %s \n", app.Env)
		godotenv.Load(".env." + app.Env + ".local")
		godotenv.Load(".env." + app.Env)
	} else {
		fmt.Println("app.Env NON settato, carico i dati dal file .env")
		godotenv.Load() // The Original .env
		app.Env = os.Getenv("Env")
		fmt.Printf("app.Env                 : %s \n", app.Env)
	}

	app.InverterPort = os.Getenv("inverter.port")
	inverterLoggerSerial, _ := strconv.Atoi(os.Getenv("inverter.loggerSerial"))
	app.InverterLoggerSerial = uint(inverterLoggerSerial)
	app.InverterReadInterval, _ = strconv.Atoi(os.Getenv("inverter.readInterval"))

	app.MQTTURL = os.Getenv("mqtt.url")
	app.MQTTUser = os.Getenv("mqtt.user")
	app.MQTTPassword = os.Getenv("mqtt.password")
	app.MQTTTopicName = os.Getenv("mqtt.prefix")

	fmt.Printf("app.InverterPort        : %s \n", app.InverterPort)
	fmt.Printf("app.InverterLoggerSerial: %d \n", app.InverterLoggerSerial)
	fmt.Printf("app.InverterReadInterval: %d \n", app.InverterReadInterval)
	fmt.Printf("app.MQTTURL             : %s \n", app.MQTTURL)
	fmt.Printf("app.MQTTUser            : %s \n", app.MQTTUser)
	fmt.Printf("app.MQTTPassword        : %s \n", app.MQTTPassword)
	fmt.Printf("app.MQTTTopicName       : %s \n", app.MQTTTopicName)

	var err error
	config, err = NewConfig(app)
	if err != nil {
		log.Fatalln(err)
	}

	hasMQTT = config.Mqtt.Url != "" && config.Mqtt.Prefix != ""

	port = tcpip.New(config.Inverter.Port)
	log.Printf("using TCP/IP communications port %s", config.Inverter.Port)

	if hasMQTT {
		mqtt, err = mosquitto.New(&config.Mqtt)
		if err != nil {
			log.Fatalf("MQTT connection failed: %s", err)
		}

		log.Printf("using MQTT at URL %s", config.Mqtt.Url)
	}

	device = invt.NewInvtLogger(config.Inverter.LoggerSerial, port)
}

func main() {
	for {
		log.Printf("performing measurements")
		timeStart := time.Now()

		// //
		// // Measurements
		// //
		// measurements, err := device.Query()
		// if err != nil {
		// 	log.Printf("failed to perform measurements: %s", err)
		// 	failedConnections++

		// 	if failedConnections > maximumFailedConnections {
		// 		time.Sleep(time.Duration(config.Inverter.ReadInterval) * time.Second)
		// 	}

		// 	continue
		// }

		// log.Println("Inverter Measurement: ", measurements)

		// failedConnections = 0

		// if hasMQTT {
		// 	go func() {
		// 		err = mqtt.InsertRecord(measurements)
		// 		if err != nil {
		// 			log.Printf("failed to insert record to MQTT: %s\n", err)
		// 		} else {
		// 			log.Println("measurements pushed to MQTT")
		// 		}
		// 	}()
		// }

		err := loadStation()
		if err != nil {
			continue
		}

		err = loadEnergyTodayTotals()
		if err != nil {
			continue
		}

		err = loadGridOutput()
		if err != nil {
			continue
		}

		err = loadInverterInfo()
		if err != nil {
			continue
		}

		err = loadLoadInfo()
		if err != nil {
			continue
		}

		err = loadBatteryOutput()
		if err != nil {
			continue
		}

		err = loadPVOutput()
		if err != nil {
			continue
		}

		duration := time.Since(timeStart)

		delay := time.Duration(config.Inverter.ReadInterval)*time.Second - duration
		if delay <= 0 {
			delay = 1 * time.Second
		}

		time.Sleep(delay)
	}

}

// func isSerialPort(portName string) bool {
// 	return strings.HasPrefix(portName, "/")
// }

func loadStation() error {
	failedConnections := 0

	//
	// Station
	//
	measurementsStation, err := device.QueryStation()

	if err != nil {
		log.Printf("failed to perform measurementsStation: %s", err)
		failedConnections++

		if failedConnections > maximumFailedConnections {
			time.Sleep(time.Duration(config.Inverter.ReadInterval) * time.Second)
		}

		return err
	}

	log.Println("Station Measurement: ", measurementsStation)

	failedConnections = 0

	if hasMQTT {
		go func() {
			err = mqtt.InsertGenericRecord("station", measurementsStation)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsStation pushed to MQTT")
			}
		}()
	}

	return nil

}

func loadEnergyTodayTotals() error {
	failedConnections := 0

	//
	// Energy Today Totals
	//
	measurementsEnergyTodayTotals, err := device.QueryEnergyTodayTotals()

	if err != nil {
		log.Printf("failed to perform measurementsEnergyTodayTotals: %s", err)
		failedConnections++

		if failedConnections > maximumFailedConnections {
			time.Sleep(time.Duration(config.Inverter.ReadInterval) * time.Second)
		}

		return err
	}

	pv := make(map[string]interface{})
	grid := make(map[string]interface{})
	load := make(map[string]interface{})
	purchase := make(map[string]interface{})
	batCharge := make(map[string]interface{})
	batDischarge := make(map[string]interface{})
	root := make(map[string]interface{})

	pv["PV Day Energy"] = measurementsEnergyTodayTotals["PV Day Energy"]
	pv["PV Month Energy"] = measurementsEnergyTodayTotals["PV Month Energy"]
	pv["PV Year Energy"] = measurementsEnergyTodayTotals["PV Year Energy"]
	pv["PV Total Energy"] = measurementsEnergyTodayTotals["PV Total Energy"]

	grid["Grid Day Energy"] = measurementsEnergyTodayTotals["Grid Day Energy"]
	grid["Grid Month Energy"] = measurementsEnergyTodayTotals["Grid Month Energy"]
	grid["Grid Year Energy"] = measurementsEnergyTodayTotals["Grid Year Energy"]
	grid["Grid Total Energy"] = measurementsEnergyTodayTotals["Grid Total Energy"]

	load["Load Day Energy"] = measurementsEnergyTodayTotals["Load Day Energy"]
	load["Load Month Energy"] = measurementsEnergyTodayTotals["Load Month Energy"]
	load["Load Year Energy"] = measurementsEnergyTodayTotals["Load Year Energy"]
	load["Load Total Energy"] = measurementsEnergyTodayTotals["Load Total Energy"]

	purchase["Purchasing Day Energy"] = measurementsEnergyTodayTotals["Purchasing Day Energy"]
	purchase["Purchasing Month Energy"] = measurementsEnergyTodayTotals["Purchasing Month Energy"]
	purchase["Purchasing Year Energy"] = measurementsEnergyTodayTotals["Purchasing Year Energy"]
	purchase["Purchasing Total Energy"] = measurementsEnergyTodayTotals["Purchasing Total Energy"]

	batCharge["BAT Charge Day Energy"] = measurementsEnergyTodayTotals["BAT Charge Day Energy"]
	batCharge["BAT Charge Month Energy"] = measurementsEnergyTodayTotals["BAT Charge Month Energy"]
	batCharge["BAT Charge Year Energy"] = measurementsEnergyTodayTotals["BAT Charge Year Energy"]
	batCharge["BAT Charge Total Energy"] = measurementsEnergyTodayTotals["BAT Charge Total Energy"]

	batDischarge["BAT Discharge Day Energy"] = measurementsEnergyTodayTotals["BAT Discharge Day Energy"]
	batDischarge["BAT Discharge Month Energy"] = measurementsEnergyTodayTotals["BAT Discharge Month Energy"]
	batDischarge["BAT Discharge Year Energy"] = measurementsEnergyTodayTotals["BAT Discharge Year Energy"]
	batDischarge["BAT Discharge Total Energy"] = measurementsEnergyTodayTotals["BAT Discharge Total Energy"]

	root["S BUS Voltage"] = measurementsEnergyTodayTotals["S BUS Voltage"]
	root["N BUS Voltage"] = measurementsEnergyTodayTotals["N BUS Voltage"]
	root["DC DC Temperature"] = measurementsEnergyTodayTotals["DC DC Temperature"]

	log.Println("measurementsEnergyTodayTotals: ", measurementsEnergyTodayTotals)

	failedConnections = 0

	if hasMQTT {
		err = mqtt.InsertGenericRecord("EnergyTodayTotals", root)
		if err != nil {
			log.Printf("failed to insert record to MQTT: %s\n", err)
		} else {
			log.Println("measurementsEnergyTodayTotals pushed to MQTT")
		}

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsEnergyTodayTotals pushed to MQTT")
			}
		}("EnergyTodayTotals/PV", pv)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsEnergyTodayTotals pushed to MQTT")
			}
		}("EnergyTodayTotals/Grid", grid)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsEnergyTodayTotals pushed to MQTT")
			}
		}("EnergyTodayTotals/Load", load)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsEnergyTodayTotals pushed to MQTT")
			}
		}("EnergyTodayTotals/Purchase", purchase)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsEnergyTodayTotals pushed to MQTT")
			}
		}("EnergyTodayTotals/Battery Charge", batCharge)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsEnergyTodayTotals pushed to MQTT")
			}
		}("EnergyTodayTotals/Battery Discharge", batDischarge)
	}

	return nil

}

func loadGridOutput() error {
	failedConnections := 0

	//
	// Grid Output
	//
	measurementsGridOutput, err := device.QueryGridOutput()

	if err != nil {
		log.Printf("failed to perform measurementsGridOutput: %s", err)
		failedConnections++

		if failedConnections > maximumFailedConnections {
			time.Sleep(time.Duration(config.Inverter.ReadInterval) * time.Second)
		}

		return err
	}

	gridA := make(map[string]interface{})
	gridB := make(map[string]interface{})
	gridC := make(map[string]interface{})
	root := make(map[string]interface{})

	gridA["Inv A Voltage"] = measurementsGridOutput["Grid A Voltage"]
	gridA["Inv A Current"] = measurementsGridOutput["Grid A Current"]
	gridA["Inv A Power"] = measurementsGridOutput["Grid A Power"]

	gridB["Inv B Voltage"] = measurementsGridOutput["Grid B Voltage"]
	gridB["Inv B Current"] = measurementsGridOutput["Grid B Current"]
	gridB["Inv B Power"] = measurementsGridOutput["Grid B Power"]

	gridC["Inv C Voltage"] = measurementsGridOutput["Grid C Voltage"]
	gridC["Inv C Current"] = measurementsGridOutput["Grid C Current"]
	gridC["Inv C Power"] = measurementsGridOutput["Grid C Power"]

	root["Grid Freq"] = measurementsGridOutput["Grid Freq"]
	root["Inv 1 Temperature"] = measurementsGridOutput["Inv 1 Temperature"]
	root["Inv 2 Temperature"] = measurementsGridOutput["Inv 2 Temperature"]

	log.Println("measurementsGridOutput: ", measurementsGridOutput)

	failedConnections = 0

	if hasMQTT {
		err = mqtt.InsertGenericRecord("GridOutput", root)
		if err != nil {
			log.Printf("failed to insert record to MQTT: %s\n", err)
		} else {
			log.Println("measurementsGridOutput pushed to MQTT")
		}

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsGridOutput pushed to MQTT")
			}
		}("GridOutput/Grid A", gridA)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsGridOutput pushed to MQTT")
			}
		}("GridOutput/Grid B", gridB)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsGridOutput pushed to MQTT")
			}
		}("GridOutput/Grid C", gridC)
	}

	return nil

}

func loadInverterInfo() error {
	failedConnections := 0

	//
	// Inverter Info
	//
	measurementsInverterInfo, err := device.QueryInverterInfo()

	if err != nil {
		log.Printf("failed to perform measurementsInverterInfo: %s", err)
		failedConnections++

		if failedConnections > maximumFailedConnections {
			time.Sleep(time.Duration(config.Inverter.ReadInterval) * time.Second)
		}

		return err
	}

	invA := make(map[string]interface{})
	invB := make(map[string]interface{})
	invC := make(map[string]interface{})
	root := make(map[string]interface{})

	invA["Inv A Voltage"] = measurementsInverterInfo["Inv A Voltage"]
	invA["Inv A Current"] = measurementsInverterInfo["Inv A Current"]
	invA["Inv A Power"] = measurementsInverterInfo["Inv A Power"]
	invA["Inv A Freq"] = measurementsInverterInfo["Inv A Freq"]

	invB["Inv B Voltage"] = measurementsInverterInfo["Inv B Voltage"]
	invB["Inv B Current"] = measurementsInverterInfo["Inv B Current"]
	invB["Inv B Power"] = measurementsInverterInfo["Inv B Power"]
	invB["Inv B Freq"] = measurementsInverterInfo["Inv B Freq"]

	invC["Inv C Voltage"] = measurementsInverterInfo["Inv C Voltage"]
	invC["Inv C Current"] = measurementsInverterInfo["Inv C Current"]
	invC["Inv C Power"] = measurementsInverterInfo["Inv C Power"]
	invC["Inv C Freq"] = measurementsInverterInfo["Inv C Freq"]

	root["Leak Current"] = measurementsInverterInfo["Leak Current"]

	log.Println("measurementsInverterInfo: ", measurementsInverterInfo)

	failedConnections = 0

	if hasMQTT {
		err = mqtt.InsertGenericRecord("InverterInfo", root)
		if err != nil {
			log.Printf("failed to insert record to MQTT: %s\n", err)
		} else {
			log.Println("measurementsInverterInfo pushed to MQTT")
		}

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsInverterInfo pushed to MQTT")
			}
		}("InverterInfo/INV A", invA)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsInverterInfo pushed to MQTT")
			}
		}("InverterInfo/INV B", invB)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsInverterInfo pushed to MQTT")
			}
		}("InverterInfo/INV C", invC)

	}

	return nil

}

func loadLoadInfo() error {
	failedConnections := 0

	//
	// Load Info
	//
	measurementsLoadInfo, err := device.QueryLoadInfo()

	if err != nil {
		log.Printf("failed to perform measurementsLoadInfo: %s", err)
		failedConnections++

		if failedConnections > maximumFailedConnections {
			time.Sleep(time.Duration(config.Inverter.ReadInterval) * time.Second)
		}

		return err
	}

	loadA := make(map[string]interface{})
	loadB := make(map[string]interface{})
	loadC := make(map[string]interface{})
	root := make(map[string]interface{})

	loadA["Load A Voltage"] = measurementsLoadInfo["Load A Voltage"]
	loadA["Load A Current"] = measurementsLoadInfo["Load A Current"]
	loadA["Load A Power"] = measurementsLoadInfo["Load A Power"]
	loadA["Load A Rate"] = measurementsLoadInfo["Load A Rate"]

	loadB["Load B Voltage"] = measurementsLoadInfo["Load B Voltage"]
	loadB["Load B Current"] = measurementsLoadInfo["Load B Current"]
	loadB["Load B Power"] = measurementsLoadInfo["Load B Power"]
	loadB["Load B Rate"] = measurementsLoadInfo["Load B Rate"]

	loadC["Load C Voltage"] = measurementsLoadInfo["Load C Voltage"]
	loadC["Load C Current"] = measurementsLoadInfo["Load C Current"]
	loadC["Load C Power"] = measurementsLoadInfo["Load C Power"]
	loadC["Load C Rate"] = measurementsLoadInfo["Load C Rate"]

	root["Generator Port Voltage A"] = measurementsLoadInfo["Generator Port Voltage A"]
	root["Generator Port Voltage B"] = measurementsLoadInfo["Generator Port Voltage B"]
	root["Generator Port Voltage C"] = measurementsLoadInfo["Generator Port Voltage C"]

	log.Println("measurementsLoadInfo: ", measurementsLoadInfo)

	failedConnections = 0

	if hasMQTT {
		err = mqtt.InsertGenericRecord("LoadInfo", root)
		if err != nil {
			log.Printf("failed to insert record to MQTT: %s\n", err)
		} else {
			log.Println("measurementsLoadInfo pushed to MQTT")
		}

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsLoadInfo pushed to MQTT")
			}
		}("LoadInfo/Load A", loadA)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsLoadInfo pushed to MQTT")
			}
		}("LoadInfo/Load B", loadB)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsLoadInfo pushed to MQTT")
			}
		}("LoadInfo/Load C", loadC)
	}

	return nil

}

func loadBatteryOutput() error {
	failedConnections := 0

	//
	// Battery Output
	//
	measurementsBatteryOutput, err := device.QueryBatteryOutput()

	if err != nil {
		log.Printf("failed to perform measurementsBatteryOutput: %s", err)
		failedConnections++

		if failedConnections > maximumFailedConnections {
			time.Sleep(time.Duration(config.Inverter.ReadInterval) * time.Second)
		}

		return err
	}

	bat := make(map[string]interface{})
	bmsBAT := make(map[string]interface{})

	bat["BAT Voltage"] = measurementsBatteryOutput["BAT Voltage"]
	bat["BAT Current"] = measurementsBatteryOutput["BAT Current"]
	bat["BAT 1 Current"] = measurementsBatteryOutput["BAT 1 Current"]
	bat["BAT 2 Current"] = measurementsBatteryOutput["BAT 2 Current"]
	bat["BAT 3 Current"] = measurementsBatteryOutput["BAT 3 Current"]
	bat["BAT SOC"] = measurementsBatteryOutput["BAT SOC"]
	bat["BAT Temperature"] = measurementsBatteryOutput["BAT Temperature"]
	bat["BAT Charge Voltage"] = measurementsBatteryOutput["BAT Charge Voltage"]
	bat["BAT Charge Current Limit"] = measurementsBatteryOutput["BAT Charge Current Limit"]
	bat["BAT Discharge Current Limit"] = measurementsBatteryOutput["BAT Discharge Current Limit"]
	bat["BAT Power"] = measurementsBatteryOutput["BAT Power"]

	bmsBAT["BMS BAT Voltage"] = measurementsBatteryOutput["BMS BAT Voltage"]
	bmsBAT["BMS BAT Current"] = measurementsBatteryOutput["BMS BAT Current"]
	bmsBAT["BMS BAT Cell Max Voltage"] = measurementsBatteryOutput["BMS BAT Cell Max Voltage"]
	bmsBAT["BMS BAT Cell Min Voltage"] = measurementsBatteryOutput["BMS BAT Cell Min Voltage"]
	bmsBAT["BMS BAT Cell Max Temperature"] = measurementsBatteryOutput["BMS BAT Cell Max Temperature"]
	bmsBAT["BMS BAT Cell Min Temperature"] = measurementsBatteryOutput["BMS BAT Cell Min Temperature"]

	log.Println("measurementsBatteryOutput: ", measurementsBatteryOutput)

	failedConnections = 0

	if hasMQTT {
		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsBatteryOutput pushed to MQTT")
			}
		}("BatteryOutput/BAT", bat)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsBatteryOutput pushed to MQTT")
			}
		}("BatteryOutput/BMS BAT", bmsBAT)
	}

	return nil

}

func loadPVOutput() error {
	failedConnections := 0

	//
	// PV Output
	//
	measurementsPVOutput, err := device.QueryPVOutput()

	if err != nil {
		log.Printf("failed to perform measurementsPVOutput: %s", err)
		failedConnections++

		if failedConnections > maximumFailedConnections {
			time.Sleep(time.Duration(config.Inverter.ReadInterval) * time.Second)
		}

		return err
	}

	PV1 := make(map[string]interface{})
	PV2 := make(map[string]interface{})

	PV1["Voltage PV 1"] = measurementsPVOutput["Voltage PV 1"]
	PV1["Current PV 1"] = measurementsPVOutput["Current PV 1"]
	PV1["Power PV 1"] = measurementsPVOutput["Power PV 1"]

	PV2["Voltage PV 2"] = measurementsPVOutput["Voltage PV 2"]
	PV2["Current PV 2"] = measurementsPVOutput["Current PV 2"]
	PV2["Power PV 2"] = measurementsPVOutput["Power PV 2"]

	log.Println("measurementsPVOutput: ", measurementsPVOutput)

	failedConnections = 0

	if hasMQTT {
		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsPVOutput pushed to MQTT")
			}
		}("PVOutput/PV1", PV1)

		go func(topic string, data map[string]interface{}) {
			err = mqtt.InsertGenericRecord(topic, data)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsPVOutput pushed to MQTT")
			}
		}("PVOutput/PV2", PV2)
	}

	return nil

}
