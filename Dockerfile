FROM amd64/alpine:latest

WORKDIR /work

ADD ./bin/linux-amd64-hui-monitor /work/main

CMD ["./main"]

