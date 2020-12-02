package libhttpserver

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func readRequestFromConnection(conn net.Conn) ([]byte, error) {
	temp := make([]byte, buffSize)
	data := make([]byte, 0)
	length := 0

	for {
		n, err := conn.Read(temp)

		if err != nil {
			break
		}

		data = append(data, temp[:n]...)
		length += n
		if n < buffSize {
			break
		}
	}

	return data, nil
}

func LogInfo(logString string) {
	if verboseLogging {
		log.Println(logString)
	}
}

func findRoute(parsedRequest *Request) (handlerFn, string) {
	paths := strings.Split(parsedRequest.route, "/")
	if len(paths) > 2 {
		return routeMap[parsedRequest.Method]["/"], parsedRequest.route
	}
	return routeMap[parsedRequest.Method]["/"], paths[len(paths)-1]
}

func parsePacket(data []byte) UDPPacket {
	pType := data[0]
	seqNo := data[1:5]
	peerAddr := data[5:9]
	peerPort := data[9:11]
	payload := data[11:]

	return UDPPacket{
		pType:    []byte{pType},
		seqNo:    seqNo,
		peerAddr: peerAddr,
		peerPort: peerPort,
		payload:  payload,
	}
}

func getBytesFromPacket(packet UDPPacket) []byte {
	packetBytes := append(packet.pType, packet.seqNo...)
	packetBytes = append(packetBytes, packet.peerAddr...)
	packetBytes = append(packetBytes, packet.peerPort...)
	return packetBytes
}

func MakePacket(pType uint32, seqNo uint32, addr string, port uint16, payload string) UDPPacket {

	// pType, one of the following: 0 - Data, 1- ACK, 2 - SYN, 3 - SYN-ACK, 4 - NAK; 1 byte
	pTypeByte := []byte{byte(pType)}

	// seqNo, for SYN it is the initial pNo during 3WH -- else incremental packet numbers; 4 bytes BE
	seqNoBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(seqNoBytes, seqNo)

	// peerAddr, either sender/receiver -- translated by router!; 4 bytes
	peerAddrBytes := make([]byte, 4)

	peerAddrSplit := strings.Split(addr, ".")
	for i, section := range peerAddrSplit {
		intSection, _ := strconv.Atoi(section)
		peerAddrBytes[i] = byte(intSection)
	}

	// peerPort, either sender/receiver -- translated by router!; 2 bytes BE
	peerPortBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(peerPortBytes, port)

	// payload; max 1013 bytes
	// TODO handle size constraints/breaking somehow...
	payloadBytes := []byte(payload)

	// Packet Size Range: 11 (no payload) to 1024 (full payload)
	return UDPPacket{
		pType:    pTypeByte,
		seqNo:    seqNoBytes,
		peerAddr: peerAddrBytes,
		peerPort: peerPortBytes,
		payload:  payloadBytes,
	}
}

func inNaks(seqNo uint32, naks []uint32) bool {
	for _, nakSeq := range naks {
		if nakSeq == seqNo {
			return true
		}
	}
	return false
}

func inAcks(seqNo uint32, acks []uint32) bool {
	for _, ackSeq := range acks {
		if ackSeq == seqNo {
			return true
		}
	}
	return false
}

func getAddressFromBytes(packet UDPPacket) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		packet.peerAddr[0], packet.peerAddr[1], packet.peerAddr[2], packet.peerAddr[3])
}

func handleConnection(curConn net.Conn) {
	LogInfo(fmt.Sprintf("Handling client %s", curConn.RemoteAddr().String()))
	defer curConn.Close()

	requestData, err := readRequestFromConnection(curConn)
	var response string
	var statusCode int
	var headers string

	if err != nil {
		LogInfo("Read request error!")
	}

	parsedRequest := parseRequestData(string(requestData))
	handler := routeMap[parsedRequest.Method][parsedRequest.route]

	if handler != nil {
		response, statusCode, headers = handler(parsedRequest, nil, &rootDirectory)
	} else {
		handler, pathParam := findRoute(parsedRequest)
		response, statusCode, headers = handler(parsedRequest, &pathParam, &rootDirectory)
	}

	httpResponse := constructStructuredResponse(response, statusCode, headers)
	_, writeErr := curConn.Write([]byte(httpResponse))
	if writeErr != nil {
		LogInfo("Connection write error!")
	}
	LogInfo(fmt.Sprintf("Responded to %s with status code %d", curConn.RemoteAddr().String(), statusCode))
}

