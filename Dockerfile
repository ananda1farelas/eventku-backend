# syntax=docker/dockerfile:1

# Tahap 1: build. Image besar berisi Go compiler, HANYA untuk compile.
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app .

# Tahap 2: runtime. Hanya binary, tanpa compiler dan tanpa source code.
FROM alpine:3.24
RUN addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=build /out/app ./app

# Logger menulis ke logs/app.log, jadi foldernya harus bisa ditulis user app.
RUN mkdir logs && chown app:app logs
USER app

EXPOSE 3000
ENTRYPOINT ["./app"]