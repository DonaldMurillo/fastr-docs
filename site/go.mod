module github.com/DonaldMurillo/fastr-docs/site

go 1.27.0

require (
	github.com/DonaldMurillo/fastr-docs v0.1.0
	github.com/DonaldMurillo/gofastr v0.82.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	golang.org/x/image v0.45.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Use the checked-out fastr-docs package while this self-hosted site is
// developed alongside the library.
replace github.com/DonaldMurillo/fastr-docs => ..
