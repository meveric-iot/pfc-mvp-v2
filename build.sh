#!/bin/bash 
export PATH=$PATH:/usr/local/go/bin # put into ~/.profile
env GOOS=linux GOARCH=arm GOARM=7 go build pfcserver.go worker.go config.go periphcontrol.go wificontrol.go logging.go chartsData.go
