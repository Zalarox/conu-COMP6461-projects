package main

import (
	"flag"
	"fmt"
	"httpc/pkg/libhttpc"
	"io/ioutil"
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
	outputPtr := cmdHttpc.String("o", libhttpc.BlankString, libhttpc.HelpTextOutput)
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

		_ = cmdHttpc.Parse(os.Args[2:])
		headers := map[string]string{}
		url := ""
		tail := cmdHttpc.Args()
		method := os.Args[1]

		for _, headerString := range headerPtr {
			headerSet := strings.Split(headerString, ":")
			headers[headerSet[0]] = headerSet[1]
		}

		if strings.ToLower(method) == "get" {
			if len(tail) != 0 {
				url = tail[len(tail)-1]
			} else {
				fmt.Println(libhttpc.HelpTextGet)
			}

			res, getErr := libhttpc.Get(url, headers, *verbosePtr)
			if getErr != nil {
				fmt.Println(getErr)
				return
			}
			response, parsingErr := libhttpc.FromString(res)
			if parsingErr != nil {
				fmt.Println(parsingErr)
				return
			}

			if response.StatusCode >= 300 && response.StatusCode <= 302 {
				response = handleRedirects(response, libhttpc.DefaultRedirectURI, headers, *verbosePtr, 0)
			}

			if *outputPtr != "" {
				ioutil.WriteFile(*outputPtr, []byte(response.Body), os.FileMode(os.O_RDWR))
			} else {
				fmt.Println(response.Body)
			}

		} else if strings.ToLower(method) == "post" {
			requestBody := []byte(*dataPtr)

			if *filePtr != "" && len(requestBody) == 0 {
				fileContent, err := ioutil.ReadFile(*filePtr)
				if err != nil {
					fmt.Println(err)
					return
				}
				requestBody = fileContent
			}

			if len(tail) != 0 {
				url = tail[len(tail)-1]
			} else {
				fmt.Println(libhttpc.HelpTextPost)
			}

			res, postErr := libhttpc.Post(url, headers, requestBody, *verbosePtr)
			if postErr != nil {
				fmt.Println(postErr)
				return
			}
			response, parsingErr := libhttpc.FromString(res)
			if parsingErr != nil {
				fmt.Println(parsingErr)
				return
			}

			if *outputPtr != "" {
				ioutil.WriteFile(*outputPtr, []byte(response.Body), os.FileMode(os.O_RDWR))
			} else {
				fmt.Println(response.Body)
			}
		} else {
			// error
			fmt.Println(libhttpc.HelpTextMain)
		}
	}
}

func handleRedirects(response *libhttpc.Response, inputUrl string, headers libhttpc.RequestHeader, isVerbose bool, redirectCount int) *libhttpc.Response {
	if redirectCount < 5 && response.StatusCode >= 300 && response.StatusCode <= 302 {
		res, _ := libhttpc.Get(inputUrl, headers, isVerbose)
		response, _ = libhttpc.FromString(res)
		return handleRedirects(response, libhttpc.DefaultRedirectURI, headers, isVerbose, redirectCount+1)
	} else {
		return response
	}
}

func main() {
	parseArgs()
}
