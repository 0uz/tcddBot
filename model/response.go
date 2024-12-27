package model

type TCDDResponse struct {
	TrainLegs []TrainLegs `json:"trainLegs"`
}

type TrainLegs struct {
	TrainAvailabilities []TrainAvailabilities `json:"trainAvailabilities"`
	ResultCount         int                   `json:"resultCount"`
}

type TrainAvailabilities struct {
	Trains        []Trains `json:"trains"`
	TotalTripTime int      `json:"totalTripTime"`
	MinPrice      float64  `json:"minPrice"`
	Connection    bool     `json:"connection"`
	DayChanged    bool     `json:"dayChanged"`
}

type Trains struct {
	ID                       int                        `json:"id"`
	Number                   string                     `json:"number"`
	Name                     string                     `json:"name"`
	CommercialName           string                     `json:"commercialName"`
	Type                     string                     `json:"type"`
	Line                     any                        `json:"line"`
	Reversed                 bool                       `json:"reversed"`
	ScheduleID               int                        `json:"scheduleId"`
	DepartureStationID       int                        `json:"departureStationId"`
	ArrivalStationID         int                        `json:"arrivalStationId"`
	MinPrice                 MinPrice                   `json:"minPrice"`
	ReservationLockTime      int                        `json:"reservationLockTime"`
	Reservable               bool                       `json:"reservable"`
	BookingClassCapacities   []BookingClassCapacities   `json:"bookingClassCapacities"`
	Segments                 []Segments                 `json:"segments"`
	Cars                     []Cars                     `json:"cars"`
	TrainSegments            []TrainSegments            `json:"trainSegments"`
	TotalDistance            float64                    `json:"totalDistance"`
	AvailableFareInfo        []AvailableFareInfo        `json:"availableFareInfo"`
	CabinClassAvailabilities []CabinClassAvailabilities `json:"cabinClassAvailabilities"`
	TrainDate                int64                      `json:"trainDate"`
	TrainNumber              string                     `json:"trainNumber"`
	SkipsDay                 bool                       `json:"skipsDay"`
}

type TrainSegments struct {
	DepartureTime string `json:"departureTime"`
	ArrivalTime   string `json:"arrivalTime"`
	DepartureStationID int    `json:"departureStationId"`
	ArrivalStationID   int    `json:"arrivalStationId"`
}

type CabinClassAvailabilities struct {
	CabinClass        CabinClass `json:"cabinClass"`
	AvailabilityCount int        `json:"availabilityCount"`
}

type CabinClass struct {
	ID                      int    `json:"id"`
	Code                    string `json:"code"`
	Name                    string `json:"name"`
	AdditionalServices      any    `json:"additionalServices"`
	BookingClassModels      any    `json:"bookingClassModels"`
	ShowAvailabilityOnQuery bool   `json:"showAvailabilityOnQuery"`
}

type MinPrice struct {
	Type          any     `json:"type"`
	PriceAmount   float64 `json:"priceAmount"`
	PriceCurrency string  `json:"priceCurrency"`
}

type BookingClassCapacities struct {
	ID             int `json:"id"`
	TrainID        int `json:"trainId"`
	BookingClassID int `json:"bookingClassId"`
	Capacity       int `json:"capacity"`
}

type StationStatus struct {
	ID     int `json:"id"`
	Name   any `json:"name"`
	Detail any `json:"detail"`
}

type StationType struct {
	ID     int `json:"id"`
	Name   any `json:"name"`
	Detail any `json:"detail"`
}

type DepartureStation struct {
	ID                    int           `json:"id"`
	StationNumber         string        `json:"stationNumber"`
	AreaCode              int           `json:"areaCode"`
	Name                  string        `json:"name"`
	StationStatus         StationStatus `json:"stationStatus"`
	StationType           StationType   `json:"stationType"`
	UnitID                int           `json:"unitId"`
	CityID                int           `json:"cityId"`
	DistrictID            int           `json:"districtId"`
	NeighbourhoodID       int           `json:"neighbourhoodId"`
	UicCode               any           `json:"uicCode"`
	TechnicalUnit         string        `json:"technicalUnit"`
	StationChefID         int           `json:"stationChefId"`
	Detail                string        `json:"detail"`
	ShowOnQuery           bool          `json:"showOnQuery"`
	PassengerDrop         bool          `json:"passengerDrop"`
	TicketSaleActive      bool          `json:"ticketSaleActive"`
	Active                bool          `json:"active"`
	Email                 string        `json:"email"`
	OrangeDeskEmail       string        `json:"orangeDeskEmail"`
	Address               string        `json:"address"`
	Longitude             float64       `json:"longitude"`
	Latitude              float64       `json:"latitude"`
	Altitude              float64       `json:"altitude"`
	StartKm               float64       `json:"startKm"`
	EndKm                 float64       `json:"endKm"`
	ShowOnMap             bool          `json:"showOnMap"`
	PassengerAdmission    bool          `json:"passengerAdmission"`
	DisabledAccessibility bool          `json:"disabledAccessibility"`
	Phones                any           `json:"phones"`
	WorkingDays           any           `json:"workingDays"`
	Hardwares             any           `json:"hardwares"`
	PhysicalProperties    any           `json:"physicalProperties"`
	StationPlatforms      any           `json:"stationPlatforms"`
	SalesChannels         any           `json:"salesChannels"`
	IATACode              any           `json:"IATACode"`
}

