package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/go-home-io/server/plugins/common"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/ipv4"
)

// Discovery provider.
type discoverUPNP struct {
	logger     common.ILoggerProvider
	advAddress string

	connection *net.UDPConn
}

// Start starts new UPNP server.
func (d *discoverUPNP) Start() error {
	l, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 0),
		Port: 1900,
	})

	if err != nil {
		d.logger.Error("Failed to start UPNP server", err)
		return err
	}

	d.connection = l
	p := ipv4.NewPacketConn(d.connection)

	interfaces, err := net.Interfaces()
	if err != nil {
		d.logger.Error("No available interfaces", err)
		return err
	}

	addr := &net.UDPAddr{
		IP: net.IPv4(239, 255, 255, 250),
	}

	var joined []string
	for _, v := range interfaces {
		if v.Flags&net.FlagMulticast == 0 {
			continue
		}

		err = p.JoinGroup(&v, addr)
		if err != nil {
			continue
		}

		joined = append(joined, v.Name)
	}

	if len(joined) == 0 {
		err = errors.New("no available multicast interfaces")
		d.logger.Error("Failed to start UPNP", err)
		return err
	}

	d.logger.Debug("Started UPNP server", "addresses", strings.Join(joined, " "))

	go d.listen()
	return nil
}

// Stop stops running UPNP server.
func (d *discoverUPNP) Stop() {
	d.connection.Close()
}

// Waits for incoming UDP messages.
func (d *discoverUPNP) listen() {
	var b [1500]byte
	for {
		n, add, err := d.connection.ReadFromUDP(b[:])
		if err != nil {
			d.logger.Error("Failed to read from UPND", err)
			return
		}

		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(b[:n])))
		if err != nil {
			d.logger.Error("UPNP request failed", err)
			continue
		}

		if req.Method != "M-SEARCH" || req.URL.Path != "*" ||
			req.Header.Get("Man") != `"ssdp:discover"` ||
			req.Header.Get("St") != `urn:schemas-upnp-org:device:basic:1` {
			continue
		}

		d.logger.Info("Received discovery request", "address", add.String())
		d.discoveryRespond(add)
	}
}

// Responds to discovery message with advertising address.
func (d *discoverUPNP) discoveryRespond(addr *net.UDPAddr) {
	c, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		d.logger.Error("Discovery respond error", err)
		return
	}
	defer c.Close()

	url := fmt.Sprintf("http://%s/upnp/setup.xml", d.advAddress)

	var buf bytes.Buffer
	buf.WriteString("HTTP/1.1 200 OK\r\n")
	http.Header{
		"Cache-Control": {`max-age=300`},
		"Ext":           {``},
		"Location":      {url},
		"Opt":           {`"http://schemas.upnp.org/upnp/1/0/"; ns=01`},
		"St":            {`urn:schemas-upnp-org:device:basic:1`},
		"Usn":           {`uuid:f6543a06-800d-48ba-8d8f-bc2949eddc33`},
	}.Write(&buf)
	buf.WriteString("\r\n")

	_, err = c.Write(buf.Bytes())
	if err != nil {
		d.logger.Error("Error writing UPnP response:", err)
	}

}

// Setup replies on initial HTTP discovery request
// with emulator details in XML format.
func (d *discoverUPNP) Setup(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	type Root struct {
		XMLName struct{} `xml:"urn:schemas-upnp-org:device-1-0 root"`

		Major int `xml:"specVersion>major"`
		Minor int `xml:"specVersion>minor"`

		URLBase string

		DeviceType   string `xml:"device>deviceType"`
		FriendlyName string `xml:"device>friendlyName"`
		Manufacturer string `xml:"device>manufacturer"`
		ModelName    string `xml:"device>modelName"`
		ModelNumber  string `xml:"device>modelNumber"`
		UDN          string `xml:"device>UDN"`
	}

	x := Root{
		Major: 1,
		Minor: 0,

		URLBase: "http://" + d.advAddress + "/",

		DeviceType:   "urn:schemas-upnp-org:device:Basic:1",
		FriendlyName: "go-home",
		Manufacturer: "Royal Philips Electronics",
		ModelName:    "Philips hue bridge 2012",
		ModelNumber:  "929000226503",
		UDN:          "uuid:f6543a06-800d-48ba-8d8f-bc2949eddc33",
	}

	w.Header().Set("Content-Type", "application/xml")
	io.WriteString(w, xml.Header)
	if err := xml.NewEncoder(w).Encode(&x); err != nil {
		d.logger.Error("Encoder error", err)
		return
	}
	io.WriteString(w, "\n")
}
