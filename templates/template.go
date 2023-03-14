package templates

import (
	"embed"
	"fmt"
	"github.com/ystv/stv_web/structs"
	"html/template"
	"io"
	"path/filepath"
	"strings"
	"time"
)

const TemplatePrefix = "templates/"

//go:embed *.tmpl
var tmpls1 embed.FS

type Templater struct {
	dashboard *template.Template
}

var (
	funcs = template.FuncMap{
		"cleantime": cleanTime,
	}
	BaseTemplates = []string{
		"_base.tmpl",
		"_top.tmpl",
		"_footer.tmpl",
	}
)

func cleanTime(t time.Time) string {
	return t.Format(time.RFC1123Z)
}

func (t *Templater) Page(w io.Writer, p structs.PageParams) error {
	return t.dashboard.Execute(w, p)
}

func (t *Templater) RenderTemplate(w io.Writer, context structs.PageParams, data interface{}, mainTmpl string, addTmpls ...string) error {
	_ = tmpls1
	var err error

	td := structs.Globals{
		PageParams: context,
		PageData:   data,
	}

	ownTmpls := append(addTmpls, mainTmpl)
	baseTmpls := append(BaseTemplates, ownTmpls...)

	var tmpls []string
	for _, baseTmpl := range baseTmpls {
		tmpls = append(tmpls, filepath.Join(TemplatePrefix, baseTmpl))
	}

	t1 := template.New("_base.tmpl")
	t1.Funcs(template.FuncMap{
		"html":      renderHTML,
		"stripHtml": StripHTML,
		"formatDuration": func(d time.Duration) string {
			days := int64(d.Hours()) / 24
			hours := int64(d.Hours()) % 24
			minutes := int64(d.Minutes()) % 60
			seconds := int64(d.Seconds()) % 60

			segments := []struct {
				name  string
				value int64
			}{
				{"Day", days},
				{"Hour", hours},
				{"Min", minutes},
				{"Sec", seconds},
			}

			parts := []string{}

			for _, s := range segments {
				if s.value == 0 {
					continue
				}
				plural := ""
				if s.value != 1 {
					plural = "s"
				}

				parts = append(parts, fmt.Sprintf("%d %s%s", s.value, s.name, plural))
			}
			return strings.Join(parts, " ")
		},
		"formatTime": func(fmt string, t time.Time) string {
			return t.Format(fmt)
		},
		"now": func() time.Time {
			return time.Now()
		},
		"thisYear": func() int {
			return time.Now().Year()
		},
		"add": func(a, b int) int {
			return a + b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			return a / b
		},
		"inc": func(a int) int {
			return a + 1
		},
		"even": func(a int) bool {
			return a%2 == 0
		},
	})
	t1, err = t1.ParseFiles(tmpls...)
	if err != nil {
		return err
	}

	return t1.Execute(w, td)
}

func renderHTML(value interface{}) template.HTML {
	return template.HTML(fmt.Sprint(value))
}
