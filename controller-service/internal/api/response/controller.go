package response

type ConfigResponse struct {
	AgentID      string `json:"agent_id,omitempty"`
	PollURL      string `json:"poll_url"`
	PollInterval int    `json:"poll_interval"`
}
