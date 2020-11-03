package libhttpserver

import (
	"fmt"
	"log"
	"net"
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

func handleConnection(curConn net.Conn) {
	fmt.Printf("Handling client %s\n", curConn.RemoteAddr().String())
	defer curConn.Close()

	requestData, err := readRequestFromConnection(curConn)
	var response string

	if err != nil {
		log.Fatalln(err)
	}

	parsedRequest := parseRequestData(string(requestData))
	handler := routeMap[parsedRequest.method][parsedRequest.route]

	if handler != nil {
		response = handler(parsedRequest)
	} else {
		log.Fatalln("No Registered Handler!")
		return
	}

	_, writeErr := curConn.Write([]byte(response))
	if writeErr != nil {
		log.Fatalln(writeErr)
	}
}

func parseRequestData(request string) *Request {
	requestLines := strings.Split(request, CRLF)
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
		parsedRequest.method = "POST"
		parsedRequest.body = &cleanedRequestLines[len(cleanedRequestLines)-1]
	} else {
		parsedRequest.method = "GET"
	}

	return &parsedRequest
}

func RegisterHandler(method string, route string, handler handlerFn) {
	if routeMap[method] == nil {
		routeMap[method] = map[string]handlerFn{}
	}
	routeMap[method][route] = handler
}

func StartServer(port string, directory string, verbose bool) {
	listener, err := net.Listen("tcp", port)

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
