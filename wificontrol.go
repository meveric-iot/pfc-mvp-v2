package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
)

var wifiController PFCwifiControl

// PFCwifiControl - type for wifi controller
type PFCwifiControl struct {
}

func (c *PFCwifiControl) initApPlusStationMode() error {
	// install packages: dnsmasq, hostapd
	// append startup script to rc.local
	// edit interfaces
	// edit dhcpcd.conf
	// edit dnsmasq.conf
	// call writeSettings for write initial setting
	// exec sudo reboot
	fmt.Println("Initialize WiFi controling...")
	exec.Command("/bin/sh", "-c", "sudo apt-get update").CombinedOutput()
	exec.Command("/bin/sh", "-c", "sudo apt-get remove -y --purge dnsmasq hostapd").CombinedOutput()
	exec.Command("/bin/sh", "-c", "sudo apt-get -y install dnsmasq hostapd").CombinedOutput()
	exec.Command("/bin/sh", "-c", "sudo apt-get -y purge dns-root-data").CombinedOutput()
	exec.Command("/bin/sh", "-c", "sudo systemctl disable hostapd").CombinedOutput()
	exec.Command("/bin/sh", "-c", "sudo systemctl disable dnsmasq").CombinedOutput()
	fmt.Println("Packages installed...")

	rcLocal := []byte(`#!/bin/sh -e
# Print the IP address
_IP=$(hostname -I) || true
if [ "$_IP" ]; then
printf "My IP address is %s\n" "$_IP"
fi	

sudo ifconfig wlan0 up

cd /home/pi/go_exp/ && sudo /home/pi/go_exp/pfcserver

exit 0`)
	err := ioutil.WriteFile("/etc/rc.local", rcLocal, 0666)
	if err != nil {
		return err
	}
	fmt.Println("rc.local writed")

	dhcpcdConf := []byte(`# A sample configuration for dhcpcd.
# See dhcpcd.conf(5) for details.

# Allow users of this group to interact with dhcpcd via the control socket.
#controlgroup wheel

# Inform the DHCP server of our hostname for DDNS.
hostname

# Use the hardware address of the interface for the Client ID.
clientid
# or
# Use the same DUID + IAID as set in DHCPv6 for DHCPv4 ClientID as per RFC4361.
# Some non-RFC compliant DHCP servers do not reply with this set.
# In this case, comment out duid and enable clientid above.
#duid

# Persist interface configuration when dhcpcd exits.
persistent

# Rapid commit support.
# Safe to enable by default because it requires the equivalent option set
# on the server to actually work.
option rapid_commit

# A list of options to request from the DHCP server.
option domain_name_servers, domain_name, domain_search, host_name
option classless_static_routes
# Most distributions have NTP support.
option ntp_servers
# Respect the network MTU. This is applied to DHCP routes.
option interface_mtu

# A ServerID is required by RFC2131.
require dhcp_server_identifier

# Generate Stable Private IPv6 Addresses instead of hardware based ones
slaac private

# Example static IP configuration:
#interface eth0
#static ip_address=192.168.0.10/24
#static ip6_address=fd51:42f8:caae:d92e::ff/64
#static routers=192.168.0.1
#static domain_name_servers=192.168.0.1 8.8.8.8 fd51:42f8:caae:d92e::1

# It is possible to fall back to a static IP if DHCP fails:
# define static profile
#profile static_eth0
#static ip_address=192.168.1.23/24
#static routers=192.168.1.1
#static domain_name_servers=192.168.1.1

# fallback to static profile on eth0
#interface eth0
#fallback static_eth0
nohook wpa_supplicant
`)
	err = ioutil.WriteFile("/etc/dhcpcd.conf", dhcpcdConf, 0666)
	if err != nil {
		return err
	}
	fmt.Println("dhcpcd.conf writed")

	dnsmasqConf := []byte(`# Delays sending DHCPOFFER and proxydhcp replies for at least the specified number of seconds.
dhcp-mac=set:client_is_a_pi,B8:27:EB:*:*:*
dhcp-reply-delay=tag:client_is_a_pi,2
#AutoHotspot config
interface=wlan0
bind-dynamic 
server=8.8.8.8
domain-needed
bogus-priv
dhcp-range=192.168.50.150,192.168.50.200,255.255.255.0,12h
`)
	err = ioutil.WriteFile("/etc/dnsmasq.conf", dnsmasqConf, 0666)
	if err != nil {
		return err
	}
	fmt.Println("dnsmasq.conf writed")

	err = c.writeSettings("PFC MVP v1.0", "12345670", "bbc", "12345678")
	if err != nil {
		return err
	}
	fmt.Println("settings writed")

	interfaces := []byte(`# interfaces(5) file used by ifup(8) and ifdown(8)

# Please note that this file is written to be used with dhcpcd
# For static IP, consult /etc/dhcpcd.conf and 'man dhcpcd.conf'

# Include files from /etc/network/interfaces.d:
source-directory /etc/network/interfaces.d
`)
	err = ioutil.WriteFile("/etc/network/interfaces", interfaces, 0666)
	if err != nil {
		return err
	}
	fmt.Println("interfaces writed")

	service := []byte(`[Unit]
Description=Automatically generates an internet Hotspot when a valid ssid is not in range
After=multi-user.target
[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/bin/autohotspotN
[Install]
WantedBy=multi-user.target`)
	err = ioutil.WriteFile("/etc/systemd/system/autohotspot.service", service, 0666)
	if err != nil {
		return err
	}
	fmt.Println("autohotspot.service writed")
	exec.Command("/bin/sh", "-c", "sudo systemctl enable autohotspot.service").CombinedOutput()

	script := []byte(`#!/bin/bash
#version 0.95-4-N/HS-I

#You may share this script on the condition a reference to RaspberryConnect.com 
#must be included in copies or derivatives of this script. 

#Network Wifi & Hotspot with Internet
#A script to switch between a wifi network and an Internet routed Hotspot
#A Raspberry Pi with a network port required for Internet in hotspot mode.
#Works at startup or with a seperate timer or manually without a reboot
#Other setup required find out more at
#http://www.raspberryconnect.com

wifidev="wlan0" #device name to use. Default is wlan0.
ethdev="eth0" #Ethernet port to use with IP tables
#use the command: iw dev ,to see wifi interface name 

IFSdef=$IFS
cnt=0
#These four lines capture the wifi networks the RPi is setup to use
wpassid=$(awk '/ssid="/{ print $0 }' /etc/wpa_supplicant/wpa_supplicant.conf | awk -F'ssid=' '{ print $2 }' ORS=',' | sed 's/\"/''/g' | sed 's/,$//')
IFS=","
ssids=($wpassid)
IFS=$IFSdef #reset back to defaults


#Note:If you only want to check for certain SSIDs
#Remove the # in in front of ssids=('mySSID1'.... below and put a # infront of all four lines above
# separated by a space, eg ('mySSID1' 'mySSID2')
#ssids=('mySSID1' 'mySSID2' 'mySSID3')

#Enter the Routers Mac Addresses for hidden SSIDs, seperated by spaces ie 
#( '11:22:33:44:55:66' 'aa:bb:cc:dd:ee:ff' ) 
mac=()

ssidsmac=("${ssids[@]}" "${mac[@]}") #combines ssid and MAC for checking

createAdHocNetwork()
{
	echo "Creating Hotspot"
	ip link set dev "$wifidev" down
	ip a add 192.168.50.5/24 brd + dev "$wifidev"
	ip link set dev "$wifidev" up
	dhcpcd -k "$wifidev" >/dev/null 2>&1
	iptables -t nat -A POSTROUTING -o "$ethdev" -j MASQUERADE
	iptables -A FORWARD -i "$ethdev" -o "$wifidev" -m state --state RELATED,ESTABLISHED -j ACCEPT
	iptables -A FORWARD -i "$wifidev" -o "$ethdev" -j ACCEPT
	systemctl start dnsmasq
	systemctl start hostapd
	echo 1 > /proc/sys/net/ipv4/ip_forward
}

KillHotspot()
{
	echo "Shutting Down Hotspot"
	ip link set dev "$wifidev" down
	systemctl stop hostapd
	systemctl stop dnsmasq
	iptables -D FORWARD -i "$ethdev" -o "$wifidev" -m state --state RELATED,ESTABLISHED -j ACCEPT
	iptables -D FORWARD -i "$wifidev" -o "$ethdev" -j ACCEPT
	echo 0 > /proc/sys/net/ipv4/ip_forward
	ip addr flush dev "$wifidev"
	ip link set dev "$wifidev" up
	dhcpcd  -n "$wifidev" >/dev/null 2>&1
}

ChkWifiUp()
{
	echo "Checking WiFi connection ok"
		sleep 20 #give time for connection to be completed to router
	if ! wpa_cli -i "$wifidev" status | grep 'ip_address' >/dev/null 2>&1
		then #Failed to connect to wifi (check your wifi settings, password etc)
			echo 'Wifi failed to connect, falling back to Hotspot.'
				wpa_cli terminate "$wifidev" >/dev/null 2>&1
			createAdHocNetwork
	fi
}


FindSSID()
{
#Check to see what SSID's and MAC addresses are in range
ssidChk=('NoSSid')
i=0; j=0
until [ $i -eq 1 ] #wait for wifi if busy, usb wifi is slower.
do
		ssidreply=$((iw dev "$wifidev" scan ap-force | egrep "^BSS|SSID:") 2>&1) >/dev/null 2>&1 
		echo "SSid's in range: " $ssidreply
		echo "Device Available Check try " $j
		if (($j >= 10)); then #if busy 10 times goto hotspot
					echo "Device busy or unavailable 10 times, going to Hotspot"
					ssidreply=""
					i=1
	elif echo "$ssidreply" | grep "No such device (-19)" >/dev/null 2>&1; then
				echo "No Device Reported, try " $j
		NoDevice
		elif echo "$ssidreply" | grep "Network is down (-100)" >/dev/null 2>&1 ; then
				echo "Network Not available, trying again" $j
				j=$((j + 1))
				sleep 2
	elif echo "$ssidreplay" | grep "Read-only file system (-30)" >/dev/null 2>&1 ; then
		echo "Temporary Read only file system, trying again"
		j=$((j + 1))
		sleep 2
	elif ! echo "$ssidreply" | grep "resource busy (-16)"  >/dev/null 2>&1 ; then
				echo "Device Available, checking SSid Results"
		i=1
	else #see if device not busy in 2 seconds
				echo "Device unavailable checking again, try " $j
		j=$((j + 1))
		sleep 2
	fi
done

for ssid in "${ssidsmac[@]}"
do
		if (echo "$ssidreply" | grep "$ssid") >/dev/null 2>&1
		then
			#Valid SSid found, passing to script
				echo "Valid SSID Detected, assesing Wifi status"
				ssidChk=$ssid
				return 0
		else
			#No Network found, NoSSid issued"
				echo "No SSid found, assessing WiFi status"
				ssidChk='NoSSid'
		fi
done
}

NoDevice()
{
	#if no wifi device,ie usb wifi removed, activate wifi so when it is
	#reconnected wifi to a router will be available
	echo "No wifi device connected"
	wpa_supplicant -B -i "$wifidev" -c /etc/wpa_supplicant/wpa_supplicant.conf >/dev/null 2>&1
	exit 1
}

FindSSID

#Create Hotspot or connect to valid wifi networks
if [ "$ssidChk" != "NoSSid" ] 
then
		echo 0 > /proc/sys/net/ipv4/ip_forward #deactivate ip forwarding
		if systemctl status hostapd | grep "(running)" >/dev/null 2>&1
		then #hotspot running and ssid in range
				KillHotspot
				echo "Hotspot Deactivated, Bringing Wifi Up"
				wpa_supplicant -B -i "$wifidev" -c /etc/wpa_supplicant/wpa_supplicant.conf >/dev/null 2>&1
				ChkWifiUp
		elif { wpa_cli -i "$wifidev" status | grep 'ip_address'; } >/dev/null 2>&1
		then #Already connected
				echo "Wifi already connected to a network"
		else #ssid exists and no hotspot running connect to wifi network
				echo "Connecting to the WiFi Network"
				wpa_supplicant -B -i "$wifidev" -c /etc/wpa_supplicant/wpa_supplicant.conf >/dev/null 2>&1
				ChkWifiUp
		fi
else #ssid or MAC address not in range
		if systemctl status hostapd | grep "(running)" >/dev/null 2>&1
		then
				echo "Hostspot already active"
		elif { wpa_cli status | grep "$wifidev"; } >/dev/null 2>&1
		then
				echo "Cleaning wifi files and Activating Hotspot"
				wpa_cli terminate >/dev/null 2>&1
				ip addr flush "$wifidev"
				ip link set dev "$wifidev" down
				rm -r /var/run/wpa_supplicant >/dev/null 2>&1
				createAdHocNetwork
		else #"No SSID, activating Hotspot"
				createAdHocNetwork
		fi
fi`)
	err = ioutil.WriteFile("/usr/bin/autohotspotN", script, 0666)
	if err != nil {
		return err
	}
	fmt.Println("autohotspot script writed")
	exec.Command("/bin/sh", "-c", "sudo chmod +x /usr/bin/autohotspotN").CombinedOutput()

	defhostapd := []byte(`# Defaults for hostapd initscript
#
# See /usr/share/doc/hostapd/README.Debian for information about alternative
# methods of managing hostapd.
#
# Uncomment and set DAEMON_CONF to the absolute path of a hostapd configuration
# file and hostapd will be started during system boot. An example configuration
# file can be found at /usr/share/doc/hostapd/examples/hostapd.conf.gz
#
#DAEMON_CONF=""

# Additional daemon options to be appended to hostapd command:-
# 	-d   show more debug messages (-dd for even more)
# 	-K   include key data in debug messages
# 	-t   include timestamps in some debug messages
#
# Note that -B (daemon mode) and -P (pidfile) options are automatically
# configured by the init.d script and must not be added to DAEMON_OPTS.
#
DAEMON_CONF="/etc/hostapd/hostapd.conf"`)
	err = ioutil.WriteFile("/etc/default/hostapd", defhostapd, 0666)
	if err != nil {
		return err
	}
	fmt.Println("/default/hostapd writed")

	sysctl := []byte(`#
# /etc/sysctl.conf - Configuration file for setting system variables
# See /etc/sysctl.d/ for additional system variables.
# See sysctl.conf (5) for information.
#

#kernel.domainname = example.com

# Uncomment the following to stop low-level messages on console
#kernel.printk = 3 4 1 3

##############################################################3
# Functions previously found in netbase
#

# Uncomment the next two lines to enable Spoof protection (reverse-path filter)
# Turn on Source Address Verification in all interfaces to
# prevent some spoofing attacks
#net.ipv4.conf.default.rp_filter=1
#net.ipv4.conf.all.rp_filter=1

# Uncomment the next line to enable TCP/IP SYN cookies
# See http://lwn.net/Articles/277146/
# Note: This may impact IPv6 TCP sessions too
#net.ipv4.tcp_syncookies=1

# Uncomment the next line to enable packet forwarding for IPv4
net.ipv4.ip_forward=1

# Uncomment the next line to enable packet forwarding for IPv6
#  Enabling this option disables Stateless Address Autoconfiguration
#  based on Router Advertisements for this host
#net.ipv6.conf.all.forwarding=1


###################################################################
# Additional settings - these settings can improve the network
# security of the host and prevent against some network attacks
# including spoofing attacks and man in the middle attacks through
# redirection. Some network environments, however, require that these
# settings are disabled so review and enable them as needed.
#
# Do not accept ICMP redirects (prevent MITM attacks)
#net.ipv4.conf.all.accept_redirects = 0
#net.ipv6.conf.all.accept_redirects = 0
# _or_
# Accept ICMP redirects only for gateways listed in our default
# gateway list (enabled by default)
# net.ipv4.conf.all.secure_redirects = 1
#
# Do not send ICMP redirects (we are not a router)
#net.ipv4.conf.all.send_redirects = 0
#
# Do not accept IP source route packets (we are not a router)
#net.ipv4.conf.all.accept_source_route = 0
#net.ipv6.conf.all.accept_source_route = 0
#
# Log Martian Packets
#net.ipv4.conf.all.log_martians = 1
#

###################################################################
# Magic system request Key
# 0=disable, 1=enable all
# Debian kernels have this set to 0 (disable the key)
# See https://www.kernel.org/doc/Documentation/sysrq.txt
# for what other values do
#kernel.sysrq=1

###################################################################
# Protected links
#
# Protects against creating or following links under certain conditions
# Debian kernels have both set to 1 (restricted) 
# See https://www.kernel.org/doc/Documentation/sysctl/fs.txt
#fs.protected_hardlinks=0
#fs.protected_symlinks=0
net.ipv4.ip_forward=1
`)
	err = ioutil.WriteFile("/etc/sysctl.conf", sysctl, 0666)
	if err != nil {
		return err
	}
	fmt.Println("sysctl.conf writed")

	c.writeSettings("PFC MVP v1.1", "12345678", "sopl", "")

	exec.Command("/bin/sh", "-c", "sudo reboot").CombinedOutput()

	return err
}

