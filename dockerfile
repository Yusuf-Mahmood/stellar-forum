FROM golang:latest

LABEL version="1.0"
LABEL description="Stellar Forum is a website that allows users to communicate through posts and comments."

WORKDIR /app

COPY go.mod ./

COPY . .

RUN go build -o main ./cmd/main.go

EXPOSE 8080

CMD [ "./main" ]