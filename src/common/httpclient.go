package common

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient struct {
	BaseURL string
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
	}
}

func (c *HTTPClient) Get(endpoint string) ([]byte, error) {
	url := c.BaseURL + endpoint
	response, err := http.Get(url)
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("get kernel info from error , status code: " + fmt.Sprint(response.StatusCode))
	}
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *HTTPClient) Post(endpoint string, data []byte) ([]byte, error) {
	url := c.BaseURL + endpoint
	response, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *HTTPClient) Delete(endpoint string) error {
	url := c.BaseURL + endpoint
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNoContent {
		return nil
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return nil
}