func constructStructuredResponse(response string, statusCode int, headers string) string {
	statusLine := fmt.Sprintf("HTTP/1.0 %d %s %s", statusCode, reasonPhrase[statusCode], CRLF)
	return fmt.Sprintf("%s%s%s%s", statusLine, headers, CRLF+CRLF, response)
}

func parseRequestData(request string) *Request {
	initialSplit := strings.SplitN(request, CRLF+CRLF, 2)
	requestLines := strings.Split(initialSplit[0], CRLF)
	cleanedRequestLines := []string{}
	parsedRequest := Request{}

	for _, line := range requestLines {
		if line != blankString {
			cleanedRequestLines = append(cleanedRequestLines, line)
		}
	}

	firstReqLine := strings.Split(cleanedRequestLines[0], " ")
	parsedRequest.route = firstReqLine[1]

	if strings.Contains(cleanedRequestLines[0], "POST") {
		parsedRequest.Method = "POST"
		headers := strings.Join(cleanedRequestLines[1:len(cleanedRequestLines)], CRLF)
		parsedRequest.headers = &headers
		parsedRequest.Body = &initialSplit[1]
	} else {
		parsedRequest.Method = "GET"
		if len(cleanedRequestLines) > 1 {
			headers := strings.Join(cleanedRequestLines[1:len(cleanedRequestLines)-1], CRLF)
			parsedRequest.headers = &headers
		}
	}

	return &parsedRequest
}

func RegisterHandler(method string, route string, handler handlerFn) {
	if routeMap[method] == nil {
		routeMap[method] = map[string]handlerFn{}
	}
	routeMap[method][route] = handler
}

