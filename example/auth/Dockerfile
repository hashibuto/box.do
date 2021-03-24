FROM golang:1.16.2-alpine3.12 AS builder

COPY auth/ /app/auth/
WORKDIR /app/auth
RUN go build

FROM alpine:3.7
WORKDIR /app/auth
COPY --from=builder /app/auth/auth .

ENTRYPOINT ["auth"]