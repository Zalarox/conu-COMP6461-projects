package libhttpserver

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
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

func inNaks(seqNo uint32, receiver *Receiver) bool {
	for _, nakSeq := range receiver.naks {
		if nakSeq == seqNo {
			return true
		}
	}
	return false
}

func handleUDPConnection(reqData []byte, addr *net.UDPAddr, conn *net.UDPConn) {
	packet := parsePacket(reqData)
	hostAddr := getAddressFromBytes(packet)
	if packet.pType[0] == 2 {
		// SYN
		receivedSeq := binary.BigEndian.Uint32(packet.seqNo)
		synAck := MakePacket(3, receivedSeq+1, hostAddr, binary.BigEndian.Uint16(packet.peerPort), "")
		packetBytes := getBytesFromPacket(synAck)
		_, writeErr := conn.WriteToUDP(packetBytes, addr)
		if writeErr != nil {
			log.Fatalln(writeErr)
		}
		LogInfo(fmt.Sprintf("Got SYN packet: %d", receivedSeq))
		return
	} else if packet.pType[0] == 1 {
		// ACK
		receivedSeq := binary.BigEndian.Uint32(packet.seqNo)
		LogInfo(fmt.Sprintf("Received ACK: %d", receivedSeq))
		// modify the window
		return
	} else {
		// put into receiver window for client_id
		clientId := fmt.Sprintf("%s:%d", hostAddr, binary.BigEndian.Uint16(packet.peerPort))
		clientReceiverInterface, _ := clients.Load(clientId)
		lastReceivedSeqNo := binary.BigEndian.Uint32(packet.seqNo)
		if clientReceiverInterface == nil {
			clients.Store(clientId, Receiver{
				lastReceivedPacketNum: lastReceivedSeqNo,
				expectedPacketNum:     lastReceivedSeqNo + 1,
				receiverWindow:        []UDPPacket{},
			})
		} else {
			clientReceiver := clientReceiverInterface.(Receiver)
			if clientReceiver.expectedPacketNum == lastReceivedSeqNo {
				// happy path
				setLastReceivedPacketNum(lastReceivedSeqNo, &clientReceiver)
				setExpectedPacketNum(lastReceivedSeqNo, &clientReceiver)
				appendToReceiverWindow(&packet, &clientReceiver)
				// send ACK
				ackPacket := MakePacket(2, lastReceivedSeqNo, hostAddr, binary.BigEndian.Uint16(packet.peerPort), "")
				packetBytes := getBytesFromPacket(ackPacket)
				_, writeErr := conn.WriteToUDP(packetBytes, addr)
				if writeErr != nil {
					log.Fatalln(writeErr)
				}
			} else {
				// either a retransmitted packet from client
				if lastReceivedSeqNo < clientReceiver.expectedPacketNum && inNaks(lastReceivedSeqNo, &clientReceiver) {
					setExpectedPacketNum(clientReceiver.lastReceivedPacketNum+1, &clientReceiver)
					appendToReceiverWindow(&packet, &clientReceiver)
					removeFromNaks(lastReceivedSeqNo, &clientReceiver)
				} else if lastReceivedSeqNo > clientReceiver.expectedPacketNum {
					// or a packet which is much ahead of the expected seq. no.
					for i := clientReceiver.expectedPacketNum; i < lastReceivedSeqNo; i++ {
						// put in NAKs
						addToNaks(i, &clientReceiver)
						// make NAK pack and send client
						nakPacket := MakePacket(4, i, hostAddr, binary.BigEndian.Uint16(packet.peerPort), "")
						packetBytes := getBytesFromPacket(nakPacket)
						_, writeErr := conn.WriteToUDP(packetBytes, addr)
						if writeErr != nil {
							log.Fatalln(writeErr)
						}
					}
				}
			}
		}
	}

	//var response string
	//var statusCode int
	//var headers string
	//
	//parsedRequest := parseRequestData(string(packet.payload))
	//handler := routeMap[parsedRequest.Method][parsedRequest.route]
	//
	//if handler != nil {
	//	response, statusCode, headers = handler(parsedRequest, nil, &rootDirectory)
	//} else {
	//	handler, pathParam := findRoute(parsedRequest)
	//	response, statusCode, headers = handler(parsedRequest, &pathParam, &rootDirectory)
	//}
	//
	//httpResponse := constructStructuredResponse(response, statusCode, headers)
	//responsePacket := MakePacket(0, 1, hostAddr, binary.BigEndian.Uint16(packet.peerPort), httpResponse)
	//
	//packetBytes := getBytesFromPacket(responsePacket)
	//
	//n, writeErr := conn.WriteToUDP(packetBytes, addr)
	//if writeErr != nil {
	//	log.Fatalln(writeErr)
	//}
	//LogInfo(fmt.Sprintf("Responded to %s with status code %d, written %d", addr, statusCode, n))
}

func addToNaks(nakSeq uint32, clientReceiver *Receiver) {
	clientReceiver.naks = append(clientReceiver.naks, nakSeq)
}

func removeFromNaks(nakToRemove uint32, clientReceiver *Receiver) {
	naks := clientReceiver.naks
	for i, nakSeq := range naks {
		if nakToRemove == nakSeq {
			naks[len(naks)-1], naks[i] = naks[i], naks[len(naks)-1]
			naks = naks[:len(naks)-1]
			break
		}
	}
}

func setLastReceivedPacketNum(seqNo uint32, clientReceiver *Receiver) {
	clientReceiver.lastReceivedPacketNum = seqNo
}

func setExpectedPacketNum(seqNo uint32, clientReceiver *Receiver) {
	clientReceiver.expectedPacketNum = seqNo + 1
}

func appendToReceiverWindow(receivedPacket *UDPPacket, clientReceiver *Receiver) {
	clientReceiver.receiverWindow = append(clientReceiver.receiverWindow, *receivedPacket)
	// throw in additional logic to handle window overflow
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
		log.Fatalln(err)
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
		log.Fatalln(writeErr)
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

	for {
		buffer := make([]byte, 1024)
		_, addr, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleUDPConnection(buffer, addr, udpConn)
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
