#!usr/bin/python
import sys
import Adafruit_DHT

while True:
	humidity, temp = Adafruit_DHT.read_retry(11, 4)
	print 't={0:0.1f}\nh={1:0.1f}'.format(temp, humidity)
	break
