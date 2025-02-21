FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CONFIG_PATH=configs/values_local_docker.yaml

RUN go build -o /build ./cmd/server

EXPOSE 8080

CMD ["/build"]