type ArrivalStation struct {
	ID                    int           `json:"id"`
	StationNumber         string        `json:"stationNumber"`
	AreaCode              int           `json:"areaCode"`
	Name                  string        `json:"name"`
	StationStatus         StationStatus `json:"stationStatus"`
	StationType           StationType   `json:"stationType"`
	UnitID                int           `json:"unitId"`
	CityID                int           `json:"cityId"`
	DistrictID            int           `json:"districtId"`
	NeighbourhoodID       int           `json:"neighbourhoodId"`
	UicCode               any           `json:"uicCode"`
	TechnicalUnit         string        `json:"technicalUnit"`
	StationChefID         int           `json:"stationChefId"`
	Detail                string        `json:"detail"`
	ShowOnQuery           bool          `json:"showOnQuery"`
	PassengerDrop         bool          `json:"passengerDrop"`
	TicketSaleActive      bool          `json:"ticketSaleActive"`
	Active                bool          `json:"active"`
	Email                 string        `json:"email"`
	OrangeDeskEmail       string        `json:"orangeDeskEmail"`
	Address               string        `json:"address"`
	Longitude             float64       `json:"longitude"`
	Latitude              float64       `json:"latitude"`
	Altitude              float64       `json:"altitude"`
	StartKm               float64       `json:"startKm"`
	EndKm                 float64       `json:"endKm"`
	ShowOnMap             bool          `json:"showOnMap"`
	PassengerAdmission    bool          `json:"passengerAdmission"`
	DisabledAccessibility bool          `json:"disabledAccessibility"`
	Phones                any           `json:"phones"`
	WorkingDays           any           `json:"workingDays"`
	Hardwares             any           `json:"hardwares"`
	PhysicalProperties    any           `json:"physicalProperties"`
	StationPlatforms      any           `json:"stationPlatforms"`
	SalesChannels         any           `json:"salesChannels"`
	IATACode              any           `json:"IATACode"`
}

type Segment struct {
	ID               int              `json:"id"`
	Name             string           `json:"name"`
	DepartureStation DepartureStation `json:"departureStation"`
	ArrivalStation   ArrivalStation   `json:"arrivalStation"`
	LineID           int              `json:"lineId"`
	LineOrder        int              `json:"lineOrder"`
}

type Segments struct {
	ID            int     `json:"id"`
	DepartureTime int64   `json:"departureTime"`
	ArrivalTime   int64   `json:"arrivalTime"`
	Stops         bool    `json:"stops"`
	Duration      int     `json:"duration"`
	StopDuration  int     `json:"stopDuration"`
	Distance      float64 `json:"distance"`
	Segment       Segment `json:"segment"`
}

type FareFamily struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type BookingClass struct {
	ID         int        `json:"id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	CabinClass CabinClass `json:"cabinClass"`
	FareFamily FareFamily `json:"fareFamily"`
}

type Price struct {
	Type          any     `json:"type"`
	PriceAmount   float64 `json:"priceAmount"`
	PriceCurrency string  `json:"priceCurrency"`
}

type FareBasis struct {
	Code   string  `json:"code"`
	Factor float64 `json:"factor"`
	Price  Price   `json:"price"`
}

type BasePrice struct {
	Type          any     `json:"type"`
	PriceAmount   float64 `json:"priceAmount"`
	PriceCurrency string  `json:"priceCurrency"`
}

type CrudePrice struct {
	Type          any     `json:"type"`
	PriceAmount   float64 `json:"priceAmount"`
	PriceCurrency string  `json:"priceCurrency"`
}

type BaseTransportationCost struct {
	Type          any     `json:"type"`
	PriceAmount   float64 `json:"priceAmount"`
	PriceCurrency string  `json:"priceCurrency"`
}

type PricingList struct {
	BasePricingID          int                    `json:"basePricingId"`
	BookingClass           BookingClass           `json:"bookingClass"`
	CabinClassID           int                    `json:"cabinClassId"`
	BasePricingType        string                 `json:"basePricingType"`
	FareBasis              FareBasis              `json:"fareBasis"`
	BasePrice              BasePrice              `json:"basePrice"`
	CrudePrice             CrudePrice             `json:"crudePrice"`
	BaseTransportationCost BaseTransportationCost `json:"baseTransportationCost"`
	Availability           int                    `json:"availability"`
}

type Availabilities struct {
	TrainCarID         int           `json:"trainCarId"`
	TrainCarName       any           `json:"trainCarName"`
	CabinClass         CabinClass    `json:"cabinClass"`
	Availability       int           `json:"availability"`
	PricingList        []PricingList `json:"pricingList"`
	AdditionalServices []any         `json:"additionalServices"`
}

type Cars struct {
	ID             int              `json:"id"`
	Name           string           `json:"name"`
	TrainID        int              `json:"trainId"`
	TemplateID     int              `json:"templateId"`
	CarIndex       int              `json:"carIndex"`
	Unlabeled      bool             `json:"unlabeled"`
	Capacity       int              `json:"capacity"`
	CabinClassID   int              `json:"cabinClassId"`
	Availabilities []Availabilities `json:"availabilities"`
}

type BookingClassAvailabilities struct {
	BookingClass BookingClass `json:"bookingClass"`
	Price        float64      `json:"price"`
	Availability int          `json:"availability"`
}

type CabinClasses struct {
	CabinClass                 CabinClass                   `json:"cabinClass"`
	AvailabilityCount          int                          `json:"availabilityCount"`
	MinPrice                   float64                      `json:"minPrice"`
	BookingClassAvailabilities []BookingClassAvailabilities `json:"bookingClassAvailabilities"`
}

type AvailableFareInfo struct {
	FareFamily   FareFamily     `json:"fareFamily"`
	CabinClasses []CabinClasses `json:"cabinClasses"`
}