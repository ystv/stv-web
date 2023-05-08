package templates

import (
	"embed"
	"fmt"
	"github.com/ystv/stv_web/structs"
	"html/template"
	"io"
	"strings"
	"time"
)

//go:embed *.tmpl
var tmpls embed.FS

type (
	Templater struct{}
	Template  string
)

const (
	AdminTemplate             Template = "admin.tmpl"
	ElectionTemplate          Template = "election.tmpl"
	ElectionsTemplate         Template = "elections.tmpl"
	EmailTemplate             Template = "email.tmpl"
	ErrorTemplate             Template = "errors.tmpl"
	HomeTemplate              Template = "home.tmpl"
	QRTemplate                Template = "qr.tmpl"
	RegisteredTemplate        Template = "registered.tmpl"
	RegistrationTemplate      Template = "registration.tmpl"
	RegistrationErrorTemplate Template = "registrationError.tmpl"
	VoteTemplate              Template = "vote.tmpl"
	VotedTemplate             Template = "voted.tmpl"
	VoteErrorTemplate         Template = "voteError.tmpl"
	VotersTemplate            Template = "voters.tmpl"
)

func (t Template) GetString() string {
	return string(t)
}

func (t *Templater) RenderTemplate(w io.Writer, data interface{}, mainTmpl Template) error {
	var err error

	td := structs.Globals{
		PageData: data,
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

			var parts []string

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
		"inc": func(a int) int {
			return a + 1
		},
		"even": func(a int) bool {
			return a%2 == 0
		},
		"incUInt64": func(a uint64) uint64 {
			return a + 1
		},
	})

	t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_top.tmpl", "_footer.tmpl", string(mainTmpl))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return t1.Execute(w, td)
}

func (t *Templater) RenderEmail(emailTemplate Template) *template.Template {
	return template.Must(template.New("email.tmpl").ParseFS(tmpls, emailTemplate.GetString()))
}

func renderHTML(value interface{}) template.HTML {
	return template.HTML(fmt.Sprint(value))
}
