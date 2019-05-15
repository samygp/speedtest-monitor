FROM alpine

RUN apk add ca-certificates

ADD speedtest-monitor /
EXPOSE 8080



ENTRYPOINT ["/speedtest-monitor"]