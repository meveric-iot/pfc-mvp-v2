﻿package main

import (
	"io/ioutil"
)

func saveSettings(cfg *[]byte, filename string) error {
	err := ioutil.WriteFile(filename, *cfg, 0666)
	return err
}

func readSettings(cfg *[]byte, filename string) error {
	dat, err := ioutil.ReadFile(filename)
	*cfg = dat
	return err
}

func fillSettingsDefault(cfg *map[string]string) {
	settings = make(map[string]string)
	settings["light_on_time"] = "08:00:00"
	settings["light_off_time"] = "22:00:00"
	settings["pump_on_time"] = "00:00:05"
	settings["pump_pause_time"] = "06:00:00"
	settings["fan_on_time"] = "00:00:30"
	settings["fan_pause_time"] = "00:30:00"
	settings["temperature_threshold"] = "23.0"
	settings["ssid"] = "sopl"
	settings["pass"] = ""
	settings["ap_ssid"] = "PFC V1.2 Prototype"
	settings["ap_pass"] = "12345678"
	*cfg = settings
}
