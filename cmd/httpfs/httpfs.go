package main

import (
	"flag"
	"fmt"
	"httpc/pkg/libhttpserver"
)

func myFunc(reqData *libhttpserver.Request) string {
	// structure as a valid HTTP response
	return "My Response!"
}

func parseArgs() {
	verbosePtr := flag.Bool("v", false, libhttpserver.HelpTextVerbose)
	dirPtr := flag.String("d", "/", libhttpserver.HelpTextDir)
	portPtr := flag.String("p", "8080", libhttpserver.HelpTextPort)

	flag.Parse()
	fmt.Println("v:", *verbosePtr)
	fmt.Println("d:", *dirPtr)
	fmt.Println("port:", *portPtr)

	PORT := ":" + *portPtr
	fmt.Println("PORT:", PORT)

	libhttpserver.RegisterHandler("POST", "/", myFunc)
	libhttpserver.StartServer(PORT, *dirPtr, *verbosePtr)
}

func main() {
	parseArgs()
}
