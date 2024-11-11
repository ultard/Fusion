FROM golang:alpine AS build

WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    --mount=type=bind,target=. \
    go build -o /tmp/main ./app/cmd/server/main.go

FROM alpine:edge

WORKDIR /src

COPY ./app/templates ./templates
COPY --from=build /tmp/main .
RUN apk --no-cache add ca-certificates

ENTRYPOINT ["/src/main"]
