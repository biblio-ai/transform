package main

import (
	//"flag"
	"context"
	"database/sql"
	"fmt"
	//"github.com/Azure/azure-sdk-for-go/services/preview/cognitiveservices/v3.0-preview/computervision"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v2.0/computervision"
	"github.com/Azure/go-autorest/autorest"
	"log"
	"os"
	//"strings"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ilyakaznacheev/cleanenv"
	//"github.com/satori/go.uuid"
	//"net/http"
	_ "net/url"
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
	Port     int    `yaml:"port" env-default:"5431"  env-description:"Database host"`
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
	fmt.Println("Get Staging Data")

	//stg_rows, err := env.db.Query("SELECT url as file_url,header_identifier, date_latest, metadata_identifier, metadata_identifier_handle_id, metadata_identifier_cms_id, metadata_identifier_accession_id, metadata_identifier_file_id from stg_slv_primo where metadata_identifier like 'IE%'")
	stg_rows_url, err := env.db.Query("SELECT distinct url as stg_url from stg_slv_primo")
	if err != nil {
		// handle this error better than this
		panic(err)
	}
	defer stg_rows_url.Close()
	var (
		stg_url                          string
		header_identifier                string
		date_latest                      string
		metadata_identifier              string
		metadata_identifier_handle_id    string
		metadata_identifier_cms_id       string
		metadata_identifier_accession_id string
		metadata_identifier_file_id      string
	)

	sum := 0
	for stg_rows_url.Next() {

		err = stg_rows_url.Scan(&stg_url)
		if err != nil {
			// handle this error
			panic(err)
		}

		sum += 1
		fmt.Println("Row:")
		fmt.Println(sum)

		var item_id string
		var inserted_item_id string

		imageURL := stg_url

		var url string

		stmt := `select id, url from item where url = $1 limit 1`
		row := env.db.QueryRow(stmt, imageURL)
		switch err := row.Scan(&item_id, &url); err {
		case sql.ErrNoRows:
			fmt.Println("No rows were returned!")
		case nil:
			fmt.Println(item_id, url)
		default:
			panic(err)
		}
		if len(item_id) == 0 {

			fmt.Println("URL not found - getting all data")
			fmt.Println(imageURL)

			stmt_all := `SELECT url,header_identifier, date_latest, metadata_identifier, metadata_identifier_handle_id, metadata_identifier_cms_id, metadata_identifier_accession_id, metadata_identifier_file_id from stg_slv_primo where url = $1 limit 1`
			stg_rows_all := env.db.QueryRow(stmt_all, stg_url)
			err = stg_rows_all.Scan(&url, &header_identifier, &date_latest, &metadata_identifier, &metadata_identifier_handle_id, &metadata_identifier_cms_id, &metadata_identifier_accession_id, &metadata_identifier_file_id)
			//fmt.Println(value)
			if err != nil {
				// handle this error
				panic(err)
			}

			fmt.Println("Prepare insert")
			fmt.Println(url)
			statement := `INSERT INTO item (url) VALUES ($1) RETURNING id`
			err = env.db.QueryRow(statement, imageURL).Scan(&inserted_item_id)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Last Insert ID")
			fmt.Println(inserted_item_id)

			fmt.Println("Prepare insert - file_url")
			metadata_key := "url"
			metadata_value := url
			fmt.Println(imageURL)
			statement_metadata := `INSERT INTO item_metadata (item_id, metadata_key, metadata_value) VALUES ($1,$2,$3)`
			_, err = env.db.Exec(statement_metadata, inserted_item_id, metadata_key, metadata_value)
			if err != nil {
				fmt.Println(err)
				return
			}

			if len(url) > 0 {
				fmt.Println("Prepare insert - url")
				metadata_key := "url"
				metadata_value := url
				statement_metadata := `INSERT INTO item_metadata (item_id, metadata_key, metadata_value) VALUES ($1,$2,$3)`
				_, err = env.db.Exec(statement_metadata, inserted_item_id, metadata_key, metadata_value)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			if len(header_identifier) > 0 {
				fmt.Println("Prepare insert - header_identifier")
				metadata_key := "header_identifier"
				metadata_value := header_identifier
				statement_metadata := `INSERT INTO item_metadata (item_id, metadata_key, metadata_value) VALUES ($1,$2,$3)`
				_, err = env.db.Exec(statement_metadata, inserted_item_id, metadata_key, metadata_value)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if len(date_latest) > 0 {
				fmt.Println("Prepare insert - date_latest")
				metadata_key := "date_latest"
				metadata_value := date_latest
				statement_metadata := `INSERT INTO item_metadata (item_id, metadata_key, metadata_value) VALUES ($1,$2,$3)`
				_, err = env.db.Exec(statement_metadata, inserted_item_id, metadata_key, metadata_value)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if len(metadata_identifier) > 0 {
				fmt.Println("Prepare insert - metadata_identifier")
				metadata_key := "metadata_identifier"
				metadata_value := metadata_identifier
				statement_metadata := `INSERT INTO item_metadata (item_id, metadata_key, metadata_value) VALUES ($1,$2,$3)`
				_, err = env.db.Exec(statement_metadata, inserted_item_id, metadata_key, metadata_value)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if len(metadata_identifier_handle_id) > 0 {
				fmt.Println("Prepare insert - metadata_identifier_handle_id")
				metadata_key := "metadata_identifier_handle_id"
				metadata_value := metadata_identifier_handle_id
				statement_metadata := `INSERT INTO item_metadata (item_id, metadata_key, metadata_value) VALUES ($1,$2,$3)`
				_, err = env.db.Exec(statement_metadata, inserted_item_id, metadata_key, metadata_value)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if len(metadata_identifier_cms_id) > 0 {
				fmt.Println("Prepare insert - metadata_identifier_cms_id")
				metadata_key := "metadata_identifier_cms_id"
				metadata_value := metadata_identifier_cms_id
				statement_metadata := `INSERT INTO item_metadata (item_id, metadata_key, metadata_value) VALUES ($1,$2,$3)`
				_, err = env.db.Exec(statement_metadata, inserted_item_id, metadata_key, metadata_value)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if len(metadata_identifier_accession_id) > 0 {
				fmt.Println("Prepare insert - metadata_identifier_accession_id")
				metadata_key := "metadata_identifier_accession_id"
				metadata_value := metadata_identifier_accession_id
				statement_metadata := `INSERT INTO item_metadata (item_id, metadata_key, metadata_value) VALUES ($1,$2,$3)`
				_, err = env.db.Exec(statement_metadata, inserted_item_id, metadata_key, metadata_value)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			if len(metadata_identifier_file_id) > 0 {
				fmt.Println("Prepare insert - metadata_identifier_file_id")
				metadata_key := "metadata_identifier_file_id"
				metadata_value := metadata_identifier_file_id
				statement_metadata := `INSERT INTO item_metadata (item_id, metadata_key, metadata_value) VALUES ($1,$2,$3)`
				_, err = env.db.Exec(statement_metadata, inserted_item_id, metadata_key, metadata_value)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
		if len(item_id) > 0 {

			fmt.Println("Skip item - already in table insert")
			fmt.Println("Len:" + fmt.Sprint(len(item_id)))
			fmt.Println("Item ID:" + item_id)
			fmt.Println("URL:" + imageURL)
			continue
		}

		fmt.Println("Item ID")
		fmt.Println(inserted_item_id)
		fmt.Println("Item URL")
		fmt.Println(imageURL)
		timestamp := time.Now().Unix()
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
		BatchReadFileRemoteImage(env, computerVisionClient, imageURL, inserted_item_id, timestamp)

		// Analyze features of an image, remote
		DescribeRemoteImage(env, computerVisionClient, imageURL, inserted_item_id, timestamp)
		CategorizeRemoteImage(env, computerVisionClient, imageURL, inserted_item_id, timestamp)
		TagRemoteImage(env, computerVisionClient, imageURL, inserted_item_id, timestamp)
		DetectFacesRemoteImage(env, computerVisionClient, imageURL, inserted_item_id, timestamp)
		DetectObjectsRemoteImage(env, computerVisionClient, imageURL, inserted_item_id, timestamp)
		// DetectBrandsRemoteImage(env,computerVisionClient, imageURL, inserted_item_id, timestamp)
		//DetectAdultOrRacyContentRemoteImage(env,computerVisionClient, imageURL, inserted_item_id, timestamp)
		//DetectColorSchemeRemoteImage(env,computerVisionClient, imageURL, inserted_item_id, timestamp)
		//DetectDomainSpecificContentRemoteImage(env,computerVisionClient, imageURL, inserted_item_id, timestamp)

		//text_analytic(inserted_item_id)
		DetectLanguage(env, inserted_item_id, timestamp)
		ExtractEntities(env, inserted_item_id, timestamp)
		//SentimentAnalysis(env,inserted_item_id, timestamp)
		ExtractKeyPhrases(env, inserted_item_id, timestamp)
		Get(env, inserted_item_id)

	}

}
