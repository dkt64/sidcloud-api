FROM golang

WORKDIR /go/src/github.com/dkt64/sidcloud-api
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["sidcloud-api"]

EXPOSE 80
