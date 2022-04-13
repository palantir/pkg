module github.com/palantir/pkg/safeyaml

go 1.18

require (
	github.com/palantir/pkg v1.0.1
	github.com/palantir/pkg/transform v1.0.1
	github.com/stretchr/testify v1.4.0
	gopkg.in/yaml.v2 v2.2.2
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.1
