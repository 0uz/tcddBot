package handlers

const (
    StateNone = iota
    StateSelectDeparture
    StateSelectArrival
    StateSelectDate
)

const (
    CallbackDateToday     = "date_today"
    CallbackDateTomorrow  = "date_tomorrow"
    CallbackDateCustom    = "date_custom"
    CallbackStationPrefix = "station_"
    MaxStationsPerPage    = 5
)

type UserState struct {
    State            int
    DepartureStation string
    ArrivalStation   string
    CurrentPage      int
}