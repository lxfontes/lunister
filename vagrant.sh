#!/bin/bash
sudo yum install -y bridge-utils epel-release
sudo yum install -y wireshark ngrep net-tools

# create namespaces + interfaces for testing
sudo ip link add trunkpeer type veth peer name trunketh
sudo ip link add clientpeer type veth peer name clienteth

sudo ip netns add trunkns
sudo ip link set trunketh netns trunkns

sudo ip netns add clientns
sudo ip link set clienteth netns clientns

sudo ip netns exec trunkns ifconfig trunketh 10.0.0.1/24 up
sudo ip netns exec clientns ifconfig clienteth 10.0.0.2/24 up

brctl addbr trunk
brctl addif trunk trunkpeer

brctl addbr client
brctl addif client clientpeer

ifconfig trunkpeer up
ifconfig clientpeer up

ifconfig trunk up
ifconfig client up


ip tuntap add dev trunktap mode tap
ip tuntap add dev clienttap mode tap

brctl addif trunk trunktap
brctl addif client clienttap

ifconfig trunktap up
ifconfig clienttap up
