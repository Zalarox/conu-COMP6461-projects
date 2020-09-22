package libhttpc

import (
	"fmt"
	"golang.org/x/tools/go/ssa/interp/testdata/src/errors"
	"io"
	"net"
	"net/url"
)

const ProtocolVersion = "HTTP/1.0"
const CRLF = "\r\n"

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
	// add more on the fly should there be need
}

type Request struct {
	Method        string
	URL           string
	Proto         string
	Header        Header
	Body          io.ReadCloser
	ContentLength int64
	Host          string
	// add more on the fly should there be need
}

func Get(inputUrl string, headers Header) (*Response, error) {
	parsedURL, error := url.Parse(inputUrl)
	parsedHeaders := stringifyHeaders(headers)

	if error != nil {
		//
	}

	port := parsedURL.Port()
	if port == "" {
		port = "80"
	}

	host := fmt.Sprintf("%s:%s", parsedURL.Hostname(), port)

	conn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("Failed to create a TCP connection")
	}
	recvBuf := make([]byte, 1024)
	requestString := fmt.Sprintf("GET %s %s%s%s%s%s", parsedURL.RequestURI(), ProtocolVersion, CRLF, parsedHeaders, CRLF, CRLF)
	fmt.Fprintf(conn, requestString)
	n, err := conn.Read(recvBuf[:])
	if n > 0 {
		fmt.Println("From server -- " + string(recvBuf))
	}
	return nil, nil
}

func stringifyHeaders(headers Header) string {
	headersString := ""
	for headerKey, headerValue := range headers {
		headersString += fmt.Sprintf("%s:%s%s", headerKey, headerValue, CRLF)
	}
	return headersString
}

func Post(inputUrl string, headers Header, body []byte) (*Response, error) {
	parsedURL, err := url.Parse(inputUrl)
	headers["Content-Length"] = fmt.Sprintf("%d", len(body))
	parsedHeaders := stringifyHeaders(headers)

	if err != nil {
		//
	}

	port := parsedURL.Port()
	if port == "" {
		port = "80"
	}

	host := fmt.Sprintf("%s:%s", parsedURL.Hostname(), port)

	conn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("Failed to create a TCP connection")
	}
	recvBuf := make([]byte, 1024)
	requestString := fmt.Sprintf("POST %s %s%s%s%s%s%s%s", parsedURL.RequestURI(), ProtocolVersion, CRLF, parsedHeaders, CRLF, CRLF, body, CRLF)
	fmt.Println(requestString)
	fmt.Fprintf(conn, requestString)
	n, err := conn.Read(recvBuf[:])
	if n > 0 {
		fmt.Println("From server -- " + string(recvBuf))
	}
	conn.Close()
	return nil, nil
}

func executeRequest() (*Response, error) {
	return nil, nil
}
