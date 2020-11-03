package libhttpserver

const CRLF = "\r\n"

const HelpTextVerbose = `Prints debugging messages.`

const HelpTextDir = `Specifies the directory that the server will use to read/write requested files. 
Default is the current directory when launching the application.`

const HelpTextPort = `Specifies the port number that the server will listen and serve at. Default is 8080.`

const buffSize = 1024
const blankString = ""

type handlerFn func(request *Request, pathParam *string, root *string) (string, int, string)

var rootDirectory string

var reasonPhrase = map[int]string{
	200: "OK",
	201: "Created",
	202: "Accepted",
	204: "No Content",
	301: "Moved Permanently",
	302: "Moved Temporarily",
	304: "Not Modified",
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
}

type Request struct {
	method  string
	route   string
	headers *string
	body    *string
}

var routeMap = map[string]map[string]handlerFn{}
