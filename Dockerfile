FROM golang:1.26-alpine AS build

LABEL maintainer="Kypello <dev@kypello.io>"

ENV CGO_ENABLED=0

WORKDIR /src

RUN apk add -U --no-cache ca-certificates git

# The LICENSE and CREDITS shipped in the image come from this tree, not from a
# curl of minio/mc master: the fork carries its own NOTICE, and pulling those
# files over the network made the image contents depend on an upstream branch.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -tags kqueue \
      -ldflags "$(go run buildscripts/gen-ldflags.go)" \
      -o /go/bin/kc .

FROM scratch

LABEL maintainer="Kypello <dev@kypello.io>"

# go-homedir shells out to `getent` when HOME is unset, and a scratch image has
# no getent -- kc then fails to resolve its config directory and exits 1 before
# running anything. Setting HOME keeps that lookup in-process.
ENV HOME=/

COPY --from=build /go/bin/kc      /usr/bin/kc
COPY --from=build /src/CREDITS    /licenses/CREDITS
COPY --from=build /src/LICENSE    /licenses/LICENSE
COPY --from=build /src/NOTICE     /licenses/NOTICE
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["kc"]
