FROM golang:alpine as build-env
COPY . /src
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 go build main.go

FROM alpine
WORKDIR /
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY --from=build-env /src/main /usr/local/bin/app
RUN chmod +x /usr/local/bin/app
ENTRYPOINT [ "/usr/local/bin/app" ]