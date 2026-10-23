# Imagen única para la nube: compila el frontend (React + PWA) y el backend (Go)
# y los sirve desde el mismo servidor en el puerto $PORT.
#   docker build -t multicatalogo .
#   docker run -p 3000:3000 -e DATABASE_URL=... -e JWT_SECRET=... multicatalogo

# 1. Frontend
FROM node:22-alpine AS frontend
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY index.html vite.config.js eslint.config.js pwa-assets.config.js ./
COPY public ./public
COPY src ./src
# Mismo origen: la app llama a /api en el mismo dominio.
ENV VITE_API_URL=
RUN npm run build

# 2. Backend
FROM golang:1.26-alpine AS backend
WORKDIR /app
COPY multicatalogo-backend/go.mod multicatalogo-backend/go.sum ./
RUN go mod download
COPY multicatalogo-backend/ ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server .

# 3. Imagen final: solo el ejecutable y la carpeta web
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /server /app/server
COPY --from=frontend /app/dist /app/web
ENV STATIC_DIR=/app/web \
    APP_ENV=prod \
    PORT=3000
EXPOSE 3000
CMD ["/app/server"]
