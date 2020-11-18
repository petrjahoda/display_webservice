[![developed_using](https://img.shields.io/badge/developed%20using-Jetbrains%20Goland-lightgrey)](https://www.jetbrains.com/go/)
<br/>
![GitHub](https://img.shields.io/github/license/petrjahoda/display_webservice)
[![GitHub last commit](https://img.shields.io/github/last-commit/petrjahoda/display_webservice)](https://github.com/petrjahoda/display_webservice/commits/master)
[![GitHub issues](https://img.shields.io/github/issues/petrjahoda/display_webservice)](https://github.com/petrjahoda/display_webservice/issues)
<br/>
![GitHub language count](https://img.shields.io/github/languages/count/petrjahoda/display_webservice)
![GitHub top language](https://img.shields.io/github/languages/top/petrjahoda/display_webservice)
![GitHub repo size](https://img.shields.io/github/repo-size/petrjahoda/display_webservice)
<br/>
[![Docker Pulls](https://img.shields.io/docker/pulls/petrjahoda/display_webservice)](https://hub.docker.com/r/petrjahoda/display_webservice)
[![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/petrjahoda/display_webservice?sort=date)](https://hub.docker.com/r/petrjahoda/display_webservice/tags)
<br/>
[![developed_using](https://img.shields.io/badge/database-PostgreSQL-red)](https://www.postgresql.org) [![developed_using](https://img.shields.io/badge/runtime-Docker-red)](https://www.docker.com)

# Display WebService

## Description
Go web service, that shows web pages on port 81
* `/display_1` shows all workplaces with their statuses in small tiles

## Installation Information
Install under docker runtime using [this dockerfile image](https://github.com/petrjahoda/system/tree/master/latest) with this command: ```docker-compose up -d```

## Implementation Information
Check the software running with this command: ```docker stats```. <br/>
Display_webservice has to be running.

## Developer Information
Use software only as a [part of a system](https://github.com/petrjahoda/system) using Docker runtime.<br/>
 Do not run under linux, windows or mac on its own.

© 2020 Petr Jahoda

Example
![Display](screenshots/actual.png)
