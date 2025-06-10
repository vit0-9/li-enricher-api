FROM golang:1.24.4-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux
ENV GO111MODULE=on

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -a -installsuffix cgo -o main .


FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser

COPY --from=builder /app/main .

EXPOSE 3000

CMD ["./main"]