func StartUDPServer(port string, directory string, verbose bool) {
	portInt, _ := strconv.Atoi(port)
	serverIP, _, _ := net.ParseCIDR("127.0.0.1/8")
	serverAddr := net.UDPAddr{
		IP:   serverIP,
		Port: portInt,
	}
	udpConn, err := net.ListenUDP("udp", &serverAddr)

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		log.Fatal("Directory not found.")
	}

	rootDirectory = directory
	verboseLogging = verbose

	if err != nil {
		fmt.Println(err)
		return
	}

	defer udpConn.Close()
	clients := sync.Map{}
	for {
		buffer := make([]byte, 1024)
		n, addr, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}

		packet := parsePacket(buffer[:n])
		hostAddr := getAddressFromBytes(packet)
		clientPackets, loaded := clients.LoadOrStore(hostAddr, make(chan UDPPacket))

		if !loaded {
			go func() {
				// add a timer that resets everytime a packet comes in
				// but times out when it crosses a threshold
				for {
					timeout := 1 * time.Second
					deadline := time.Now().Add(timeout)
					wTimeoutErr := udpConn.SetWriteDeadline(deadline)
					//rTimeoutErr := udpConn.SetReadDeadline(deadline)
					//if rTimeoutErr != nil || wTimeoutErr != nil {
					if wTimeoutErr != nil {
						fmt.Println("wTimeoutErr!")
					}

					var expectedSeqNo uint32
					expectedSeqNo = 4
					acks := make([]uint32, 5)
					naks := make([]uint32, 5)
					httpPayload := make([]string, 1024)

					for {
						for packet := range clientPackets.(chan UDPPacket) {
							receivedSeqNo := binary.BigEndian.Uint32(packet.seqNo)
							if packet.pType[0] == 0 {
								if inAcks(receivedSeqNo, acks) {
									continue
								}
								if receivedSeqNo == expectedSeqNo {
									acks = append(acks, receivedSeqNo)
									// SEND ACK
									ackPacket := MakePacket(1, receivedSeqNo, hostAddr, binary.BigEndian.Uint16(packet.peerPort), "")
									packetBytes := getBytesFromPacket(ackPacket)
									_, writeErr := udpConn.WriteToUDP(packetBytes, addr)
									if writeErr != nil {
										LogInfo("Timeout packet 0!")
									}
									// STORE payload in proper structure
									httpPayload[int(receivedSeqNo)] = string(packet.payload)
									LogInfo(fmt.Sprintf("ACK'd packet %d", receivedSeqNo))
									expectedSeqNo += 1
									break
								} else if receivedSeqNo < expectedSeqNo {
									// retransmitted packet from client
									// SEND ACK
									ackPacket := MakePacket(1, receivedSeqNo, hostAddr, binary.BigEndian.Uint16(packet.peerPort), "")
									packetBytes := getBytesFromPacket(ackPacket)
									_, writeErr := udpConn.WriteToUDP(packetBytes, addr)
									if writeErr != nil {
										LogInfo("Timeout for retransmitted!")
									}
									LogInfo(fmt.Sprintf("ACK'd packet %d", receivedSeqNo))
									// CHECK IF IN naks
									if inNaks(receivedSeqNo, naks) {
										// STORE payload in proper structure
										httpPayload[int(receivedSeqNo)] = string(packet.payload)
									}
									// else DISCARD PACKET
								} else {
									// SEND ACK
									ackPacket := MakePacket(1, receivedSeqNo, hostAddr, binary.BigEndian.Uint16(packet.peerPort), "")
									packetBytes := getBytesFromPacket(ackPacket)
									_, writeErr := udpConn.WriteToUDP(packetBytes, addr)
									if writeErr != nil {
										LogInfo("Timeout for higher seqNo!")
									}
									LogInfo(fmt.Sprintf("ACK'd packet %d", receivedSeqNo))
									for packetNum := expectedSeqNo; packetNum < receivedSeqNo; packetNum++ {
										naks = append(naks, packetNum)
										nakPacket := MakePacket(4, packetNum, hostAddr, binary.BigEndian.Uint16(packet.peerPort), "")
										packetBytes := getBytesFromPacket(nakPacket)
										_, writeErr := udpConn.WriteToUDP(packetBytes, addr)
										if writeErr != nil {
											LogInfo("Timeout writing NAKs!")
										}
										LogInfo(fmt.Sprintf("NAK'd packet %d", packetNum))
									}
									expectedSeqNo = receivedSeqNo + 1
								}
							}
							handleHandshakePacket(packet, addr, udpConn)
						}
					}
				}
			}()
		}

		clientPackets.(chan UDPPacket) <- packet
	}
}

func handleHandshakePacket(packet UDPPacket, addr *net.UDPAddr, conn *net.UDPConn) {
	hostAddr := getAddressFromBytes(packet)
	if packet.pType[0] == 2 {
		// SYN
		receivedSeq := binary.BigEndian.Uint32(packet.seqNo)
		synAck := MakePacket(3, receivedSeq+1, hostAddr, binary.BigEndian.Uint16(packet.peerPort), "")
		packetBytes := getBytesFromPacket(synAck)
		for {
			timeout := 2 * time.Second
			deadline := time.Now().Add(timeout)
			_ = conn.SetWriteDeadline(deadline)

			_, writeErr := conn.WriteToUDP(packetBytes, addr)
			if writeErr != nil {
				LogInfo("Timeout handshaking!")
				continue
			}
			break
		}
		LogInfo(fmt.Sprintf("Got SYN packet: %d", receivedSeq))
		return
	} else if packet.pType[0] == 1 {
		// ACK
		receivedSeq := binary.BigEndian.Uint32(packet.seqNo)
		LogInfo(fmt.Sprintf("Received ACK: %d", receivedSeq))
		return
	}
}

func StartServer(port string, directory string, verbose bool) {
	listener, err := net.Listen("tcp", port)

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		log.Fatal("Directory not found.")
	}

	rootDirectory = directory
	verboseLogging = verbose

	if err != nil {
		fmt.Println(err)
		return
	}

	defer listener.Close()

	for {
		curConn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(curConn)
	}
}
