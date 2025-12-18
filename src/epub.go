package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	epub "github.com/go-shiori/go-epub"
)

// BookBuilder gestisce la creazione e configurazione di un libro ePUB.
// Fornisce metodi per aggiungere metadati, copertina, stili, immagini e capitoli.
type BookBuilder struct {
	// book è l'istanza dell'ePUB fornita dalla libreria go-epub
	book *epub.Epub

	// config contiene la configurazione dell'applicazione
	config *Config
}

// NewBookBuilder crea un nuovo builder per il libro ePUB.
// Inizializza l'istanza ePUB con il titolo predefinito.
func NewBookBuilder(cfg *Config) (*BookBuilder, error) {
	book, err := epub.NewEpub("Il Libro Open Source")
	if err != nil {
		return nil, fmt.Errorf("errore durante la creazione del libro EPUB: %w", err)
	}

	return &BookBuilder{
		book:   book,
		config: cfg,
	}, nil
}

// SetupMetadata configura i metadati del libro ePUB.
// Include titolo, autore, identificatore UUID e lingua.
func (bb *BookBuilder) SetupMetadata() error {
	bb.book.SetTitle("Il manuale del buon dev")
	bb.book.SetAuthor("Community")
	bb.book.SetIdentifier("urn:uuid:" + bb.config.Uuid)
	bb.book.SetLang("it")

	slog.Info("Metadati del libro configurati")
	return nil
}

// AddCover aggiunge l'immagine di copertina al libro ePUB.
// Il percorso dell'immagine è preso dalla configurazione.
func (bb *BookBuilder) AddCover() error {
	coverPath, err := bb.book.AddImage(bb.config.Cover, "cover.jpg")
	if err != nil {
		return fmt.Errorf("errore nell'aggiunta della copertina: %w", err)
	}

	bb.book.SetCover(coverPath, "")
	slog.Info("Copertina aggiunta con successo")
	return nil
}

// AddStylesheet aggiunge il foglio di stile CSS al libro ePUB.
// Restituisce il percorso interno del CSS per l'uso nei capitoli.
func (bb *BookBuilder) AddStylesheet() (string, error) {
	cssPath, err := bb.book.AddCSS(bb.config.Style, "style.css")
	if err != nil {
		return "", fmt.Errorf("errore nell'aggiunta del foglio di stile: %w", err)
	}

	slog.Info("Foglio di stile aggiunto con successo")
	return cssPath, nil
}

// AddImages aggiunge tutte le immagini dei capitoli al libro con elaborazione parallela
func (bb *BookBuilder) AddImages(chapters []*Chapter, actPath string) error {
	slog.Debug("Inizio aggiunta immagini")

	// Raccoglie tutte le immagini uniche
	uniqueImages := bb.collectUniqueImages(chapters)

	// Processa le immagini in parallelo
	passedImages := make(map[string]string)
	var missingImages []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(uniqueImages))

	// Worker pool per limitare la concorrenza
	semaphore := make(chan struct{}, 10) // max 10 immagini in parallelo

	for image := range uniqueImages {
		wg.Add(1)
		go func(img string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquisisce
			defer func() { <-semaphore }() // Rilascia

			fsPath := strings.Replace(img, "/book", actPath, 1)

			// Validazione esistenza
			if _, err := os.Stat(fsPath); os.IsNotExist(err) {
				slog.Warn("Immagine non trovata, skip", "path", fsPath)
				mu.Lock()
				missingImages = append(missingImages, fsPath)
				mu.Unlock()
				return
			}

			fileName := strings.ReplaceAll(strings.TrimLeft(img, "/"), "/", "_")
			intPath, err := bb.book.AddImage(fsPath, fileName)
			if err != nil {
				errChan <- fmt.Errorf("errore nell'aggiunta dell'immagine %s: %w", fsPath, err)
				return
			}

			mu.Lock()
			passedImages[img] = intPath
			mu.Unlock()

			slog.Debug("Immagine aggiunta", "path", fsPath)
		}(image)
	}

	wg.Wait()
	close(errChan)

	// Verifica errori
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Sostituisce i path delle immagini nell'HTML
	bb.replaceImagePaths(chapters, passedImages)

	slog.Info("Immagini elaborate",
		"totale", len(passedImages),
		"mancanti", len(missingImages))

	if len(missingImages) > 0 {
		slog.Warn("Alcune immagini non sono state trovate", "count", len(missingImages))
	}

	return nil
}

// collectUniqueImages raccoglie tutte le immagini uniche dai capitoli
func (bb *BookBuilder) collectUniqueImages(chapters []*Chapter) map[string]bool {
	uniqueImages := make(map[string]bool)
	var collect func([]*Chapter)

	collect = func(chs []*Chapter) {
		for _, ch := range chs {
			for _, img := range ch.Images {
				uniqueImages[img] = true
			}
			if len(ch.Children) > 0 {
				collect(ch.Children)
			}
		}
	}

	collect(chapters)
	return uniqueImages
}

// replaceImagePaths sostituisce i path delle immagini nell'HTML dei capitoli
func (bb *BookBuilder) replaceImagePaths(chapters []*Chapter, passedImages map[string]string) {
	var replace func([]*Chapter)

	replace = func(chs []*Chapter) {
		for _, ch := range chs {
			for origPath, intPath := range passedImages {
				ch.Html = strings.ReplaceAll(ch.Html, origPath, intPath)
			}
			if len(ch.Children) > 0 {
				replace(ch.Children)
			}
		}
	}

	replace(chapters)
}

// CreateChapters crea i capitoli nel libro ePUB.
// Processa ricorsivamente la struttura dei capitoli per creare la gerarchia.
func (bb *BookBuilder) CreateChapters(chapters []*Chapter, cssPath string) error {
	slog.Info("Inizio creazione capitoli")

	if err := bb.createChaptersRecursive(chapters, cssPath, ""); err != nil {
		return err
	}

	slog.Info("Capitoli creati con successo")
	return nil
}

// createChaptersRecursive crea i capitoli ricorsivamente.
// Gestisce sia capitoli di primo livello che sotto-sezioni annidate.
func (bb *BookBuilder) createChaptersRecursive(chapters []*Chapter, cssPath, parent string) error {
	for _, chapter := range chapters {
		baseName := ""
		if parent != "" {
			baseName = parent + "__"
		}
		baseName += chapter.Filename
		fileName := baseName + ".xhtml"

		var err error
		if parent != "" {
			_, err = bb.book.AddSubSection(parent, chapter.Html, chapter.Meta.Title, fileName, cssPath)
		} else {
			_, err = bb.book.AddSection(chapter.Html, chapter.Meta.Title, fileName, cssPath)
		}

		if err != nil {
			return fmt.Errorf("errore nella creazione del capitolo %s: %w", chapter.Meta.Title, err)
		}

		if len(chapter.Children) > 0 {
			if err := bb.createChaptersRecursive(chapter.Children, cssPath, fileName); err != nil {
				return err
			}
		}
	}
	return nil
}

// Save salva il libro ePUB su file.
// Il percorso di output è preso dalla configurazione.
func (bb *BookBuilder) Save() error {
	if err := bb.book.Write(bb.config.Output); err != nil {
		return fmt.Errorf("errore durante il salvataggio del libro EPUB: %w", err)
	}

	slog.Info("Libro EPUB salvato con successo", "output", bb.config.Output)
	return nil
}
