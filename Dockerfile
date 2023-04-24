FROM golang:1.20 as build
WORKDIR /workdir
COPY go.* /workdir/
RUN go mod download
COPY *.go /workdir/

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o tsprox .


FROM gcr.io/distroless/static-debian11
COPY --from=build /workdir/tsprox /tsprox
CMD ["/tsprox"]
