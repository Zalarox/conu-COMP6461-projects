package libhttpserver

import (
	"fmt"
	"log"
	"net"
)

func readResponseFromConnection(conn net.Conn) ([]byte, error) {
	temp := make([]byte, 1024)
	data := make([]byte, 0)
	length := 0

	for {
		n, err := conn.Read(temp)
		if err != nil {
			break
		}

		data = append(data, temp[:n]...)
		length += n
		if n < 1024 {
			break
		}
	}

	return data, nil
}

func handleConnection(curConn net.Conn) {
	fmt.Printf("Handling client %s\n", curConn.RemoteAddr().String())
	defer curConn.Close()

	responseData, err := readResponseFromConnection(curConn)

	if err != nil {
		log.Fatalln(err)
	}

	_, writeErr := curConn.Write(responseData)
	if writeErr != nil {
		log.Fatalln(writeErr)
	}
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
