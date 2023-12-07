package main

import (
	"errors"

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

func (c *Config) validate() error {
	if c.Inverter.Port == "" {
		return errors.New("missing required inverter.port config")
	}

	if c.Inverter.LoggerSerial == 0 {
		return errors.New("missing required inverter.loggerSerial config")
	}

	return nil
}

func NewConfig(app Application) (*Config, error) {
	config := &Config{}

	config.Inverter.Port = app.InverterPort
	config.Inverter.LoggerSerial = app.InverterLoggerSerial
	config.Inverter.ReadInterval = app.InverterReadInterval

	config.Mqtt.Url = app.MQTTURL
	config.Mqtt.User = app.MQTTUser
	config.Mqtt.Password = app.MQTTPassword

	return config, nil
}
