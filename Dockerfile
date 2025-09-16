# --- build stage ---
    FROM golang:1.24 AS build
    WORKDIR /app
    COPY go.mod go.sum ./
    RUN go mod download
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /lg-cs2-bot
    
    # --- runtime stage ---
    FROM gcr.io/distroless/static:nonroot
    WORKDIR /app
    # Use nonroot user provided by distroless, but we will switch to root in compose if needed
    COPY .env /app/.env
    COPY data/faceit_player_names.json /app/data/faceit_player_names.json
    COPY --from=build /lg-cs2-bot /app/lg-cs2-bot
    ENTRYPOINT ["/app/lg-cs2-bot"]
    