package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/julienschmidt/sse"
	"github.com/kardianos/service"
	"net/http"
	"os"
)

const version = "2020.4.2.17"
const serviceName = "Display WebService"
const serviceDescription = "Display webpages, for use with big televisions and displays"
const config = "user=postgres password=Zps05..... dbname=version3 host=database port=5432 sslmode=disable"

type program struct{}

func main() {
	logInfo("MAIN", serviceName+" ["+version+"] starting...")
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		logError("MAIN", "Cannot start: "+err.Error())
	}
	err = s.Run()
	if err != nil {
		logError("MAIN", "Cannot start: "+err.Error())
	}
}

func (p *program) Start(service.Service) error {
	logInfo("MAIN", serviceName+" ["+version+"] started")
	go p.run()
	return nil
}

func (p *program) Stop(service.Service) error {
	logInfo("MAIN", serviceName+" ["+version+"] stopped")
	return nil
}

func (p *program) run() {
	updateProgramVersion()
	router := httprouter.New()
	timer := sse.New()
	workplaces := sse.New()
	overview := sse.New()
	router.GET("/display_1", display1)
	router.GET("/display_2", display2)
	router.GET("/css/darcula.css", darcula)
	router.GET("/js/metro.min.js", metrojs)
	router.GET("/css/metro-all.css", metrocss)
	router.GET("/js/metro.min.js.map", metrominjsmap)
	router.Handler("GET", "/time", timer)
	router.Handler("GET", "/workplaces", workplaces)
	router.Handler("GET", "/overview", overview)
	timezone := readTimeZoneFromDatabase()
	go streamTime(timer, timezone)
	go streamWorkplaces(workplaces)
	go streamOverview(overview)
	err := http.ListenAndServe(":81", router)
	if err != nil {
		logError("MAIN", "Problem starting service: "+err.Error())
		os.Exit(-1)
	}
	logInfo("MAIN", serviceName+" ["+version+"] running")
}

func metrominjsmap(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	http.ServeFile(writer, request, "js/metro.min.js.map")
}
