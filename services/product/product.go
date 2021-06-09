package main

import (
	"log"
	"bytes"
	"net/http"
	"strings"
	"strconv"
	"encoding/json"
	"time"

	"github.com/gorilla/mux"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

type Inner struct {
	Name			string		`json:"name,omitempty"`
	ImageURL		string		`json:"imageURL,omitempty"`
	Desc			string		`json:"description,omitempty"`
	Price			string		`json:"price,omitempty"`
	TotalReviews	int			`json:"totalReviews,omitempty"`
}

type Outer struct {
	URL				string	 	`json:"url,omitempty"`
	LastUpdate		time.Time 	`json:"lastUpdate,omitempty"`
	Product			Inner	 	`json:"product,omitempty"`
}

type StatusObject struct {
	InsertedID 		string		`json:"InsertedID,omitempty"`
	MatchedCount	int			`json:"MatchedCount,omitempty"`
    ModifiedCount	int			`json:"ModifiedCount,omitempty"`
}

type Response struct {
	URL 			string 		`json:"url"`
	Status			string		`json:status`
}

func scraper(url string) (res *colly.Response, err error, outer Outer) {

	log.Println("[product] Scraping Page, URL: ", url)
	product := Inner{}
	
	domain := "www.amazon.in"
	c := colly.NewCollector(
		colly.AllowedDomains(domain),
		colly.Async(true),
	)

	cookie := c.Cookies(url)
	err = c.SetCookies(url, cookie)
	if err != nil {
		return
	}

	//Randomize useragent to avoid bot detection
	extensions.RandomUserAgent(c)

	//Limit number of thread, prevents bot detection
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*amazon.*",
		Parallelism: 2,
		Delay:       2 * time.Second,
	})

	c.OnError(func(r *colly.Response, rerr error) {
		res = r
		err = rerr
		outer = Outer {}
		return
	})

	//Scraping tile
	c.OnHTML("#title",
		func(e *colly.HTMLElement) {
			_name := e.ChildText("#productTitle")
			if _name != "" {
				_name = ConvertHTMLEntities(_name)
				product.Name =  _name
			}
		})
	
	//Scraping Image URL
	c.OnHTML("#imgTagWrapperId",
		func(e *colly.HTMLElement) {
			_imageUrl := e.ChildAttr("img", "src")
			if _imageUrl != "" {
				_imageUrl = ConvertHTMLEntities(_imageUrl)
				product.ImageURL =  _imageUrl
			}
		})

	//Scraping Descriptions	
	c.OnHTML("#feature-bullets ul li",
		func(e *colly.HTMLElement) {
			_desc := e.ChildText("span")
			if _desc != "" {
				_desc = ConvertHTMLEntities(_desc)
				product.Desc =  _desc
			}
		})

	//Scraping Price
	c.OnHTML("#priceblock_ourprice",
		func(e *colly.HTMLElement) {
			_price := e.Text
			if _price != "" {
				_price = ConvertHTMLEntities(_price)
				product.Price =  _price
			}
		})

	//Scraping total reviews
	c.OnHTML("#acrCustomerReviewText",
		func(e *colly.HTMLElement) {
			_review := e.Text
			_temp := strings.ReplaceAll(_review, ",", "")
			_temp = strings.Split(_temp, " ")[0]
			totalReviews, _ := strconv.Atoi(_temp)
			product.TotalReviews = totalReviews
		})

	c.Visit(url)
	c.Wait()

	//Populating the structure
	outer = Outer {
		URL: 		url,
		Product:	product,
	}
	return
}

func handleScrape(writer http.ResponseWriter, request *http.Request){
	
	log.Println("[product] Request received; processing")
	decoder := json.NewDecoder(request.Body)
	form_data := Outer{}
	err := decoder.Decode(&form_data)
    if err != nil {
        log.Panic("[product] Unable to decode request: ", err)
	}

	res, err, data := scraper(form_data.URL)
	if err != nil {
		log.Println("[product]", res)
		log.Println("[product] Error in scrape func: ", err)
	}

	data.LastUpdate = time.Now()
	_data, err := json.Marshal(data)
	if err != nil {
		log.Fatal("[product] json.Marshal failed due to the error: ", err)
	}

	db_url := "http://database:8086/v1/db/create"
    requestObject, err := http.NewRequest("POST", db_url, bytes.NewBuffer(_data))
    requestObject.Header.Set("content-type", "application/json")

    client := &http.Client{}
    response, err := client.Do(requestObject)
    if err != nil {
        log.Panic("[product] Database request failed: ", err)
    }
    defer response.Body.Close()

	var status StatusObject
	_ = json.NewDecoder(response.Body).Decode(&status)

	httpStatus := Response {
		URL: 	form_data.URL,
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)

	if status.MatchedCount == 0 {
		httpStatus.Status = string("inserted")
		json.NewEncoder(writer).Encode(httpStatus)
	}

}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/v1/pd/scrape", handleScrape).Methods("POST")
    log.Fatal(http.ListenAndServe(":8085", router))
}
