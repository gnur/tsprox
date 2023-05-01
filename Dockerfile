FROM golang:1.20 as build

RUN apt update && apt-get install upx -y
WORKDIR /workdir
COPY go.* /workdir/
RUN go mod download
COPY *.go /workdir/
COPY *.html /workdir/

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -a -installsuffix cgo -o tsprox .
RUN upx tsprox


FROM gcr.io/distroless/static-debian11
COPY --from=build /workdir/tsprox /tsprox
CMD ["/tsprox"]
