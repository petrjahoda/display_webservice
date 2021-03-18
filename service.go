package main

import (
	"github.com/davidscholberg/go-durationfmt"
	"github.com/goodsign/monday"
	"github.com/julienschmidt/sse"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func readTimeZoneFromDatabase() string {
	logInfo("MAIN", "Reading timezone from database")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		return ""
	}
	var settings database.Setting
	db.Where("name=?", "timezone").Find(&settings)
	logInfo("MAIN", "Timezone read in "+time.Since(timer).String())
	return settings.Value
}

func updateProgramVersion() {
	logInfo("MAIN", "Writing program version into settings")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	var existingSettings database.Setting
	db.Where("name=?", serviceName).Find(&existingSettings)
	existingSettings.Name = serviceName
	existingSettings.Value = version
	db.Save(&existingSettings)
	logInfo("MAIN", "Program version written into settings in "+time.Since(timer).String())
}

func streamOverview(streamer *sse.Streamer) {
	logInfo("SSE", "Streaming overview process started")
	var workplaces []database.Workplace
	for {
		logInfo("SSE", "Streaming overview started")
		timer := time.Now()
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		workplaces = nil
		if err != nil {
			logError("SSE", "Problem opening database: "+err.Error())
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
			streamer.SendString("", "overview", "Production 0%;Downtime 0%;Poweroff 0%")
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
		logInfo("SSE", "Production: "+strconv.Itoa(productionPercent)+", Downtime: "+strconv.Itoa(downtimePercent)+", Offline: "+strconv.Itoa(offlinePercent))
		streamer.SendString("", "overview", "Production "+strconv.Itoa(productionPercent)+"%;Downtime "+strconv.Itoa(downtimePercent)+"%;Poweroff "+strconv.Itoa(offlinePercent)+"%")
		sqlDB, _ := db.DB()
		sqlDB.Close()
		logInfo("SSE", "Streaming overview ended in "+time.Since(timer).String())
		time.Sleep(10 * time.Second)
	}
}

func streamWorkplaces(streamer *sse.Streamer) {
	logInfo("SSE", "Streaming workplaces process started")
	var workplaces []database.Workplace
	for {
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		workplaces = nil
		if err != nil {
			logError("SSE", "Problem opening database: "+err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		sqlDB, _ := db.DB()
		logInfo("SSE", "Streaming workplaces running")
		timer := time.Now()
		db.Find(&workplaces)
		logInfo("SSE", "Workplaces count: "+strconv.Itoa(len(workplaces)))
		for _, workplace := range workplaces {
			stateRecord := database.StateRecord{}
			db.Where("workplace_id = ?", workplace.ID).Last(&stateRecord)
			orderRecord := database.OrderRecord{}
			db.Where("workplace_id = ?", workplace.ID).Where("date_time_end is null").Last(&orderRecord)
			workplaceHasOpenOrder := orderRecord.ID > 0
			downtimeRecord := database.DowntimeRecord{}
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
			logInfo(workplace.Name, "Workplace color: "+color+", order: "+order.Name+", downtime: "+downtime.Name+", user: "+userName)
			duration, err := durationfmt.Format(time.Now().Sub(stateRecord.DateTimeStart), "%dd %hh %mm")
			if err != nil {
				logError(workplace.Name, "Problem parsing datetime: "+err.Error())
			}
			streamer.SendString("", "workplaces", workplace.Name+";<b>"+workplace.Name+"</b><br>"+userName+"<br>"+order.Name+"<br>"+downtime.Name+"<br><br><sub>"+duration+"</sub>;"+color)
		}

		sqlDB.Close()
		logInfo("SSE", "Workplaces streamed in "+time.Since(timer).String())
		time.Sleep(10 * time.Second)
	}
}

func streamTime(streamer *sse.Streamer, timezone string) {
	logInfo("SSE", "Streaming time process started")
	for {
		if timezone == "" {
			timezone = readTimeZoneFromDatabase()
		}
		location, err := time.LoadLocation(timezone)
		if err != nil {
			logError("MAIN", "Problem loading location: "+timezone)
		} else {
			streamer.SendString("", "time", monday.Format(time.Now().In(location), "Monday, 2. January 2006, 15:04:05", monday.LocaleEnUS))
			time.Sleep(1 * time.Second)
		}
	}
}
