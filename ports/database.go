package ports

import mqtt "github.com/eclipse/paho.mqtt.golang"

type Database interface {
	InsertRecord(measurement map[string]interface{}) error
	InsertGenericRecord(topicName string, measurement map[string]interface{}) error
}

type DatabaseWithListener interface {
	Database
	Subscribe(topic string, callback mqtt.MessageHandler)
}
