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

type CVJson struct {
	Categories []struct {
		Name  string  `json:"name"`
		Score float64 `json:"score"`
	} `json:"categories"`
	Description struct {
		Captions []struct {
			Confidence float64 `json:"confidence"`
			Text       string  `json:"text"`
		} `json:"captions"`
		Tags []string `json:"tags"`
	} `json:"description"`
	Faces []struct {
		Age           int64 `json:"age"`
		FaceRectangle struct {
			Height int64 `json:"height"`
			Left   int64 `json:"left"`
			Top    int64 `json:"top"`
			Width  int64 `json:"width"`
		} `json:"faceRectangle"`
		Gender string `json:"gender"`
	} `json:"faces"`
	Metadata struct {
		Format string `json:"format"`
		Height int64  `json:"height"`
		Width  int64  `json:"width"`
	} `json:"metadata"`
	ModelVersion string `json:"modelVersion"`
	Objects      []struct {
		Confidence float64 `json:"confidence"`
		Object     string  `json:"object"`
		Rectangle  struct {
			H int64 `json:"h"`
			W int64 `json:"w"`
			X int64 `json:"x"`
			Y int64 `json:"y"`
		} `json:"rectangle"`
	} `json:"objects"`
	RequestID string `json:"requestId"`
	Tags      []struct {
		Confidence float64 `json:"confidence"`
		Name       string  `json:"name"`
	} `json:"tags"`
}

func main() {

	// Add your Computer Vision subscription key and endpoint to your environment variables.
	subscriptionKey := os.Getenv("COMPUTER_VISION_SUBSCRIPTION_KEY")
	endpoint := os.Getenv("COMPUTER_VISION_ENDPOINT")

	uriBase := endpoint + "vision/v3.2/describe"
	const imageUrl = "https://rosetta.slv.vic.gov.au/delivery/DeliveryManagerServlet?dps_func=stream&dps_pid=FL16345007"

	const params = "?maxCanidates=5&model-version=2021-04-01"
	uri := uriBase + params
	const imageUrlEnc = "{\"url\":\"" + imageUrl + "\"}"

	fmt.Println(imageUrlEnc)
	fmt.Println(uri)

	reader := strings.NewReader(imageUrlEnc)

	// Create the HTTP client
	client := &http.Client{
		Timeout: time.Second * 10,
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
	var cvjson CVJson
	json.Unmarshal(data, &cvjson)

	fmt.Println(cvjson.RequestID)
	for _, cvtag := range cvjson.Tags {
		fmt.Println("Name: " + cvtag.Name)
		fmt.Println("Confidence: " + fmt.Sprint(cvtag.Confidence))
	}
	for _, cvdesc := range cvjson.Description.Captions {
		fmt.Println("Text: " + cvdesc.Text)
		fmt.Println("Confidence: " + fmt.Sprint(cvdesc.Confidence))
	}
}
