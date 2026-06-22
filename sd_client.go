package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type SDClient struct {
	baseURL     string
	httpClient  *http.Client
	isConnected bool
}

func NewSDClient(timeout time.Duration) *SDClient {
	return &SDClient{
		isConnected: false,
		httpClient: &http.Client{
			Timeout: timeout, // Интерфейс зависнет максимум на это время
		},
	}
}

// Задаем адрес сервера
func (c *SDClient) SetAddress(ip, port string) {
	ip = strings.TrimSpace(ip)
	port = strings.TrimSpace(port)

	if !strings.HasPrefix(ip, "http://") && !strings.HasPrefix(ip, "https://") {
		ip = "http://" + ip
	}

	if port != "" {
		c.baseURL = fmt.Sprintf("%s:%s", ip, port)
	} else {
		c.baseURL = ip
	}
}

// Прямая проверка соединения (блокирующий вызов)
func (c *SDClient) CheckConnection() (*CapabilitiesResponse, error) {
	if c.baseURL == "" {
		c.isConnected = false
		return nil, fmt.Errorf("Server address not set")
	}

	url := fmt.Sprintf("%s/sdcpp/v1/capabilities", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		c.isConnected = false
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.isConnected = false
		return nil, fmt.Errorf("Server return code: %d", resp.StatusCode)
	}

	var bodyBytes []byte
	if bodyBytes, err = io.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("Error read reply: %w", err)
	}
	fmt.Printf("GetJobStatus response body: %s\n", string(bodyBytes))

	var caps CapabilitiesResponse
	if err := json.Unmarshal(bodyBytes, &caps); err != nil {
		c.isConnected = false
		return nil, fmt.Errorf("Error parse JSON: %w", err)
	}

	c.isConnected = true
	return &caps, nil
}

func (c *SDClient) ImgGetRequest(req ImgGenRequest) (*GenResponse, error) {
	if c.isConnected == false {
		return nil, fmt.Errorf("Server is not connected")
	}

	url := fmt.Sprintf("%s/sdcpp/v1/img_gen", c.baseURL)
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("Error make JSON: %w", err)
	}

	fmt.Printf("json: %+v\n", string(payload))

	resp, err := c.httpClient.Post(url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		errMsg := readErrorMessage(resp)
		if errMsg != "" {
			return nil, fmt.Errorf("Server ret code: %d: %s", resp.StatusCode, errMsg)
		}
		return nil, fmt.Errorf("Server ret code: %d", resp.StatusCode)
	}

	var genResp GenResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, fmt.Errorf("Error parse JSON: %w", err)
	}

	return &genResp, nil
}

func readErrorMessage(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error read reply: %v", err)
	}
	return extractErrorMessage(body)
}

func extractErrorMessage(body []byte) string {
	var errObj struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errObj); err == nil && errObj.Error != "" {
		return errObj.Error
	}
	return strings.TrimSpace(string(body))
}

func formatErrorBody(body []byte) string {
	errMsg := extractErrorMessage(body)
	if errMsg == "" {
		return ""
	}
	return ": " + errMsg
}

func (c *SDClient) GetJobStatus(id string) (*JobResponse, error) {
	if c.isConnected == false {
		return nil, fmt.Errorf("Server is not connected")
	}

	url := fmt.Sprintf("%s/sdcpp/v1/jobs/%s", c.baseURL, strings.TrimSpace(id))
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bodyBytes []byte
	if bodyBytes, err = io.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("Error read reply: %w", err)
	}
	//fmt.Printf("GetJobStatus response body: %s\n", string(bodyBytes))

	switch resp.StatusCode {
	case http.StatusOK:
		var jobResp JobResponse
		if err := json.Unmarshal(bodyBytes, &jobResp); err != nil {
			return nil, fmt.Errorf("Error parse JSON: %w", err)
		}
		return &jobResp, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("Job not found: %s%s", id, formatErrorBody(bodyBytes))
	case http.StatusGone:
		return nil, fmt.Errorf("Job %s is gone %s", id, formatErrorBody(bodyBytes))
	default:
		if msg := formatErrorBody(bodyBytes); msg != "" {
			return nil, fmt.Errorf("Server ret code: %d%s", resp.StatusCode, msg)
		}
		return nil, fmt.Errorf("Server ret code: %d", resp.StatusCode)
	}
}

func (c *SDClient) CancelJob(id string) (*GenResponse, error) {
	if c.isConnected == false {
		return nil, fmt.Errorf("Server is not connected")
	}

	url := fmt.Sprintf("%s/sdcpp/v1/jobs/%s/cancel", c.baseURL, strings.TrimSpace(id))
	resp, err := c.httpClient.Post(url, "application/json", strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error read reply: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var genResp GenResponse
		if err := json.Unmarshal(bodyBytes, &genResp); err != nil {
			return nil, fmt.Errorf("Error parse JSON: %w", err)
		}
		return &genResp, nil
	case http.StatusNotFound:
		return nil, fmt.Errorf("Job not found: %s%s", id, formatErrorBody(bodyBytes))
	case http.StatusConflict:
		return nil, fmt.Errorf("You can't cancel a job %s: %s", id, formatErrorBody(bodyBytes))
	case http.StatusGone:
		return nil, fmt.Errorf("Job %s is gone %s", id, formatErrorBody(bodyBytes))
	default:
		if msg := formatErrorBody(bodyBytes); msg != "" {
			return nil, fmt.Errorf("Server ret code: %d%s", resp.StatusCode, msg)
		}
		return nil, fmt.Errorf("Server ret code: %d", resp.StatusCode)
	}
}
