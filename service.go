package main

import (
	"github.com/davidscholberg/go-durationfmt"
	"github.com/goodsign/monday"
	"github.com/julienschmidt/httprouter"
	"github.com/julienschmidt/sse"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

func ReadTimeZoneFromDatabase() string {
	LogInfo("MAIN", "Reading timezone from database")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
		return ""
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var settings database.Setting
	db.Where("name=?", "timezone").Find(&settings)
	LogInfo("MAIN", "Timezone read in "+time.Since(timer).String())
	return settings.Value
}

func UpdateProgramVersion() {
	LogInfo("MAIN", "Writing program version into settings")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var existingSettings database.Setting
	db.Where("name=?", serviceName).Find(&existingSettings)
	existingSettings.Name = serviceName
	existingSettings.Value = version
	db.Save(&existingSettings)
	LogInfo("MAIN", "Program version written into settings in "+time.Since(timer).String())
}

func StreamOverview(streamer *sse.Streamer) {
	LogInfo("SSE", "Streaming overview process started")
	var workplaces []database.Workplace
	for {
		LogInfo("SSE", "Streaming overview started")
		timer := time.Now()
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		workplaces = nil
		if err != nil {
			LogError("SSE", "Problem opening database: "+err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		db.Find(&workplaces)
		production := 0
		downtime := 0
		offline := 0
		for _, workplace := range workplaces {
			stateRecord := database.StateRecord{}
			db.Where("workplace_id = ?", workplace.ID).Last(&stateRecord)
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
		LogInfo("SSE", "Production: "+strconv.Itoa(productionPercent)+", Downtime: "+strconv.Itoa(downtimePercent)+", Offline: "+strconv.Itoa(offlinePercent))
		streamer.SendString("", "overview", "Produkce "+strconv.Itoa(productionPercent)+"%;Prostoj "+strconv.Itoa(downtimePercent)+"%;Vypnuto "+strconv.Itoa(offlinePercent)+"%")
		sqlDB, err := db.DB()
		sqlDB.Close()
		LogInfo("SSE", "Streaming overview ended in "+time.Since(timer).String())
		time.Sleep(10 * time.Second)
	}
}

func StreamWorkplaces(streamer *sse.Streamer) {
	LogInfo("SSE", "Streaming workplaces process started")
	var workplaces []database.Workplace
	for {
		LogInfo("SSE", "Streaming workplaces running")
		timer := time.Now()
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		workplaces = nil
		if err != nil {
			LogError("SSE", "Problem opening database: "+err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		db.Find(&workplaces)
		LogInfo("SSE", "Workplaces count: "+strconv.Itoa(len(workplaces)))
		for _, workplace := range workplaces {
			stateRecord := database.StateRecord{}
			db.Where("workplace_id = ?", workplace.ID).Last(&stateRecord)
			orderRecord := database.OrderRecord{}
			db.Where("workplace_id = ?", workplace.ID).Where("date_time_end is null").Last(&orderRecord)
			workplaceHasOpenOrder := orderRecord.ID > 0
			downtimeRecord := database.DownTimeRecord{}
			db.Where("workplace_id = ?", workplace.ID).Where("date_time_end is null").Last(&downtimeRecord)
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
			streamer.SendString("", "workplaces", workplace.Name+";"+workplace.Name+"<br>"+userName+"<br>"+downtime.Name+"<br>"+order.Name+"<span class=\"badge-bottom\">"+duration+"</span>;"+color)
		}
		sqlDB, err := db.DB()
		sqlDB.Close()
		LogInfo("SSE", "Workplaces streamed in "+time.Since(timer).String())
		time.Sleep(10 * time.Second)
	}
}

func StreamTime(streamer *sse.Streamer, timezone string) {
	LogInfo("SSE", "Streaming time process started")
	for {
		if timezone == "" {
			timezone = ReadTimeZoneFromDatabase()
		}
		location, err := time.LoadLocation(timezone)
		if err != nil {
			LogError("MAIN", "Problem loading location: "+timezone)
		} else {
			streamer.SendString("", "time", monday.Format(time.Now().In(location), "Monday, 2. January 2006, 15:04:05", monday.LocaleCsCZ))
			time.Sleep(1 * time.Second)
		}
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
