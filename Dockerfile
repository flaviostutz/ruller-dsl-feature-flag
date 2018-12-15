#COMPILE DSL TOOL
FROM golang:1.10 AS DSLTOOL

#doing dependency build separated from source build optimizes time for developer, but is not required
#install external dependencies first
ADD /main.go $GOPATH/src/ruller-sample-dsl/main.go
RUN go get -v ruller-sample-dsl

#now build source code
ADD ruller-sample-dsl $GOPATH/src/ruller-sample-dsl
RUN go get -v ruller-sample-dsl
#RUN go test -v ruller-sample-dsl


#GENERATE CODE FROM SAMPLE DSL USING DSLTOOL
FROM golang:1.10 as DSL2CODE

ENV LOG_LEVEL 'info'

COPY --from=DSLTOOL /go/bin/* /bin/
ADD ruller-sample-dsl/rules.tmpl /tmp
ADD ruller-sample-dsl/simple.tmpl /tmp
ADD ruller-sample-dsl/complex.tmpl /tmp

RUN ruller-sample-dsl --log-level=$LOG_LEVEL


#COMPILE GENERATED CODE
FROM golang:1.10 as CODECOMPILE

#just for build cache optimization
ADD /main.go $GOPATH/src/sample-rules/main.go
RUN go get -v sample-rules

COPY --from=DSL2CODE /opt/main.go $GOPATH/src/sample-rules/
RUN go get -v sample-rules

ADD /startup.sh /

CMD [ "/startup.sh" ]

