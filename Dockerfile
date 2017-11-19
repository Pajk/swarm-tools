FROM golang:1 as builder

WORKDIR /go/src/app

COPY get-deps.sh .

RUN bash get-deps.sh

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

COPY main.go .

RUN go-wrapper download
RUN go build -v -ldflags '-extldflags "-static" -s ' -o app

FROM scratch

COPY --from=builder /go/src/app/app /app

EXPOSE 80
CMD ["./app"]
