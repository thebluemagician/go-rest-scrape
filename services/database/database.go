package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	elastic "github.com/olivere/elastic/v7"
)

var elasticClient *elastic.Client

type Outer struct {
	URL string `json:"url,omitempty"`
}

type ResponseJson struct {
	Array []string
}

const mapping = `
{
    "settings": {
        "number_of_shards": 1,
        "number_of_replicas": 1
    }
}`

func GetESClient() {

	elasticClient, _ = elastic.NewClient(elastic.SetURL("http://elasticsearch:9200"),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false))

	log.Println("[elastic] Initialized...")
}

func createIndex() {

	ctx := context.Background()
	exists, err := elasticClient.IndexExists("amazonproducts").Do(ctx)
	if err != nil {
		log.Fatal("[elastic] Unable to check Index ", err)
	}
	if !exists {
		log.Println("[elastic] Creating Index")
		_, err := elasticClient.CreateIndex("amazonproducts").BodyString(mapping).Do(ctx)
		if err != nil {
			log.Panic("[elastic] Error in Index creation: ", err)
		}
	}
}

func InsertDocument(response http.ResponseWriter, request *http.Request) {

	createIndex()
	ctx := context.Background()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(response, "Error reading request body",
			http.StatusInternalServerError)
	}

	var data Outer
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println("[elastic] Error:", err)
	}

	termQuery := elastic.NewTermQuery("url.keyword", data.URL)
	generalQ := elastic.NewBoolQuery().Should().
		Filter(termQuery)

	searchResult, err := elasticClient.Search().
		Index("amazonproducts").
		Query(generalQ).
		From(0).Size(10).
		Pretty(true).
		Do(ctx)
	if err != nil {
		panic(err)
	}

	if searchResult.Hits.TotalHits.Value == 1 {
		for _, hit := range searchResult.Hits.Hits {
			log.Println("[elastic] Updating Id ", hit.Id)
			_, err = elasticClient.Index().
				Index("amazonproducts").
				Id(hit.Id).
				BodyJson(string(body)).
				Do(ctx)

			if err != nil {
				log.Panic("[elastic] Updating error ", err)
			}
		}
	} else {
		log.Println("[elastic] Creating index")
		_, err = elasticClient.Index().
			Index("amazonproducts").
			BodyJson(string(body)).
			Do(ctx)

		if err != nil {
			log.Panic("[elastic] Indexing error ", err)
		}
	}

	log.Println("[elastic] Insertion/update Successful")
}

func GetAllDocument(writer http.ResponseWriter, request *http.Request) {

	createIndex()
	ctx := context.Background()

	matchAllQ := elastic.NewMatchAllQuery()
	fsc := elastic.NewFetchSourceContext(true).Include("product", "url").Exclude("lastUpdate")

	searchResult, err := elasticClient.Search().
		Index("amazonproducts").
		Query(matchAllQ).
		FetchSourceContext(fsc).
		Pretty(true).
		Do(ctx)

	if err != nil {
		log.Panic("[elastic] Match query failed", err)
	}

	var rawJSONSlice []json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		rawJSONSlice = append(rawJSONSlice, hit.Source)
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(rawJSONSlice)
}

func main() {

	GetESClient()
	router := mux.NewRouter()
	router.HandleFunc("/v1/db/create", InsertDocument).Methods("POST")
	router.HandleFunc("/v1/db/product", GetAllDocument).Methods("GET")
	http.ListenAndServe(":8086", router)
}
