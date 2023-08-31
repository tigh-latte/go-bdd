# Go BDD

A cucumber framework for writing integration tests via [go-dog](https://github.com/cucumber/godog).

This library's goals are to provide:
* An easily importable/executable bdd interface.
* An set of basic test steps which can be uses, or have new tests built from.
* An config which can be overridden to point be ran against different environments.

### Creating the test runner

Getting setup is simple enough, you create a directory in your project to house your tests:
```sh
$ mkdir -p test/integration
```

Next, you would create a runner file ending in `_test.go`, and insert the following into it:
```go
//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/tigh-latte/go-bdd/cucumber"
)

func TestMain(t *testing.M) {
	suite := cucumber.NewSuite(
		"my-cool-service",
	)

	os.Exit(suite.Run())
}
```
