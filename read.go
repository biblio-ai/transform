package main

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v2.0/computervision"
	"github.com/Azure/go-autorest/autorest"
	"log"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
)

type ReadJSON struct {
	AnalyzeResult struct {
		ModelVersion string `json:"modelVersion"`
		ReadResults  []struct {
			Angle  float64 `json:"angle"`
			Height int     `json:"height"`
			Lines  []struct {
				Appearance struct {
					Style struct {
						Confidence float64 `json:"confidence"`
						Name       string  `json:"name"`
					} `json:"style"`
				} `json:"appearance"`
				BoundingBox []int  `json:"boundingBox"`
				Text        string `json:"text"`
				Words       []struct {
					BoundingBox []int   `json:"boundingBox"`
					Confidence  float64 `json:"confidence"`
					Text        string  `json:"text"`
				} `json:"words"`
			} `json:"lines"`
			Page  int    `json:"page"`
			Unit  string `json:"unit"`
			Width int    `json:"width"`
		} `json:"readResults"`
		Version string `json:"version"`
	} `json:"analyzeResult"`
	CreatedDateTime     time.Time `json:"createdDateTime"`
	LastUpdatedDateTime time.Time `json:"lastUpdatedDateTime"`
	Status              string    `json:"status"`
}

func BatchReadFileRemoteImage(env *Env, client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
	fmt.Println()
	fmt.Println("-----------------------------------------")
	fmt.Println("BATCH READ FILE - remote")
	fmt.Println(remoteImageURL)
	fmt.Println(item_id)
	item_uuid, err := uuid.FromString(item_id)
	var remoteImage computervision.ImageURL
	remoteImage.URL = &remoteImageURL

	// The response contains a field called "Operation-Location",
	// which is a URL with an ID that you'll use for GetReadOperationResult to access OCR results.
	textHeaders, err := client.BatchReadFile(computerVisionContext, remoteImage)
	if err != nil {
		log.Fatal(err)
	}

	// Use ExtractHeader from the autorest library to get the Operation-Location URL
	operationLocation := autorest.ExtractHeaderValue("Operation-Location", textHeaders.Response)

	numberOfCharsInOperationId := 36
	operationId := string(operationLocation[len(operationLocation)-numberOfCharsInOperationId : len(operationLocation)])
	// </snippet_read_call>

	// <snippet_read_response>
	readOperationResult, err := client.GetReadOperationResult(computerVisionContext, operationId)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for the operation to complete.
	i := 0
	maxRetries := 35

	fmt.Println("Recognizing text in a remote image with the batch Read API ...")
	for readOperationResult.Status != computervision.Failed &&
		readOperationResult.Status != computervision.Succeeded {
		if i >= maxRetries {
			break
		}
		i++

		fmt.Printf("Server status: %v, waiting %v seconds...\n", readOperationResult.Status, i)
		time.Sleep(1 * time.Second)

		readOperationResult, err = client.GetReadOperationResult(computerVisionContext, operationId)
		if err != nil {
			log.Fatal(err)
		}
	}
	// </snippet_read_response>

	// <snippet_read_display>
	// Display the results.
	fmt.Println()
	for _, recResult := range *(readOperationResult.RecognitionResults) {
		for index, line := range *recResult.Lines {
			fmt.Println(*line.Text)
			fmt.Println(*line.BoundingBox)
			var iid string
			statement_t := `INSERT INTO item_text (item_id, timestamp, line, value, box) VALUES ($1, $2, $3, $4, $5) RETURNING id`
			err = env.db.QueryRow(statement_t, item_uuid, timestamp, index, *line.Text, fmt.Sprintf("%v", *line.BoundingBox)).Scan(&iid)

			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Entity - Last Insert ID")
			fmt.Println(iid)
			/*
			   for _, linew := range *line.Words {
			     fmt.Println("Print Word")
			     fmt.Println(*linew.Text)
			     fmt.Println(linew.Text)
			     fmt.Println(*linew.BoundingBox)
			     fmt.Println(linew.BoundingBox)
			     fmt.Println(linew.Confidence)
			   }
			*/
		}
	}
	// </snippet_read_display>
	fmt.Println()
}
func AnalyzeImage(env *Env, client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
	// Add your Computer Vision subscription key and endpoint to your environment variables.
	subscriptionKey := os.Getenv("COMPUTER_VISION_SUBSCRIPTION_KEY")
	endpoint := os.Getenv("COMPUTER_VISION_ENDPOINT")

	uriBase := endpoint + "vision/v3.2/analyze"
	var imageUrl = remoteImageURL
	item_uuid, err := uuid.FromString(item_id)

	const params = "?visualFeatures=Categories,Description,Faces,Objects,Tags&model-version=latest"
	uri := uriBase + params
	var imageUrlEnc = "{\"url\":\"" + imageUrl + "\"}"

	fmt.Println(imageUrlEnc)
	fmt.Println(uri)

	reader := strings.NewReader(imageUrlEnc)

	// Create the HTTP client
	hclient := &http.Client{
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
	resp, err := hclient.Do(req)
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

	fmt.Println("Tags in the remote image: ")
	if len(cvjson.Tags) == 0 {
		fmt.Println("No tags detected.")
	} else {
		for _, tag := range cvjson.Tags {
			fmt.Printf("'%v' with confidence %.2f%%\n", tag.Name, tag.Confidence*100)

			var idt string
			statement_tag := `INSERT INTO item_tag (item_id, timestamp, value, score) VALUES ($1, $2, $3, $4) RETURNING id`
			err = env.db.QueryRow(statement_tag, item_uuid, timestamp, tag.Name, tag.Confidence).Scan(&idt)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	for _, cvdesc := range cvjson.Description.Captions {
		fmt.Println("Text: " + cvdesc.Text)
		fmt.Println("Confidence: " + fmt.Sprint(cvdesc.Confidence))
	}
	fmt.Println("Captions from remote image: ")
	if len(cvjson.Description.Captions) == 0 {
		fmt.Println("No captions detected.")
	} else {
		for _, caption := range cvjson.Description.Captions {
			fmt.Printf("'%v' with confidence %.2f%%\n", caption.Text, caption.Confidence*100)

                        var idd string
                        statement_t := `INSERT INTO item_description (item_id, timestamp, value, score) VALUES ($1, $2, $3, $4) RETURNING id`
                        err = env.db.QueryRow(statement_t, item_uuid, timestamp, caption.Text, caption.Confidence).Scan(&idd)
                        if err != nil {
                          fmt.Println(err)
                          return
                        }

                        fmt.Println("Entity - Last Insert ID")
                        fmt.Println(idd)
		}
	}
	fmt.Println()
}
