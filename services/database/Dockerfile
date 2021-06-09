FROM golang:1.16.5-alpine

WORKDIR $GOPATH/src/github.com/thebluemagician/go-rest-scrape/services/database
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
EXPOSE 8086

CMD ["database"]
