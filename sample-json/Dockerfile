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
ADD ../main.go $GOPATH/src/sample-rules/main.go
RUN go get -v sample-rules

RUN cp /opt/main.go $GOPATH/src/sample-rules/
RUN go get -v sample-rules



#RUNTIME CONTAINER
FROM scratch
COPY --from=RULES_BUILD /go/bin/* /bin/
ADD /startup.sh /
CMD [ "/startup.sh" ]

