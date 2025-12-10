#Multi-stage сборка, сначала готовим бинарники
FROM golang:1.24.4-alpine AS stage

RUN mkdir /app
WORKDIR /app


COPY ./config /app/config 
COPY ./go.mod /app
COPY ./go.sum /app
RUN go mod download

COPY ./db /app/db 
COPY ./migrations /app/migrations
COPY ./forms /app/forms
COPY ./rabbit /app/rabbit
COPY ./models /app/models 
COPY ./handlers /app/handlers 
COPY ./main.go /app


RUN go build

# Создаем образ с бинарником и без исходников кода
FROM alpine:latest

WORKDIR /app

COPY --from=stage /app/web_backend_v2 /app/web_backend_v2

COPY --from=stage  /app/migrations /app/migrations

EXPOSE 8080
CMD /app/web_backend_v2