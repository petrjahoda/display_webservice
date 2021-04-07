package main

import (
	"github.com/goodsign/monday"
	"github.com/julienschmidt/httprouter"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type LcdWorkplaces struct {
	LcdWorkplaces     []LcdWorkplace
	ProductionPercent string
	DowntimePercent   string
	OfflinePercent    string
	ProductionColor   string
	DowntimeColor     string
	PowerOffColor     string
	Time              string
	Version           string
}

type LcdWorkplace struct {
	StateColor      string
	StateColorStyle string
	Name            string
	User            string
	Order           string
	Downtime        string
	Duration        string
}

func display1(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	logInfo("HTML", "Display 1 process started")
	timer := time.Now()
	tmpl := template.Must(template.ParseFiles("html/display_1.html"))
	var workplaces []database.Workplace
	lcdWorkplaces := LcdWorkplaces{}
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("HTML", "Problem opening database: "+err.Error())
	}
	db.Order("Name asc").Find(&workplaces)
	location := downloadActualLocation()
	loc, _ := time.LoadLocation(location)
	for _, workplace := range workplaces {
		stateRecord := database.StateRecord{}
		db.Where("workplace_id = ?", workplace.ID).Last(&stateRecord)
		state := database.State{}
		db.Where("id = ?", stateRecord.StateID).Find(&state)
		orderRecord := database.OrderRecord{}
		db.Where("workplace_id = ?", workplace.ID).Where("date_time_end is null").Last(&orderRecord)
		workplaceHasOpenOrder := orderRecord.ID > 0
		downtimeRecord := database.DowntimeRecord{}
		db.Where("workplace_id = ?", workplace.ID).Where("date_time_end is null").Last(&downtimeRecord)
		downtime := database.Downtime{}
		db.Where("id = ?", downtimeRecord.DowntimeID).Find(&downtime)
		var userName string
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
		colorStyle := "background: " + state.Color

		duration := time.Now().In(loc).Sub(stateRecord.DateTimeStart.In(loc)).Round(1 * time.Second).String()
		if err != nil {
			logError(workplace.Name, "Problem parsing datetime: "+err.Error())
		}
		logInfo(workplace.Name, duration)
		logInfo(workplace.Name, "Workplace color: "+color+", order: "+order.Name+", downtime: "+downtime.Name+", user: "+userName)
		logInfo("HTML", "Adding workplace: "+workplace.Name)
		lcdWorkplace := LcdWorkplace{Name: workplace.Name, User: userName, StateColor: color, Duration: duration, Downtime: downtime.Name, Order: order.Name, StateColorStyle: colorStyle}
		lcdWorkplaces.LcdWorkplaces = append(lcdWorkplaces.LcdWorkplaces, lcdWorkplace)
	}
	var states []database.State
	db.Find(&states)
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
	lcdWorkplaces.ProductionPercent = "Production " + strconv.Itoa(productionPercent) + "%"
	lcdWorkplaces.DowntimePercent = "Downtime " + strconv.Itoa(downtimePercent) + "%"
	lcdWorkplaces.OfflinePercent = "Poweroff " + strconv.Itoa(offlinePercent) + "%"
	for _, state := range states {
		switch state.Name {
		case "Production":
			lcdWorkplaces.ProductionColor = state.Color
		case "Downtime":
			lcdWorkplaces.DowntimeColor = state.Color
		case "Poweroff":
			lcdWorkplaces.PowerOffColor = state.Color
		}
	}
	lcdWorkplaces.Version = "version: " + version
	timezone := readTimeZoneFromDatabase()
	if timezone == "" {
		timezone = readTimeZoneFromDatabase()
	}
	if err != nil {
		logError("MAIN", "Problem loading location: "+timezone)
	} else {
		lcdWorkplaces.Time = monday.Format(time.Now().In(loc), "Monday, 2. January 2006, 15:04:05", monday.LocaleEnUS)
	}
	_ = tmpl.Execute(writer, lcdWorkplaces)
	logInfo("HTML", "Display 1 process ended in "+time.Since(timer).String())
}
