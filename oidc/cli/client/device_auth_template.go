package client

import (
	"io"
	"text/template"
)

const deviceAuthTemplateText = `
╔════════════════════════════════════════════════════════════════╗
║                    Device Authentication                       ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  To sign in, open the following URL in your browser:           ║
║                                                                ║
║    {{ .VerificationURI }}
║                                                                ║
║  And enter the code:                                           ║
║                                                                ║
║    >>>  {{ .UserCode }}  <<<
║                                                                ║
║  This code expires in {{ .ExpiresInMinutes }} minutes.
║                                                                ║
╚════════════════════════════════════════════════════════════════╝

Waiting for authentication...`

var deviceAuthTemplate = template.Must(template.New("deviceAuth").Parse(deviceAuthTemplateText))

// deviceAuthTemplateData holds the data for rendering the device auth template
type deviceAuthTemplateData struct {
	VerificationURI  string
	UserCode         string
	ExpiresInMinutes int
}

// renderDeviceAuthTemplate renders the device authentication template to the given writer
func renderDeviceAuthTemplate(w io.Writer, data *deviceAuthTemplateData) error {
	return deviceAuthTemplate.Execute(w, data)
}
