package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/lxfontes/lunister/messages"
	"github.com/mdlayher/arp"
	"github.com/mdlayher/ethernet"
	"github.com/songgao/water/waterutil"
)

type Bridge struct {
	ifaceClients   *Interface
	ifaceTrunk     *Interface
	broadcastQueue *LeakyBucket
	wg             sync.WaitGroup
	workerCount    int
	quitC          chan bool
}

func NewBridge(client string, trunk string, broadcastWorkers int) (*Bridge, error) {
	var err error
	b := &Bridge{
		broadcastQueue: NewLeakyBucket(500),
		quitC:          make(chan bool),
		workerCount:    broadcastWorkers,
	}

	b.ifaceClients, err = NewInterface(client, 1520, b)
	if err != nil {
		return nil, err
	}

	b.ifaceTrunk, err = NewInterface(trunk, 1520, b)
	return b, err
}

func (bridge *Bridge) Start() {
	bridge.ifaceClients.ReadLoop()
	bridge.ifaceTrunk.ReadLoop()

	for i := 0; i < bridge.workerCount; i++ {
		fmt.Println("Starting worker")
		go bridge.broadcastWorker()
	}
}

func (bridge *Bridge) Stop() {
	bridge.ifaceClients.Stop()
	bridge.ifaceTrunk.Stop()
	close(bridge.quitC)
}

func (bridge *Bridge) ReceivePacket(tap *Interface, packet []byte) {
	switch tap {
	case bridge.ifaceClients:
		bridge.handleClientPacket(packet)
	case bridge.ifaceTrunk:
		bridge.ifaceClients.Write(packet)
	}
}

func (bridge *Bridge) handleClientPacket(packet []byte) {
	sourceMAC := waterutil.MACSource(packet)
	destMAC := waterutil.MACDestination(packet)

	ethType := waterutil.MACEthertype(packet)
	if ethType == waterutil.ARP {
		bridge.handleClientARP(packet)
		return
	}

	ethPayload := waterutil.MACPayload(packet)

	if !(waterutil.IsIPv4(ethPayload)) {
		// ignore packet
		fmt.Println("ignoring non-ipv4 packet")
		return
	}

	sourceIP := waterutil.IPv4Source(ethPayload)
	destIP := waterutil.IPv4Destination(ethPayload)

	fmt.Printf("srcMac %v dstMac %v srcIP %v dstIP %v\n", sourceMAC, destMAC, sourceIP, destIP)
	bridge.ifaceTrunk.Write(packet)
}

func (bridge *Bridge) handleClientARP(packet []byte) {
	// check if broadcast
	// y: ask controller (pass / respond / block)
	fmt.Printf("ARP Queue %v\n", packet)

	arpReq := arp.Packet{}
	etherPayload := waterutil.MACPayload(packet)

	err := arpReq.UnmarshalBinary(etherPayload)
	if err != nil {
		fmt.Println("error parsing arp", err)
		return
	}

	if arpReq.Operation == arp.OperationRequest {
		bridge.broadcastQueue.Push(packet)
	} else {
		// this guy is responding to external arp.. not sure what to do yet
		bridge.ifaceTrunk.Write(packet)
	}
}

func (bridge *Bridge) ReceivePacketError(tap *Interface, err error) {
	fmt.Println("Error", err)
}

func (bridge *Bridge) InterfaceStart(tap *Interface) {
	fmt.Println("Started", tap.Name)
}

func (bridge *Bridge) InterfaceStop(tap *Interface) {
	fmt.Println("Stopped", tap.Name)
}

func (bridge *Bridge) broadcastWorker() {
	bridge.wg.Add(1)
	defer bridge.wg.Done()
	running := true
	for running {
		select {
		case pkt := <-bridge.broadcastQueue.Queue():
			bridge.extRequest(pkt.([]byte))
		case <-bridge.quitC:
			running = false
		}
	}
}

func (bridge *Bridge) extRequest(packet []byte) {
	arpReq := arp.Packet{}
	etherPayload := waterutil.MACPayload(packet)

	err := arpReq.UnmarshalBinary(etherPayload)
	if err != nil {
		fmt.Println("error requesting broadcast", err)
		return
	}

	req := &messages.ARPRequest{
		SourceAddress: arpReq.SenderHardwareAddr.String(),
		ProtocolType:  "IPv4",
		SourceIP:      arpReq.SenderIP.String(),
		DestinationIP: arpReq.TargetIP.String(),
	}

	reqJson := new(bytes.Buffer)
	json.NewEncoder(reqJson).Encode(req)
	response, err := http.Post("http://127.0.0.1:8080/broadcast", "application/json", reqJson)

	if err != nil {
		fmt.Println("Failed to process arp", err)
		return
	}

	if response.StatusCode != 200 {
		fmt.Println("Invalid response status", response.Status)
		return
	}

	var resp messages.ARPResponse
	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		fmt.Println("Couldn't understand upstream format")
		return
	}

	fmt.Println("Should reply with", resp.DestinationAddress)

	arpResp, err := arp.NewPacket(arp.OperationReply, resp.DestinationEther(), arpReq.TargetIP, arpReq.SenderHardwareAddr, arpReq.SenderIP)

	if err != nil {
		fmt.Println("Failed to assemble response", err)
		return
	}

	var ethResp ethernet.Frame
	ethResp.EtherType = ethernet.EtherTypeARP
	ethResp.Source = resp.DestinationEther()
	ethResp.Destination = arpReq.SenderHardwareAddr
	ethResp.Payload, err = arpResp.MarshalBinary()
	if err != nil {
		fmt.Println("Failed to assemble response", err)
		return
	}

	fakeResp, err := ethResp.MarshalBinary()
	if err != nil {
		fmt.Println("Failed to assemble response", err)
		return
	}

	fmt.Println(fakeResp)
	// fake the arp back
	bridge.ifaceClients.Write(fakeResp)
}
