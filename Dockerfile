FROM alpine:latest

WORKDIR /work

ADD ./huimonitor /work/main

CMD ["./main"]

