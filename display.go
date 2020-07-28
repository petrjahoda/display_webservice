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

func Display1(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	LogInfo("MAIN", "Displaying Display 1")
	tmpl := template.Must(template.ParseFiles("html/display_1.html"))
	var workplaces []database.Workplace
	lcdWorkplaces := LcdWorkplaces{}
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
	}
	db.Order("Name asc").Find(&workplaces)
	for _, workplace := range workplaces {
		LogInfo("MAIN", "Adding workplace: "+workplace.Name)
		lcdWorkplace := LcdWorkplace{Name: workplace.Name, User: "loading...", StateColor: "", Duration: 0 * time.Hour, Downtime: "", Order: ""}
		lcdWorkplaces.LcdWorkplaces = append(lcdWorkplaces.LcdWorkplaces, lcdWorkplace)
	}
	lcdWorkplaces.Version = "version: " + version
	_ = tmpl.Execute(writer, lcdWorkplaces)
}

func Display2(writer http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	LogInfo("MAIN", "Displaying Display 1")
	tmpl := template.Must(template.ParseFiles("html/display_2.html"))
	var workplaces []database.Workplace
	lcdWorkplaces := LcdWorkplaces{}
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
	}
	db.Order("Name asc").Find(&workplaces)
	for _, workplace := range workplaces {
		LogInfo("MAIN", "Adding workplace: "+workplace.Name)
		lcdWorkplace := LcdWorkplace{Name: workplace.Name, User: "loading...", StateColor: "", Duration: 0 * time.Hour, Downtime: "", Order: ""}
		lcdWorkplaces.LcdWorkplaces = append(lcdWorkplaces.LcdWorkplaces, lcdWorkplace)
	}
	lcdWorkplaces.Version = "version: " + version
	_ = tmpl.Execute(writer, lcdWorkplaces)
}
