package libhttpserver

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
)

func Read(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	var buffer bytes.Buffer
	for {
		ba, _, err := reader.ReadLine()
		if err != nil {
			// if the error is an End Of File this is still good
			if err == io.EOF {
				break
			}
			return "", err
		}
		buffer.Write(ba)
		//if !isPrefix {
		//	break
		//}
	}
	return buffer.String(), nil
}

func handleConnection(curConn net.Conn) {
	fmt.Printf("Handling client %s\n", curConn.RemoteAddr().String())

	data, err := Read(curConn)
	//connData, err := bufio.NewReader(curConn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(data)

	//temp := strings.TrimSpace(string(connData))

	//if temp == "STOP" {
	//	break
	//}

	//result := strconv.Itoa(random()) + "\n"
	//curConn.Write([]byte(string(result)))
	curConn.Close()
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
