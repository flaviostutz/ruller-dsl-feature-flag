FROM golang:1.10

#doing dependency build separated from source build optimizes time for developer, but is not required
#install external dependencies first
ADD /main.dep $GOPATH/src/ruller-dsl-feature-flag/main.go
RUN go get -v ruller-dsl-feature-flag

#now build source code
ADD ruller-dsl-feature-flag $GOPATH/src/ruller-dsl-feature-flag
ADD ruller-dsl-feature-flag/templates /opt/templates
RUN go get -v ruller-dsl-feature-flag

RUN cp /go/bin/* /bin/
