FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache curl make

COPY . .

RUN make bin/migrate
RUN make build OUT=/app/bin/udb

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache curl make

COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/Makefile /app/Makefile
COPY --from=builder /app/bin/migrate /app/bin/migrate
COPY --from=builder /app/bin/udb /app/bin/udb

CMD ["/app/bin/udb"]
