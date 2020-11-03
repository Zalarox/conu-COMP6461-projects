package libhttpserver

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
)

//
//func dropCR(data []byte) []byte {
//	if len(data) > 0 && data[len(data)-1] == '\r' {
//		return data[0 : len(data)-1]
//	}
//	return data
//}

func ScanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\r', '\n'}); i >= 0 {
		// We have a full newline-terminated line.
		return i + 2, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func handleConnection(curConn net.Conn) string {
	fmt.Printf("Handling client %s\n", curConn.RemoteAddr().String())
	defer curConn.Close()

	reader := bufio.NewReader(curConn)
	scanner := bufio.NewScanner(reader)
	scanner.Split(ScanCRLF)
	responseData := ""

	for scanner.Scan() {
		responseData += scanner.Text()
	}

	fmt.Println(responseData)

	return responseData
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
