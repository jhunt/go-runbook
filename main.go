package main

import (
	fmt "github.com/jhunt/go-ansi"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/golang-commonmark/markdown"
	"github.com/jhunt/go-cli"
	"gopkg.in/yaml.v2"
)

type Runbook struct {
	Title    string `yaml:"title"`
	Subtitle string `yaml:"subtitle"`
	URL      string `yaml:"url"`
	Source   string `yaml:"source"`
	Intro    string `yaml:"intro"`
	Contents string
}

type Manifest struct {
	Runbooks []*Runbook `yaml:"runbooks"`
}

func bail(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "@R{error: %s}\n", err)
		os.Exit(1)
	}
}

func main() {
	var opts struct {
		Index string `cli:"-i, --index"`
		Topic string `cli:"-t, --topic"`
		Root  string `cli:"-r, --root"`
	}

	_, args, err := cli.Parse(&opts)
	bail(err)

	if len(args) != 1 || opts.Index == "" || opts.Topic == "" || opts.Root == "" {
		fmt.Fprintf(os.Stderr, "USAGE: @Y{runbook} -i index.tpl -t topic.tpl -r output/dir/ path/to/toc.yml\n")
		os.Exit(1)
	}

	b, err := ioutil.ReadFile(opts.Index)
	bail(err)

	index, err := template.New("index").Parse(string(b))
	bail(err)

	b, err = ioutil.ReadFile(opts.Topic)
	bail(err)

	topic, err := template.New("topic").Parse(string(b))
	bail(err)

	b, err = ioutil.ReadFile(args[0])
	bail(err)

	var manifest Manifest
	err = yaml.Unmarshal(b, &manifest)
	bail(err)

	md := markdown.New()
	for _, book := range manifest.Runbooks {
		if book.URL == "" {
			book.URL = fmt.Sprintf("%s.html", strings.TrimSuffix(book.Source, ".md"))
		}
		book.Intro = md.RenderToString([]byte(book.Intro))

		b, err = ioutil.ReadFile(book.Source)
		bail(err)

		book.Contents = md.RenderToString(b)
	}

	out, err := os.Create(fmt.Sprintf("%s/index.html", opts.Root))
	bail(err)

	err = index.Execute(out, manifest)
	bail(err)

	out.Close()

	for _, book := range manifest.Runbooks {
		out, err = os.Create(fmt.Sprintf("%s/%s", opts.Root, book.URL))
		bail(err)

		err = topic.Execute(out, book)
		bail(err)

		out.Close()
	}
}
