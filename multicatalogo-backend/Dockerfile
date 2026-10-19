# Imagen del backend en Go (la usa el servicio "backend" de docker-compose.yml).

# Etapa 1: compilación.
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server .

# Etapa 2: imagen final, solo con el ejecutable.
FROM alpine:3.20
COPY --from=build /server /server
EXPOSE 3000
CMD ["/server"]
