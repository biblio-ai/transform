package main

import (
  "context"
  "encoding/json"
  "fmt"
  "github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v2.1/textanalytics"
  "github.com/Azure/go-autorest/autorest"
  "github.com/Azure/go-autorest/autorest/to"
  "log"
  "os"

  _ "github.com/mattn/go-sqlite3"
  "github.com/satori/go.uuid"
)

func GetTextAnalyticsClient() textanalytics.BaseClient {
  var subscriptionKey string = os.Getenv("TEXT_ANALYTICS_KEY")
  var endpoint string = os.Getenv("TEXT_ANALYTICS_ENDPOINT")
  textAnalyticsClient := textanalytics.New(endpoint)
  textAnalyticsClient.Authorizer = autorest.NewCognitiveServicesAuthorizer(subscriptionKey)

  return textAnalyticsClient
}

func ExtractEntities(env *Env,item_id string, timestamp int64) {
  //item_uuid,err := uuid.FromString(item_id)
  item_text := GetItemText(env,item_id)
  fmt.Println("--- Item Text - Chunch ---")
  // get any error encountered during iteration
  /*
  for _, c := range cs {
    fmt.Println("--- Item Text - Chunch ---")
    fmt.Println(c)
  }
  err = rows.Err()
  if err != nil {
    panic(err)
  }
  */

  fmt.Println("--- Item Text ---")
  stmt2, err := env.db.Prepare("select code from item_text_language where item_id = $1 order by timestamp desc limit 1")
  if err != nil {
    log.Fatal(err)
  }
  defer stmt2.Close()
  var lang_code string
  err = stmt2.QueryRow(item_id).Scan(&lang_code)

  fmt.Println("--- Item Text 2 Analytics ---")
  const uriPath = "/text/analytics/v2.1/entities"

  fmt.Printf("TextAnalytics: \n")
  fmt.Printf("Language code: %s\n", lang_code)


  cs := Chunks(item_text, 5120)
  for _, c := range cs {
    fmt.Println("--- Item Text - Chunch ---")
    fmt.Println(c)

    textAnalyticsClient := GetTextAnalyticsClient()
    ctx := context.Background()
    inputDocuments := []textanalytics.MultiLanguageInput{
      {
        ID:       to.StringPtr("1"),
        Language: to.StringPtr(lang_code),
        Text:     to.StringPtr(c),
      },
    }

    batchInput := textanalytics.MultiLanguageBatchInput{Documents: &inputDocuments}
    result, _ := textAnalyticsClient.Entities(ctx, to.BoolPtr(false), &batchInput)

    // Printing extracted entities results
    fmt.Println()
    for _, document := range *result.Documents {
      for _, entity := range *document.Entities {
        fmt.Printf("Name: %s\tType: %s", *entity.Name, *entity.Type)

        var entity_sub_type string = ""
        if entity.SubType != nil {
          fmt.Printf("Sub-Type: %s\n", *entity.SubType)
          entity_sub_type = *entity.SubType
        }
        for _, match := range *entity.Matches {

          var match_wikipedia_score float64 = 0
          var match_entity_type_score float64 = 0
          if match.EntityTypeScore != nil {
            fmt.Printf("EntityTypeScore: %v\n", *match.EntityTypeScore)
            match_entity_type_score = *match.EntityTypeScore
          }
          if match.WikipediaScore != nil {
            fmt.Printf("WikipediaScore: %v\n", *match.WikipediaScore)
            match_wikipedia_score = *match.WikipediaScore
          }
          var match_text string = ""
          if match.Text != nil {
            fmt.Printf("Match-Tex: %s\n", *match.Text)
            match_text = *match.Text
          }
          fmt.Printf("\t\t\tOffset: %v\tLength: %v\tScore: %f\n", *match.Offset, *match.Length, match_entity_type_score)
          fmt.Printf("\t\t\tName: %v\tType: %v\tSubType: %v\n", *entity.Name,*entity.Type, entity_sub_type)
          fmt.Printf("\t\t\tText: %v\tWikipedia: %v\n", match_text, match_wikipedia_score)
          fmt.Printf("\t\t\tItemID: %v\tTimestamp: %v\n", item_id, timestamp)

          //statement, _ := db.Prepare("INSERT INTO item_text_entity (item_id, timestamp,value, text_length,  text_offset,  text_type,  text_sub_type, match_text,match_wikipeida_score, text_score ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)")
          statement, _ := env.db.Prepare("INSERT INTO item_text_entity (item_id, timestamp,value, text_length,  text_offset,  text_type,  text_sub_type, match_text, text_score ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)")
          _, err := statement.Exec(
              item_id,
              timestamp,
              *entity.Name,
              *match.Length,
              *match.Offset,
              *entity.Type,
              entity_sub_type,
              match_text,
              //match_wikipedia_score,
              match_entity_type_score)
          if err != nil {
            fmt.Println(err)
            return
          }
        }
      }
      fmt.Println()
    }

    // Printing document errors
    fmt.Println("Document Errors - Extract Entities")
    for _, err := range *result.Errors {
      fmt.Printf("Document ID: %s Message : %s\n", *err.ID, *err.Message)
    }
  }

}
func DetectLanguage(env *Env,item_id string, timestamp int64) {

  fmt.Println("Detect Language")
  //stmt, err := db.Prepare("select value from item_text where item_id = ? order by timestamp DESC limit 1")
  /*
  stmt, err := db.Prepare("select value from item_text where item_id = $1 limit 1")
  if err != nil {
    log.Fatal(err)
  }
  defer stmt.Close()
  var item_text string
  err = stmt.QueryRow(item_id).Scan(&item_text)
  */
  item_text := GetItemText(env,item_id)
  cs := Chunks(item_text, 5120)
  //var last_insert_id int
  //azure
  textAnalyticsClient := GetTextAnalyticsClient()
  ctx := context.Background()
  inputDocuments := []textanalytics.LanguageInput{
    {
      ID:   to.StringPtr("0"),
      Text: to.StringPtr(cs[0]),
    },
  }

  batchInput := textanalytics.LanguageBatchInput{Documents: &inputDocuments}
  result, _ := textAnalyticsClient.DetectLanguage(ctx, to.BoolPtr(false), &batchInput)

  // Printing language detection results
  for _, document := range *result.Documents {
    fmt.Printf("Document ID: %s ", *document.ID)
    fmt.Printf("Detected Languages with Score: ")
    for _, language := range *document.DetectedLanguages {
      fmt.Printf("%s %s %f,", *language.Name, *language.Iso6391Name,*language.Score)

      statement, _ := env.db.Prepare("INSERT INTO item_text_language (item_id, timestamp, value, code, score ) VALUES ($1,$2,$3,$4,$5)")
      _, err := statement.Exec(item_id, timestamp, *language.Name, *language.Iso6391Name, *language.Score)
      if err != nil {
        fmt.Println(err)
        return
      }
    }
    fmt.Println()
  }

  // Printing document errors
  for _, err := range *result.Errors {
    fmt.Println("Document Errors - detech Lang")
    fmt.Printf("Document ID: %s Message : %s\n", *err.ID, *err.Message)
  }
}

