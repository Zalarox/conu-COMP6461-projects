package libhttpc

type RequestHeader map[string]string

type Response struct {
	StatusCode int
	Protocol   string
	Headers    string
	Body       string
}

type UDPPacket struct {
	pType    []byte
	seqNo    []byte
	peerAddr []byte
	peerPort []byte
	payload  []byte
}

const ProtocolVersion = "HTTP/1.0"

const CRLF = "\r\n"

const BlankString = ""

const HelpTextMain = `httpc is a curl-like application but supports HTTP protocol only.

Usage:
httpc command [arguments]

The commands are:

get executes a HTTP GET request and prints the response.

post executes a HTTP POST request and prints the response.

help prints this screen.

Use "httpc help [command]" for more information about a command.`

const HelpTextGet = `usage: httpc get [-v] [-h key:value] URL

Get executes a HTTP GET request for a given URL.
 -v Prints the detail of the response such as protocol, status, and headers.
 -h key:value Associates headers to HTTP Request with the format 'key:value'.`

const HelpTextPost = `usage: httpc post [-v] [-h key:value] [-d inline-data] [-f file] URL

Post executes a HTTP POST request for a given URL with inline data or from file.
 -v Prints the detail of the response such as protocol, status, and headers.
 -h key:value Associates headers to HTTP Request with the format 'key:value'.
 -d string Associates an inline data to the body HTTP POST request.
 -f file Associates the content of a file to the body HTTP POST request.
 -o Writes the response out to a file.

Either [-d] or [-f] can be used but not both.`

const HelpTextVerbose = `Prints the detail of the response such as protocol, status, and headers.`

const HelpTextData = `Associates an inline data to the body HTTP POST request.`

const HelpTextFile = `Associates the content of a file to the body HTTP POST request.`

const HelpTextHeader = `Associates headers to HTTP Request with the format 'key:value'.`

const HelpTextOutput = `Writes the response of the HTTP request to a file.`

const DefaultRedirectURI = "http://google.com"

const RouterAddr = "127.0.0.1"

const RouterPort = "3000"
