package main

import (
  "fmt"
  "github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v2.0/computervision"
  "log"

  _ "github.com/mattn/go-sqlite3"
  "github.com/satori/go.uuid"
)

func  DescribeRemoteImage(env *Env,client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
  fmt.Println("-----------------------------------------")
  fmt.Println("DESCRIBE IMAGE - remote")
  fmt.Println()
  var remoteImage computervision.ImageURL
  var caption_value string
  var caption_confidence float64
  remoteImage.URL = &remoteImageURL
  item_uuid,err := uuid.FromString(item_id)
  if err != nil {
    fmt.Printf("Something went wrong: %s", err)
    return
  }

  maxNumberDescriptionCandidates := new(int32)
  *maxNumberDescriptionCandidates = 1

  remoteImageDescription, err := client.DescribeImage(
    computerVisionContext,
    remoteImage,
    maxNumberDescriptionCandidates,
    "") // language
    if err != nil {
      log.Fatal(err)
    }

    fmt.Println("Captions from remote image: ")
    if len(*remoteImageDescription.Captions) == 0 {
      fmt.Println("No captions detected.")
    } else {
      for _, caption := range *remoteImageDescription.Captions {
        fmt.Printf("'%v' with confidence %.2f%%\n", *caption.Text, *caption.Confidence*100)
        caption_value = *caption.Text
        caption_confidence = *caption.Confidence
      }
    }
    fmt.Println()

    /*
    statement, _ := db.Prepare("INSERT INTO item_description (item_id, timestamp, value, score) VALUES (?, ?, ?, ?)")
    result, err := statement.Exec(item_id, timestamp, caption_value, caption_confidence)
    if err != nil {
      fmt.Println(err)
      return
    }
    */
    var idd string
    statement_t := `INSERT INTO item_description (item_id, timestamp, value, score) VALUES ($1, $2, $3, $4) RETURNING id`
    err = env.db.QueryRow(statement_t, item_uuid, timestamp, caption_value, caption_confidence).Scan(&idd)
    if err != nil {
      fmt.Println(err)
      return
    }


    fmt.Println("Entity - Last Insert ID")
    fmt.Println(idd)

  }

  func CategorizeRemoteImage(env *Env,client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
    fmt.Println("-----------------------------------------")
    fmt.Println("CATEGORIZE IMAGE - remote")
    fmt.Println()
    var remoteImage computervision.ImageURL
    remoteImage.URL = &remoteImageURL
    item_uuid,err := uuid.FromString(item_id)
    if err != nil {
      fmt.Printf("Something went wrong: %s", err)
      return
    }

    features := []computervision.VisualFeatureTypes{computervision.VisualFeatureTypesCategories}
    imageAnalysis, err := client.AnalyzeImage(
      computerVisionContext,
      remoteImage,
      features,
      []computervision.Details{},
      "")
      if err != nil {
        log.Fatal(err)
      }

      fmt.Println("Categories from remote image: ")
      if len(*imageAnalysis.Categories) == 0 {
        fmt.Println("No categories detected.")
      } else {
        for _, category := range *imageAnalysis.Categories {
          fmt.Printf("'%v' with confidence %.2f%%\n", *category.Name, *category.Score*100)
          /*
          statement, _ := db.Prepare("INSERT INTO item_category (item_id, timestamp,  value, score) VALUES (?, ?, ?, ?)")
          result, err := statement.Exec(item_id, timestamp,  *category.Name, *category.Score)
          fmt.Println("Entity - Last Insert ID")
          iid, err := result.LastInsertId()
          fmt.Println(iid)
          if err != nil {
            fmt.Println(err)
            return
          }
          */
          var idc string
          statement_c := `INSERT INTO item_category (item_id, timestamp, value, score) VALUES ($1, $2, $3, $4) RETURNING id`
          err = env.db.QueryRow(statement_c, item_uuid, timestamp, *category.Name, *category.Score).Scan(&idc)
          if err != nil {
            fmt.Println(err)
            return
          }
        }
      }
      fmt.Println()

    }
    func TagRemoteImage(env *Env,client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
      fmt.Println("-----------------------------------------")
      fmt.Println("TAG IMAGE - remote")
      fmt.Println()
      var remoteImage computervision.ImageURL
      remoteImage.URL = &remoteImageURL
      item_uuid,err := uuid.FromString(item_id)
      if err != nil {
        fmt.Printf("Something went wrong: %s", err)
        return
      }

      remoteImageTags, err := client.TagImage(
        computerVisionContext,
        remoteImage,
        "")
        if err != nil {
          log.Fatal(err)
        }

        fmt.Println("Tags in the remote image: ")
        if len(*remoteImageTags.Tags) == 0 {
          fmt.Println("No tags detected.")
        } else {
          for _, tag := range *remoteImageTags.Tags {
            fmt.Printf("'%v' with confidence %.2f%%\n", *tag.Name, *tag.Confidence*100)

            /*
            statement, _ := db.Prepare("INSERT INTO item_tag (item_id, timestamp,  value, score) VALUES (?, ?, ?, ?)")
            result, err := statement.Exec(item_id, timestamp,  *tag.Name, *tag.Confidence)
            fmt.Println("Entity - Last Insert ID")
            iid, err := result.LastInsertId()
            fmt.Println(iid)
            if err != nil {
              fmt.Println(err)
              return
            }
            */
            var idt string
            statement_tag := `INSERT INTO item_tag (item_id, timestamp, value, score) VALUES ($1, $2, $3, $4) RETURNING id`
            err = env.db.QueryRow(statement_tag, item_uuid, timestamp, *tag.Name, *tag.Confidence).Scan(&idt)
            if err != nil {
              fmt.Println(err)
              return
            }
          }
        }
        fmt.Println()
      }

      func DetectObjectsRemoteImage(env *Env,client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
        fmt.Println("-----------------------------------------")
        fmt.Println("DETECT OBJECTS - remote")
        fmt.Println()
        var remoteImage computervision.ImageURL
        remoteImage.URL = &remoteImageURL
        item_uuid,err := uuid.FromString(item_id)
        if err != nil {
          fmt.Printf("Something went wrong: %s", err)
          return
        }

        imageAnalysis, err := client.DetectObjects(
          computerVisionContext,
          remoteImage,
        )
        if err != nil {
          log.Fatal(err)
        }

        fmt.Println("Detecting objects in remote image: ")
        if len(*imageAnalysis.Objects) == 0 {
          fmt.Println("No objects detected.")
        } else {
          // Print the objects found with confidence level and bounding box locations.
          for _, object := range *imageAnalysis.Objects {
            fmt.Printf("'%v' with confidence %.2f%% at location (%v, %v), (%v, %v)\n",
            *object.Object, *object.Confidence*100,
            *object.Rectangle.X, *object.Rectangle.X+*object.Rectangle.W,
            *object.Rectangle.Y, *object.Rectangle.Y+*object.Rectangle.H)

            /*
            statement, _ := db.Prepare("INSERT INTO item_object (item_id, timestamp,  value, x, y, width, height, score) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
            result, err := statement.Exec(item_uuid, timestamp,  *object.Object, *object.Rectangle.X, *object.Rectangle.Y, *object.Rectangle.W, *object.Rectangle.H, *object.Confidence)
            fmt.Println("Entity - Last Insert ID")
            iid, err := result.LastInsertId()
            fmt.Println(iid)
            if err != nil {
              fmt.Println(err)
              return
            }
            */
            var iio string
            statement_o := `INSERT INTO item_object (item_id, timestamp, value, x, y, width, height, score ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
            err = env.db.QueryRow(statement_o,item_uuid, timestamp, *object.Object, *object.Rectangle.X, *object.Rectangle.Y, *object.Rectangle.W, *object.Rectangle.H, *object.Confidence).Scan(&iio)
            if err != nil {
              fmt.Println(err)
              return
            }
          }
        }
        fmt.Println()
      }

      func  DetectBrandsRemoteImage(env *Env,client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
        fmt.Println("-----------------------------------------")
        fmt.Println("DETECT BRANDS - remote")
        fmt.Println()
        var remoteImage computervision.ImageURL
        remoteImage.URL = &remoteImageURL
        item_uuid,err := uuid.FromString(item_id)
        if err != nil {
          fmt.Printf("Something went wrong: %s", err)
          return
        }

        // Define the kinds of features you want returned.
        features := []computervision.VisualFeatureTypes{computervision.VisualFeatureTypesBrands}

        imageAnalysis, err := client.AnalyzeImage(
          computerVisionContext,
          remoteImage,
          features,
          []computervision.Details{},
          "en")
          if err != nil {
            log.Fatal(err)
          }

          fmt.Println("Detecting brands in remote image: ")
          if len(*imageAnalysis.Brands) == 0 {
            fmt.Println("No brands detected.")
          } else {
            // Get bounding box around the brand and confidence level it's correctly identified.
            for _, brand := range *imageAnalysis.Brands {
              fmt.Printf("'%v' with confidence %.2f%% at location (%v, %v), (%v, %v)\n",
              *brand.Name, *brand.Confidence*100,
              *brand.Rectangle.X, *brand.Rectangle.X+*brand.Rectangle.W,
              *brand.Rectangle.Y, *brand.Rectangle.Y+*brand.Rectangle.H)

              /*
              statement, _ := db.Prepare("INSERT INTO item_brand (item_id,  timestamp, value, x, y, width, height, score) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
              result, err := statement.Exec(item_uuid,  timestamp, *brand.Name, *brand.Rectangle.X, *brand.Rectangle.Y, *brand.Rectangle.W, *brand.Rectangle.H, *brand.Confidence)
              fmt.Println("Entity - Last Insert ID")
              iid, err := result.LastInsertId()
              fmt.Println(iid)
              if err != nil {
                fmt.Println(err)
                return
              }
              */
              var iio string
              statement_brand := `INSERT INTO item_brand (item_id, timestamp, value, x, y, width, height, score ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
              err = env.db.QueryRow(statement_brand,item_uuid, timestamp, *brand.Name, *brand.Rectangle.X, *brand.Rectangle.Y, *brand.Rectangle.W, *brand.Rectangle.H, *brand.Confidence).Scan(&iio)
              if err != nil {
                fmt.Println(err)
                return
              }
            }
          }
          fmt.Println()
        }

        func DetectFacesRemoteImage(env *Env, client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
          fmt.Println("-----------------------------------------")
          fmt.Println("DETECT FACES - remote")
          fmt.Println()
          var remoteImage computervision.ImageURL
          remoteImage.URL = &remoteImageURL
          item_uuid,err := uuid.FromString(item_id)
          if err != nil {
            fmt.Printf("Something went wrong: %s", err)
            return
          }

          // Define the features you want returned with the API call.
          features := []computervision.VisualFeatureTypes{computervision.VisualFeatureTypesFaces}
          imageAnalysis, err := client.AnalyzeImage(
            computerVisionContext,
            remoteImage,
            features,
            []computervision.Details{},
            "")
            if err != nil {
              log.Fatal(err)
            }

            fmt.Println("Detecting faces in a remote image ...")
            if len(*imageAnalysis.Faces) == 0 {
              fmt.Println("No faces detected.")
            } else {
              // Print the bounding box locations of the found faces.
              for _, face := range *imageAnalysis.Faces {
                fmt.Printf("'%v' of age %v at location (%v, %v), (%v, %v)\n",
                face.Gender, *face.Age,
                *face.FaceRectangle.Left, *face.FaceRectangle.Top,
                *face.FaceRectangle.Left+*face.FaceRectangle.Width,
                *face.FaceRectangle.Top+*face.FaceRectangle.Height)

                //statement, _ := db.Prepare("INSERT INTO item_face (item_id, timestamp,  gender, age, position_left, position_top, position_width, position_height) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
                // result, err := statement.Exec(item_uuid,  timestamp, face.Gender, *face.Age, *face.FaceRectangle.Left, *face.FaceRectangle.Top, *face.FaceRectangle.Width, *face.FaceRectangle.Height)
                var idt string
                statement_tag := `INSERT INTO item_face (item_id, timestamp, gender, age, position_left, position_top, position_width, position_height) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
                err = env.db.QueryRow(statement_tag, item_uuid, timestamp, face.Gender, *face.Age,  *face.FaceRectangle.Left, *face.FaceRectangle.Top, *face.FaceRectangle.Width, *face.FaceRectangle.Height).Scan(&idt)
                if err != nil {
                  fmt.Println(err)
                  return
                }
              }
            }
            fmt.Println()
          }

          func DetectAdultOrRacyContentRemoteImage(env *Env,client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
            fmt.Println("-----------------------------------------")
            fmt.Println("DETECT ADULT OR RACY CONTENT - remote")
            fmt.Println()
            var remoteImage computervision.ImageURL
            remoteImage.URL = &remoteImageURL
            item_uuid,err := uuid.FromString(item_id)
            if err != nil {
              fmt.Printf("Something went wrong: %s", err)
              return
            }

            // Define the features you want returned from the API call.
            features := []computervision.VisualFeatureTypes{computervision.VisualFeatureTypesAdult}
            imageAnalysis, err := client.AnalyzeImage(
              computerVisionContext,
              remoteImage,
              features,
              []computervision.Details{},
              "") // language, English is default
              if err != nil {
                log.Fatal(err)
              }

              // Print whether or not there is questionable content.
              // Confidence levels: low means content is OK, high means it's not.
              fmt.Println("Analyzing remote image for adult or racy content: ")
              fmt.Printf("Is adult content: %v with confidence %.2f%%\n", *imageAnalysis.Adult.IsAdultContent, *imageAnalysis.Adult.AdultScore*100)


              var idt string
              statement_adult := `INSERT INTO item_adult (item_id, timestamp, value, score) VALUES ($1, $2, $3, $4) RETURNING id`
              err = env.db.QueryRow(statement_adult, item_uuid, timestamp, *imageAnalysis.Adult.IsAdultContent, *imageAnalysis.Adult.AdultScore).Scan(&idt)
              if err != nil {
                fmt.Println(err)
                return
              }

              fmt.Printf("Has racy content: %v with confidence %.2f%%\n", *imageAnalysis.Adult.IsRacyContent, *imageAnalysis.Adult.RacyScore*100)

              var idr string
              statement_racy := `INSERT INTO item_racy (item_id, timestamp, value, score) VALUES ($1, $2, $3, $4) RETURNING id`
              err = env.db.QueryRow(statement_racy, item_uuid, timestamp, *imageAnalysis.Adult.IsRacyContent, *imageAnalysis.Adult.RacyScore).Scan(&idr)
              if err != nil {
                fmt.Println(err)
                return
              }

              fmt.Println()
            }

            func DetectColorSchemeRemoteImage(env *Env, client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
              fmt.Println("-----------------------------------------")
              fmt.Println("DETECT COLOR SCHEME - remote")
              fmt.Println()
              var remoteImage computervision.ImageURL
              remoteImage.URL = &remoteImageURL
              item_uuid,err := uuid.FromString(item_id)
              if err != nil {
                fmt.Printf("Something went wrong: %s", err)
                return
              }

              // Define the features you'd like returned with the result.
              features := []computervision.VisualFeatureTypes{computervision.VisualFeatureTypesColor}
              imageAnalysis, err := client.AnalyzeImage(
                computerVisionContext,
                remoteImage,
                features,
                []computervision.Details{},
                "") // language, English is default
                if err != nil {
                  log.Fatal(err)
                }

                fmt.Println("Color scheme of the remote image: ")
                fmt.Printf("Is black and white: %v\n", *imageAnalysis.Color.IsBWImg)
                fmt.Printf("Accent color: 0x%v\n", *imageAnalysis.Color.AccentColor)
                fmt.Printf("Dominant background color: %v\n", *imageAnalysis.Color.DominantColorBackground)
                fmt.Printf("Dominant foreground color: %v\n", *imageAnalysis.Color.DominantColorForeground)
                fmt.Printf("Dominant colors: %v\n", strings.Join(*imageAnalysis.Color.DominantColors, ", "))
                fmt.Println()

                var idc string
                statement_color := `INSERT INTO item_color (item_id, timestamp,  black_and_white, accent_color, dominant_color_background, dominant_color_foreground, dominant_colors) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
                //err = db.QueryRow(statement_racy, item_uuid, timestamp, *imageAnalysis.Adult.IsRacyContent, *imageAnalysis.Adult.RacyScore).Scan(&idr)
                err = env.db.QueryRow(statement_color,item_uuid, timestamp,  *imageAnalysis.Color.IsBWImg, *imageAnalysis.Color.AccentColor, *imageAnalysis.Color.DominantColorBackground, *imageAnalysis.Color.DominantColorForeground, strings.Join(*imageAnalysis.Color.DominantColors, ", ")).Scan(&idc)
                if err != nil {
                  fmt.Println(err)
                  return
                }

              }

              func DetectDomainSpecificContentRemoteImage(env *Env, client computervision.BaseClient, remoteImageURL string, item_id string, timestamp int64) {
                fmt.Println("-----------------------------------------")
                fmt.Println("DETECT DOMAIN-SPECIFIC CONTENT - remote")
                fmt.Println()
                var remoteImage computervision.ImageURL
                remoteImage.URL = &remoteImageURL
                item_uuid,err := uuid.FromString(item_id)
                if err != nil {
                  fmt.Printf("Something went wrong: %s", err)
                  return
                }

                fmt.Println("Detecting domain-specific content in the local image ...")

                // Check if there are any celebrities in the image.
                celebrities, err := client.AnalyzeImageByDomain(
                  computerVisionContext,
                  "celebrities",
                  remoteImage,
                  "") // language, English is default
                  if err != nil {
                    log.Fatal(err)
                  }

                  fmt.Println("\nCelebrities: ")

                  // Marshal the output from AnalyzeImageByDomain into JSON.
                  data, err := json.MarshalIndent(celebrities.Result, "", "\t")
                  fmt.Println(string(data))

                  // Define structs for which to unmarshal the JSON.
                  type Celebrities struct {
                    Name       string  `json:"name"`
                    Confidence float64 `json:"confidence"`
                  }

                  type CelebrityResult struct {
                    Celebrities []Celebrities `json:"celebrities"`
                  }

                  var celebrityResult CelebrityResult

                  // Unmarshal the data.
                  err = json.Unmarshal(data, &celebrityResult)
                  if err != nil {
                    log.Fatal(err)
                  }

                  //	Check if any celebrities detected.
                  if len(celebrityResult.Celebrities) == 0 {
                    fmt.Println("No celebrities detected.")
                  } else {
                    for _, celebrity := range celebrityResult.Celebrities {
                      fmt.Printf("name: %v\n", celebrity.Name)
                      fmt.Printf("confidence: %.2f%%\n", celebrity.Confidence)

                      /*
                      statement, _ := db.Prepare("INSERT INTO item_celebrity (item_id, timestamp,  value, score ) VALUES (?, ?, ?, ?)")
                      result, err := statement.Exec(item_uuid, timestamp,  celebrity.Name, celebrity.Confidence)
                      fmt.Println("Entity - Last Insert ID")
                      iid, err := result.LastInsertId()
                      fmt.Println(iid)
                      if err != nil {
                        fmt.Println(err)
                        return
                      }
                      */
                      var iio string
                      statement_celebrity := `INSERT INTO item_celebrity (item_id, timestamp, value, score ) VALUES ($1, $2, $3, $4) RETURNING id`
                      err = env.db.QueryRow(statement_celebrity,item_uuid, timestamp, celebrity.Name, celebrity.Confidence).Scan(&iio)
                      if err != nil {
                        fmt.Println(err)
                        return
                      }
                    }
                  }
                  fmt.Println("\nLandmarks: ")

                  // Check if there are any landmarks in the image.
                  landmarks, err := client.AnalyzeImageByDomain(
                    computerVisionContext,
                    "landmarks",
                    remoteImage,
                    "")
                    if err != nil {
                      log.Fatal(err)
                    }

                    // Marshal the output from AnalyzeImageByDomain into JSON.
                    data, err = json.MarshalIndent(landmarks.Result, "", "\t")

                    // Define structs for which to unmarshal the JSON.
                    type Landmarks struct {
                      Name       string  `json:"name"`
                      Confidence float64 `json:"confidence"`
                    }

                    type LandmarkResult struct {
                      Landmarks []Landmarks `json:"landmarks"`
                    }

                    var landmarkResult LandmarkResult

                    // Unmarshal the data.
                    err = json.Unmarshal(data, &landmarkResult)
                    if err != nil {
                      log.Fatal(err)
                    }

                    // Check if any celebrities detected.
                    if len(landmarkResult.Landmarks) == 0 {
                      fmt.Println("No landmarks detected.")
                    } else {
                      for _, landmark := range landmarkResult.Landmarks {
                        fmt.Printf("name: %v\n", landmark.Name)
                        /*
                        statement, _ := db.Prepare("INSERT INTO item_landmark (item_id,  timestamp, value, score ) VALUES (?, ?, ?, ?)")
                        result, err := statement.Exec(item_uuid,  timestamp, landmark.Name, landmark.Confidence)
                        fmt.Println("Entity - Last Insert ID")
                        iid, err := result.LastInsertId()
                        fmt.Println(iid)
                        if err != nil {
                          fmt.Println(err)
                          return
                        }
                        */
                        var iio string
                        statement_landmark := `INSERT INTO item_lanmark (item_id, timestamp, value, score ) VALUES ($1, $2, $3, $4) RETURNING id`
                        err = env.db.QueryRow(statement_landmark,item_uuid, timestamp, landmark.Name, landmark.Confidence).Scan(&iio)
                        if err != nil {
                          fmt.Println(err)
                          return
                        }
                      }
                    }
                    fmt.Println()
                  }
