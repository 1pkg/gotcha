FROM golang:1.15

RUN mkdir -p $GOPATH/src/github.com/1pkg/gotcha
WORKDIR $GOPATH/src/github.com/1pkg/gotcha
ADD ./* ./
ADD ./vendor ./vendor
RUN go build -mod=vendor -o /var/gotcha

CMD ["/var/gotcha"]