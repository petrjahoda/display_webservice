package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"net/http"
	"time"
)

type LcdWorkplaces struct {
	LcdWorkplaces []LcdWorkplace
	Version       string
}

type LcdWorkplace struct {
	StateColor string
	Name       string
	User       string
	Order      string
	Downtime   string
	Duration   time.Duration
}

func display1(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	logInfo("HTML", "Display 1 process started")
	timer := time.Now()
	tmpl := template.Must(template.ParseFiles("html/display_1.html"))
	var workplaces []database.Workplace
	lcdWorkplaces := LcdWorkplaces{}
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		logError("HTML", "Problem opening database: "+err.Error())
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	db.Order("Name asc").Find(&workplaces)
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
		logInfo(workplace.Name, "Workplace color: "+color+", order: "+order.Name+", downtime: "+downtime.Name+", user: "+userName)
		logInfo("HTML", "Adding workplace: "+workplace.Name)
		lcdWorkplace := LcdWorkplace{Name: workplace.Name, User: userName, StateColor: color, Duration: time.Now().Sub(stateRecord.DateTimeStart), Downtime: downtime.Name, Order: order.Name}
		lcdWorkplaces.LcdWorkplaces = append(lcdWorkplaces.LcdWorkplaces, lcdWorkplace)
	}
	lcdWorkplaces.Version = "version: " + version
	_ = tmpl.Execute(writer, lcdWorkplaces)
	logInfo("HTML", "Display 1 process ended in "+time.Since(timer).String())
}

func display2(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	logInfo("HTML", "Display 2 process started")
	timer := time.Now()
	tmpl := template.Must(template.ParseFiles("html/display_2.html"))
	var workplaces []database.Workplace
	lcdWorkplaces := LcdWorkplaces{}
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		logError("HTML", "Problem opening database: "+err.Error())
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	db.Order("Name asc").Find(&workplaces)
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
		logInfo(workplace.Name, "Workplace color: "+color+", order: "+order.Name+", downtime: "+downtime.Name+", user: "+userName)
		logInfo("HTML", "Adding workplace: "+workplace.Name)
		lcdWorkplace := LcdWorkplace{Name: workplace.Name, User: userName, StateColor: color, Duration: time.Now().Sub(stateRecord.DateTimeStart), Downtime: downtime.Name, Order: order.Name}
		lcdWorkplaces.LcdWorkplaces = append(lcdWorkplaces.LcdWorkplaces, lcdWorkplace)
	}
	lcdWorkplaces.Version = "version: " + version
	_ = tmpl.Execute(writer, lcdWorkplaces)
	logInfo("HTML", "Display 2 process ended in "+time.Since(timer).String())
}
