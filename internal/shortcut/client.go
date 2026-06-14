package shortcut

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const apiBase string = "https://api.app.shortcut.com/api/v3"

type Client struct {
	client   http.Client
	apiToken string
}

type Iteration struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Status string `json:"status"`
}

func NewClient(apiToken string) Client {
	client := http.Client {}
	return Client{apiToken: apiToken, client: client}
}

func (it *Iteration) IsStarted() bool {
	return it.Status == "started"
}

func (c *Client) shortcutGet(endpoint string, buf any) error {
	req, err := http.NewRequest("GET", apiBase+endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Shortcut-Token", c.apiToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to GET %s", endpoint)
	}

	return json.NewDecoder(resp.Body).Decode(buf)
}

func (c *Client) GetAllIterations() ([]Iteration, error) {
	var iterations []Iteration
	err := c.shortcutGet("/iterations", &iterations)
	if err != nil {
		return nil, err
	}

	return iterations, nil
}

func (c *Client) GetActiveIterations() ([]Iteration, error) {
	iterations, err := c.GetAllIterations()
	if err != nil {
		return nil, err
	}

	// should always be three active iterations
	active := make([]Iteration, 3)
	for _, it := range iterations {
		if it.IsStarted() {
			active = append(active, it)
		}
	}

	return active, nil
}
