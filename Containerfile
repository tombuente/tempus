FROM golang:latest

WORKDIR /bot

COPY . ./

RUN go mod download

RUN go build ./cmd/bot/bot.go

CMD ["./bot"]
