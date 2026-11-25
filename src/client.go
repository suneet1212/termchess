package src

import (
	"fmt"
	"net"
	"os"
	"time"
)

func StartClient() {
	// Step 1: identify ipv4 and ipv6 address
	ipv6Addresses, err := getValidIpv6Addresses()
	if err != nil {
		fmt.Println(err.Error())
	}

	isIPV6Present := len(ipv6Addresses) > 0
	isEasyNAT := false

	// Step 2: STUN call
	if !isIPV6Present {
		// Step 3: if ipv4 only, check for nat type
		serverIp1 := os.Getenv("server_1_ip")
		serverPort1 := os.Getenv("server_1_port")
		serverIp2 := os.Getenv("server_2_ip")
		serverPort2 := os.Getenv("server_2_port")

		ipv4_1, port_1, err := CallStun(serverIp1, serverPort1)
		ipv4_2, port_2, err := CallStun(serverIp2, serverPort2)
		if err != nil {
			fmt.Println(err.Error())
		}
		if ipv4_1.Equal(ipv4_2) && port_1 == port_2 {
			isEasyNAT = true
			fmt.Println("Easy NAT!!!!!")
		} else {
			fmt.Printf("%d.%d.%d.%d:%d\n", ipv4_1[0], ipv4_1[1], ipv4_1[2], ipv4_1[3], port_1)
			fmt.Printf("%d.%d.%d.%d:%d\n", ipv4_2[0], ipv4_2[1], ipv4_2[2], ipv4_2[3], port_2)
		}
	}

	// Step 4:
	if isIPV6Present || isEasyNAT {
		// directly accessible, share data to Server and call for simultaneous break
		// Step 5: Coordinate with Server for the exact time to make udp request
		// TODO - defining values for now, call server to fetch these values
		timer := time.NewTimer(1 * time.Second)
		opponentIp := "127.0.0.1"
		opponentPort := "18745"

		// Step 6: Make UDP request to break firewall
		var firewallBroken = false
		for range 3 {
			timer = time.NewTimer(1 * time.Second)
			firewallBroken, _ = udpCall(timer.C, opponentIp, opponentPort)
			if firewallBroken {
				break
			}
		}
		if firewallBroken {
			fmt.Println("Lessgoo!!")
		}

		// Step 7: Make Wireguard connection
	} else {
		// Hard NAT and no IPV6
		// Need a relay here
	}
}

func udpCall(timer <-chan time.Time, udpIp, udpPort string) (bool, error) {
	<-timer
	fmt.Println("Sending message to opponent")
	conn, err := net.DialTimeout("udp", udpIp+":"+udpPort, 3*time.Second)
	if err != nil {
		fmt.Println("Failed to resolve server address:", err)
		return false, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	req := []byte("hello from udp")
	_, err = conn.Write(req)
	if err != nil {
		fmt.Println("Cannot write the udp call: ", err)
		return false, err
	}

	response := make([]byte, 2048)
	x, err := conn.Read(response)
	if err != nil {
		fmt.Println("Error in reading response from udp call", err)
		return false, err
	}
	response = response[:x]

	fmt.Println("recd following message:" + string(response))
	return x > 0, nil
}

func getValidIpv6Addresses() ([]net.Addr, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var validIPV6 []net.Addr
	for _, addr := range addrs {
		// fmt.Printf("%T -> %s -> %d\n", addr, addr, addr)
		ip, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		if ip.IP.To4() == nil && ip.IP.IsGlobalUnicast() {
			validIPV6 = append(validIPV6, addr)
		}
	}
	return validIPV6, nil
}
