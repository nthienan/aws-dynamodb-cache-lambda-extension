package version

import (
	"bytes"
	"runtime"
	"strings"
	"text/template"
)

var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion = runtime.Version()
	GoOS      = runtime.GOOS
	GoArch    = runtime.GOARCH
)

func Print(program string) string {
	var versionTemplate = `
{{.program}}, version: {{.version}} (branch: {{.branch}}, revision: {{.revision}})
    build user:   {{.buildUser}}
    build date:   {{.buildDate}}
    go version:   {{.goVersion}}
    platform:     {{.platform}}
`

	m := map[string]string{
		"program":   program,
		"version":   Version,
		"branch":    Branch,
		"revision":  Revision,
		"buildUser": BuildUser,
		"buildDate": BuildDate,
		"goVersion": GoVersion,
		"platform":  GoOS + "/" + GoArch,
	}

	t := template.Must(template.New("version").Parse(versionTemplate))
	var buffer bytes.Buffer
	if err := t.ExecuteTemplate(&buffer, "version", m); err != nil {
		panic(err)
	}

	return strings.TrimSpace(buffer.String())
}
