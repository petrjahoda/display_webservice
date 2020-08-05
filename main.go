package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/julienschmidt/sse"
	"github.com/kardianos/service"
	"net/http"
	"os"
)

const version = "2020.3.2.5"
const serviceName = "Display WebService"
const serviceDescription = "Display webpages, for use with big televisions and displays"
const config = "user=postgres password=Zps05..... dbname=version3 host=database port=5432 sslmode=disable"

type program struct{}

func main() {
	LogInfo("MAIN", serviceName+" ["+version+"] starting...")
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		LogError("MAIN", "Cannot start: "+err.Error())
	}
	err = s.Run()
	if err != nil {
		LogError("MAIN", "Cannot start: "+err.Error())
	}
}

func (p *program) Start(service.Service) error {
	LogInfo("MAIN", serviceName+" ["+version+"] started")
	go p.run()
	return nil
}

func (p *program) Stop(service.Service) error {
	LogInfo("MAIN", serviceName+" ["+version+"] stopped")
	return nil
}

func (p *program) run() {
	UpdateProgramVersion()
	router := httprouter.New()
	timer := sse.New()
	workplaces := sse.New()
	overview := sse.New()
	router.GET("/display_1", Display1)
	router.GET("/display_2", Display2)
	router.GET("/css/darcula.css", darcula)
	router.GET("/js/metro.min.js", metrojs)
	router.GET("/css/metro-all.css", metrocss)
	router.Handler("GET", "/time", timer)
	router.Handler("GET", "/workplaces", workplaces)
	router.Handler("GET", "/overview", overview)
	timezone := ReadTimeZoneFromDatabase()
	go StreamTime(timer, timezone)
	go StreamWorkplaces(workplaces)
	go StreamOverview(overview)
	err := http.ListenAndServe(":81", router)
	if err != nil {
		LogError("MAIN", "Problem starting service: "+err.Error())
		os.Exit(-1)
	}
	LogInfo("MAIN", serviceName+" ["+version+"] running")
}
