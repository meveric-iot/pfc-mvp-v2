package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type PFCperiph struct {
	fanState,
	chillerState,
	lightState,
	pumpState bool
}

var periph PFCperiph

func (p *PFCperiph) InitGpio() {
	exec.Command("/bin/sh", "-c", "gpio mode 0 out").CombinedOutput()
	time.Sleep(time.Millisecond * 200)
	exec.Command("/bin/sh", "-c", "gpio write 0 1").CombinedOutput()
	time.Sleep(time.Millisecond * 800)
	exec.Command("/bin/sh", "-c", "gpio mode 2 out").CombinedOutput()
	time.Sleep(time.Millisecond * 200)
	exec.Command("/bin/sh", "-c", "gpio write 2 1").CombinedOutput()
	time.Sleep(time.Millisecond * 800)
	exec.Command("/bin/sh", "-c", "gpio mode 3 out").CombinedOutput()
	time.Sleep(time.Millisecond * 200)
	exec.Command("/bin/sh", "-c", "gpio write 3 1").CombinedOutput()
	time.Sleep(time.Millisecond * 800)
	exec.Command("/bin/sh", "-c", "gpio mode 12 out").CombinedOutput()
	time.Sleep(time.Millisecond * 200)
	exec.Command("/bin/sh", "-c", "gpio write 12 1").CombinedOutput()
	time.Sleep(time.Millisecond * 800)
}

func (p *PFCperiph) SetLightState(state bool) {
	var s string
	if state == true {
		s = "0"
	} else {
		s = "1"
	}
	exec.Command("/bin/sh", "-c", "gpio write 0 "+s).CombinedOutput()
}

func (p *PFCperiph) SetFanState(state bool) {
	var s string
	if state == true {
		s = "0"
	} else {
		s = "1"
	}
	exec.Command("/bin/sh", "-c", "gpio write 2 "+s).CombinedOutput()
}

func (p *PFCperiph) SetPumpState(state bool) {
	var s string
	if state == true {
		s = "0"
	} else {
		s = "1"
	}
	exec.Command("/bin/sh", "-c", "gpio write 3 "+s).CombinedOutput()
}

func (p *PFCperiph) SetChillerState(state bool) {
	var s string
	if state == true {
		s = "0"
	} else {
		s = "1"
	}
	exec.Command("/bin/sh", "-c", "gpio write 12 "+s).CombinedOutput()
}

func (p *PFCperiph) ReadHumTempSensor() (h, t float64) {
	output, err := exec.Command("/bin/sh", "-c", "python /home/pi/go_exp/dht_read.py").CombinedOutput()
	//var err error = nil
	//output := "t=22.1\nh=40.5"
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return 0.0, 0.0
	} else {
		s := string(output)
		os.Stderr.WriteString(s)
		lines := strings.Split(s, "\n")
		strT := lines[0][2:]
		strH := lines[1][2:]
		t, _ := strconv.ParseFloat(strT, 64)
		h, _ := strconv.ParseFloat(strH, 64)
		//fmt.Println(t, h)
		return h, t
	}
}

/*

output, err := exec.Command("sudo", "ifdown wlan0").CombinedOutput()
if err != nil {
	os.Stderr.WriteString(err.Error())
}

exec.Command("/bin/sh", "-c", "sudo reboot").CombinedOutput()

*/
