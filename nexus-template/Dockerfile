# Build context: nexus-backend root; one image supports every role of this service.
FROM golang:1.26.8-alpine AS build
ARG SERVICE
ARG RELEASE=development
RUN apk add --no-cache ca-certificates git
WORKDIR /src
COPY . .
RUN test -n "$SERVICE" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service ./services/${SERVICE}-service

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 nexus
WORKDIR /app
COPY --from=build /out/service /app/service
ARG SERVICE
COPY services/${SERVICE}-service/migrations /app/migrations
USER 10001:10001
EXPOSE 9100
ENTRYPOINT ["/app/service"]
