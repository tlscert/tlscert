FROM golang:1.24@sha256:d9db32125db0c3a680cfb7a1afcaefb89c898a075ec148fdc2f0f646cc2ed509

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server github.com/tlscert/tlscert/server/cmd

EXPOSE 8080
EXPOSE 50051

CMD ["/server"]