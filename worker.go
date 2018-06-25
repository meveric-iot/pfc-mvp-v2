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
	fan_on_time           uint32
	fan_pause_time        uint32
	ssid                  string
	pass                  string
	ap_ssid               string
	ap_pass               string
	fLightOn,
	fPumpOn,
	fPumpEnabled,
	fChillerOn,
	fFanOn,
	fFanEnabled bool
	pumpNextTime, fanNextTime int64
	fLightSwitched,
	fPumpSwitched,
	fFanSwitched,
	fChillerSwitched bool
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

	t3 := strToHHMMss(cfg["fan_on_time"])
	t4 := strToHHMMss(cfg["fan_pause_time"])

	s.pump_on_time = uint32(t1.Hour()*3600 + t1.Minute()*60 + t1.Second())
	s.pump_pause_time = uint32(t2.Hour()*3600 + t2.Minute()*60 + t2.Second())

	s.fan_on_time = uint32(t3.Hour()*3600 + t3.Minute()*60 + t3.Second())
	s.fan_pause_time = uint32(t4.Hour()*3600 + t4.Minute()*60 + t4.Second())

	s.pumpNextTime = 0
	s.fanNextTime = 0
	s.fPumpEnabled = false
	s.fFanEnabled = false

	s.light_on_time = strToHHMMss(cfg["light_on_time"])
	s.light_off_time = strToHHMMss(cfg["light_off_time"])

	s.fLightSwitched = false
	s.fPumpSwitched = false
	s.fFanSwitched = false
	s.fChillerSwitched = false
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

func (s *PFC_settings) GetTargetChillerState() bool {
	return s.fChillerOn
}

func (s *PFC_settings) SetLightStateManual(state bool) {
	s.restartMModeTimer()
	if s.fLightOn != state {
		s.fLightSwitched = true
	}
	s.fLightOn = state
}

func (s *PFC_settings) SetPumpStateManual(state bool) {
	s.restartMModeTimer()
	if s.fPumpOn != state {
		s.fPumpSwitched = true
	}
	s.fPumpOn = state
}

func (s *PFC_settings) SetFanStateManual(state bool) {
	s.restartMModeTimer()
	if s.fFanOn != state {
		s.fFanSwitched = true
	}
	s.fFanOn = state
}

func (s *PFC_settings) SetChillerStateManual(state bool) {
	s.restartMModeTimer()
	if s.fChillerOn != state {
		s.fChillerSwitched = true
	}
	s.fChillerOn = state
}

func (s *PFC_settings) ReadLightSwitchedFlag() bool {
	if s.fLightSwitched == true {
		s.fLightSwitched = false
		return true
	}
	return false
}

func (s *PFC_settings) ReadPumpSwitchedFlag() bool {
	if s.fPumpSwitched == true {
		s.fPumpSwitched = false
		return true
	}
	return false
}

func (s *PFC_settings) ReadFanSwitchedFlag() bool {
	if s.fFanSwitched == true {
		s.fFanSwitched = false
		return true
	}
	return false
}

func (s *PFC_settings) ReadChillerSwitchedFlag() bool {
	if s.fChillerSwitched == true {
		s.fChillerSwitched = false
		return true
	}
	return false
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
			if s.fLightOn != invert {
				s.fLightSwitched = true
			}
			s.fLightOn = invert
		} else {
			if s.fLightOn != !invert {
				s.fLightSwitched = true
			}
			s.fLightOn = !invert
		}

		if s.pumpNextTime == 0 {
			s.fPumpOn = true
			s.fPumpEnabled = true
			s.pumpNextTime = time_now.Unix() + int64(s.pump_on_time)
		} else {
			if time_now.Unix() >= s.pumpNextTime {
				if s.fPumpEnabled {
					s.fPumpSwitched = true
					s.fPumpEnabled = false
					s.pumpNextTime = time_now.Unix() + int64(s.pump_pause_time)
					s.fPumpOn = false
				} else {
					s.fPumpSwitched = true
					s.fPumpEnabled = true
					s.pumpNextTime = time_now.Unix() + int64(s.pump_on_time)
					s.fPumpOn = true
				}
			}
		}

		if s.fanNextTime == 0 {
			s.fFanOn = true
			s.fFanEnabled = true
			s.fanNextTime = time_now.Unix() + int64(s.fan_on_time)
		} else {
			if time_now.Unix() >= s.fanNextTime {
				if s.fFanEnabled {
					s.fFanSwitched = true
					s.fFanEnabled = false
					s.fanNextTime = time_now.Unix() + int64(s.fan_pause_time)
					s.fFanOn = false
				} else {
					s.fFanSwitched = true
					s.fFanEnabled = true
					s.fanNextTime = time_now.Unix() + int64(s.fan_on_time)
					s.fFanOn = true
				}
			}
		}

		if t > s.temperature_threshold {
			if s.fChillerOn != true {
				s.fChillerSwitched = true
			}
			s.fChillerOn = true
		} else {
			if s.fChillerOn != false {
				s.fChillerSwitched = true
			}
			s.fChillerOn = false
		}
	}

}
