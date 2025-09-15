FROM golang:1.24@sha256:87916acb3242b6259a26deaa7953bdc6a3a6762a28d340e4f1448e7b5c27c009

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server github.com/tlscert/tlscert/server/cmd

EXPOSE 50051

CMD ["/server"]