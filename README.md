# pfc-mvp-v2
Controller and web-interface server for MIT Personal Food Computer MVP, written in Golang;
RPi connecting to Wi-Fi network and establish web-interface for controlling;
If It can't connect to network It starting own AP for configurating; URL: http://192.168.50.5

At begin, you should configure RPi:
* connect your RPi to internet
* install Python and Adafruit DHT library
* install wiringpi 
* clone repo in /home/pi/go_exp/, run "sudo bash build.sh", run "sudo /home/pi/go_exp/pfcserver"

if something goes wrong, try this manual:
http://www.raspberryconnect.com/network/item/333-raspberry-pi-hotspot-access-point-dhcpcd-method