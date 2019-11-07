# Sidcloud docker image

FROM golang

RUN mkdir -p /root/.local/share/sidplayfp

WORKDIR /go/src/github.com/dkt64/sidcloud-api

COPY . .

COPY ./sidplayfp/libsidplayfp.so.5 /usr/local/lib

COPY ./sidplayfp/kernal /root/.local/share/sidplayfp
COPY ./sidplayfp/basic /root/.local/share/sidplayfp
COPY ./sidplayfp/chargen /root/.local/share/sidplayfp
COPY ./sidplayfp/Songlengths.txt /root/.local/share/sidplayfp

RUN ldconfig

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["sidcloud-api"]

EXPOSE 80
