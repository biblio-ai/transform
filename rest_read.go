package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	// Add your Computer Vision subscription key and endpoint to your environment variables.
	subscriptionKey := os.Getenv("COMPUTER_VISION_SUBSCRIPTION_KEY")
	endpoint := os.Getenv("COMPUTER_VISION_ENDPOINT")

	uriBase := endpoint + "vision/v3.2/read/analyze"
	//uriBase := endpoint + "vision/v3.2/ocr"

	//const imageUrl = "https://rosetta.slv.vic.gov.au/delivery/DeliveryManagerServlet?dps_func=stream&dps_pid=FL19637103"
	const imageUrl = "https://rosetta.slv.vic.gov.au/delivery/DeliveryManagerServlet?dps_func=stream&dps_pid=FL16345007"

	const params = "?readingOrder=natural&model-version=2021-09-30-preview"
	//const params = "?readingOrder=natural&model-version=latest"
	//const params = "?model-version=latest"
	//const params = "?detechOrientation=true&model-version=latest"

	uri := uriBase + params
	fmt.Println(uri)
	const imageUrlEnc = "{\"url\":\"" + imageUrl + "\"}"

	reader := strings.NewReader(imageUrlEnc)

	fmt.Println(imageUrlEnc)
	fmt.Println(uri)

	// Create the HTTP client
	client := &http.Client{
		Timeout: time.Second * 20,
        } 

	// Create the POST request, passing the image URL in the request body
	req, err := http.NewRequest("POST", uri, reader)
	if err != nil {
		panic(err)
	}

	// Add request headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Ocp-Apim-Subscription-Key", subscriptionKey)

	// Send the request and retrieve the response
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp.Header.Get("Operation-Location"))
	fmt.Println(resp.Header.Get("apim-request-id"))

	defer resp.Body.Close()

	// Read the response body
	// Note, data is a byte array
	data, err := ioutil.ReadAll(resp.Body)
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
