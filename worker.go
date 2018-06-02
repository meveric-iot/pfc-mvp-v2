package main

import (
	"fmt"
	"strconv"
	"time"
)

var timerMM *time.Timer
var fMM = false

// time in seconds
type PFC_settings struct {
	light_on_time         time.Time
	light_off_time        time.Time
	temperature_threshold float64
	pump_on_time          uint32
	pump_pause_time       uint32
	ssid                  string
	pass                  string
	ap_ssid               string
	ap_pass               string
	fLightOn,
	fPumpOn,
	fPumpEnabled,
	fFanOn bool
	pumpNextTime int64
}

func strToHHMMss(str string) time.Time {
	res, e := time.Parse("15:04:05", str)
	if e != nil {
		res, _ = time.Parse("15:04:05", "00:00:05")
	}
	return res
}

func (s *PFC_settings) makeFromMap(cfg map[string]string) {
	s.ap_pass = cfg["ap_pass"]
	s.ap_ssid = cfg["ap_ssid"]
	s.ssid = cfg["ssid"]
	s.pass = cfg["pass"]
	s.temperature_threshold, _ = strconv.ParseFloat(cfg["temperature_threshold"], 64)

	t1 := strToHHMMss(cfg["pump_on_time"])
	t2 := strToHHMMss(cfg["pump_pause_time"])

	s.pump_on_time = uint32(t1.Hour()*3600 + t1.Minute()*60 + t1.Second())
	s.pump_pause_time = uint32(t2.Hour()*3600 + t2.Minute()*60 + t2.Second())

	s.light_on_time = strToHHMMss(cfg["light_on_time"])
	s.light_off_time = strToHHMMss(cfg["light_off_time"])
}

func (s *PFC_settings) GetTargetLightState() bool {
	return s.fLightOn
}

func (s *PFC_settings) GetTargetPumpState() bool {
	return s.fPumpOn
}

func (s *PFC_settings) GetTargetFanState() bool {
	return s.fFanOn
}

func (s *PFC_settings) SetLightStateManual(state bool) {
	s.restartMModeTimer()
	s.fLightOn = state
}

func (s *PFC_settings) SetPumpStateManual(state bool) {
	s.restartMModeTimer()
	s.fPumpOn = state
}

func (s *PFC_settings) SetFanStateManual(state bool) {
	s.restartMModeTimer()
	s.fFanOn = state
}

func (s *PFC_settings) restartMModeTimer() {
	fmt.Println("Manual mode...")
	if timerMM != nil {
		timerMM.Stop()
	}
	fMM = true
	timerMM = time.NewTimer(time.Second * 60)
	go func() {
		<-timerMM.C
		timerMM = nil
		fMM = false
		fmt.Println("Restore automatic mode")
	}()
}

// UpdateFlags recalculate flags from settings, current time and humidity and tamperature
func (s *PFC_settings) UpdateFlags(h, t float64) {
	//
	time_now := time.Now()

	light_from_sec := s.light_on_time.Hour()*3600 + s.light_on_time.Minute()*60
	light_to_sec := s.light_off_time.Hour()*3600 + s.light_off_time.Minute()*60
	sec_from_midnight := time_now.Hour()*3600 + time_now.Minute()*60 + time_now.Second()

	if fMM == false {
		invert := light_from_sec < light_to_sec                                     // at example, (07:00 < 22:00) => invert = true
		if sec_from_midnight > light_from_sec && sec_from_midnight < light_to_sec { // sec_from_midnight between _from and _to
			s.fLightOn = invert
		} else {
			s.fLightOn = !invert
		}

		if s.pumpNextTime == 0 {
			s.fPumpOn = true
			s.fPumpEnabled = true
			s.pumpNextTime = time_now.Unix() + int64(s.pump_on_time)
		} else {
			if time_now.Unix() >= s.pumpNextTime {
				if s.fPumpEnabled {
					s.fPumpEnabled = false
					s.pumpNextTime = time_now.Unix() + int64(s.pump_pause_time)
					s.fPumpOn = false
				} else {
					s.fPumpEnabled = true
					s.pumpNextTime = time_now.Unix() + int64(s.pump_on_time)
					s.fPumpOn = true
				}
			}
		}

		if t > s.temperature_threshold {
			s.fFanOn = true
		} else {
			s.fFanOn = false
		}
	}

}
