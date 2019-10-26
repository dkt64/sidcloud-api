# Sidcloud docker image

FROM golang

WORKDIR /go/src/github.com/dkt64/sidcloud-api

COPY . .

COPY ./sidplayfp/libsidplayfp.so.5 /usr/local/lib
RUN ldconfig

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["sidcloud-api"]

EXPOSE 80
