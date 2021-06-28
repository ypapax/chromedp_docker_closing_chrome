FROM golang:latest as build-env
RUN mkdir $GOPATH/src/app
WORKDIR $GOPATH/src/app
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
COPY main.go .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /root/app ./*.go

FROM zenika/alpine-chrome
#RUN apk update
#RUN apk upgrade
#RUN apk add bash
COPY --from=build-env /root/app /
COPY entrypoint.sh entrypoint.sh
RUN ls && pwd
#RUN chmod +x entrypoint.sh
#RUN mkdir -p /headless-shell/swiftshader/ \
#    && cd /headless-shell/swiftshader/ \
#    && ln -s ../libEGL.so libEGL.so \
#    && ln -s ../libGLESv2.so libGLESv2.so
ENTRYPOINT ["/usrc/src/app/entrypoint.sh"]