package homeassistiant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HAClient represents a Home Assistant API client
type HAClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// State represents a Home Assistant entity state
type State struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged time.Time              `json:"last_changed"`
	LastUpdated time.Time              `json:"last_updated"`
}

// Calendar represents a Home Assistant calendar entity
type Calendar struct {
	EntityID string `json:"entity_id"`
	Name     string `json:"name"`
}

// CalendarEvent represents a calendar event
type CalendarEvent struct {
	Summary     string `json:"summary"`
	Start       Time   `json:"start"`
	End         Time   `json:"end"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
}

// Time represents a calendar event time
type Time struct {
	DateTime string `json:"dateTime,omitempty"`
	Date     string `json:"date,omitempty"`
}

// ServiceResponse represents the response data from a service call
type ServiceResponse map[string]interface{}

// ChangedState represents a state change during service execution
type ChangedState struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged string                 `json:"last_changed"`
}

// ServiceResult represents the complete response from a service call
type ServiceResult struct {
	ChangedStates   []ChangedState  `json:"changed_states"`
	ServiceResponse ServiceResponse `json:"service_response,omitempty"`
}

// NewClient creates a new Home Assistant client
func NewClient(baseURL string, token string) *HAClient {
	return &HAClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// GetStates retrieves all entity states
func (haClient *HAClient) GetStates() ([]State, error) {
	path := "/api/states"
	req, err := haClient.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := haClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var states []State
	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return nil, err
	}

	return states, err
}

// GetState retrieves state for a specific entity
func (haClient *HAClient) GetState(entityID string) (*State, error) {
	path := fmt.Sprintf("/api/states/%s", entityID)
	req, err := haClient.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := haClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var state State
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, err
	}

	return &state, err
}

// GetErrorLog retrieves the error log
func (haClient *HAClient) GetErrorLog() (string, error) {
	path := "/api/error_log"

	req, err := haClient.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	resp, err := haClient.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	logText, err := io.ReadAll(resp.Body)
	return string(logText), err
}

// GetCameraProxy retrieves camera image
func (haClient *HAClient) GetCameraProxy(entityID string) ([]byte, error) {
	return haClient.getRaw(fmt.Sprintf("/api/camera_proxy/%s", entityID))
}

// GetCalendars retrieves all calendars
func (haClient *HAClient) GetCalendars() ([]Calendar, error) {
	path := "/api/calendars"
	req, err := haClient.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := haClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var calendars []Calendar
	if err := json.NewDecoder(resp.Body).Decode(&calendars); err != nil {
		return nil, err
	}

	return calendars, err
}

// GetCalendarEvents retrieves calendar events
func (haClient *HAClient) GetCalendarEvents(entityID string, start, end time.Time) ([]CalendarEvent, error) {
	path := fmt.Sprintf("/api/calendars/%s?start=%s&end=%s",
		entityID,
		start.Format(time.RFC3339),
		end.Format(time.RFC3339))

	req, err := haClient.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := haClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var events []CalendarEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	return events, err
}

// CallService calls a service within a specific domain
func (haClient *HAClient) CallService(domain, service string, data interface{}, returnResponse bool) (*ServiceResult, error) {
	path := fmt.Sprintf("/api/services/%s/%s", domain, service)
	if returnResponse {
		path += "?return_response"
	}

	req, err := haClient.makeRequest(http.MethodPost, path, data)
	if err != nil {
		return nil, err
	}

	resp, err := haClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var result ServiceResult
	if returnResponse {
		var tempResult ServiceResult
		if err := json.NewDecoder(resp.Body).Decode(&tempResult); err != nil {
			return nil, err
		}

		result = tempResult
	} else {
		var tempResult []ChangedState
		if err := json.NewDecoder(resp.Body).Decode(&tempResult); err != nil {
			return nil, err
		}
		result.ChangedStates = tempResult
	}

	return &result, nil
}

// UpdateState updates an entity state
func (haClient *HAClient) UpdateState(entityID string, state *State) (*State, error) {
	path := fmt.Sprintf("/api/states/%s", entityID)

	req, err := haClient.makeRequest(http.MethodPost, path, state)
	if err != nil {
		return nil, err
	}

	resp, err := haClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var updatedState State
	if err := json.NewDecoder(resp.Body).Decode(&updatedState); err != nil {
		return nil, err
	}

	return &updatedState, err
}

//func (c *HAClient) get(path string) (*http.Response, error) {
//	return c.doRequest(http.MethodGet, path, nil)
//}

func (haClient *HAClient) getRaw(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, haClient.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+haClient.token)
	resp, err := haClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s", resp.Status)
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	return buf.Bytes(), err
}

//func (c *HAClient) _post(path string, body interface{}) (*http.Response, error) {
//	req, err := c.makeRequest(http.MethodPost, path, body)
//	resp, err := c.httpClient.Do(req)
//	return c.doRequest(http.MethodPost, path, body)
//}

func (haClient *HAClient) _doRequest(method, path string, body interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, haClient.baseURL+path, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+haClient.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := haClient.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with %s: %s", resp.Status, b)
	}
	return resp, nil
}

func (haClient *HAClient) makeRequest(method, path string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, haClient.baseURL+path, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+haClient.token)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
