package main

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

// Config rappresenta la configurazione dell'applicazione
// Config contiene la configurazione dell'applicazione caricata dalle variabili d'ambiente.
// Tutti i percorsi sono validati all'avvio per assicurarsi che file e directory esistano.
type Config struct {
	Input  string `env:"INPUT" envDefault:"/tmp/book" validate:"required"`
	Output string `env:"OUTPUT" envDefault:"./il-manuale-del-buon-dev.epub" validate:"required"`
	Cover  string `env:"COVER" envDefault:"./assets/cover.jpg" validate:"required"`
	Style  string `env:"STYLE" envDefault:"./assets/style.css" validate:"required"`
	Uuid   string `env:"UUID" validate:"required"`
}

// LoadConfig carica e valida la configurazione dalle variabili d'ambiente
// LoadConfig carica e valida la configurazione dalle variabili d'ambiente.
// Ritorna un errore se la configurazione non è valida o se i file richiesti non esistono.
func LoadConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("errore durante il parsing delle variabili d'ambiente: %w", err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("errore durante la validazione della configurazione: %w", err)
	}

	return &cfg, nil
}
