FROM 1.24-alpine⁠ AS builder

WORKDIR /app/


RUN apk add --no-cache vips-dev alpine-sdk make inkscape

RUN apk add --no-cache \
    font-terminus font-inconsolata font-dejavu font-noto font-noto-cjk \
    font-awesome font-noto-extra font-vollkorn font-misc-cyrillic \
    font-mutt-misc font-screen-cyrillic font-winitzki-cyrillic \
    font-cronyx-cyrillic tzdata

COPY . .

RUN make bin/migrate
RUN make build OUT=/app/bin/udb

FROM alpine:3.19

WORKDIR /app/

RUN apk add --no-cache curl make inkscape font-terminus font-inconsolata font-dejavu font-noto font-noto-cjk \
    font-awesome font-noto-extra font-vollkorn font-misc-cyrillic \
    font-mutt-misc font-screen-cyrillic font-winitzki-cyrillic \
    font-cronyx-cyrillic tzdata

COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/Makefile /app/Makefile
COPY --from=builder /app/bin/migrate /app/bin/migrate
COPY --from=builder /app/bin/udb /app/bin/udb
COPY --from=builder /app/root.crt /app/root.crt
CMD ["/app/bin/udb"]
