FROM golang:1.24@sha256:991aa6a6e4431f2f01e869a812934bd60fbc87fb939e4a1ea54b8494ab9d2fc6

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server github.com/tlscert/tlscert/server/cmd

EXPOSE 8080
EXPOSE 50051

CMD ["/server"]