package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var settings = make(map[string]string)
var worker PFC_settings
var tickerCam, tickerSensors *time.Ticker
var hum float64
var temp float64

func tryTakePhoto(prv bool) bool {
	var timer *time.Timer
	fmt.Println("Cam try capture")
	dir, err := GetTargetPathByDate()
	if err == nil {
		flagDelay := false
		go func() {
			<-timer.C
			flagDelay = true
		}()
		timer = time.NewTimer(time.Millisecond * 990)

		if prv == true {
			exec.Command("/bin/sh", "-c", "sudo fswebcam -r 960x720 --jpeg 90 -D 1 /home/pi/go_exp/static/img.jpg").CombinedOutput()
		} else {
			exec.Command("/bin/sh", "-c", "sudo fswebcam -r 960x720 --jpeg 90 -D 1 "+dir+GenerateStringByCurrentTime()+".jpg").CombinedOutput()
		}
		timer.Stop()
		time.Sleep(time.Millisecond * 1500)

		return flagDelay
	}
	return false
}

func camCapture() {
	for range tickerCam.C {
		dir, _ := GetTargetPathByDate()
		loadTempHumPointsFromLogFile(dir + "sensors.log")
		if tryTakePhoto(false) == false {
			if tryTakePhoto(false) == false {
				if tryTakePhoto(false) == false {
					tryTakePhoto(false)
				}
			}
		}
	}
}

func writeSensorsToLog() {
	for range tickerSensors.C {
		dir, _ := GetTargetPathByDate()
		AppendLineToLog(dir+"sensors.log", GenerateTimestamp()+" t "+strconv.FormatFloat(temp, 'f', 1, 64)+" h "+strconv.FormatFloat(hum, 'f', 1, 64)+"\r\n")
	}
}

func handlePeriphStates() {
	var h, t float64
	var dir string

	fmt.Println("handlePeriphStates started")

	for true {
		dir, _ = GetTargetPathByDate()
		h, t = periph.ReadHumTempSensor()
		worker.UpdateFlags(h, t)
		// Add Lock
		hum = h
		temp = t
		periph.SetLightState(worker.GetTargetLightState())
		periph.SetPumpState(worker.GetTargetPumpState())
		periph.SetFanState(worker.GetTargetFanState())
		if worker.ReadLightSwitchedFlag() {
			AppendLineToLog(dir+"switchings.log", GenerateTimestamp()+" LightEnabled -> "+BoolToString(worker.GetTargetLightState())+"\r\n")
		}
		if worker.ReadPumpSwitchedFlag() {
			AppendLineToLog(dir+"switchings.log", GenerateTimestamp()+" PumpEnabled  -> "+BoolToString(worker.GetTargetPumpState())+"\r\n")
		}
		if worker.ReadFanSwitchedFlag() {
			AppendLineToLog(dir+"switchings.log", GenerateTimestamp()+" FanEnabled   -> "+BoolToString(worker.GetTargetFanState())+"\r\n")
		}

		time.Sleep(time.Millisecond * 500)
	}
}

func mainDataHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL
	state := false
	//fmt.Println(params.Path)

	if params.Path == "/mainData" { // out t, h, states
		parameters := make(map[string]string)
		parameters["temp_val"] = strconv.FormatFloat(temp, 'f', 1, 64)
		parameters["hum_val"] = strconv.FormatFloat(hum, 'f', 1, 64)
		parameters["light_state"] = BoolToString(worker.GetTargetLightState())
		parameters["pump_state"] = BoolToString(worker.GetTargetPumpState())
		parameters["fan_state"] = BoolToString(worker.GetTargetFanState())
		current := time.Now()
		parameters["date_time"] = fmt.Sprintf("%02d:%02d:%02d %02d.%02d.%04d", current.Hour(), current.Minute(), current.Second(), current.Day(), current.Month(), current.Year())
		jsonString, _ := json.Marshal(parameters)
		fmt.Fprint(w, string(jsonString))
		return
	}

	if params.Path == "/mainData/toggleLight" {
		state = worker.GetTargetLightState()
		worker.SetLightStateManual(!state)
		fmt.Fprint(w, "60")
	} else if params.Path == "/mainData/togglePump" {
		state = worker.GetTargetPumpState()
		worker.SetPumpStateManual(!state)
		fmt.Fprint(w, "60")
	} else if params.Path == "/mainData/toggleFan" {
		state = worker.GetTargetFanState()
		worker.SetFanStateManual(!state)
		fmt.Fprint(w, "60")
	} else if params.Path == "/mainData/getGraphHumData" {
		str := getCharHumDataJSONStr()
		fmt.Fprint(w, str)
	} else if params.Path == "/mainData/getGraphTempData" {
		str := getCharTempDataJSONStr()
		fmt.Fprint(w, str)
	} else if params.Path == "/mainData/shutdown" {
		exec.Command("/bin/sh", "-c", "sudo shutdown 0").CombinedOutput()
	} else if params.Path == "/mainData/updatePhoto" {
		//exec.Command("/bin/sh", "-c", "sudo rm /home/pi/go_exp/img.jpg").CombinedOutput()
		if tryTakePhoto(true) == false {
			if tryTakePhoto(true) == false {
				if tryTakePhoto(true) == false {
					tryTakePhoto(true)
				}
			}
		}
		fmt.Fprint(w, "")
	}
}

func growingSettingsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	if len(body) == 0 { // empty request, return settings in JSON
		parameters := make(map[string]string)
		parameters["valLightOnTime"] = settings["light_on_time"]
		parameters["valLightOffTime"] = settings["light_off_time"]
		parameters["valPumpPauseTime"] = settings["pump_pause_time"]
		parameters["valPumpOnTime"] = settings["pump_on_time"]
		parameters["valFanOnThreshold"] = settings["temperature_threshold"]
		jsonString, _ := json.Marshal(parameters)
		fmt.Fprint(w, string(jsonString))
		return
	}
	fmt.Println(body)
	var t map[string]string
	err = json.Unmarshal(body, &t)
	if err != nil {
		return
	}
	settings["light_on_time"] = t["valLightOnTime"]
	settings["light_off_time"] = t["valLightOffTime"]
	settings["pump_on_time"] = t["valPumpOnTime"]
	settings["pump_pause_time"] = t["valPumpPauseTime"]
	settings["temperature_threshold"] = t["valFanOnThreshold"]

	jsonString, _ := json.Marshal(settings)
	worker.makeFromMap(settings)
	saveSettings(&jsonString, "settings.txt")

}

func systemSettingsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	if len(body) == 0 { // empty request, return settings in JSON
		parameters := make(map[string]string)
		parameters["valSsid"] = settings["ssid"]
		parameters["valAPName"] = settings["ap_ssid"]
		parameters["valPass"] = settings["pass"]
		jsonString, _ := json.Marshal(parameters)
		fmt.Fprint(w, string(jsonString))
		return
	}
	var t map[string]string
	err = json.Unmarshal(body, &t)
	if err != nil {
		return
	}
	settings["ssid"] = t["valSsid"]
	settings["ap_ssid"] = t["valAPName"]
	settings["pass"] = t["valPass"]

	jsonString, _ := json.Marshal(settings)
	worker.makeFromMap(settings)
	saveSettings(&jsonString, "settings.txt")
}

func main() {
	var tmp []byte

	if readSettings(&tmp, "settings.txt") != nil {
		fillSettingsDefault(&settings)
		jsonString, _ := json.Marshal(&settings)
		saveSettings(&jsonString, "settings.txt")
		wifiController.initApPlusStationMode()
	} else {
		json.Unmarshal(tmp, &settings)
	}

	periph.InitGpio()

	go handlePeriphStates()

	tickerCam = time.NewTicker(30 * time.Minute)
	go camCapture()

	tickerSensors = time.NewTicker(1 * time.Minute)
	go writeSensorsToLog()

	dir, _ := GetTargetPathByDate()
	loadTempHumPointsFromLogFile(dir + "sensors.log")

	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("./out"))))
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/mainData", mainDataHandler)
	http.HandleFunc("/mainData/", mainDataHandler)
	http.HandleFunc("/growingSettings", growingSettingsHandler)
	http.HandleFunc("/systemSettings", systemSettingsHandler)

	worker.makeFromMap(settings)

	fmt.Println("Starting server...")
	http.ListenAndServe(":80", nil)
}

func noDirListing(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || r.URL.Path == "" {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}
