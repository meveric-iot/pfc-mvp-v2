package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

var settings = make(map[string]string)
var worker PFC_settings

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

func handlePeriphStates() {
	var h, t float64

	fmt.Println("handlePeriphStates started")

	for true {
		h, t = periph.ReadHumTempSensor()
		worker.UpdateFlags(h, t)
		// Add Lock
		hum = h
		temp = t
		periph.SetLightState(worker.GetTargetLightState())
		periph.SetPumpState(worker.GetTargetPumpState())
		periph.SetFanState(worker.GetTargetFanState())

		time.Sleep(time.Millisecond * 1000)
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

	go handlePeriphStates()

	fmt.Println(settings)
	http.HandleFunc("/do/", doHandler)

	worker.makeFromMap(settings)

	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)

	fmt.Println("Starting server...")
	http.ListenAndServe(":80", nil)
}
