# Dockerfile
# docker build -t feedocean-rss2full:latest .
# docker run -d -p 8088:8088 --name rss2full feedocean-rss2full
FROM golang AS builder

ARG RSS2FULL_VERSION="0.1.2"

ENV CGO_ENABLED 0

RUN curl -fsSLO https://github.com/feedocean/rss2full/archive/v${RSS2FULL_VERSION}.tar.gz && \
      tar zvxf v${RSS2FULL_VERSION}.tar.gz -C /go/src/ && \
      mv /go/src/rss2full-${RSS2FULL_VERSION} /go/src/rss2full

# Compile rss2full
RUN cd /go/src/rss2full && go build

FROM alpine
MAINTAINER feedocean.com
WORKDIR /root
COPY --from=builder /go/src/rss2full/wwwroot /root/wwwroot
COPY --from=builder /go/src/rss2full/rss2full /root/rss2full

# Server port to listen
EXPOSE 8088
CMD ["/root/rss2full"]