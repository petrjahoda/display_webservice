#!/usr/bin/env bash
cd linux
upx display_webservice_linux
cd ..
docker rmi -f petrjahoda/display_webservice:latest
docker build -t petrjahoda/display_webservice:latest .
docker push petrjahoda/display_webservice:latest

docker rmi -f petrjahoda/display_webservice:2020.3.3
docker build -t petrjahoda/display_webservice:2020.3.3 .
docker push petrjahoda/display_webservice:2020.3.3
