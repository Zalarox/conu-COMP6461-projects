package main

import (
	"flag"
	"fmt"
	"httpc/pkg/libhttpserver"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var fileMutex sync.Mutex

func listFiles(path string) []string {
	allFiles, err := ioutil.ReadDir(path)
	files := []string{}
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range allFiles {
		if !file.IsDir() {
			files = append(files, file.Name())
		}
	}

	return files
}

func makeHeaders(responseBody string, responseHeaders []string) string {
	bodyLength := fmt.Sprintf("Content-Length:%d", len(responseBody))
	responseHeaders = append(responseHeaders, bodyLength)
	responseHeaders = append(responseHeaders, "Content-Disposition:inline")
	headers := strings.Join(responseHeaders, libhttpserver.CRLF)
	return headers
}

func getTypeHeader(fileType string) string {
	key := "Content-Type:"
	switch fileType {
	case ".txt":
		return key + "text/plain"
	case ".json":
		return key + "application/json"
	case ".xml":
		return key + "application/xml"
	case ".html":
		return key + "text/html"
	}
	return key + "text/plain"
}

func getHandler(reqData *libhttpserver.Request, pathParam *string, root *string) (string, int, string) {

	fileMutex.Lock() // LOCK
	libhttpserver.LogInfo("Acquired Lock")
	defer func() {
		libhttpserver.LogInfo("Released Lock")
		fileMutex.Unlock()
	}() // UNLOCK ON RETURN

	if reqData.Method == "GET" {
		if pathParam == nil {
			libhttpserver.LogInfo("Responding to request for directory listing.")
			files := listFiles(*root)
			body := strings.Join(files, ",")
			responseHeaders := makeHeaders(body, []string{})
			return body, 200, responseHeaders
		}

		if strings.Contains(*pathParam, "/") {
			errStr := fmt.Sprintf("Access Forbidden: '%s' is outside server root directory", *pathParam)
			libhttpserver.LogInfo("Access Denied to request.")
			return errStr, 403, makeHeaders(errStr, []string{})
		}

		dat, err := ioutil.ReadFile(filepath.Join(*root, *pathParam))
		stringDat := string(dat)
		stringDat = strings.ReplaceAll(stringDat, "\r\n", "\n")
		getHeaders := makeHeaders(stringDat, []string{})
		ext := filepath.Ext(*pathParam)
		typeHeader := getTypeHeader(ext)
		getHeaders = getHeaders + libhttpserver.CRLF + typeHeader

		if err != nil {
			errStr := fmt.Sprintf("No file exists with name '%s'", *pathParam)
			return errStr, 404, makeHeaders(errStr, []string{})
		}
		libhttpserver.LogInfo("Returning requested file " + *pathParam)
		return stringDat, 200, getHeaders
	} else if reqData.Method == "POST" {
		if strings.Contains(*pathParam, "/") {
			errStr := fmt.Sprintf("Access Forbidden: '%s' is outside server root directory", *pathParam)
			libhttpserver.LogInfo("Access Denied to request.")
			return errStr, 403, makeHeaders(errStr, []string{})
		}
		err := ioutil.WriteFile(filepath.Join(*root, *pathParam), []byte(*reqData.Body), 0644)
		if err != nil {
			errStr := fmt.Sprintf("Failed to write to file '%s'", *pathParam)
			return errStr, 500, makeHeaders(errStr, []string{})
		} else {
			successStr := "Successfully written content to file"
			libhttpserver.LogInfo("Successfully written out " + *pathParam)
			return successStr, 200, makeHeaders(successStr, []string{})
		}
	}
	return "", 500, makeHeaders("", []string{})
}

func parseArgs() {
	currDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	verbosePtr := flag.Bool("v", false, libhttpserver.HelpTextVerbose)
	dirPtr := flag.String("d", currDir, libhttpserver.HelpTextDir)
	portPtr := flag.String("p", "8080", libhttpserver.HelpTextPort)

	flag.Parse()
	fmt.Printf("Server listening on port: %s\nDirectory Served: %s\nVerbose Logging:%t\n\n", *portPtr, *dirPtr, *verbosePtr)

	PORT := ":" + *portPtr

	libhttpserver.RegisterHandler("POST", "/", getHandler)
	libhttpserver.RegisterHandler("GET", "/", getHandler)
	libhttpserver.StartServer(PORT, *dirPtr, *verbosePtr)
}

func main() {
	parseArgs()
}
