FROM scratch
# MAINTAINER Ilia Martynov <ilyanix@gmail.com>
## 
## 
# STEP 1 build executable binary
FROM golang:alpine as builder
RUN apk add git

## get code
COPY . $GOPATH/src/github.com/ilyanix/nills
#RUN go get github.com/ilyanix/nills
#RUN git clone https://github.com/ilyanix/nills.git $GOPATH/src/github.com/ilyanix/nills
WORKDIR $GOPATH/src/github.com/ilyanix/nills

## get dependancies
RUN go get -d -v
#build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/nills
## 
# STEP 2 build a small image
# start from scratch
FROM scratch
EXPOSE 9080
ENTRYPOINT ["/go/bin/nills"]

# Copy our static executable
COPY --from=builder /go/bin/nills /go/bin/nills

### FROM scratch
### 
### 
### EXPOSE 9080
### 
### ENTRYPOINT ["/go/bin/nills"]
### 
### ADD nills /go/bin/nills
