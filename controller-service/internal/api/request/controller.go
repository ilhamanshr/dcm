package request

import "errors"

type RegisterAgentRequest struct {
	Name string `json:"name"`
}

type UpdateConfigRequest struct {
	URL          string `json:"url"`
	PollInterval int    `json:"poll_interval"`
}

func (r UpdateConfigRequest) Validate() error {
	if r.URL == "" {
		return errors.New("url is required")
	}
	if r.PollInterval <= 0 {
		return errors.New("poll_interval must be greater than 0")
	}
	return nil
}
