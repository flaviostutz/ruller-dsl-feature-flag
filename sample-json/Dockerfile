#GENERATE CODE FROM SAMPLE JSON USING DSLTOOL
FROM flaviostutz/ruller-dsl-feature-flag as RULES_BUILD
ENV LOG_LEVEL 'info'
ADD /domains.json /opt/
ADD /menu.json /opt/
RUN ruller-dsl-feature-flag \
    --log-level=$LOG_LEVEL \
    --source=/opt/domains.json,/opt/menu.json \
    --target=/opt/main.go

#just for build cache optimization
ADD /main.dep $GOPATH/src/sample-json/main.go
RUN go get -v sample-json

#now build generated code
RUN cat -n /opt/main.go
RUN cp /opt/main.go $GOPATH/src/sample-json/
RUN go get -v sample-json



#RUNTIME CONTAINER
# FROM scratch
FROM golang:1.10
COPY --from=RULES_BUILD /go/bin/* /bin/
ADD /group1.txt /opt/group1.txt
ADD /startup.sh /
CMD [ "sh", "-C", "/startup.sh" ]

