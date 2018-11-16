FROM alpine:3.8

ADD ./peek /peek

ENTRYPOINT ["/peek"]
