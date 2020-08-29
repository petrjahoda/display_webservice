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
		logInfo("HTML", "Adding workplace: "+workplace.Name)
		lcdWorkplace := LcdWorkplace{Name: workplace.Name, User: "loading...", StateColor: "", Duration: 0 * time.Hour, Downtime: "", Order: ""}
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
		logInfo("HTML", "Adding workplace: "+workplace.Name)
		lcdWorkplace := LcdWorkplace{Name: workplace.Name, User: "loading...", StateColor: "", Duration: 0 * time.Hour, Downtime: "", Order: ""}
		lcdWorkplaces.LcdWorkplaces = append(lcdWorkplaces.LcdWorkplaces, lcdWorkplace)
	}
	lcdWorkplaces.Version = "version: " + version
	_ = tmpl.Execute(writer, lcdWorkplaces)
	logInfo("HTML", "Display 2 process ended in "+time.Since(timer).String())
}
