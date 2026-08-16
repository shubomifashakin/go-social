package templates

import _ "embed"

//go:embed welcome.html
var WelcomeTemplate string

//go:embed delete-request.html
var DeleteRequestTemplate string
