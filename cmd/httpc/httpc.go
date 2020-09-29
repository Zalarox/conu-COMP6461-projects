package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"httpc/pkg/libhttpc"
	"os"
	"strings"
)

type flagList []string

// implements interface
func (flags *flagList) String() string {
	return strings.Join(*flags, ", ")
}

func (flags *flagList) Set(value string) error {
	*flags = append(*flags, strings.TrimSpace(value))
	return nil
}

func parseArgs() {
	cmdHelp := flag.NewFlagSet("help", flag.ExitOnError)
	cmdHttpc := flag.NewFlagSet("httpc", flag.ExitOnError)

	var headerPtr flagList

	verbosePtr := cmdHttpc.Bool("v", true, libhttpc.HelpTextVerbose)
	dataPtr := cmdHttpc.String("d", libhttpc.BlankString, libhttpc.HelpTextData)
	filePtr := cmdHttpc.String("f", libhttpc.BlankString, libhttpc.HelpTextFile)
	cmdHttpc.Var(&headerPtr, "h", libhttpc.HelpTextHeader)

	if len(os.Args) == 1 {
		fmt.Println(libhttpc.HelpTextMain)
		return
	}

	switch strings.ToLower(os.Args[1]) {
	case "help":
		_ = cmdHelp.Parse(os.Args[2:])
		helpFor := cmdHelp.Args()

		if len(helpFor) == 0 {
			fmt.Println(libhttpc.HelpTextMain)
		} else {
			if strings.ToLower(helpFor[0]) == "get" {
				fmt.Println(libhttpc.HelpTextGet)
			} else if strings.ToLower(helpFor[0]) == "post" {
				fmt.Println(libhttpc.HelpTextPost)
			} else {
				fmt.Println(libhttpc.HelpTextMain)
			}
		}
	default:

		fmt.Println(cmdHttpc.Args())

		method := os.Args[1]
		if strings.ToLower(method) == "get" {
			fmt.Println("Got GET")
		} else if strings.ToLower(method) == "post" {
			fmt.Println("Got POST")
		} else {
			// error
			fmt.Println("Got INVALID")
		}

		_ = cmdHttpc.Parse(os.Args[2:])

		fmt.Println("verbose:", *verbosePtr)
		fmt.Println("header:", headerPtr)
		fmt.Println("data:", *dataPtr)
		fmt.Println("file:", *filePtr)
		fmt.Println("tail:", cmdHttpc.Args())
	}
}

func testProgram() {
	// 	Sample Headers
	sampleHeaders := libhttpc.RequestHeader{
		"Content-Type":  "application/json",
		"Authorization": "None",
	}

	fmt.Println("Making GET request:")
	resp, err := libhttpc.Get("https://httpbin.org/get", sampleHeaders)
	response, err := libhttpc.FromString(resp)
	if response != nil {
		fmt.Println("GET response:")
		fmt.Println(response.Body)
	}

	// Sample Body
	sampleBody := map[string]string{
		"Luke Skywalker": "Peacekeeper",
	}

	reqBody, _ := json.Marshal(sampleBody)

	fmt.Println("Making POST request:")
	resp, err = libhttpc.Post("https://httpbin.org/post", sampleHeaders, reqBody)
	if err != nil {
		fmt.Println(err)
	}

	response, err = libhttpc.FromString(resp)
	if response != nil {
		fmt.Println("POST response:")
		fmt.Println(response.Body)
	}
}

func main() {
	parseArgs()
	//testProgram()
}
