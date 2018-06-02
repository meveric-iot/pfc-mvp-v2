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
	//exec.Command("/bin/sh", "-c", "sudo apt-get remove -y --purge dnsmasq hostapd").CombinedOutput()
	//exec.Command("/bin/sh", "-c", "sudo apt-get -y install dnsmasq hostapd").CombinedOutput()

	fmt.Println("Packages installed...")

	rcLocal := []byte(`#!/bin/sh -e
# Print the IP address
_IP=$(hostname -I) || true
if [ "$_IP" ]; then
printf "My IP address is %s\n" "$_IP"
fi	

sudo /bin/bash /usr/local/bin/hostapdstart

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
# Respect the network MTU.
# Some interface drivers reset when changing the MTU so disabled by default.
#option interface_mtu

# A ServerID is required by RFC2131.
require dhcp_server_identifier

# Generate Stable Private IPv6 Addresses instead of hardware based ones
slaac private

# A hook script is provided to lookup the hostname if not set by the DHCP
# server, but it should not be run by default.
nohook lookup-hostname

#Ethernet connection
#static ip_address=10.0.0.205/24
#static routers=10.0.0.1
#static domain_name_servers=10.0.0.1
#denyinterfaces wlan0
interface uap0`)
	err = ioutil.WriteFile("/etc/dhcpcd.conf", dhcpcdConf, 0666)
	if err != nil {
		return err
	}
	fmt.Println("dhcpcd.conf writed")

	dnsmasqConf := []byte(`# Delays sending DHCPOFFER and proxydhcp replies for at least the specified number of seconds.
dhcp-mac=set:client_is_a_pi,B8:27:EB:*:*:*
dhcp-reply-delay=tag:client_is_a_pi,2
interface=lo,uap0
no-dhcp-interface=lo,wlan0
bind-interfaces
server=8.8.8.8
domain-needed
bogus-priv
dhcp-range=192.168.50.50,192.168.50.150,12h`)
	err = ioutil.WriteFile("/etc/dnsmasq.conf", dnsmasqConf, 0666)
	if err != nil {
		return err
	}
	fmt.Println("dnsmasq.conf writed")

	wpaSupplicant := []byte(`ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
country=RU
	
network={
  scan_ssid=1  
}`)
	err = ioutil.WriteFile("/etc/wpa_supplicant/wpa_supplicant.conf", wpaSupplicant, 0666)
	if err != nil {
		return err
	}
	fmt.Println("wpa_supplicant.conf writed")

	err = c.writeSettings("PFC MVP v1.0", "12345670", "bbc", "12345678")
	if err != nil {
		return err
	}
	fmt.Println("settings writed")

	return err
}

func (c *PFCwifiControl) writeSettings(apSsid, apPass, ssid, pass string) error {

	hostapdstart := []byte(`#!/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
iw dev wlan0 interface add uap0 type __ap
sysctl net.ipv4.ip_forward=1
iptables -t nat -A POSTROUTING -s 192.168.50.0/24 ! -d 192.168.50.0/24 -j MASQUERADE
# write the first part of the conf file
echo -en 'interface=uap0\nssid=` + apSsid + `\ndriver=nl80211\nhw_mode=g\nieee80211n=1\nchannel=' > /etc/hostapd/hostapd.conf
iwlist channel 2> /dev/null | awk '/Current/ {print substr($5,1,length($5) - 1)}' >> /etc/hostapd/hostapd.conf
echo -e 'macaddr_acl=0\nauth_algs=1\nignore_broadcast_ssid=0\nwpa=2\nctrl_interface=/var/run/hostapd\nctrl_interface_group=0' >> /etc/hostapd/hostapd.conf
echo -e 'wpa_passphrase=` + apPass + `\nwpa_key_mgmt=WPA-PSK\nwpa_pairwise=TKIP\nrsn_pairwise=CCMP' >> /etc/hostapd/hostapd.conf
ifdown wlan0
ip link set dev uap0 up
ip addr add 192.168.50.1/24 broadcast 192.168.50.255 dev uap0
sleep 1
# start hostapd
hostapd -B -P /run/hostapd.pid /etc/hostapd/hostapd.conf &
sleep 1
service hostapd restart
ifup wlan0
service dnsmasq restart`)
	err := ioutil.WriteFile("/usr/local/bin/hostapdstart", hostapdstart, 0755)
	if err != nil {
		return err
	}
	fmt.Println("hostapdstart writed")

	interfaces := []byte(`source-directory /etc/network/interfaces.d
auto wlan0
iface wlan0 inet dhcp
  wpa-ssid "` + ssid + `"
  wpa-psk "` + pass + `"`)
	err = ioutil.WriteFile("/etc/network/interfaces", interfaces, 0666)
	if err != nil {
		return err
	}
	fmt.Println("interfaces writed")

	exec.Command("/bin/sh", "-c", "sudo reboot").CombinedOutput()

	return err
}

/*output, err := exec.Command("sudo", "ifdown wlan0").CombinedOutput()
if err != nil {
	os.Stderr.WriteString(err.Error())
}*/
