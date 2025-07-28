FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "test", "./internal/scenarios/...", "-v"]
