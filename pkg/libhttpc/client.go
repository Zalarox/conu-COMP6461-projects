package libhttpc

import (
	"fmt"
	"io"
	"net"
	"net/url"
)

const ProtocolVersion = "HTTP/1.0"
const CRLF = "\r\n"
const BlankString = ""

type Header map[string]string

type Response struct {
	Status        string
	StatusCode    int
	Proto         string
	ProtoMajor    int
	ProtoMinor    int
	Header        Header
	Body          io.ReadCloser
	ContentLength int64
}

func Get(inputUrl string, headers Header) (string, error) {
	parsedURL, parsedHeaders, conn, err := connectHandler(inputUrl, headers)
	defer conn.Close()

	if err != nil {
		return BlankString, err
	}

	requestString := fmt.Sprintf("GET %s %s%s%s%s%s", parsedURL.RequestURI(), ProtocolVersion, CRLF, parsedHeaders, CRLF, CRLF)
	fmt.Fprintf(conn, requestString)

	response, err := readResponseFromConnection(conn)

	if err != nil {
		return "", nil
	}

	return string(response), nil
}

func Post(inputUrl string, headers Header, body []byte) (string, error) {
	headers["Content-Length"] = fmt.Sprintf("%d", len(body))
	parsedURL, parsedHeaders, conn, err := connectHandler(inputUrl, headers)
	defer conn.Close()

	if err != nil {
		return BlankString, err
	}

	requestString := fmt.Sprintf("POST %s %s%s%s%s%s%s%s", parsedURL.RequestURI(), ProtocolVersion, CRLF, parsedHeaders, CRLF, body, CRLF)
	fmt.Fprintf(conn, requestString)

	response, err := readResponseFromConnection(conn)

	if err != nil {
		return "", nil
	}

	return string(response), nil
}

func readResponseFromConnection(conn net.Conn) ([]byte, error) {
	temp := make([]byte, 1024)
	data := make([]byte, 0)
	length := 0

	for {
		n, err := conn.Read(temp)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		data = append(data, temp[:n]...)
		length += n
	}

	return data, nil
}

func connectHandler(inputUrl string, headers Header) (*url.URL, string, net.Conn, error) {
	parsedURL, error := url.Parse(inputUrl)
	parsedHeaders := stringifyHeaders(headers)

	if error != nil {
		return nil, "", nil, error
	}

	port := parsedURL.Port()
	if port == "" {
		port = "80"
	}

	host := fmt.Sprintf("%s:%s", parsedURL.Hostname(), port)

	conn, err := net.Dial("tcp", host)
	return parsedURL, parsedHeaders, conn, err
}

func stringifyHeaders(headers Header) string {
	headersString := ""
	for headerKey, headerValue := range headers {
		headersString += fmt.Sprintf("%s:%s%s", headerKey, headerValue, CRLF)
	}
	return headersString
}
