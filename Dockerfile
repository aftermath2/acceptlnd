FROM golang:1.21-alpine3.19 as builder

WORKDIR /app

COPY go.mod .

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go build -o acceptlnd -ldflags="-s -w" .

# ---

FROM scratch

COPY --from=builder /app/acceptlnd /acceptlnd

ENTRYPOINT ["/acceptlnd"]
