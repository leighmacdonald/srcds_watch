FROM golang:1.20-alpine as build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build

FROM alpine:latest
LABEL maintainer="Leigh MacDonald <leigh.macdonald@gmail.com>"
LABEL org.opencontainers.image.source="https://github.com/leighmacdonald/srcds_watch"
EXPOSE 8877
RUN apk add dumb-init
WORKDIR /app
COPY --from=build /build/srcds_watch .
ENTRYPOINT ["dumb-init", "--"]
CMD ["./srcds_watch"]
