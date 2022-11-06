FROM golang:1.19-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /discord-bot-pi-locator

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /discord-bot-pi-locator /discord-bot-pi-locator

USER nonroot:nonroot

ENTRYPOINT ["/discord-bot-pi-locator"]