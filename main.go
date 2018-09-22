package main

import (
	"bufio"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/russross/blackfriday"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	// load markdown file
	mdFile, err := ioutil.ReadFile("./example.md")
	if err != nil {
		log.Fatal(err)
	}
	// convert markdown to html
	htmlSrc := blackfriday.MarkdownCommon(mdFile)
	// replace code-parts with syntax-highlighted parts
	replaced, err := replaceCodeParts(htmlSrc)
	if err != nil {
		log.Fatal(err)
	}
	// read template
	t, err := template.ParseFiles("./template.html")
	if err != nil {
		log.Fatal(err)
	}
	// write css
	hlbuf := bytes.Buffer{}
	hlw := bufio.NewWriter(&hlbuf)
	formatter := html.New(html.WithClasses())
	if err := formatter.WriteCSS(hlw, styles.MonokaiLight); err != nil {
		log.Fatal(err)
	}
	hlw.Flush()
	// write html output
	if err := t.Execute(os.Stdout, struct {
		Content template.HTML
		Style   template.CSS
	}{
		Content: template.HTML(replaced),
		Style:   template.CSS(hlbuf.String()),
	}); err != nil {
		log.Fatal(err)
	}
}

func replaceCodeParts(mdFile []byte) (string, error) {
	byteReader := bytes.NewReader(mdFile)
	doc, err := goquery.NewDocumentFromReader(byteReader)
	if err != nil {
		return "", err
	}
	// find code-parts via selector and replace them with highlighted versions
	var hlErr error
	doc.Find("code[class*=\"language-\"]").Each(func(i int, s *goquery.Selection) {
		if hlErr != nil {
			return
		}
		class, _ := s.Attr("class")
		lang := strings.TrimPrefix(class, "language-")
		oldCode := s.Text()
		lexer := lexers.Get(lang)
		formatter := html.New(html.WithClasses())
		iterator, err := lexer.Tokenise(nil, string(oldCode))
		if err != nil {
			hlErr = err
			return
		}
		b := bytes.Buffer{}
		buf := bufio.NewWriter(&b)
		if err := formatter.Format(buf, styles.GitHub, iterator); err != nil {
			hlErr = err
			return
		}
		if err := buf.Flush(); err != nil {
			hlErr = err
			return
		}
		s.SetHtml(b.String())
	})
	if hlErr != nil {
		return "", hlErr
	}
	new, err := doc.Html()
	if err != nil {
		return "", err
	}
	// replace unnecessarily added html tags
	new = strings.Replace(new, "<html><head></head><body>", "", 1)
	new = strings.Replace(new, "</body></html>", "", 1)
	return new, nil
}
