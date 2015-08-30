FROM alpine:3.2
RUN apk add --update ca-certificates # Certificates for SSL
ADD build/cmd/try6d_linux_amd64.bin /usr/local/bin/try6d
ENTRYPOINT /usr/local/bin/try6d
VOLUME ["/var/lib/try6", "/etc/try6d/certs"]
EXPOSE 9000

#FROM golang:1.5.0

#ENV APP github.com/jllopis/try6
#ADD . /go/src/${APP}
#WORKDIR /go/src/${APP}
#RUN GO15VENDOREXPERIMENT=1 go install ${APP}/cmd/try6d \
#    && mkdir /var/lib/try6

#VOLUME ["/var/lib/try6", "/etc/try6d/certs"]

#EXPOSE 9000
#ENTRYPOINT /go/bin/try6d
