package main

import (
	"github.com/misterdelle/invt_logger_reader/adapters/export/mosquitto"
)

type Config struct {
	Inverter struct {
		Port         string
		LoggerSerial uint
		ReadInterval int
	}
	Mqtt mosquitto.MqttConfig
}

func NewConfig(app Application) (*Config, error) {
	config := &Config{}

	config.Inverter.Port = app.InverterPort
	config.Inverter.LoggerSerial = app.InverterLoggerSerial
	config.Inverter.ReadInterval = app.InverterReadInterval

	config.Mqtt.Url = app.MQTTURL
	config.Mqtt.User = app.MQTTUser
	config.Mqtt.Password = app.MQTTPassword
	config.Mqtt.Prefix = app.MQTTTopicName

	return config, nil
}
