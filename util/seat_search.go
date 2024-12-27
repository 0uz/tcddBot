package util

import (
    "tcddbot/model"
    "time"
)

type SeatAvailability struct {
    Train            model.Trains
    DepartureTime    time.Time
    AvailableSeats   map[string]int // Changed to map cabin class names to seat counts
    IsYHT           bool
}

func FindAvailableSeats(trainLegs []model.TrainLegs) []SeatAvailability {
    var results []SeatAvailability
    
    if len(trainLegs) == 0 {
        return results
    }

    for _, trainLeg := range trainLegs {
        for _, trainAvailability := range trainLeg.TrainAvailabilities {
            for _, train := range trainAvailability.Trains {
                seatsByClass := make(map[string]int)
                for _, cabinClassAvailability := range train.CabinClassAvailabilities {
                    if cabinClassAvailability.CabinClass.Name == "TEKERLEKLİ SANDALYE" {
                        continue
                    }
                    if cabinClassAvailability.AvailabilityCount > 0 {
                        seatsByClass[cabinClassAvailability.CabinClass.Name] = cabinClassAvailability.AvailabilityCount
                    }
                }

                if len(seatsByClass) > 0 {
                    departureTime, _ := time.Parse("2006-01-02T15:04:05", train.TrainSegments[0].DepartureTime)
                    results = append(results, SeatAvailability{
                        Train:          train,
                        DepartureTime:  departureTime,
                        AvailableSeats: seatsByClass,
                        IsYHT:         train.Type == "YHT",
                    })
                }
            }
        }
    }

    return results
}
