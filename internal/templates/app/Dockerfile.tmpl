# Simple development Dockerfile for the generated app
FROM golang:1.21-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o /bin/server ./cmd/server

FROM alpine:3.20
ENV PORT=8080
EXPOSE 8080
COPY --from=build /bin/server /server
ENTRYPOINT ["/server"]