func (c *PFCwifiControl) writeSettings(apSsid, apPass, ssid, pass string) error {

	if pass == "" {
		pass = "key_mgmt=NONE"
	} else {
		pass = `psk="` + pass + `"`
	}

	wpaSupplicant := []byte(`ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
country=GB
update_config=1

network={
ssid="` + ssid + `"
` + pass + `
}
`)
	err := ioutil.WriteFile("/etc/wpa_supplicant/wpa_supplicant.conf", wpaSupplicant, 0666)
	if err != nil {
		return err
	}
	fmt.Println("wpa_supplicant.conf writed")
	hostapd := []byte(`
#2.4GHz setup wifi 80211 b,g,n
interface=wlan0
driver=nl80211
ssid=` + apSsid + `
hw_mode=g
channel=8
wmm_enabled=0
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0
wpa=2
wpa_passphrase=` + apPass + `
wpa_key_mgmt=WPA-PSK
wpa_pairwise=CCMP TKIP
rsn_pairwise=CCMP

#80211n - Change GB to your WiFi country code
country_code=GB
ieee80211n=1
ieee80211d=1`)
	err = ioutil.WriteFile("/etc/hostapd/hostapd.conf", hostapd, 0755)
	if err != nil {
		return err
	}
	fmt.Println("hostapdstart writed")

	exec.Command("/bin/sh", "-c", "sudo /usr/bin/autohotspotN").CombinedOutput()

	return err
}

/*output, err := exec.Command("sudo", "ifdown wlan0").CombinedOutput()
if err != nil {
	os.Stderr.WriteString(err.Error())
}*/
