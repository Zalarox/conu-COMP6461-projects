package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"httpc/pkg/libhttpc"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HTTPClient struct {
	url string
}

func makeGet() {
	resp, err := http.Get("https://httpbin.org/get")
	if err != nil {
		fmt.Println("Something went wrong during the GET request")
		fmt.Println(err)
	}
	extractBody(resp)
}

func makePost(requestMap map[string]string) {
	requestBody, err := json.Marshal(requestMap)
	if err != nil {
		fmt.Println("Something went wrong marshalling JSON")
		fmt.Println(err)
	}

	resp, err := http.Post("https://httpbin.org/post", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Something went wrong during the POST request")
		fmt.Println(err)
	}
	extractBody(resp)
}

func extractBody(resp *http.Response) {
	if resp != nil {
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Something went wrong parsing response body")
			fmt.Println(err)
		}
		fmt.Println(string(body))
	}
}

func printRequestBody() {

}

func printRequestHeaders() {

}

func parseQueryParams(inputURL string) {
	parsedURL, _ := url.Parse(inputURL)
	paramsMap, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil {
		fmt.Println("Something went wrong parsing query params")
		fmt.Println(err)
	}
	fmt.Println(paramsMap)
}

func main() {
	fmt.Println("httpc: a HTTP client")
	sampleHeaders := libhttpc.Header{
		"Authorization": "None",
		"Referer":       "google.com",
		"Content-Type":  "application/json",
	}
	sampleBody := map[string]string{
		"Bleh": "bleeh",
	}
	//_, err := libhttpc.Get("http://httpbin.org/headers", sampleHeaders)
	reqBody, _ := json.Marshal(sampleBody)
	resp, err := libhttpc.Post("ttp://httpbin.org/post", sampleHeaders, reqBody)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
}
