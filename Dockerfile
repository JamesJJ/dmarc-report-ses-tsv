ARG BASE_IMAGE=gcr.io/distroless/base

#-----------------------------------------#

FROM $BASE_IMAGE:debug as os_config_image

ARG BASE_IMAGE

RUN ["adduser","-h","/","-s","/sbin/nologin","-D","-H","app_user"]

#-----------------------------------------#

FROM golang:1.16-stretch as build_image

ARG BASE_IMAGE

WORKDIR /go/src/app

COPY . .

RUN go build -o /go/bin/app -v

RUN go test -v -cover

#-----------------------------------------#

FROM $BASE_IMAGE as service_image

ARG BASE_IMAGE

LABEL base_image=$BASE_IMAGE

LABEL owner_team=OPS

COPY --from=os_config_image --chown=root:root /etc/passwd /etc/group /etc/

COPY --from=build_image --chown=root:root /go/bin/app /app

USER app_user:app_user

ENV PODREADY_VERBOSE=true

CMD ["/app"]
