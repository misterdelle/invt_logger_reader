package mosquitto

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttConfig struct {
	Url      string `yaml:"url"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Prefix   string `yaml:"prefix"`
}

type Connection struct {
	client mqtt.Client
	prefix string
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Printf("MQTT Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

func New(config *MqttConfig) (*Connection, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Url)
	opts.SetClientID("invt")
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	if config.User != "" {
		opts.SetUsername(config.User)
	}

	if config.Password != "" {
		opts.SetPassword(config.Password)
	}

	conn := &Connection{}
	conn.client = mqtt.NewClient(opts)
	conn.prefix = config.Prefix
	if token := conn.client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return conn, nil

}

// func (conn *Connection) InsertRecord(measurement map[string]interface{}) error {
// 	measurementCopy := make(map[string]interface{}, len(measurement))
// 	for k, v := range measurement {
// 		measurementCopy[k] = v
// 	}

// 	go func(measurement map[string]interface{}) {
// 		// timestamp it
// 		measurement["LastTimestamp"] = time.Now().UnixNano() / int64(time.Millisecond)
// 		m, _ := json.Marshal(measurement)
// 		measurement["All"] = string(m)

// 		for k, v := range measurement {
// 			token := conn.client.Publish(fmt.Sprintf("%s/%s", conn.prefix, k), 0, true, fmt.Sprintf("%v", v))
// 			res := token.WaitTimeout(1 * time.Second)
// 			if !res || token.Error() != nil {
// 				log.Printf("error inserting to MQTT: %s", token.Error())
// 			}
// 		}

// 	}(measurementCopy)

// 	return nil
// }

func (conn *Connection) InsertRecord(measurement map[string]interface{}) error {
	measurementCopy := make(map[string]interface{}, len(measurement))
	for k, v := range measurement {
		measurementCopy[k] = v
	}

	measurementCopy["LastTimestamp"] = time.Now().UnixNano() / int64(time.Millisecond)

	measurementCopy2 := make(map[string]interface{}, 1)
	m, _ := json.Marshal(measurementCopy)
	measurementCopy2["inverter"] = string(m)

	go func(measurement map[string]interface{}) {
		// timestamp it
		// measurement["LastTimestamp"] = time.Now().UnixNano() / int64(time.Millisecond)
		// m, _ := json.Marshal(measurement)
		// measurement["All"] = string(m)

		for k, v := range measurement {
			token := conn.client.Publish(fmt.Sprintf("%s/%s", conn.prefix, k), 0, true, fmt.Sprintf("%v", v))
			res := token.WaitTimeout(1 * time.Second)
			if !res || token.Error() != nil {
				log.Printf("error inserting to MQTT: %s", token.Error())
			}
		}

	}(measurementCopy2)

	return nil
}

func (conn *Connection) Subscribe(topic string, callback mqtt.MessageHandler) {
	conn.client.Subscribe(topic, 0, callback)
}

func (conn *Connection) InsertGenericRecord(topicName string, measurement map[string]interface{}) error {
	go func(allData map[string]interface{}) {
		for k, v := range allData {
			token := conn.client.Publish(fmt.Sprintf("%s/%s/%s", conn.prefix, topicName, k), 0, true, fmt.Sprintf("%v", v))
			res := token.WaitTimeout(1 * time.Second)
			if !res || token.Error() != nil {
				log.Printf("error inserting to MQTT: %s", token.Error())
			}
		}

		clear(measurement)
	}(measurement)

	return nil
}
