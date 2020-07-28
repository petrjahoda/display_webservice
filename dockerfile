FROM alpine:latest
RUN apk update && apk upgrade && apk add bash&& apk add nano
WORKDIR /bin
COPY /css /bin/css
COPY /html /bin/html
COPY /js /bin/js
COPY /linux /bin
ENTRYPOINT display_webservice_linux