func SentimentAnalysis(env *Env,item_id string, timestamp int64) {

  fmt.Println("Sentiment")
  //stmt, err := db.Prepare("select value from item_text where item_id = ? order by timestamp DESC limit 1")
  stmt, err := env.db.Prepare("select value from item_text where item_id = $1 limit 1")
  if err != nil {
    log.Fatal(err)
  }
  defer stmt.Close()
  var item_text string
  //var last_insert_id int
  err = stmt.QueryRow(item_id).Scan(&item_text)
  //azure
  //stmt2, err := db.Prepare("select code from item_text_language where item_id = ? order by timestamp DESC limit 1")
  stmt2, err := env.db.Prepare("select code from item_text_language where item_id = $1 order by timestamp desc limit 1")
  if err != nil {
    log.Fatal(err)
  }
  defer stmt2.Close()
  var lang_code string
  //var last_insert_id int
  err = stmt2.QueryRow(item_id).Scan(&lang_code)

  textAnalyticsClient := GetTextAnalyticsClient()
  ctx := context.Background()
  inputDocuments := []textanalytics.MultiLanguageInput{
    {
      Language: to.StringPtr(lang_code),
      ID:       to.StringPtr("0"),
      Text:     to.StringPtr(item_text),
    },
  }

  batchInput := textanalytics.MultiLanguageBatchInput{Documents: &inputDocuments}
  result, _ := textAnalyticsClient.Sentiment(ctx, to.BoolPtr(false), &batchInput)
  var batchResult textanalytics.SentimentBatchResult
  jsonString, _ := json.Marshal(result)
  _ = json.Unmarshal(jsonString, &batchResult)

  // Printing sentiment results
  for _, document := range *batchResult.Documents {
    fmt.Printf("Document ID: %s ", *document.ID)
    fmt.Printf("Sentiment Score: %f\n", *document.Score)

    statement, _ := env.db.Prepare("INSERT INTO item_text_sentiment (item_id,timestamp, score ) VALUES ($1,$2,$3)")
    result, err := statement.Exec(item_id, timestamp, *document.Score)
    fmt.Println("Entity - Last Insert ID")
    iid, err := result.LastInsertId()
    fmt.Println(iid)
    if err != nil {
      fmt.Println(err)
      return
    }
  }

  // Printing document errors
  fmt.Println("Document Errors - Sentiment")
  for _, err := range *batchResult.Errors {
    fmt.Printf("Document ID: %s Message : %s\n", *err.ID, *err.Message)
  }
}
func ExtractKeyPhrases(env *Env,item_id string, timestamp int64) {

  //stmt, err := db.Prepare("select value from item_text where item_id = ? order by timestamp DESC limit 1")
  /*
  stmt, err := db.Prepare("select value from item_text where item_id = $1 limit 1")
  if err != nil {
    log.Fatal(err)
  }
  defer stmt.Close()
  var item_text string
  */
  item_text := GetItemText(env,item_id)
  //var last_insert_id int
  //err = stmt.QueryRow(item_id).Scan(&item_text)

  //stmt2, err := db.Prepare("select code from item_text_language where item_id = ? order by timestamp DESC limit 1")
  stmt2, err := env.db.Prepare("select code from item_text_language where item_id = $1 order by timestamp desc limit 1")
  if err != nil {
    log.Fatal(err)
  }
  defer stmt2.Close()
  var lang_code string
  //var last_insert_id int
  err = stmt2.QueryRow(item_id).Scan(&lang_code)
  //azure
  cs := Chunks(item_text, 5120)
  for _, c := range cs {
    fmt.Println("--- Item Text - Chunch ---")

    textAnalyticsClient := GetTextAnalyticsClient()
    ctx := context.Background()
    inputDocuments := []textanalytics.MultiLanguageInput{
      {
        ID:       to.StringPtr("1"),
        Language: to.StringPtr(lang_code),
        Text:     to.StringPtr(c),
      },
    }

    batchInput := textanalytics.MultiLanguageBatchInput{Documents: &inputDocuments}
    result, _ := textAnalyticsClient.KeyPhrases(ctx, to.BoolPtr(false), &batchInput)

    // Printing extracted key phrases results
    for _, document := range *result.Documents {
      fmt.Printf("Document ID: %s\n", *document.ID)
      fmt.Printf("Document Language: %s\n", lang_code)
      fmt.Printf("\tExtracted Key Phrases:\n")
      for _, keyPhrase := range *document.KeyPhrases {
        fmt.Printf("\t\t%s\n", keyPhrase)
        statement, _ := env.db.Prepare("INSERT INTO item_text_key_phrase (item_id,timestamp, value ) VALUES ($1,$2,$3)")
        _, err := statement.Exec(item_id, timestamp, keyPhrase)
        if err != nil {
          fmt.Println(err)
          return
        }
        /*
        fmt.Println("Entity - Last Insert ID")
        iid, err := result.LastInsertId()
        fmt.Println(iid)
        */
      }
      fmt.Println()
    }
    // Printing document errors
    for _, err := range *result.Errors {
      fmt.Println("Document Errors - text key phrases")
      fmt.Printf("Document ID: %s Message : %s\n", *err.ID, *err.Message)
    }
  }

}
func Get(env *Env,item_id string) {

  fmt.Println("Print Items: Key Phrases")
  stmt, err := env.db.Prepare(`
  WITH RECURSIVE  x as 
  (
    select 
    max(timestamp) as max 
    from 
    item_text_key_phrase 
    where 
    item_id = $1
  )  
  select distinct value from item_text_key_phrase WHERE timestamp = (select max from x)
  `)
  result, err := stmt.Query(item_id)
  if err != nil {
    log.Fatal(err)
  }
  defer stmt.Close()

  var value string

  for result.Next() {
    err := result.Scan(&value)
    if err != nil {
      log.Fatal(err)
    }
    fmt.Println(value)
  }
}
func GetItemText(env *Env,item_id string) (value_concat string) {

  fmt.Println("Get Item Text")

  item_uuid,err := uuid.FromString(item_id)
  rows, err := env.db.Query("SELECT value from item_text where item_id = $1", item_uuid)
  if err != nil {
    // handle this error better than this
    panic(err)
  }
  defer rows.Close()
  for rows.Next() {
    var value string
    err = rows.Scan(&value)
    if err != nil {
      // handle this error
      panic(err)
    }
    //fmt.Println(value)
    value_concat += "\n"
    value_concat += value
  }
  return
}
