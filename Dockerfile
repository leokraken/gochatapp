FROM golang:1.7-alpine

WORKDIR /home/app
RUN apk update && \
    apk add git && \
    mkdir -p /home/app
ADD . /home/app/
RUN /home/app/scripts/install.sh

CMD go run /home/app/chatapp.go


