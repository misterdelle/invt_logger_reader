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
			err = mqtt.InsertGenericRecord("Station", measurementsStation)
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

	log.Println("measurementsEnergyTodayTotals: ", measurementsEnergyTodayTotals)

	failedConnections = 0

	if hasMQTT {
		go func() {
			err = mqtt.InsertGenericRecord("EnergyTodayTotals", measurementsEnergyTodayTotals)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsEnergyTodayTotals pushed to MQTT")
			}
		}()
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

	log.Println("measurementsGridOutput: ", measurementsGridOutput)

	failedConnections = 0

	if hasMQTT {
		go func() {
			err = mqtt.InsertGenericRecord("GridOutput", measurementsGridOutput)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsGridOutput pushed to MQTT")
			}
		}()
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

	log.Println("measurementsInverterInfo: ", measurementsInverterInfo)

	failedConnections = 0

	if hasMQTT {
		go func() {
			err = mqtt.InsertGenericRecord("InverterInfo", measurementsInverterInfo)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsInverterInfo pushed to MQTT")
			}
		}()
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

	log.Println("measurementsLoadInfo: ", measurementsLoadInfo)

	failedConnections = 0

	if hasMQTT {
		go func() {
			err = mqtt.InsertGenericRecord("LoadInfo", measurementsLoadInfo)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsLoadInfo pushed to MQTT")
			}
		}()
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

	log.Println("measurementsBatteryOutput: ", measurementsBatteryOutput)

	failedConnections = 0

	if hasMQTT {
		go func() {
			err = mqtt.InsertGenericRecord("BatteryOutput", measurementsBatteryOutput)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsBatteryOutput pushed to MQTT")
			}
		}()
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

	log.Println("measurementsPVOutput: ", measurementsPVOutput)

	failedConnections = 0

	if hasMQTT {
		go func() {
			err = mqtt.InsertGenericRecord("PVOutput", measurementsPVOutput)
			if err != nil {
				log.Printf("failed to insert record to MQTT: %s\n", err)
			} else {
				log.Println("measurementsPVOutput pushed to MQTT")
			}
		}()
	}

	return nil

}
