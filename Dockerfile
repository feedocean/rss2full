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

RUN apk update && apk --no-cache add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /app
COPY --from=builder /go/src/rss2full/wwwroot /app/wwwroot
COPY --from=builder /go/src/rss2full/rss2full /app/rss2full

# Server port to listen
ENV PORT 8088

ENTRYPOINT ["/app/rss2full"]

EXPOSE ${PORT}
CMD [""]