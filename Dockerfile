FROM golang
ADD . /go/src/github.com/tobyjsullivan/ues-command-api
RUN  go install github.com/tobyjsullivan/ues-command-api
CMD /go/bin/ues-command-api