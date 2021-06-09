# go-rest-scrape
This is a simple **Go** project which scrapes few details like- product name, description, price and review count from the Amazon (amazon.in) India product catalogue and stores in into document oriented database namely **ElasticSearch**.

## Project modules

### 1. Product
This module fetches data from Amazon product catalogue, processes it, structures in the required json.

### 2. Database
Handles the request from product module to store the data in database.
Also has API to get all the stored product details

## Build and Installation

### Prerequisites
The project is developed and tested on following environments and platform:
**Sno.** | **Name** | **Version/Config.**
-------: | :------: | :------------------
1 | Operating System | Kubuntu 20.10 (Linux iudx 5.8.0-25-generic)
2 | Language | go version go1.16.5 linux/amd64
3 | IDE | Visual Studio Code Version 1.56.2
4 | Containerization | Docker version 20.10.5, docker-compose version 1.28.5
5 | Document Database | elasticsearch:7.7.0
6 | API Testing | Postman v7.36.5

### Build & Install
1. Clone or download the repository
2. Change the directory
3. Run docker-compose build and up command 

```
cd /path/to/go-rest/scrape
docker-compose build
docker-compose up #run in foreground
docker-compose up -d #to run in background
docker ps #check the running containers
```

### API Usage 
#### Scrape the Product Details
##### Request
`POST /v1/pd/scrape`
```
curl --location --request POST '127.0.0.1:8085/v1/pd/scrape' \
--header 'Content-Type: application/json' \
--data-raw '{
    "url":"https://www.amazon.in/Airdopes-281-Bluetooth-Immersive-Resistance/dp/B084DRKGZ3"
}'
```
##### Response
```
{
    "url": "https://www.amazon.in/Airdopes-281-Bluetooth-Immersive-Resistance/dp/B084DRKGZ3",
    "Status": "inserted"
}
```
  </br> 
  
#### Get All the Products (DB)
##### Request
`GET /v1/db/product`
```
curl --location --request GET '127.0.0.1:8086/v1/db/product'
```
##### Response
```
[
    {
        "product": {
            "totalReviews": 9277,
            "price": "₹ 1,799.00",
            "imageURL": "https://images-eu.ssl-images-amazon.com/images/I/31SDq6jpxmL._SY300_SX300_QL70_ML2_.jpg",
            "name": "boAt Airdopes 281 Bluetooth Truly Wireless Earbuds with Mic(Furious Blue)",
            "description": "1 year warranty from the date of purchase, you can activate your warranty by giving a missed call on 9223032222. Alternatively you can claim your warranty at support.boat-lifestyle.com or reach out to us at +912249461882/info@imaginemarketingindia.com."
        },
        "url": "https://www.amazon.in/Airdopes-281-Bluetooth-Immersive-Resistance/dp/B084DRKGZ3"
    }
]
```

