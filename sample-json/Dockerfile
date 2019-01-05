#GENERATE CODE FROM SAMPLE JSON USING DSLTOOL
FROM flaviostutz/ruller-dsl-feature-flag as BUILD
ENV LOG_LEVEL 'info'

RUN curl https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz --output /opt/GeoLite2-City.tar.gz
RUN cd /opt && \
    tar -xvf GeoLite2-City.tar.gz && \
    mv */GeoLite2-City.mmdb /opt/Geolite2-City.mmdb

ADD /domains.json /opt/
ADD /menu.json /opt/
RUN ruller-dsl-feature-flag \
    --log-level=debug \
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
COPY --from=BUILD /go/bin/* /bin/
COPY --from=BUILD /opt/Geolite2-City.mmdb /opt/
ADD /group1.txt /opt/group1.txt
ADD /startup.sh /
CMD [ "sh", "-C", "/startup.sh" ]

