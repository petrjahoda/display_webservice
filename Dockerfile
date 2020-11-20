FROM alpine:latest as build
RUN apk add tzdata

FROM scratch as final
COPY /css /css
COPY /html html
COPY /js js
COPY /fonts fonts
COPY /linux /
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
CMD ["/display_webservice"]