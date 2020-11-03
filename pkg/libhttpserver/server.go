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
	var statusCode int
	var headers string

	if err != nil {
		log.Fatalln(err)
	}

	parsedRequest := parseRequestData(string(requestData))
	handler := routeMap[parsedRequest.method][parsedRequest.route]

	if handler != nil {
		response, statusCode, headers = handler(parsedRequest)
	} else {
		log.Fatalln("No Registered Handler!")
		return
	}

	httpResponse := constructStructuredResponse(response, statusCode, headers)

	_, writeErr := curConn.Write([]byte(httpResponse))
	if writeErr != nil {
		log.Fatalln(writeErr)
	}
}

func constructStructuredResponse(response string, statusCode int, headers string) string {
	statusLine := fmt.Sprintf("HTTP/1.0 %d %s %s", statusCode, reasonPhrase[statusCode], CRLF)
	return fmt.Sprintf("%s%s%s%s", statusLine, headers, CRLF, response)
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
		headers := strings.Join(cleanedRequestLines[1:len(cleanedRequestLines)-2], CRLF)
		parsedRequest.headers = &headers
		parsedRequest.body = &cleanedRequestLines[len(cleanedRequestLines)-1]
	} else {
		parsedRequest.method = "GET"
		headers := strings.Join(cleanedRequestLines[1:len(cleanedRequestLines)-1], CRLF)
		parsedRequest.headers = &headers
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
