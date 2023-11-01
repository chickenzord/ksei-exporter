FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY main.go ./
COPY internal internal
RUN CGO_ENABLED=0 go build -trimpath -o /bin/ksei-exporter .

FROM alpine:edge
RUN apk add -U curl ca-certificates && update-ca-certificates
COPY --from=build /bin/ksei-exporter /bin/ksei-exporter
CMD [ "/bin/ksei-exporter" ]
