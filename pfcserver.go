package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

var settings = make(map[string]string)
var worker PFC_settings
var tickerCam, tickerSensors *time.Ticker
var hum float64
var temp float64

func editHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("edit.html"))
	t.Execute(w, settings)
}

func doHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL
	// add cutting params and extract cmd
	fmt.Println(params.Path)
	fmt.Fprint(w, "60")

	if params.Path == "/do/setFanOn" {
		worker.SetFanStateManual(true)
		fmt.Println("Fan on")
		//wifiController.writeSettings("PFC MVP v2.0", "12345670", "bbc", "12345678")
	} else if params.Path == "/do/setFanOff" {
		worker.SetFanStateManual(false)
		fmt.Println("Fan off")
	} else if params.Path == "/do/setLightOn" {
		worker.SetLightStateManual(true)
		fmt.Println("Light on")
	} else if params.Path == "/do/setLightOff" {
		worker.SetLightStateManual(false)
		fmt.Println("Light off")
	} else if params.Path == "/do/setPumpOn" {
		worker.SetPumpStateManual(true)
		fmt.Println("Pump on")
	} else if params.Path == "/do/setPumpOff" {
		worker.SetPumpStateManual(false)
		fmt.Println("Pump off")
	} else if params.Path == "/do/takePhoto" {
		fmt.Println("Photo")
		exec.Command("/bin/sh", "-c", "sudo rm /home/pi/go_exp/img.jpg").CombinedOutput()

		if tryTakePhoto(true) == false {
			if tryTakePhoto(true) == false {
				if tryTakePhoto(true) == false {
					tryTakePhoto(true)
				}
			}
		}
	}

}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	parameters := make(map[string]string)
	parameters["light_on_time"] = r.FormValue("light_on_time")
	parameters["light_off_time"] = r.FormValue("light_off_time")
	parameters["pump_on_time"] = r.FormValue("pump_on_time")
	parameters["pump_pause_time"] = r.FormValue("pump_pause_time")
	parameters["temperature_threshold"] = r.FormValue("temperature_threshold")
	parameters["ssid"] = r.FormValue("ssid")
	parameters["ap_ssid"] = r.FormValue("ap_ssid")
	parameters["pass"] = r.FormValue("pass")
	parameters["ap_pass"] = r.FormValue("ap_pass")

	jsonString, _ := json.Marshal(parameters)
	saveSettings(&jsonString, "settings.txt")
	settings = parameters

	worker.makeFromMap(settings)

	text := `<html><head><title>ok</title></head><body>
	<script type="text/javascript">   
	function Redirect() { window.location="/edit/"; } 
	setTimeout('Redirect()', 1500);   
	</script>
	Successfully saved  
	</body>
	</html>`
	fmt.Fprint(w, text)
}

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
			exec.Command("/bin/sh", "-c", "sudo fswebcam -r 960x720 --jpeg 90 -D 1 /home/pi/go_exp/img.jpg").CombinedOutput()
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
		AppendLineToLog(dir+"sensors.log", GenerateTimestamp()+" t "+strconv.FormatFloat(temp, 'f', 2, 64)+" h "+strconv.FormatFloat(hum, 'f', 2, 64)+"\r\n")
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

func main() {
	var tmp []byte

	//H, T := <-h, t
	//fmt.Println(h, t)

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

	fmt.Println(settings)
	http.HandleFunc("/do/", doHandler)

	worker.makeFromMap(settings)

	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)

	fmt.Println("Starting server...")
	http.ListenAndServe(":80", nil)
}
