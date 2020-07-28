package main

import (
	"github.com/davidscholberg/go-durationfmt"
	"github.com/goodsign/monday"
	"github.com/julienschmidt/httprouter"
	"github.com/julienschmidt/sse"
	"github.com/kardianos/service"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

const version = "2020.3.1.28"
const programName = "Display WebService"
const programDescription = "Display webpages, for use with big televisions and displays"
const config = "user=postgres password=Zps05..... dbname=version3 host=database port=5432 sslmode=disable"

type program struct{}

func (p *program) Start(s service.Service) error {
	LogInfo("MAIN", "Starting "+programName+" on "+s.Platform())
	go p.run()
	return nil
}

func (p *program) run() {
	LogInfo("MAIN", programName+" version "+version+" started")
	WriteProgramVersionIntoSettings()
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
	go StreamTime(timer)
	go StreamWorkplaces(workplaces)
	go StreamOverview(overview)
	LogInfo("MAIN", "Server running")
	_ = http.ListenAndServe(":81", router)
}
func (p *program) Stop(s service.Service) error {
	LogInfo("MAIN", "Stopped on platform "+s.Platform())
	return nil
}

func main() {
	serviceConfig := &service.Config{
		Name:        programName,
		DisplayName: programName,
		Description: programDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		LogError("MAIN", err.Error())
	}
	err = s.Run()
	if err != nil {
		LogError("MAIN", "Problem starting "+serviceConfig.Name)
	}
}

func WriteProgramVersionIntoSettings() {
	LogInfo("MAIN", "Updating program version in database")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	var settings database.Setting
	db.Where("name=?", programName).Find(&settings)
	settings.Name = programName
	settings.Value = version
	db.Save(&settings)
	LogInfo("MAIN", "Program version updated, elapsed: "+time.Since(timer).String())
}

func StreamOverview(streamer *sse.Streamer) {
	var workplaces []database.Workplace
	for {
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		LogInfo("MAIN", "Streaming overview")
		workplaces = nil
		if err != nil {
			LogError("MAIN", "Problem opening database: "+err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		db.Find(&workplaces)
		production := 0
		downtime := 0
		offline := 0
		for _, workplace := range workplaces {
			stateRecord := database.StateRecord{}
			db.Where("workplace_id = ?", workplace.ID).Where("date_time_end is null").Find(&stateRecord)
			switch stateRecord.StateID {
			case 1:
				production++
			case 2:
				downtime++
			case 3:
				offline++
			}
		}
		sum := production + offline + downtime
		if sum == 0 {
			streamer.SendString("", "overview", "Produkce 0%;Prostoj 0%;Vypnuto 0%")
			time.Sleep(10 * time.Second)
			continue
		}
		LogInfo("MAIN", "Production: "+strconv.Itoa(production)+", Downtime: "+strconv.Itoa(downtime)+", Offline: "+strconv.Itoa(offline))
		productionPercent := production * 100 / sum
		downtimePercent := downtime * 100 / sum
		offlinePercent := 100 - productionPercent - downtimePercent
		floatPointMiscalculation := offline == 0 && offlinePercent > 0
		if floatPointMiscalculation {
			offlinePercent = 0
			if downtimePercent > productionPercent {
				downtimePercent++
			} else {
				productionPercent++
			}
		}
		LogInfo("MAIN", "Production: "+strconv.Itoa(productionPercent)+", Downtime: "+strconv.Itoa(downtimePercent)+", Offline: "+strconv.Itoa(offlinePercent))
		streamer.SendString("", "overview", "Produkce "+strconv.Itoa(productionPercent)+"%;Prostoj "+strconv.Itoa(downtimePercent)+"%;Vypnuto "+strconv.Itoa(offlinePercent)+"%")
		time.Sleep(10 * time.Second)
	}
}

func StreamWorkplaces(streamer *sse.Streamer) {
	var workplaces []database.Workplace
	for {
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		LogInfo("MAIN", "Streaming workplaces")
		workplaces = nil
		if err != nil {
			LogError("MAIN", "Problem opening database: "+err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		db.Find(&workplaces)
		LogInfo("MAIN", "Workplaces count: "+strconv.Itoa(len(workplaces)))
		for _, workplace := range workplaces {
			stateRecord := database.StateRecord{}
			db.Where("workplace_id = ?", workplace.ID).Where("date_time_end is null").Find(&stateRecord)
			orderRecord := database.OrderRecord{}
			db.Where("workplace_id = ?", workplace.ID).Where("date_time_end is null").Find(&orderRecord)
			workplaceHasOpenOrder := orderRecord.ID > 0
			downtimeRecord := database.DownTimeRecord{}
			db.Where("workplace_id = ?", workplace.ID).Where("date_time_end is null").Find(&downtimeRecord)
			downtime := database.Downtime{}
			db.Where("id = ?", downtimeRecord.DowntimeID).Find(&downtime)
			userName := ""
			order := database.Order{}
			if workplaceHasOpenOrder {
				db.Where("id = ?", orderRecord.OrderID).Find(&order)
				userRecord := database.UserRecord{}
				db.Where("order_record_id = ?", orderRecord.ID).Find(&userRecord)
				user := database.User{}
				db.Where("id = ?", userRecord.UserID).Find(&user)
				userName = user.FirstName + " " + user.SecondName
			}
			color := "green"
			switch stateRecord.StateID {
			case 1:
				color = "green"
			case 2:
				color = "orange"
			case 3:
				color = "red"
			}
			LogInfo(workplace.Name, "Workplace color: "+color+", order: "+order.Name+", downtime: "+downtime.Name+", user: "+userName)
			duration, err := durationfmt.Format(time.Now().Sub(stateRecord.DateTimeStart), "%dd %hh %mm")
			if err != nil {
				LogError(workplace.Name, "Problem parsing datetime: "+err.Error())
			}
			LogInfo(workplace.Name, "Streaming data to LCD")
			streamer.SendString("", "workplaces", workplace.Name+";"+workplace.Name+"<br>"+userName+"<br>"+downtime.Name+"<br>"+order.Name+"<span class=\"badge-bottom\">"+duration+"</span>;"+color)
			LogInfo(workplace.Name, "Data streamed")
		}
		LogInfo("MAIN", "Workplaces streamed, waiting 10 second for another run")
		time.Sleep(10 * time.Second)
	}
}

func StreamTime(streamer *sse.Streamer) {
	for {
		streamer.SendString("", "time", monday.Format(time.Now(), "Monday, 2. January 2006 15:04:05", monday.LocaleCsCZ))
		time.Sleep(1 * time.Second)
	}
}

func darcula(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	http.ServeFile(writer, request, "css/darcula.css")
}

func metrojs(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	http.ServeFile(writer, request, "js/metro.min.js")
}

func metrocss(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	http.ServeFile(writer, request, "css/metro-all.css")
}
