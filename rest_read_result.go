package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	// Add your Computer Vision subscription key and endpoint to your environment variables.
	subscriptionKey := os.Getenv("COMPUTER_VISION_SUBSCRIPTION_KEY")
	endpoint := os.Getenv("COMPUTER_VISION_ENDPOINT")

	uriBase := endpoint + "vision/v3.2/read/analyzeResults/"
	const params = "ed18ee22-982c-46cd-942a-5425e6b3370c"
	uri := uriBase + params

	fmt.Println(uri)

	// Create the HTTP client
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// Create the POST request, passing the image URL in the request body
	req_read, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err)
	}
	// Add request headers
	req_read.Header.Add("Content-Type", "application/json")
	req_read.Header.Add("Ocp-Apim-Subscription-Key", subscriptionKey)

	// Send the request and retrieve the response
	resp_read, err := client.Do(req_read)
	if err != nil {
		panic(err)
	}

	defer resp_read.Body.Close()

	// Read the response body
	// Note, data is a byte array
	data, err := ioutil.ReadAll(resp_read.Body)
	if err != nil {
		panic(err)
	}

	// Parse the JSON data from the byte array
	var f interface{}
	json.Unmarshal(data, &f)

	// Format and display the JSON result
	jsonFormatted, _ := json.MarshalIndent(f, "", "  ")
	fmt.Println(string(jsonFormatted))
}
