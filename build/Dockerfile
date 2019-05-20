FROM golang:1.12 as build

WORKDIR /go/src/web
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM gcr.io/distroless/base

#s Default mongo url
ENV MONGOURL="mongodb://myawesomecomputer:27017"

COPY --from=build /go/bin/web /
CMD ["/web"]