# build stage
FROM ubuntu:bionic AS builder

ENV GO_VERSION="1.19.3"

RUN apt-get update
RUN apt-get install -y wget git gcc

RUN wget -P /tmp "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz"

RUN tar -C /usr/local -xzf "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
RUN rm "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

WORKDIR $GOPATH

WORKDIR /app

COPY . .

RUN go build -o main .

# main stage
FROM ubuntu:bionic

WORKDIR /app

COPY --from=builder /app/main /app/commservice.linux /app/customers.csv ./

CMD ["/bin/bash", "-c", "./commservice.linux & ./main"]
