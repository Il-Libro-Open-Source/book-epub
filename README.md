# book-epub

La versione ePUB del Libro Open Source e gli script necessari a generarla.

## Descrizione

Questo repository contiene il generatore e le risorse per creare la versione ePUB del [Libro Open Source](https://github.com/Il-Libro-Open-Source/book). Il processo converte i file Markdown del libro in un file ePUB pronto per la distribuzione.

## Requisiti

- Go >= 1.24
- Git

## Build locale

1. Clona questo repository:

   ```sh
   git clone https://github.com/Il-Libro-Open-Source/book-epub.git
   cd book-epub
   ```

2. Clona il repository del libro:

   ```sh
   git clone https://github.com/Il-Libro-Open-Source/book.git /tmp/book
   ```

3. Costruisci il generatore:

   ```sh
   go build -o ./epub-generator ./src
   ```

4. Genera il file ePUB:

   ```sh
   UUID=f9298b0f-bea1-4cb6-a601-2a35027bd44e ./epub-generator
   ```

   Il file `il-manuale-del-buon-dev.epub` verrà generato nella directory corrente.

## Variabili d'ambiente

- `INPUT`: percorso alla cartella del libro (default: `/tmp/book`)
- `OUTPUT`: percorso di output del file ePUB (default: `./il-manuale-del-buon-dev.epub`)
- `COVER`: percorso dell'immagine di copertina (default: `./assets/cover.jpg`)
- `STYLE`: percorso del file CSS (default: `./assets/style.css`)
- `UUID`: UUID del libro (obbligatorio)

## Workflow GitHub Actions

Il workflow `build-and-release.yml` compila il generatore, crea il file ePUB e lo pubblica come artifact e, se presente un tag, come release su GitHub.

## Riferimenti

- [Il Libro Open Source](https://github.com/Il-Libro-Open-Source/book)
- [go-epub](https://github.com/go-shiori/go-epub)
- [goldmark](https://github.com/yuin/goldmark)
