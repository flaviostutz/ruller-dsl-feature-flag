FROM golang:1.10

RUN curl https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz --output /opt/GeoLite2-City.tar.gz
RUN cd /opt && \
    tar -xvf GeoLite2-City.tar.gz && \
    mv */GeoLite2-City.mmdb /opt/Geolite2-City.mmdb

#doing dependency build separated from source build optimizes time for developer, but is not required
#install external dependencies first
ADD main.dep $GOPATH/src/sample-scratch/main.go
RUN go get -v sample-scratch

#now build source code
ADD *.go $GOPATH/src/sample-scratch/
RUN go test -cover -v sample-scratch
RUN go get -v sample-scratch

ADD hugeids.txt /opt
ADD other.txt /opt

RUN cp /go/bin/* /bin/

CMD [ "/bin/sample-scratch" ]
