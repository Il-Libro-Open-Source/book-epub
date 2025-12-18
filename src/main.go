// Package main fornisce un generatore per convertire file Markdown del
// "Il Libro Open Source" in formato ePUB, MOBI e PDF.
//
// Il tool processa una struttura di directory di file Markdown con frontmatter YAML,
// li converte in HTML usando Goldmark, e li impacchetta in un file ePUB con
// gerarchia dei capitoli, immagini e styling CSS.
//
// Variabili d'ambiente:
//   - INPUT:  Percorso al repository del libro (default: /tmp/book)
//   - OUTPUT: Percorso del file ePUB di output (default: ./il-manuale-del-buon-dev.epub)
//   - COVER:  Percorso all'immagine di copertina (default: ./assets/cover.jpg)
//   - STYLE:  Percorso al file CSS (default: ./assets/style.css)
//   - UUID:   UUID identificatore del libro (richiesto)
package main

import (
	"log/slog"
	"path/filepath"
)

func main() {
	slog.Info("Avvio del programma")

	// Carica la configurazione
	cfg, err := LoadConfig()
	if err != nil {
		slog.Error("Errore durante il caricamento della configurazione", "error", err)
		panic(err)
	}
	slog.Info("Configurazione caricata con successo", "config", cfg)

	// Crea il convertitore markdown
	cv := NewMarkdownConverter()

	// Carica i capitoli
	mdPath := filepath.Join(cfg.Input, "docs", "it")
	chapters, err := GetChapters(&cv, mdPath)
	if err != nil {
		slog.Error("Errore durante il caricamento dei capitoli", "error", err)
		panic(err)
	}
	slog.Info("Capitoli caricati con successo", "numero capitoli", len(chapters))

	// Crea il libro EPUB
	builder, err := NewBookBuilder(cfg)
	if err != nil {
		slog.Error("Errore durante la creazione del libro EPUB", "error", err)
		panic(err)
	}

	// Configura i metadati
	if err := builder.SetupMetadata(); err != nil {
		slog.Error("Errore durante la configurazione dei metadati", "error", err)
		panic(err)
	}

	// Aggiunge la copertina
	if err := builder.AddCover(); err != nil {
		slog.Error("Errore durante l'aggiunta della copertina", "error", err)
		panic(err)
	}

	// Aggiunge il foglio di stile
	cssPath, err := builder.AddStylesheet()
	if err != nil {
		slog.Error("Errore durante l'aggiunta del foglio di stile", "error", err)
		panic(err)
	}

	// Aggiunge le immagini
	if err := builder.AddImages(chapters, cfg.Input); err != nil {
		slog.Error("Errore durante l'aggiunta delle immagini", "error", err)
		panic(err)
	}

	// Crea i capitoli
	if err := builder.CreateChapters(chapters, cssPath); err != nil {
		slog.Error("Errore durante la creazione dei capitoli", "error", err)
		panic(err)
	}

	// Salva il libro
	if err := builder.Save(); err != nil {
		slog.Error("Errore durante il salvataggio del libro EPUB", "error", err)
		panic(err)
	}

	slog.Info("Processo completato con successo")
}
