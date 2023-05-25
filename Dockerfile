FROM golang:1.19-alpine as buildEnv
ARG VERSION=0.0.0
RUN mkdir -p /opt/auth
COPY . /opt/auth/
WORKDIR /opt/auth
RUN go build -ldflags="-X main.Version=${VERSION}" -o ./build/auth ./apps/main/main.go

FROM alpine:latest
RUN mkdir -p /opt/auth
WORKDIR /opt/auth
COPY --from=buildEnv /opt/auth/build/auth ./auth
EXPOSE 9000
CMD ./auth
