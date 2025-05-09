package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog" // Import slog
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	epub "github.com/go-shiori/go-epub"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

func main() {
	slog.Info("Avvio del programma")

	type config struct {
		Input  string `env:"INPUT" envDefault:"/tmp/book" validate:"required,dir"`
		Output string `env:"OUTPUT" envDefault:"./il-manuale-del-buon-dev.epub" validate:"required"`
		Cover  string `env:"COVER" envDefault:"./assets/cover.jpg" validate:"required"`
		Style  string `env:"STYLE" envDefault:"./assets/style.css" validate:"required"`
		Uuid   string `env:"UUID" validate:"required"`
	}

	// parse
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		slog.Error("Errore durante il parsing delle variabili d'ambiente", "error", err)
		panic(err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(cfg); err != nil {
		slog.Error("Errore durante la validazione della configurazione", "error", err)
		panic(fmt.Errorf("failed to parse env vars: %w", err))
	}
	slog.Info("Configurazione caricata con successo", "config", cfg)

	input := cfg.Input
	output := cfg.Output
	uuid := cfg.Uuid

	// Create a new EPUB book
	book, err := epub.NewEpub("Il Libro Open Source")
	if err != nil {
		slog.Error("Errore durante la creazione del libro EPUB", "error", err)
		panic(err)
	}
	slog.Info("Libro EPUB creato con successo")

	// Set the title and author
	book.SetTitle("Il manuale del buon dev")
	book.SetAuthor("Community")

	coverPath, err := book.AddImage(cfg.Cover, "cover.jpg")
	if err != nil {
		panic(err)
	}

	cssPath, err := book.AddCSS(cfg.Style, "style.css")
	if err != nil {
		panic(err)
	}

	book.SetCover(coverPath, "")

	book.SetIdentifier("urn:uuid:" + uuid)
	book.SetLang("it")

	mdPath := input + "/docs/it"

	cv := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			&frontmatter.Extender{},
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
					chromahtml.WrapLongLines(true),
					chromahtml.TabWidth(2),
				),
			),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	chapters, err := getChapters(&cv, mdPath)
	if err != nil {
		slog.Error("Errore durante il caricamento dei capitoli", "error", err)
		panic(err)
	}
	slog.Info("Capitoli caricati con successo", "numero capitoli", len(chapters))

	err = addImages(book, chapters, cfg.Input)
	if err != nil {
		slog.Error("Errore durante l'aggiunta delle immagini", "error", err)
		panic(err)
	}
	slog.Info("Immagini aggiunte con successo")

	err = createChapters(book, chapters, cssPath, "")
	if err != nil {
		slog.Error("Errore durante la creazione dei capitoli", "error", err)
		panic(err)
	}
	slog.Info("Capitoli creati con successo")

	// Save the EPUB book
	err = book.Write(output)
	if err != nil {
		slog.Error("Errore durante il salvataggio del libro EPUB", "error", err)
		panic(err)
	}
	slog.Info("Libro EPUB salvato con successo", "output", output)
}

func addImages(book *epub.Epub, chapters []*Chapter, actPath string) (err error) {
	slog.Info("Inizio aggiunta immagini")
	passedImages := map[string]string{}
	for _, chapter := range chapters {
		for _, image := range chapter.Images {
			intPath, ok := passedImages[image]
			if !ok {
				fsPath := strings.Replace(image, "/book", actPath, 1)
				fileName := strings.ReplaceAll(strings.TrimLeft(image, "/"), "/", "_")
				intPath, err = book.AddImage(fsPath, fileName)
				if err != nil {
					err = fmt.Errorf("failed to add image %s: %w", fsPath, err)
					return
				}
				passedImages[image] = intPath
			}

			// Replace the image path in the HTML content
			chapter.Html = strings.ReplaceAll(chapter.Html, image, intPath)
		}

		if len(chapter.Children) > 0 {
			err = addImages(book, chapter.Children, actPath)
			if err != nil {
				return
			}
		}
	}
	slog.Info("Immagini aggiunte con successo")
	return nil
}

func createChapters(book *epub.Epub, chapters []*Chapter, cssPath string, parent string) error {
	slog.Info("Inizio creazione capitoli")
	for _, chapter := range chapters {
		// Add a chapter
		baseName := ""
		if parent != "" {
			// Add a sub-chapter
			baseName = parent + "__"
		}
		baseName += chapter.Filename

		fileName := baseName + ".xhtml"

		var err error
		if parent != "" {
			// Add a sub-chapter
			_, err = book.AddSubSection(parent, chapter.Html, chapter.Meta.Title, fileName, cssPath)
		} else {
			_, err = book.AddSection(chapter.Html, chapter.Meta.Title, fileName, cssPath)
		}

		if err != nil {
			return err
		}

		// Recursively add child chapters
		if len(chapter.Children) > 0 {
			err := createChapters(book, chapter.Children, cssPath, fileName)
			if err != nil {
				return err
			}
		}
	}
	slog.Info("Capitoli creati con successo")
	return nil
}

type Chapter struct {
	Filename string
	Meta     ChapterMeta
	Content  string
	Html     string
	Children []*Chapter
	Images   []string
}

type ChapterMeta struct {
	Title string `yaml:"title"`
	Order int    `yaml:"nav_order"`
}

func getChapters(cv *goldmark.Markdown, mdPath string) (list []*Chapter, err error) {
	slog.Info("Caricamento capitoli", "path", mdPath)
	// Read markdown files from the directory
	files, err := fs.ReadDir(os.DirFS(mdPath), ".")
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			content, err := os.ReadFile(filepath.Join(mdPath, file.Name()))
			if err != nil {
				panic(err)
			}

			// Remove Jekyll-specific codes
			tempContent := string(content)
			tempContent = regexp.MustCompile(`\{:[^}]*\}`).ReplaceAllString(tempContent, "")
			tempContent = strings.ReplaceAll(tempContent, "- TOC", "")
			content = []byte(tempContent)

			// Find all images paths
			images := []string{}
			re := regexp.MustCompile(`!\[.*?\]\((.*?)\)`)
			matches := re.FindAllSubmatch(content, -1)
			for _, match := range matches {
				if len(match) > 1 {
					images = append(images, string(match[1]))
				}
			}

			ctx := parser.NewContext()
			var buf bytes.Buffer
			if err := (*cv).Convert(content, &buf, parser.WithContext(ctx)); err != nil {
				panic(err)
			}

			d := frontmatter.Get(ctx)

			meta := ChapterMeta{}
			err = d.Decode(&meta)
			if err != nil {
				return nil, err
			}

			html := buf.String()
			html = strings.ReplaceAll(html, "<br>", "<br/>")

			fileName := strings.TrimSuffix(file.Name(), ".md")

			ch := Chapter{
				Filename: fileName,
				Meta:     meta,
				Content:  string(content),
				Html:     html,
				Images:   images,
			}

			// Add children chapters if any
			dirPath := filepath.Join(mdPath, file.Name()[:len(file.Name())-3])
			if info, err := os.Stat(dirPath); err == nil && info.IsDir() {
				children, err := getChapters(cv, dirPath)
				if err != nil {
					return nil, err
				}
				ch.Children = children
			}

			list = append(list, &ch)
		}
	}

	slices.SortFunc(list, func(a, b *Chapter) int {
		return a.Meta.Order - b.Meta.Order
	})

	slog.Info("Capitoli caricati", "numero capitoli", len(list))
	return list, nil
}
