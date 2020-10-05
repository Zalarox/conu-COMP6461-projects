package libhttpc

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func Get(inputUrl string, headers RequestHeader, isVerbose bool) (string, error) {
	parsedURL, parsedHeaders, conn, err := connectHandler(inputUrl, headers)

	if err != nil {
		return BlankString, err
	}

	defer conn.Close()

	requestString := fmt.Sprintf(
		"GET %s %s%s%s%s%s",
		parsedURL.RequestURI(), ProtocolVersion, CRLF,
		parsedHeaders, CRLF, CRLF)

	fmt.Fprintf(conn, requestString)

	if isVerbose {
		fmt.Println(requestString)
	}

	response, err := readResponseFromConnection(conn)

	if err != nil {
		return BlankString, nil
	}

	return string(response), nil
}

func Post(inputUrl string, headers RequestHeader, body []byte, isVerbose bool) (string, error) {
	headers["Content-Length"] = fmt.Sprintf("%d", len(body))
	parsedURL, parsedHeaders, conn, err := connectHandler(inputUrl, headers)

	if err != nil {
		return BlankString, err
	}

	defer conn.Close()

	requestString := fmt.Sprintf("POST %s %s%s%s%s%s%s",
		parsedURL.RequestURI(), ProtocolVersion, CRLF,
		parsedHeaders, CRLF, body, CRLF)
	fmt.Fprintf(conn, requestString)

	if isVerbose {
		fmt.Println(requestString)
	}

	response, err := readResponseFromConnection(conn)

	if err != nil {
		return BlankString, err
	}

	return string(response), nil
}

func FromString(response string) (*Response, error) {
	responseSplit := strings.Split(response, CRLF+CRLF)
	// splits between (statusLine + headers) and Body
	if len(responseSplit) == 2 {
		response := Response{}
		preBody := responseSplit[0]
		body := responseSplit[1]

		preBodySplit := strings.Split(preBody, "\n")
		if strings.HasPrefix(preBodySplit[0], "HTTP") {
			statusLineSplit := strings.Split(preBodySplit[0], " ")
			response.Protocol = statusLineSplit[0]

			statusCode, err := parseStatusCode(statusLineSplit[1])

			if err != nil {
				return nil, err
			}

			response.StatusCode = statusCode
		}

		response.Headers = strings.Join(preBodySplit[1:], "\n")

		response.Body = body

		return &response, nil
	}
	return nil, nil
}

func parseStatusCode(statusCode string) (int, error) {
	code, err := strconv.Atoi(statusCode)
	if err != nil {
		return -1, err
	}
	return code, nil
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

func connectHandler(inputUrl string, headers RequestHeader) (*url.URL, string, net.Conn, error) {
	parsedURL, urlErr := url.Parse(inputUrl)
	parsedHeaders := stringifyHeaders(headers)

	if urlErr != nil {
		return nil, BlankString, nil, urlErr
	}

	port := parsedURL.Port()
	if port == BlankString {
		port = "80"
	}

	host := fmt.Sprintf("%s:%s", parsedURL.Hostname(), port)

	conn, err := net.Dial("tcp", host)
	return parsedURL, parsedHeaders, conn, err
}

func stringifyHeaders(headers RequestHeader) string {
	headersString := BlankString
	for headerKey, headerValue := range headers {
		headersString += fmt.Sprintf("%s:%s%s", headerKey, headerValue, CRLF)
	}
	return headersString
}
