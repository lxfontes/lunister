package messages

import (
	"net"
)

type ARPRequest struct {
	SourceAddress string `json:"source_hw"`
	ProtocolType  string `json:"protocol"`
	SourceIP      string `json:"source_ip"`
	DestinationIP string `json:"dest_ip"`
}

type ARPResponse struct {
	DestinationAddress string `json:"dest_hw"`
}

func (arp ARPResponse) DestinationEther() net.HardwareAddr {
	mac, _ := net.ParseMAC(arp.DestinationAddress)
	return mac
}
