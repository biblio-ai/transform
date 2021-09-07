package main

import (
  "flag"
  "context"
  "database/sql"
  "encoding/json"
  "fmt"
  //"github.com/Azure/azure-sdk-for-go/services/preview/cognitiveservices/v3.0-preview/computervision"
  "github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v2.0/computervision"
  "github.com/Azure/go-autorest/autorest"
  "log"
  "os"
  "strings"
  "time"

  _ "github.com/mattn/go-sqlite3"
  _ "github.com/lib/pq"

  "github.com/ilyakaznacheev/cleanenv"
  "github.com/satori/go.uuid"
  "net/http"
  _"net/url"

)

// Declare global so don't have to pass it to all of the tasks.
var computerVisionContext context.Context
//var database, _ = sql.Open("sqlite3", "./azure.db")
/*
const (
  host = os.Getenv("DATABASE_HOST")
  port = os.Getenv("DATABASE_POST") 
  user = os.Getenv("DATABASE_USER")  
  password = os.Getenv("DATABASE_PASSWORD")
  dbname  = os.Getenv("DATABASE_NAME")
)
type ConfigDatabase struct {
  Port     int `yaml:"port" env-default:"5431"  env-description:"Database host"`
  Host     string `yaml:"host" env-description:"Database host"`
  Name     string `yaml:"name" env-default:"postgres"  env-description:"Database host"`
  User     string `yaml:"user"  env-default:"postgres" env-description:"Database host"`
  Password string `yaml:"password" env-description:"Database host"`
}
*/
//var cfg ConfigDatabase

type ConfigDatabase struct {
  Port     int `yaml:"port" env-default:"5431"  env-description:"Database host"`
  Host     string `yaml:"host" env-description:"Database host"`
  Name     string `yaml:"name" env-default:"postgres"  env-description:"Database host"`
  User     string `yaml:"user"  env-default:"postgres" env-description:"Database host"`
  Password string `yaml:"password" env-description:"Database host"`
}

var cfg ConfigDatabase

//var db *sql.DB
  
type Env struct {
    db *sql.DB
}

