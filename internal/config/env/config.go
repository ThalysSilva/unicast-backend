package env

import (
	"fmt"
	"os"
)

type Evolution struct {
	Host   string
	Port   string
	APIKey string
}

type Auth struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	JWESecret          string
	RegisterInviteKey  string
}

type Defaults struct {
	CountryCode string
}

type OAuth struct {
	FrontendBaseURL    string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

type Config struct {
	Evolution Evolution
	Auth      Auth
	Defaults  Defaults
	OAuth     OAuth
	Admin     struct {
		Secret string
	}
}

// Load carrega variáveis de ambiente necessárias para integrações externas.
// Retorna erro se algo obrigatório estiver vazio.
func Load() (*Config, error) {
	cfg := &Config{
		Evolution: Evolution{
			Host:   os.Getenv("EVOLUTION_HOST"),
			Port:   os.Getenv("EVOLUTION_PORT"),
			APIKey: os.Getenv("AUTHENTICATION_API_KEY"),
		},
		Auth: Auth{
			AccessTokenSecret:  os.Getenv("ACCESS_TOKEN_SECRET"),
			RefreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET"),
			JWESecret:          os.Getenv("JWE_SECRET"),
			RegisterInviteKey:  os.Getenv("REGISTER_INVITE_KEY"),
		},
		Defaults: Defaults{
			CountryCode: os.Getenv("DEFAULT_COUNTRY_CODE"),
		},
		OAuth: OAuth{
			FrontendBaseURL:    os.Getenv("FRONTEND_BASE_URL"),
			GoogleClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
			GoogleClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
			GoogleRedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"),
		},
	}

	cfg.Admin.Secret = os.Getenv("ADMIN_SECRET")

	if cfg.Defaults.CountryCode == "" {
		cfg.Defaults.CountryCode = "55"
	}
	if cfg.OAuth.FrontendBaseURL == "" {
		cfg.OAuth.FrontendBaseURL = "http://localhost:3000"
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Evolution.Host == "" || cfg.Evolution.Port == "" || cfg.Evolution.APIKey == "" {
		return fmt.Errorf("variáveis da Evolution API ausentes (EVOLUTION_HOST, EVOLUTION_PORT, AUTHENTICATION_API_KEY)")
	}
	if cfg.Auth.AccessTokenSecret == "" || cfg.Auth.RefreshTokenSecret == "" || cfg.Auth.JWESecret == "" || cfg.Auth.RegisterInviteKey == "" {
		return fmt.Errorf("segredos de autenticação ausentes (ACCESS_TOKEN_SECRET, REFRESH_TOKEN_SECRET, JWE_SECRET, REGISTER_INVITE_KEY)")
	}
	if cfg.Admin.Secret == "" {
		return fmt.Errorf("ADMIN_SECRET ausente")
	}
	return nil
}
