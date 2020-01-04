package main

import (
	"encoding/json"
	"github.com/labstack/gommon/log"
	"net/http"
	"os"
)

const ApiUrl   = "https://api.nature.global/1/devices"

type Device struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	TemperatureOffset int32        `json:"temperature_offset"`
	HumidityOffset    int32        `json:"humidity_offset"`
	CreatedAt         string       `json:"created_at"`
	UpdatedAt         string       `json:"updated_at"`
	FirmwareVersion   string       `json:"firmware_version"`
	NewestEvents      NewestEvents `json:"newest_events"`
}

type NewestEvents struct {
	Temperature SensorValue `json:"te"`
	Humidity    SensorValue `json:"hu"`
	Illuminance SensorValue `json:"il"`
}

type SensorValue struct {
	Value     float64 `json:"val"`
	CreatedAt string  `json:"created_at"`
}

func GetTemperature() float64 {
	d := &Device{}
	temperature := d.FetchValuesFromNatureRemo()
	return temperature
}

func (d *Device) FetchValuesFromNatureRemo() float64 {
	var devices []*Device

	req, err := http.NewRequest("GET", ApiUrl, nil)
	if err != nil {
		log.Error(err)
	}
	remoToken := os.Getenv("REMO_TOKEN")
	req.Header.Set("Authorization", "Bearer " + remoToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}

	err = json.NewDecoder(resp.Body).Decode(&devices)
	if err != nil {
		log.Error(err)
	}

	return devices[0].NewestEvents.Temperature.Value
}