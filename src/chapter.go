package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/frontmatter"
)

// Chapter rappresenta un singolo capitolo o sezione del libro.
// I capitoli possono essere annidati per creare una gerarchia nel sommario.
type Chapter struct {
	// Filename è il nome base senza estensione (es. "introduzione")
	Filename string

	// Meta contiene i metadati estratti dal frontmatter YAML
	Meta ChapterMeta

	// Content è il contenuto Markdown originale dopo la rimozione dei codici Jekyll
	Content string

	// Html è il contenuto HTML convertito pronto per l'inclusione nell'ePUB
	Html string

	// Children sono i sotto-capitoli annidati sotto questo capitolo
	Children []*Chapter

	// Images sono i percorsi delle immagini referenziate nel Markdown di questo capitolo
	Images []string
}

// ChapterMeta contiene i metadati estratti dal frontmatter YAML.
// Segue il formato del frontmatter Jekyll usato nel repository sorgente.
type ChapterMeta struct {
	// Title è il titolo leggibile del capitolo
	Title string `yaml:"title"`

	// Order determina la posizione nel sommario (numeri più bassi vengono prima)
	Order int `yaml:"nav_order"`
}

// GetChapters legge e converte i file markdown in capitoli
// Elabora i file in parallelo per migliorare le performance
func GetChapters(cv *goldmark.Markdown, mdPath string) ([]*Chapter, error) {
	slog.Debug("Caricamento capitoli", "path", mdPath)

	files, err := fs.ReadDir(os.DirFS(mdPath), ".")
	if err != nil {
		return nil, fmt.Errorf("errore nella lettura della directory %s: %w", mdPath, err)
	}

	// Filtra i file markdown
	var markdownFiles []fs.DirEntry
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			markdownFiles = append(markdownFiles, file)
		}
	}

	// Processa i file in parallelo
	list := make([]*Chapter, len(markdownFiles))
	var wg sync.WaitGroup
	errChan := make(chan error, len(markdownFiles))

	for i, file := range markdownFiles {
		wg.Add(1)
		go func(idx int, f fs.DirEntry) {
			defer wg.Done()
			chapter, err := processMarkdownFile(cv, mdPath, f.Name())
			if err != nil {
				errChan <- err
				return
			}
			list[idx] = chapter
		}(i, file)
	}

	wg.Wait()
	close(errChan)

	// Verifica errori
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	// Ordina i capitoli
	slices.SortFunc(list, func(a, b *Chapter) int {
		return a.Meta.Order - b.Meta.Order
	})

	slog.Info("Capitoli caricati", "numero capitoli", len(list))
	return list, nil
}

// processMarkdownFile elabora un singolo file Markdown.
// Legge il file, rimuove i codici Jekyll, estrae le immagini,
// converte in HTML e carica eventuali sotto-capitoli.
func processMarkdownFile(cv *goldmark.Markdown, mdPath, fileName string) (*Chapter, error) {
	content, err := os.ReadFile(filepath.Join(mdPath, fileName))
	if err != nil {
		return nil, fmt.Errorf("errore nella lettura del file %s: %w", fileName, err)
	}

	// Pulisce il contenuto dai codici Jekyll
	content = cleanJekyllContent(content)

	// Estrae le immagini
	images := extractImages(content)

	// Converte markdown in HTML
	html, meta, err := convertMarkdownToHTML(cv, content)
	if err != nil {
		return nil, err
	}

	chapter := &Chapter{
		Filename: strings.TrimSuffix(fileName, ".md"),
		Meta:     meta,
		Content:  string(content),
		Html:     html,
		Images:   images,
	}

	// Carica capitoli figli se esistono
	dirPath := filepath.Join(mdPath, fileName[:len(fileName)-3])
	if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
		children, err := GetChapters(cv, dirPath)
		if err != nil {
			return nil, err
		}
		chapter.Children = children
	}

	return chapter, nil
}

// cleanJekyllContent rimuove i codici specifici di Jekyll dal contenuto.
// Rimuove pattern come {: .class} e marcatori TOC che non sono compatibili con ePUB.
func cleanJekyllContent(content []byte) []byte {
	tempContent := string(content)
	tempContent = regexp.MustCompile(`\{:[^}]*\}`).ReplaceAllString(tempContent, "")
	tempContent = strings.ReplaceAll(tempContent, "- TOC", "")
	return []byte(tempContent)
}

// extractImages estrae i percorsi delle immagini dal contenuto Markdown.
// Usa una regex per trovare tutti i pattern ![alt](path) e restituisce i percorsi.
func extractImages(content []byte) []string {
	var images []string
	re := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
	matches := re.FindAllSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			images = append(images, string(match[1]))
		}
	}
	return images
}

// convertMarkdownToHTML converte il contenuto Markdown in HTML ed estrae i metadati.
// Usa Goldmark per il parsing e converte i tag <br> in <br/> per conformità XHTML.
// Restituisce l'HTML, i metadati del frontmatter e un eventuale errore.
func convertMarkdownToHTML(cv *goldmark.Markdown, content []byte) (string, ChapterMeta, error) {
	ctx := parser.NewContext()
	var buf bytes.Buffer

	if err := (*cv).Convert(content, &buf, parser.WithContext(ctx)); err != nil {
		return "", ChapterMeta{}, fmt.Errorf("errore nella conversione markdown: %w", err)
	}

	d := frontmatter.Get(ctx)
	var meta ChapterMeta
	if err := d.Decode(&meta); err != nil {
		return "", ChapterMeta{}, fmt.Errorf("errore nell'estrazione dei metadati: %w", err)
	}

	html := buf.String()
	html = strings.ReplaceAll(html, "<br>", "<br/>")

	return html, meta, nil
}
