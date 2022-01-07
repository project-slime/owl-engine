package config

import (
	"fmt"

	"github.com/spf13/pflag"
)

type ServerOptions struct {
	Mode        string `json:"mode,omitempty" yaml:"mode"`
	Bind        string `json:"bind,omitempty" yaml:"bind"`
	Port        int    `json:"port" yaml:"port"`
	Secret      string `json:"secret" yaml:"secret"`
	EnableProxy bool   `json:"enable_proxy" yaml:"enableProxy"`
	Proxy       string `json:"proxy" yaml:"proxy"`
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Mode:        "dev",
		Bind:        "0.0.0.0",
		Port:        9530,
		Secret:      "",
		EnableProxy: false,
		Proxy:       "",
	}
}

func (s *ServerOptions) Validate() []error {
	errors := make([]error, 0)

	if s.Port < 0 || s.Port > 65535 {
		errors = append(errors, fmt.Errorf("secure port must be between 0 and 65535"))
	}

	return errors
}

func (s *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Bind, "bind-address", "0.0.0.0", "server bind address")
	fs.IntVar(&s.Port, "secure-port", 9530, "secure port number")
	fs.StringVar(&s.Mode, "server-mode", "dev", "Specify deployment mode. eg: dev,test,prod")
}
