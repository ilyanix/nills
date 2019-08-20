FROM scratch
MAINTAINER Ilia Martynov <ilyanix@gmail.com>


# STEP 1 build executable binary
FROM golang:alpine as builder
RUN apk add git
COPY . $GOPATH/src/github.com/ilyanix/nills
WORKDIR $GOPATH/src/github.com/ilyanix/nills
#get dependancies
RUN go get -d -v
#build the binary
RUN go build -o /go/bin/nills

# STEP 2 build a small image
# start from scratch
FROM alpine
# Copy our static executable
COPY --from=builder /go/bin/nills /go/bin/nills

EXPOSE 9080

ENTRYPOINT ["/go/bin/nills"]
