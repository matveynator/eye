package client

import (
	"encoding/json"
	"fmt"
	neturl "net/url"

	"github.com/nl2go/hrobot-go/models"
)

func (c *Client) ResetGet(ip string) (*models.Reset, error) {
	url := fmt.Sprintf(c.baseURL+"/reset/%s", ip)
	bytes, err := c.doGetRequest(url)
	if err != nil {
		return nil, err
	}

	var resetResp models.ResetResponse
	err = json.Unmarshal(bytes, &resetResp)
	if err != nil {
		return nil, err
	}

	return &resetResp.Reset, nil
}

func (c *Client) ResetSet(ip string, input *models.ResetSetInput) (*models.ResetPost, error) {
	url := fmt.Sprintf(c.baseURL+"/reset/%s", ip)

	formData := neturl.Values{}
	formData.Set("type", input.Type)

	bytes, err := c.doPostFormRequest(url, formData)
	if err != nil {
		return nil, err
	}

	var resetResp models.ResetPostResponse
	err = json.Unmarshal(bytes, &resetResp)
	if err != nil {
		return nil, err
	}

	return &resetResp.Reset, nil
}
