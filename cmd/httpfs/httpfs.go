package main

import (
	"flag"
	"fmt"
	"httpc/pkg/libhttpserver"
)

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

	libhttpserver.StartServer(PORT, *dirPtr, *verbosePtr)
}

func main() {
	parseArgs()
}
