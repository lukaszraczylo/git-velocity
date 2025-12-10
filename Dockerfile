FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata git

COPY git-velocity /usr/local/bin/git-velocity
COPY config.example.yaml /etc/git-velocity/config.example.yaml

RUN chmod +x /usr/local/bin/git-velocity

ENTRYPOINT ["/usr/local/bin/git-velocity"]
CMD ["--help"]
