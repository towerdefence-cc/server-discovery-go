FROM golang:alpine as go
WORKDIR /app
ENV GO111MODULE=on

COPY go.mod .
RUN go mod download

COPY . .
RUN go build -o server-discovery .

FROM alpine

COPY --from=go /app/server-discovery /app/server-discovery
CMD ["/app/server-discovery"]