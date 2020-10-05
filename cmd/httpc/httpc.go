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

	verbosePtr := cmdHttpc.Bool("v", false, libhttpc.HelpTextVerbose)
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
		_ = cmdHttpc.Parse(os.Args[2:])
		headers := map[string]string{}
		url := ""
		tail := cmdHttpc.Args()

		method := os.Args[1]
		if strings.ToLower(method) == "get" {
			for _, headerString := range headerPtr {
				headerSet := strings.Split(headerString, ":")
				headers[headerSet[0]] = headerSet[1]
			}
			//fmt.Println("file:", *filePtr)

			url := ""
			if len(tail) != 0 {
				url = cmdHttpc.Args()[0]
			} else {
				fmt.Println(libhttpc.HelpTextGet)
			}

			res, _ := libhttpc.Get(url, headers, *verbosePtr)
			response, _ := libhttpc.FromString(res)
			fmt.Println(response.Body)

		} else if strings.ToLower(method) == "post" {
			fmt.Println("file:", *filePtr)

			if len(tail) != 0 {
				url = cmdHttpc.Args()[0]
			} else {
				fmt.Println(libhttpc.HelpTextPost)
			}

			res, _ := libhttpc.Post(url, headers, []byte(*dataPtr), *verbosePtr)
			response, _ := libhttpc.FromString(res)
			fmt.Println(response.Body)

		} else {
			// error
			fmt.Println("Got INVALID")
		}
	}
}

func testProgram() {
	// 	Sample Headers
	sampleHeaders := libhttpc.RequestHeader{
		"Content-Type":  "application/json",
		"Authorization": "None",
	}

	fmt.Println("Making GET request:")
	resp, err := libhttpc.Get("https://httpbin.org/get", sampleHeaders, true)
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
	resp, err = libhttpc.Post("https://httpbin.org/post", sampleHeaders, reqBody, true)
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
