FROM golang:1.14.3-alpine3.11

WORKDIR /app
ADD /go.mod /app/
ADD /go.sum /app/
RUN go mod download

#now build source code
ADD / /app/
RUN go build -o /bin/ruller-dsl-feature-flag

ENV CONDITION_DEBUG false

