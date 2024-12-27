package service

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "tcddbot/config"
    "tcddbot/model"
    "time"
)

type TrainService struct {
    cfg    *config.Config
    client *http.Client
}

func NewTrainService(cfg *config.Config) *TrainService {
    return &TrainService{
        cfg: cfg,
        client: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (s *TrainService) CheckAvailability(ctx context.Context, departureID, arrivalID int, date string) (*model.TCDDResponse, error) {
    adjustedDate, err := s.adjustDate(date)
    if (err != nil) {
        return nil, fmt.Errorf("date adjustment failed: %w", err)
    }

    reqBody := map[string]interface{}{
        "searchRoutes": []map[string]interface{}{
            {
                "departureStationId": departureID,
                "arrivalStationId":   arrivalID,
                "departureDate":      adjustedDate,
            },
        },
        "passengerTypeCounts": []map[string]interface{}{
            {"id": 0, "count": 1},
        },
        "searchReservation": false,
    }

    return s.makeRequest(ctx, reqBody)
}

func (s *TrainService) CheckTrainAvailability(departureStationID, arrivalStationID int, travelDate string) (bool, error) {
    adjustedDate, err := s.adjustDate(travelDate)
    if err != nil {
        return false, fmt.Errorf("date adjustment failed: %w", err)
    }

    reqBody := map[string]interface{}{
        "searchRoutes": []map[string]interface{}{
            {
                "departureStationId": departureStationID,
                "arrivalStationId":   arrivalStationID,
                "departureDate":      adjustedDate,
            },
        },
        "passengerTypeCounts": []map[string]interface{}{
            {"id": 0, "count": 1},
        },
        "searchReservation": false,
    }

    resp, err := s.makeRequest(context.Background(), reqBody)
    if err != nil {
        return false, err
    }

    // Check if response indicates no availability
    if len(resp.TrainLegs) == 0 {
        return false, nil
    }

    return true, nil
}

type ErrorResponse struct {
    Timestamp string `json:"timestamp"`
    TraceId   string `json:"traceId"`
    Status    string `json:"status"`
    Type      string `json:"type"`
    Code      int    `json:"code"`
    Message   string `json:"message"`
    Detail    string `json:"detail"`
    Title     string `json:"title"`
}

func (s *TrainService) makeRequest(ctx context.Context, reqBody interface{}) (*model.TCDDResponse, error) {
    jsonBody, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("marshal request body: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", s.cfg.APIEndpoint+"/train-availability", bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    req.Header.Set("Accept", "application/json")
    req.Header.Set("Authorization", s.cfg.AuthToken)
    req.Header.Set("unit-id", s.cfg.UnitID)
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("do request: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read response body: %w", err)
    }

    // First try to unmarshal as error response
    var errorResp ErrorResponse
    if err := json.Unmarshal(body, &errorResp); err == nil {
        if errorResp.Code == 604 {
            return nil, fmt.Errorf("no trains available: %s", errorResp.Message)
        }
    }

    var response model.TCDDResponse
    if err := json.Unmarshal(body, &response); err != nil {
        return nil, fmt.Errorf("unmarshal response: %w", err)
    }

    return &response, nil
}

func (s *TrainService) adjustDate(date string) (string, error) {
    parsedDate, err := time.Parse("02-01-2006", date)
    if err != nil {
        return "", fmt.Errorf("parse date: %w", err)
    }
    adjustedDate := parsedDate.AddDate(0, 0, -1)
    return adjustedDate.Format("02-01-2006") + " 21:00:00", nil
}
