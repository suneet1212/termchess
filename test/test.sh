#!/usr/bin/env bash

# check if we have sudo access, if not exit
if [[ $EUID -ne 0 ]]; then
    echo "run as root: sudo $0 $*" >&2
    exit 1
fi

INTERNET_NS="internet"

create_internet() {
    ip netns add "${INTERNET_NS}"
    ip netns exec "${INTERNET_NS}" ip link add br0 type bridge
    ip netns exec "${INTERNET_NS}" ip link set br0 up
}

create_router_and_connect_to_ns() {
    local router_id="$1"
    local public_ip="$2"  # give a /8 IP address
    local connection_ns=$3
    local router_ns=r"${router_id}"

    ip netns add "${router_ns}"
    ip netns exec "${router_ns}" ip link add br0 type bridge
    ip netns exec "${router_ns}" ip addr add 192.168.1.1/24 dev br0
    ip netns exec "${router_ns}" ip link set br0 up

    # add a link from the router to the namespace
    ip link add v-int netns "${router_ns}" type veth peer name v"${router_ns}" netns "${connection_ns}"
    ip netns exec "${connection_ns}" ip link set v"${router_ns}" master br0
    ip netns exec "${router_ns}" ip addr add "${public_ip}" dev v-int
    ip netns exec "${connection_ns}" ip link set v"${router_ns}" up
    ip netns exec "${router_ns}" ip link set v-int up

    # setup NAT on the router
    ip netns exec "${router_ns}" iptables -t nat -A POSTROUTING -o v-int -j MASQUERADE
}

connect_container_to_router() {
    local pid=$1
    local container_ip="$2"  # give a 192.168.1.0/24 IP address
    local router_id="$3"
    local router_ns=r"${router_id}"

    # connect the container to the router
    ip link add v-router netns "${pid}" type veth peer name v"${pid}" netns "${router_ns}"
    nsenter -t "${pid}" -n ip addr add "${container_ip}" dev v-router
    nsenter -t "${pid}" -n ip link set v-router up
    ip netns exec "${router_ns}" ip link set v"${pid}" master br0
    ip netns exec "${router_ns}" ip link set v"${pid}" up

    # set default route to router
    nsenter -t "${pid}" -n ip route add default via 192.168.1.1 dev v-router
}

create_direct_connection() {
    local pid="$1"
    local public_ip="$2"  # give a /8 IP address

    # add a link from the container to the internet
    ip link add v-int netns "${pid}" type veth peer name v"c${pid}" netns "${INTERNET_NS}"
    ip netns exec "${INTERNET_NS}" ip link set v"c${pid}" master br0
    nsenter -t "${pid}" -n ip addr add "${public_ip}" dev v-int
    ip netns exec "${INTERNET_NS}" ip link set v"c${pid}" up
    nsenter -t "${pid}" -n ip link set v-int up
}

setup_firewall_in_container() {
    local pid="$1"
    nsenter -t "${pid}" -n iptables -A INPUT -m conntrack --ctstate INVALID,NEW,UNTRACKED -j DROP
}

# spin up the docker setup
docker compose up -d

# find the PIDs for the containers so that we can find the netns
PEER_ONLY_FIREWALL_1_PID="$(docker inspect -f '{{ .State.Pid }}' peer-only-firewall-1)"
PEER_ONLY_FIREWALL_2_PID="$(docker inspect -f '{{ .State.Pid }}' peer-only-firewall-2)"
PEER_NAT_1_PID="$(docker inspect -f '{{ .State.Pid }}' peer-nat-1)"
PEER_NAT_2_PID="$(docker inspect -f '{{ .State.Pid }}' peer-nat-2)"
PEER_NAT_AND_FIREWALL_1_PID="$(docker inspect -f '{{ .State.Pid }}' peer-nat-and-firewall-1)"
PEER_NAT_AND_FIREWALL_2_PID="$(docker inspect -f '{{ .State.Pid }}' peer-nat-and-firewall-2)"
PEER_CGNAT_1_PID="$(docker inspect -f '{{ .State.Pid }}' peer-cgnat-1)"
PEER_CGNAT_2_PID="$(docker inspect -f '{{ .State.Pid }}' peer-cgnat-2)"
PEER_ALL_REACHABLE_PID="$(docker inspect -f '{{ .State.Pid }}' peer-all-reachable)"

# setup internet
create_internet

# setup firewall in containers that require firewall
setup_firewall_in_container "${PEER_ONLY_FIREWALL_1_PID}"
setup_firewall_in_container "${PEER_ONLY_FIREWALL_2_PID}"
setup_firewall_in_container "${PEER_NAT_AND_FIREWALL_1_PID}"
setup_firewall_in_container "${PEER_NAT_AND_FIREWALL_2_PID}"

# directly connect all reachable container
create_direct_connection  "${PEER_ALL_REACHABLE_PID}" "10.0.0.2/8"

# directly connect the only firewall containers to internet (no NAT so no routers in between)
create_direct_connection "${PEER_ONLY_FIREWALL_1_PID}" "10.0.1.2/8"
create_direct_connection "${PEER_ONLY_FIREWALL_2_PID}" "10.0.1.3/8"

# create NAT peers
# first router
create_router_and_connect_to_ns "1" "10.0.2.2/8" "${INTERNET_NS}"
connect_container_to_router "${PEER_NAT_1_PID}" "192.168.1.2/24" "1"

# second router
create_router_and_connect_to_ns "2" "10.0.2.3/8" "${INTERNET_NS}"
connect_container_to_router "${PEER_NAT_2_PID}" "192.168.1.2/24" "2"

