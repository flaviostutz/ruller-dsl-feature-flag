FROM golang:1.10

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
