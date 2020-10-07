package main

import (
	"flag"
	"fmt"
	"httpc/pkg/libhttpc"
	"io/ioutil"
	"os"
	"regexp"
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

func writeOutput(outputPtr *string, toWrite []byte) {
	if *outputPtr != "" {
		err := ioutil.WriteFile(*outputPtr, toWrite, os.FileMode(os.O_RDWR))
		if err == nil {
			fmt.Printf("Successfully written result to %s\n", *outputPtr)
		} else {
			fmt.Printf("Error encountered: %s", err.Error())
		}
	} else {
		fmt.Println(string(toWrite))
	}
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
				match, _ := regexp.MatchString("^http(s?)://", url)
				if match == false {
					url = "https://" + url
				}
			} else {
				fmt.Println(libhttpc.HelpTextGet)
				return
			}

			res, getErr := libhttpc.Get(url, headers)

			if getErr != nil {
				writeOutput(outputPtr, []byte(getErr.Error()))
				return
			}
			response, parsingErr := libhttpc.FromString(res)
			if parsingErr != nil {
				writeOutput(outputPtr, []byte(parsingErr.Error()))
				return
			}

			if response.StatusCode >= 300 && response.StatusCode <= 302 {
				resString, redirectErr := libhttpc.HandleRedirects(response, res, headers, 0)
				if redirectErr != nil {
					writeOutput(outputPtr, []byte(redirectErr.Error()))
					return
				}
				res = resString
				responseAfterRedirect, parsingErr := libhttpc.FromString(res)
				if parsingErr != nil {
					writeOutput(outputPtr, []byte(parsingErr.Error()))
					return
				}
				response = responseAfterRedirect
			}

			if *verbosePtr {
				writeOutput(outputPtr, []byte(res))
				return
			}

			writeOutput(outputPtr, []byte(response.Body))

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
				match, _ := regexp.MatchString("^http(s?)://", url)
				if match == false {
					url = "https://" + url
				}
			} else {
				fmt.Println(libhttpc.HelpTextPost)
				return
			}

			res, postErr := libhttpc.Post(url, headers, requestBody)
			if *verbosePtr {
				writeOutput(outputPtr, []byte(res))
				return
			}

			if postErr != nil {
				writeOutput(outputPtr, []byte(postErr.Error()))
				return
			}
			response, parsingErr := libhttpc.FromString(res)
			if parsingErr != nil {
				writeOutput(outputPtr, []byte(parsingErr.Error()))
				return
			}

			writeOutput(outputPtr, []byte(response.Body))
		} else {
			// error
			fmt.Println(libhttpc.HelpTextMain)
		}
	}
}

func main() {
	parseArgs()
}
