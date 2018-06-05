#!/bin/bash 
export PATH=$PATH:/usr/local/go/bin # put into ~/.profile
go run pfcserver.go worker.go config.go periphcontrol.go wificontrol.go logging.go chartsData.go