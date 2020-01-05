package main

import (
	"github.com/labstack/gommon/log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const ApiBaseUrl   = "https://api.nature.global/1/appliances/"

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

func PutAcSettings(mode string) {
	d := &Device{}
	d.PostAcSettingsToNatureRemo(mode)
}

func (d *Device) PostAcSettingsToNatureRemo(mode string) {
	var ApiUrl   = ApiBaseUrl + os.Getenv("APPLIANCE_ID") + "/aircon_settings"
	values := url.Values{}

	if mode == "power-off" {
		values.Add("button", mode)
	} else {
		values.Add("operation_mode", mode)
	}

	req, err := http.NewRequest("POST", ApiUrl, strings.NewReader(values.Encode()))
	if err != nil {
		log.Error(err)
	}
	remoToken := os.Getenv("REMO_TOKEN")
	req.Header.Set("Authorization", "Bearer " + remoToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}
	defer resp.Body.Close()
}