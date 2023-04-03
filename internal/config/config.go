package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Server struct {
	BindHost  string   `envconfig:"bind_host" default:"0.0.0.0"`
	BindPort  int      `envconfig:"bind_port" default:"8080"`
	BasicAuth []string `envconfig:"basic_auth"`
}

func (s *Server) BindAddress() string {
	return fmt.Sprintf("%s:%d", s.BindHost, s.BindPort)
}

func (s *Server) BasicAuthCredentials() map[string]string {
	credentials := map[string]string{}

	for _, pair := range s.BasicAuth {
		frags := strings.SplitN(pair, ":", 2)
		credentials[frags[0]] = frags[1]
	}

	return credentials
}

type KSEI struct {
	Accounts        Accounts      `envconfig:"accounts"`
	AuthDir         string        `envconfig:"auth_dir" default:".goksei-auth"`
	RefreshInterval time.Duration `envconfig:"refresh_interval" default:"1h"`
	RefreshJitter   float32       `envconfig:"refresh_jitter" default:"0.2"`
}

type Config struct {
	Server Server `envconfig:"server"`
	KSEI   KSEI   `envconfig:"ksei"`
}

func FromEnv() (*Config, error) {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		return nil, err
	}

	return &config, nil
}
