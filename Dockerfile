FROM golang:alpine as build-env
RUN apk --no-cache add git
RUN go get github.com/btcsuite/btcd/rpcclient
RUN go get github.com/btcsuite/btcd/chaincfg/chainhash
RUN go get github.com/btcsuite/btcd/wire
RUN go get github.com/gorilla/mux
RUN go get github.com/paulbellamy/ratecounter
RUN mkdir -p /go/src/github.com/gertjaap/ocm-backend
ADD . /go/src/github.com/gertjaap/ocm-backend
WORKDIR /go/src/github.com/gertjaap/ocm-backend
RUN go get ./...
RUN go build -o ocm-backend

# final stage
FROM alpine
RUN apk --no-cache add ca-certificates libzmq
WORKDIR /app
COPY --from=build-env /go/src/github.com/gertjaap/ocm-backend/ocm-backend /app/
ENTRYPOINT ./ocm-backend