func main() {

  var config_file string
  if fileExists("config.custom.yml") {
    config_file = "./config.custom.yml"
  } else {
    config_file = "./config.yml"
  }
  err := cleanenv.ReadConfig(config_file, &cfg)
  fmt.Printf("%+v", cfg)
  fmt.Println(err)
  psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
  "password=%s dbname=%s sslmode=disable",
  cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
  db, err := sql.Open("postgres", psqlInfo)
  if err != nil {
    fmt.Println("Successfully NOT connected!")
    panic(err)
  } else {
    fmt.Println("Successfully connected! - main")
  }
   // Create an instance of Env containing the connection pool.
    env := &Env{db: db}
  defer env.db.Close()

  createBiblioDB(env)

  err = env.db.Ping()
  if err != nil {
    fmt.Println("Successfully NOT connected!")
    panic(err)
  }
  //create(&db)
  stmt := `select id, url from item where url = $1 limit 1`
  var item_id string
  //var item_uuid uuid.UUID
  //var last_insert_id int

  urlFlag := flag.String("url","https://rosetta.slv.vic.gov.au/delivery/DeliveryManagerServlet?dps_func=stream&dps_pid=FL20697763","URL here")
  flag.Parse()

  if strings.Contains(*urlFlag, "handle") {// true

  resp, err := http.Get(*urlFlag)

  if err != nil {
    // handle err
  }
  defer resp.Body.Close()
  }

  imageURL := *urlFlag

  var url string

  row := env.db.QueryRow(stmt,imageURL)
  switch err := row.Scan(&item_id, &url); err {
  case sql.ErrNoRows:
    fmt.Println("No rows were returned!")
  case nil:
    fmt.Println(item_id, url)
  default:
    panic(err)
  }
  if len(item_id) ==0 {

    fmt.Println("Prepare insert")
    fmt.Println(imageURL)
    statement := `INSERT INTO item (url) VALUES ($1) RETURNING id`
    err = env.db.QueryRow(statement,imageURL).Scan(&item_id)
    if err != nil {
      fmt.Println(err)
      return
    }

    fmt.Println("Last Insert ID")
    fmt.Println(item_id)
  }
  fmt.Println("Item ID")
  fmt.Println(item_id)
  fmt.Println("Item URL")
  fmt.Println(imageURL)
  timestamp  := time.Now().Unix()
  /*
  * Configure the Computer Vision client
  * Set environment variables for COMPUTER_VISION_SUBSCRIPTION_KEY and COMPUTER_VISION_ENDPOINT,
  * then restart your command shell or your IDE for changes to take effect.
  */
  computerVisionKey := os.Getenv("COMPUTER_VISION_SUBSCRIPTION_KEY")

  if computerVisionKey == "" {
    log.Fatal("\n\nPlease set a COMPUTER_VISION_SUBSCRIPTION_KEY environment variable.\n" +
    "**You may need to restart your shell or IDE after it's set.**\n")
  }

  endpointURL := os.Getenv("COMPUTER_VISION_ENDPOINT")
  if endpointURL == "" {
    log.Fatal("\n\nPlease set a COMPUTER_VISION_ENDPOINT environment variable.\n" +
    "**You may need to restart your shell or IDE after it's set.**")
  }

  computerVisionClient := computervision.New(endpointURL)
  computerVisionClient.Authorizer = autorest.NewCognitiveServicesAuthorizer(computerVisionKey)

  computerVisionContext = context.Background()
  /*
  * END - Configure the Computer Vision client
  */
  /*printedImageURL := "https://i.imgur.com/I9r02n7.png"
  */
  //	printedImageURL := "https://s3-ap-southeast-2.amazonaws.com/awm-media/collection/PR82/193.023/large/4164690.JPG"
  //        printedImageURL := "https://i.imgur.com/6n0uxk9.png" /*SLV avoca*/
  //        printedImageURL := "https://i.imgur.com/i41tezf.jpg" /*SLV eureka*/
  //printedImageURL := "https://i.imgur.com/YkqQZfB.png" /*George Swinburne*/
  //        printedImageURL := "https://i.imgur.com/XkJUPRL.png" /*SLV eureka*/
  //printedImageURL := "https://commons.swinburne.edu.au/file/cd53e247-3e39-458e-8582-9fa0a2a2e120/1/cor-duncan_to_green_1920.jpg"
  ///* SWIn letteer
  // Analyze text in an image, remote
  BatchReadFileRemoteImage(env,computerVisionClient, imageURL, item_id, timestamp)

  // Analyze features of an image, remote
  DescribeRemoteImage(env, computerVisionClient, imageURL, item_id, timestamp)
  CategorizeRemoteImage(env,computerVisionClient, imageURL, item_id, timestamp)
  TagRemoteImage(env,computerVisionClient, imageURL, item_id, timestamp)
  DetectFacesRemoteImage(env,computerVisionClient, imageURL, item_id, timestamp)
  DetectObjectsRemoteImage(env,computerVisionClient, imageURL, item_id, timestamp)
  DetectBrandsRemoteImage(env,computerVisionClient, imageURL, item_id, timestamp)
  DetectAdultOrRacyContentRemoteImage(env,computerVisionClient, imageURL, item_id, timestamp)
  DetectColorSchemeRemoteImage(env,computerVisionClient, imageURL, item_id, timestamp)
  //DetectDomainSpecificContentRemoteImage(env,computerVisionClient, imageURL, item_id, timestamp)

  //text_analytic(item_id) 
  DetectLanguage(env,item_id, timestamp)
  ExtractEntities(env,item_id, timestamp)
  SentimentAnalysis(env,item_id, timestamp)
  ExtractKeyPhrases(env,item_id, timestamp)
  Get(env,item_id)


}


func BatchReadFileRemoteImage(env *Env, client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
  fmt.Println()
  fmt.Println("-----------------------------------------")
  fmt.Println("BATCH READ FILE - remote")
  fmt.Println(remoteImageURL)
  fmt.Println(item_id)
  item_uuid,err := uuid.FromString(item_id)
  var remoteImage computervision.ImageURL
  remoteImage.URL = &remoteImageURL

  // The response contains a field called "Operation-Location", 
  // which is a URL with an ID that you'll use for GetReadOperationResult to access OCR results.
  textHeaders, err := client.BatchReadFile(computerVisionContext, remoteImage)
  if err != nil { log.Fatal(err) }

  // Use ExtractHeader from the autorest library to get the Operation-Location URL
  operationLocation := autorest.ExtractHeaderValue("Operation-Location", textHeaders.Response)

  numberOfCharsInOperationId := 36
  operationId := string(operationLocation[len(operationLocation)-numberOfCharsInOperationId : len(operationLocation)])
  // </snippet_read_call>

  // <snippet_read_response>
  readOperationResult, err := client.GetReadOperationResult(computerVisionContext, operationId)
  if err != nil { log.Fatal(err) }

  // Wait for the operation to complete.
  i := 0
  maxRetries := 10

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
    if err != nil { log.Fatal(err) }
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
      err = env.db.QueryRow(statement_t,item_uuid, timestamp, index, *line.Text, fmt.Sprintf("%v",*line.BoundingBox)).Scan(&iid)

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
