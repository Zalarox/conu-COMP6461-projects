package libhttpserver

const CRLF = "\r\n"

const HelpTextVerbose = `Prints debugging messages.`

const HelpTextDir = `Specifies the directory that the server will use to read/write requested files. 
Default is the current directory when launching the application.`

const HelpTextPort = `Specifies the port number that the server will listen and serve at. Default is 8080.`

const buffSize = 1024
const blankString = ""

type handlerFn func(request *Request) string

type Request struct {
	method string
	route  string
	body   *string
}

var routeMap = map[string]map[string]handlerFn{}
