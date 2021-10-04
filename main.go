package main

import (
	"encoding/json"
	"github.com/hyperboloide/lk"
	"github.com/julienschmidt/httprouter"
	"github.com/julienschmidt/sse"
	"github.com/kardianos/service"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"time"
)

const version = "2021.4.1.4"
const serviceName = "Display WebService"
const serviceDescription = "Display webpages, for use with big televisions and displays"
const config = "user=postgres password=pj79.. dbname=system host=database port=5432 sslmode=disable application_name=display_webservice"

type program struct{}

func main() {
	logInfo("MAIN", serviceName+" ["+version+"] starting...")
	logInfo("MAIN", "© "+strconv.Itoa(time.Now().Year())+" Petr Jahoda")
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
	programIsActive := false
	for !programIsActive {
		programIsActive = checkActivation(programIsActive)
		logInfo("MAIN", serviceName+": licence is not valid")
		time.Sleep(10 * time.Second)
	}
	updateProgramVersion()
	router := httprouter.New()
	timer := sse.New()
	workplaces := sse.New()
	overview := sse.New()
	router.ServeFiles("/js/*filepath", http.Dir("js"))
	router.ServeFiles("/css/*filepath", http.Dir("css"))
	router.ServeFiles("/fonts/*filepath", http.Dir("fonts"))

	router.GET("/display_1", display1)

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

func checkActivation(programIsActive bool) bool {
	if programIsActive {
		return true
	}
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		return false
	}
	var customerName database.Setting
	var softwareLicense database.Setting
	db.Where("name = ?", "company").Find(&customerName)
	db.Where("name = ?", serviceName).Find(&softwareLicense)
	const publicKeyBase32 = "ARIVIK3FHZ72ERWX6FQ6Z3SIGHPSMCDBRCONFKQRWSDIUMEEESQULEKQ7J7MZVFZMJDFO6B46237GOZETQ4M2NE32C3UUNOV5EUVE3OIV72F5LQRZ6DFMM6UJPELARG7RLJWKQRATUWD5YT46Q2TKQMPPGIA===="
	publicKey, err := lk.PublicKeyFromB32String(publicKeyBase32)
	if err != nil {
		return false
	}
	license, err := lk.LicenseFromB32String(softwareLicense.Note)
	if err != nil {
		return false
	}
	if ok, err := license.Verify(publicKey); err != nil {
		return false
	} else if !ok {
		return false
	}
	result := struct {
		Software string `json:"software"`
		Customer string `json:"customer"`
	}{}
	if err := json.Unmarshal(license.Data, &result); err != nil {
		return false
	}
	if result.Customer == customerName.Value && result.Software == softwareLicense.Name {
		logInfo("MAIN", serviceName+": licence is valid")
		return true
	}
	return false
}
