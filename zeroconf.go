package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/godbus/dbus"
	"github.com/oleksandr/bonjour"
)

const avahiDbusName string = "org.freedesktop.Avahi"
const avahiInterfaceServer string = avahiDbusName + ".Server"
const avahiInterfaceEntrygroup string = avahiDbusName + ".EntryGroup"
const avahiEntrygroupAddService string = avahiInterfaceEntrygroup + ".AddService"
const avahiEntrygroupCommit = avahiInterfaceEntrygroup + ".Commit"

const avahiIfUnspec = int32(-1)
const protoUnspec = int32(-1)
const protoInet = int32(0)
const protoInet6 = int32(1)

func registerAvahiService(name, serviceType string, serverPort uint16, txt []string) bool {
	conn, err := dbus.SystemBus()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to D-BUS session bus:", err)
		return false
	}

	obj := conn.Object(avahiDbusName, "/")

	var groupPath dbus.ObjectPath
	call := obj.Call("org.freedesktop.Avahi.Server.EntryGroupNew", 0)
	if call.Err != nil {
		fmt.Fprintln(os.Stderr, "Avahi not found on D-BUS")
		return false
	}
	call.Store(&groupPath)

	group := conn.Object(avahiDbusName, groupPath)

	var txtBytes [][]byte
	for _, s := range txt {
		txtBytes = append(txtBytes, []byte(s))
	}

	var flags = uint32(0)
	var domain = "local"
	var host = string("")
	call = group.Call(avahiEntrygroupAddService, 0, avahiIfUnspec, protoUnspec, flags, name, serviceType, domain, host, serverPort, txtBytes)
	if call.Err != nil {
		fmt.Println(os.Stderr, "Failed to call", avahiEntrygroupAddService)
		return false
	}
	call = group.Call(avahiEntrygroupCommit, 0)
	if call.Err != nil {
		fmt.Println(os.Stderr, "Failed to call", avahiEntrygroupCommit)
		return false
	}

	return true
}

func registerViaBuiltin(hostname, name, serviceType string, serverPort uint16, txt []string) {
	addrs, err := net.LookupIP(hostname)
	if err != nil {
		log.Fatalln("Could not determine host IP addresses for %s", hostname)
	}

	var ip4 string
	var ip6 string
	for i := 0; i < len(addrs); i++ {
		if ipv4 := addrs[i].To4(); ipv4 != nil {
			if !addrs[i].IsLoopback() {
				ip4 = addrs[i].String()
			}
		} else if ipv6 := addrs[i].To16(); ipv6 != nil {
			if !addrs[i].IsLoopback() {
				ip6 = addrs[i].String()
			}
		}
	}

	var ip string
	if len(ip4) > 0 {
		ip = ip4
	} else if len(ip6) > 0 {
		ip = ip6
	} else {
		log.Fatalln("Neither ipv4 or ipv6 address resolved to hostname")
	}

	_, err = bonjour.RegisterProxy(name, serviceType, "local", int(serverPort), hostname, ip, txt, nil)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func installZeroconfListener(name, serviceType string, serverPort uint16) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("Failed to fetch hostname")
	}

	nameHostname := strings.ToLower(name + "@" + hostname)

	txt := []string{"host=" + hostname}
	registered := registerAvahiService(nameHostname, serviceType, serverPort, txt)
	if !registered {
		registerViaBuiltin(hostname, nameHostname, serviceType, serverPort, txt)
		log.Println("Registered via internal bonjour handler")
	} else {
		log.Println("Registered via AVAHI")
	}
}
