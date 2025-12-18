# book-epub

[![Build EPUB and Release](https://github.com/Il-Libro-Open-Source/book-epub/actions/workflows/build-and-release.yml/badge.svg)](https://github.com/Il-Libro-Open-Source/book-epub/actions/workflows/build-and-release.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Il-Libro-Open-Source/book-epub)](go.mod)

La versione ePUB del [Il Libro Open Source](https://github.com/Il-Libro-Open-Source/book) e gli script necessari a generarla.

## 📖 Descrizione

Questo repository contiene il generatore e le risorse per creare la versione ePUB del [Libro Open Source](https://github.com/Il-Libro-Open-Source/book). Il processo converte i file Markdown del libro in un file ePUB pronto per la distribuzione, con supporto anche per MOBI e PDF.

**Caratteristiche principali:**

- 🚀 Conversione Markdown → ePUB con syntax highlighting
- 📱 Supporto multi-formato (ePUB, MOBI, PDF)
- 🎨 Styling CSS personalizzato con supporto dark mode
- 🖼️ Gestione automatica di immagini e copertina
- 📑 Gerarchia capitoli con frontmatter YAML
- ⚡ Elaborazione parallela per performance ottimali
- 🔄 CI/CD automatizzato con GitHub Actions

## 🛠️ Requisiti

- **Go** >= 1.24
- **Git**
- **Calibre** (opzionale, per conversione MOBI/PDF)
- **Task** (opzionale, per usare i task predefiniti)

### Installazione Calibre

```sh
# macOS
brew install --cask calibre

# Ubuntu/Debian
sudo apt-get install calibre

# Windows
# Scarica da https://calibre-ebook.com/download
```

## 🚀 Installazione e Build

### Build locale

1. **Clona questo repository:**

   ```sh
   git clone https://github.com/Il-Libro-Open-Source/book-epub.git
   cd book-epub
   ```

2. **Clona il repository del libro:**

   ```sh
   git clone https://github.com/Il-Libro-Open-Source/book.git /tmp/book
   ```

3. **Costruisci il generatore:**

   ```sh
   go build -o ./bin/epub-generator ./src
   ```

4. **Genera il file ePUB:**

   ```sh
   UUID=f9298b0f-bea1-4cb6-a601-2a35027bd44e ./bin/epub-generator
   ```

   Il file `il-manuale-del-buon-dev.epub` verrà generato nella directory corrente.

### Build con Task

Se hai [Task](https://taskfile.dev/) installato, puoi usare i comandi predefiniti:

```sh
# Mostra tutti i task disponibili
task --list

# Genera solo ePUB
task generate

# Genera ePUB, MOBI e PDF
task generate-all

# Esegue test e validazioni
task test

# Task di sviluppo (clean, format, lint, test)
task dev
```

## ⚙️ Configurazione

### Variabili d'ambiente

Il generatore può essere configurato tramite variabili d'ambiente:

| Variabile | Descrizione | Default | Obbligatoria |
|-----------|-------------|---------|--------------|
| `INPUT` | Percorso alla cartella del libro | `/tmp/book` | No |
| `OUTPUT` | Percorso di output del file ePUB | `./il-manuale-del-buon-dev.epub` | No |
| `COVER` | Percorso dell'immagine di copertina | `./assets/cover.jpg` | No |
| `STYLE` | Percorso del file CSS | `./assets/style.css` | No |
| `UUID` | UUID del libro (formato URN) | - | **Sì** |

### Esempio di utilizzo

```sh
# Configurazione personalizzata
INPUT=/path/to/book \
OUTPUT=./output/my-book.epub \
COVER=./custom-cover.jpg \
STYLE=./custom-style.css \
UUID=f9298b0f-bea1-4cb6-a601-2a35027bd44e \
./bin/epub-generator
```

### Generazione UUID

```sh
# Su macOS/Linux
uuidgen | tr '[:upper:]' '[:lower:]'

# Con Go
go run -c "package main; import \"github.com/google/uuid\"; func main() { println(uuid.New().String()) }"
```

## 📚 Generazione Multi-Formato

### ePUB → MOBI

```sh
# Richiede Calibre
ebook-convert il-manuale-del-buon-dev.epub il-manuale-del-buon-dev.mobi --verbose
```

### ePUB → PDF

```sh
# Richiede Calibre
ebook-convert il-manuale-del-buon-dev.epub il-manuale-del-buon-dev.pdf \
  --verbose \
  --cover assets/cover-pdf.jpg \
  --remove-first-image \
  --pdf-default-font-size 14 \
  --pdf-page-numbers
```

### Con Task

```sh
# Genera tutti i formati automaticamente
task generate-all
```

## 🧪 Test e Sviluppo

### Eseguire i test

```sh
# Test unitari
go test -v ./src/...

# Test con coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./src/...
go tool cover -html=coverage.out -o coverage.html
```

### Formattazione e Linting

```sh
# Formattazione
go fmt ./src/...

# Linting
golangci-lint run ./src/...

# Con Task
task lint
task fmt
```

### Validazione ePUB

```sh
# Richiede epubcheck
epubcheck il-manuale-del-buon-dev.epub

# macOS
brew install epubcheck

# Ubuntu/Debian
sudo apt-get install epubcheck
```

## 🏗️ Architettura

Il progetto è strutturato in moduli separati per migliorare manutenibilità e testabilità:

```text
src/
├── main.go          # Entry point dell'applicazione
├── config.go        # Gestione configurazione e validazione
├── chapter.go       # Elaborazione capitoli e conversione Markdown
├── markdown.go      # Configurazione convertitore Goldmark
├── epub.go          # Costruzione e salvataggio ePUB
└── *_test.go        # Test unitari
```

### Flusso di elaborazione

1. **Caricamento configurazione** - Validazione variabili d'ambiente
2. **Lettura capitoli** - Scansione directory Markdown (parallela)
3. **Conversione Markdown → HTML** - Con syntax highlighting
4. **Elaborazione immagini** - Validazione e inclusione (parallela)
5. **Creazione ePUB** - Assemblaggio con metadati e stili
6. **Salvataggio** - Output file ePUB

## 🔧 Troubleshooting

### Errore: "cover file not found"

Assicurati che il file di copertina esista nel percorso specificato:

```sh
ls -la assets/cover.jpg
```

Se il file manca, puoi specificare un percorso alternativo:

```sh
COVER=/path/to/your/cover.jpg ./bin/epub-generator
```

### Errore: "failed to parse env vars: UUID is required"

L'UUID è obbligatorio. Generane uno nuovo:

```sh
UUID=$(uuidgen | tr '[:upper:]' '[:lower:]') ./bin/epub-generator
```

### Errore: "ebook-convert not found"

Calibre non è installato. Segui le [istruzioni di installazione](#installazione-calibre).

### Immagini mancanti nell'ePUB

Controlla i log per vedere quali immagini non sono state trovate:

```sh
LOG_LEVEL=DEBUG ./bin/epub-generator 2>&1 | grep "Immagine non trovata"
```

### Errori di validazione ePUB

Usa `epubcheck` per identificare problemi:

```sh
epubcheck il-manuale-del-buon-dev.epub
```

## 🤝 Contribuire

Contributi sono benvenuti! Per contribuire:

1. **Fork** il repository
2. **Crea** un branch per la tua feature (`git checkout -b feature/amazing-feature`)
3. **Commit** le modifiche (`git commit -m 'Add amazing feature'`)
4. **Push** al branch (`git push origin feature/amazing-feature`)
5. **Apri** una Pull Request

### Linee guida

- Assicurati che i test passino: `go test ./src/...`
- Formatta il codice: `go fmt ./src/...`
- Aggiungi test per nuove funzionalità
- Documenta le funzioni pubbliche con GoDoc
- Segui le convenzioni Go standard

### Reporting bug

Apri una [issue](https://github.com/Il-Libro-Open-Source/book-epub/issues) con:

- Descrizione del problema
- Passi per riprodurre
- Output del comando con `LOG_LEVEL=DEBUG`
- Versione di Go e sistema operativo

## 📋 Workflow GitHub Actions

Il workflow `build-and-release.yml` automatizza:

- ✅ Build del generatore
- ✅ Esecuzione dei test
- ✅ Generazione ePUB, MOBI e PDF
- ✅ Upload artifacts per ogni build
- ✅ Release automatica per i tag

### Trigger manuale

Puoi triggerare manualmente il workflow con opzione di release:

1. Vai su **Actions** → **Build EPUB and Release**
2. Clicca **Run workflow**
3. Seleziona `do_release: true` per creare una release
4. (Opzionale) Specifica un `release_tag`

## 📚 Riferimenti

- [Il Libro Open Source](https://github.com/Il-Libro-Open-Source/book)
- [go-epub](https://github.com/go-shiori/go-epub) - Libreria per generazione ePUB
- [goldmark](https://github.com/yuin/goldmark) - Parser Markdown
- [Calibre](https://calibre-ebook.com/) - Toolkit per e-book
- [ePUB Specification](https://www.w3.org/publishing/epub3/)
- [Task](https://taskfile.dev/) - Task runner

## 📄 Licenza

Questo progetto è distribuito sotto licenza MIT. Vedi il file [LICENSE](LICENSE) per maggiori dettagli.

## 👥 Autori

- **Community** - [Il Libro Open Source](https://github.com/Il-Libro-Open-Source)

---

Fatto con ❤️ dalla community di [Il Libro Open Source](https://github.com/Il-Libro-Open-Source)
