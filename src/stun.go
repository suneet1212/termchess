package src

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"time"
)

func CallStun(serverIp, serverPort string) (net.IP, int, error) {
	stunMagicCookie := 0x2112A442
	var ipv4 net.IP
	var port uint16

	n, err := rand.Int(rand.Reader, big.NewInt(900_000_000_000))
	if err != nil {
		// panic(err)
		return nil, 0, err
	}

	tid := fmt.Sprintf("%012d", n.Int64()+100_000_000_000)

	// txnId := string(tid)
	var req []byte
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, 1)
	req = append(req, buf...)
	binary.BigEndian.PutUint16(buf, 0)
	req = append(req, buf...)
	buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(stunMagicCookie))
	req = append(req, buf...)
	req = append(req, tid...)

	conn, err := net.DialTimeout("udp", serverIp+":"+serverPort, 3*time.Second)
	if err != nil {
		fmt.Println("Failed to resolve server address:", err)
		return nil, 0, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))

	_, err = conn.Write(req)
	if err != nil {
		fmt.Println("Failed to send request to stun server")
	}

	response := make([]byte, 2048)
	x, err := conn.Read(response)
	if err != nil {
		fmt.Println("Failed to receive message:", err)
		return nil, 0, err
	}
	// fmt.Println(response[:x])
	response = response[:x]
	msgType := binary.BigEndian.Uint16(response[:2])
	magic := binary.BigEndian.Uint32(response[4:8])

	if msgType == 0x0101 && int(magic) == stunMagicCookie {
		pos := 20
		for pos < len(response) {
			attrType := binary.BigEndian.Uint16(response[pos : pos+2])
			attrLen := binary.BigEndian.Uint16(response[pos+2 : pos+4])
			if attrType == 0x0020 {
				family := response[pos+5]
				xorPort := binary.BigEndian.Uint16(response[pos+6 : pos+8])
				port = xorPort ^ uint16(stunMagicCookie>>16)
				// fmt.Printf("port = %d\n", port)
				if family == 0x01 {
					// ipv4 address family
					xorIp := binary.BigEndian.Uint32(response[pos+8 : pos+12])
					ipInt := xorIp ^ uint32(stunMagicCookie)
					ipv4 = make(net.IP, 4)
					binary.BigEndian.PutUint32(ipv4, ipInt)
					// fmt.Printf("ip = %d\n", ipv4)
					// return nil, 0, err
				} else {
					fmt.Printf("new family detected - %d\n", family)
				}
			}
			pos += int(attrLen)
		}
	}
	return ipv4, int(port), err
}
