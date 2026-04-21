FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/todo-api ./

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/todo-api /app/todo-api
USER nonroot:nonroot
EXPOSE 9090
ENTRYPOINT ["/app/todo-api"]
