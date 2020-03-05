#!/usr/bin/env bash
cd linux
upx display_webservice_linux
cd ..
cd mac
upx display_webservice_mac
cd ..
cd windows
upx display_wenservice_windows.exe
cd ..
docker rmi -f petrjahoda/display_webservice:"$1"
docker build -t petrjahoda/display_webservice:"$1" .
docker push petrjahoda/display_webservice:"$1"