package config

type EventOptions struct {
	Hooks []string `json:"hooks" yaml:"hooks"`
}

func NewEventOptions() *EventOptions {
	return &EventOptions{
		Hooks: make([]string, 0),
	}
}
