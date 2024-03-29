package main

import (
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
	for {
		logInfo("SSE", "Streaming overview started")
		timer := time.Now()
		production, downtime, offline := downloadDataForStreaming()
		sum := production + offline + downtime
		if sum == 0 {
			streamer.SendString("", "overview", "Production 0%;Downtime 0%;Poweroff 0%")
			continue
		} else {
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
			streamer.SendString("", "overview", "Production "+strconv.Itoa(productionPercent)+"%;Downtime "+strconv.Itoa(downtimePercent)+"%;Poweroff "+strconv.Itoa(offlinePercent)+"%")
		}
		logInfo("SSE", "Streaming overview ended in "+time.Since(timer).String())
		time.Sleep(10 * time.Second)
	}
}

func downloadDataForStreaming() (int, int, int) {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return 0, 0, 0
	}
	production := 0
	downtime := 0
	offline := 0
	var stateRecords []database.StateRecord
	// select * from (select state_records.id as id, states.name as state, workplaces.name as workplace from state_records join workplaces on state_records.workplace_id=workplaces.id join states on state_records.state_id=states.id) as results where id in (select distinct max(id) as id from state_records group by workplace_id)
	db.Raw("select * from state_records where id in (select distinct max(id) as id from state_records group by workplace_id)").Find(&stateRecords)
	for _, record := range stateRecords {
		switch record.StateID {
		case 1:
			production++
		case 2:
			downtime++
		case 3:
			offline++
		}
	}
	return production, downtime, offline
}

func streamWorkplaces(streamer *sse.Streamer) {
	for {
		logInfo("SSE", "Streaming workplaces started")
		timer := time.Now()
		location := downloadActualLocation()
		loc, _ := time.LoadLocation(location)
		workplaces := downloadActualWorkplaces()
		if len(workplaces) > 0 {
			cachedDowntimeRecords := downloadDowntimeRecords()
			cachedOrderRecords := downloadOrderRecords()
			cachedStateRecords := downloadStateRecords()
			cachedDowntimes := downloadDowntimes()
			cachedOrders := downloadOrders()
			cachedUsers := downloadUsers()
			cachedStates := downloadStates()
			for _, workplace := range workplaces {
				workplaceDowntimeRecord := cachedDowntimeRecords[int(workplace.ID)]
				workplaceOrderRecord := cachedOrderRecords[int(workplace.ID)]
				workplaceStateRecord := cachedStateRecords[int(workplace.ID)]
				streamer.SendString("", "workplaces", workplace.Name+";<b>"+workplace.Name+"</b><br>"+cachedUsers[int(workplaceOrderRecord.UserId.Int32)].FirstName+" "+cachedUsers[int(workplaceOrderRecord.UserId.Int32)].SecondName+"<br>"+cachedOrders[workplaceOrderRecord.OrderID].Name+"<br>"+cachedDowntimes[workplaceDowntimeRecord.DowntimeID].Name+"<br><sub>"+time.Now().In(loc).Sub(workplaceStateRecord.DateTimeStart.In(loc)).Round(1*time.Minute).String()+"</sub>;"+cachedStates[workplaceStateRecord.StateID].Color)
			}
		}
		logInfo("SSE", "Streaming workplaces ended in "+time.Since(timer).String())
		time.Sleep(10 * time.Second)
	}
}

func downloadStateRecords() map[int]database.StateRecord {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return nil
	}
	var stateRecords []database.StateRecord
	db.Raw("select * from state_records where id in (select distinct max(id) as id from state_records group by workplace_id)").Find(&stateRecords)
	cachedStateRecords := make(map[int]database.StateRecord)
	for _, stateRecord := range stateRecords {
		cachedStateRecords[stateRecord.WorkplaceID] = stateRecord
	}
	return cachedStateRecords
}

func downloadOrderRecords() map[int]database.OrderRecord {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return nil
	}
	var orderRecords []database.OrderRecord
	db.Where("date_time_end is null").Find(&orderRecords)
	cachedOrderRecords := make(map[int]database.OrderRecord)
	for _, orderRecord := range orderRecords {
		cachedOrderRecords[orderRecord.WorkplaceID] = orderRecord
	}
	return cachedOrderRecords
}

func downloadStates() map[int]database.State {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return nil
	}
	var states []database.State
	db.Find(&states)
	cachedStates := make(map[int]database.State)
	for _, state := range states {
		cachedStates[int(state.ID)] = state
	}
	return cachedStates
}

func downloadUsers() map[int]database.User {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return nil
	}
	var users []database.User
	db.Find(&users)
	cachedUsers := make(map[int]database.User)
	for _, user := range users {
		cachedUsers[int(user.ID)] = user
	}
	return cachedUsers
}

func downloadOrders() map[int]database.Order {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return nil
	}
	var orders []database.Order
	db.Find(&orders)
	cachedOrders := make(map[int]database.Order)
	for _, order := range orders {
		cachedOrders[int(order.ID)] = order
	}
	return cachedOrders
}

func downloadDowntimes() map[int]database.Downtime {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return nil
	}
	var downtimes []database.Downtime
	db.Find(&downtimes)
	cachedDowntimes := make(map[int]database.Downtime)
	for _, downtime := range downtimes {
		cachedDowntimes[int(downtime.ID)] = downtime
	}
	return cachedDowntimes
}

func downloadDowntimeRecords() map[int]database.DowntimeRecord {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return nil
	}
	var downtimeRecords []database.DowntimeRecord
	db.Where("date_time_end is null").Find(&downtimeRecords)
	cachedDowntimeRecords := make(map[int]database.DowntimeRecord)
	for _, downtimeRecord := range downtimeRecords {
		cachedDowntimeRecords[downtimeRecord.WorkplaceID] = downtimeRecord
	}
	return cachedDowntimeRecords
}

func downloadActualLocation() string {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return ""
	}
	var timezone database.Setting
	db.Where("name=?", "timezone").Find(&timezone)
	return timezone.Value
}

func downloadActualWorkplaces() []database.Workplace {
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("SSE", "Problem opening database: "+err.Error())
		return nil
	}
	var workplaces []database.Workplace
	db.Find(&workplaces)
	return workplaces
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
