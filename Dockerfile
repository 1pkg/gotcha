FROM golang:1.15

RUN mkdir -p $GOPATH/src/github.com/1pkg/gotcha
WORKDIR $GOPATH/src/github.com/1pkg/gotcha
ADD ./* ./
ADD ./vendor ./vendor

CMD ["go", "test", "-mod=vendor", "-count=1", "-coverprofile", "test.cover", "./..."]
