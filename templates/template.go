package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"time"
)

//go:embed *.tmpl
var tmpls embed.FS

type (
	Templater struct{}
	Template  string
)

const (
	NotFound404Template       Template = "404NotFound.tmpl"
	AdminTemplate             Template = "admin.tmpl"
	AdminErrorTemplate        Template = "adminError.tmpl"
	ElectionTemplate          Template = "election.tmpl"
	ElectionsTemplate         Template = "elections.tmpl"
	EmailTemplate             Template = "email.tmpl"
	ErrorTemplate             Template = "error.tmpl"
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

	t1 := template.New("_base.tmpl")
	t1.Funcs(template.FuncMap{
		"thisYear": func() int { return time.Now().Year() },
		"inc": func(a int) int {
			return a + 1
		},
		"even": func(a int) bool {
			return a%2 == 0
		},
		"incUInt64": func(a uint64) uint64 {
			return a + 1
		},
		"divPercent": func(a, b uint64) string {
			return fmt.Sprintf("%03.2f%%", (float64(a)/float64(b))*float64(100))
		},
	})

	t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_top.tmpl", "_footer.tmpl", string(mainTmpl))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return t1.Execute(w, data)
}

func (t *Templater) RenderEmail(emailTemplate Template) *template.Template {
	return template.Must(template.New("email.tmpl").ParseFS(tmpls, emailTemplate.GetString()))
}

// This section is for go template linter
var (
	AllTemplates = [][]string{
		{"404NotFound.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"admin.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"adminError.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"election.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"elections.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"email.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"error.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"home.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"qr.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"registered.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"registration.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"registrationError.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"vote.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"voted.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"voteError.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
		{"voters.tmpl", "_base.tmpl", "_top.tmpl", "_footer.tmpl"},
	}

	_ = AllTemplates
)
