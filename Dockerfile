FROM golang:1.21

ENV GO111MODULE=on

ADD . /usr/local/go/src/prismDifference
WORKDIR /usr/local/go/src/prismDifference
RUN go mod download && go mod verify 
RUN go build -v

CMD ["app"]