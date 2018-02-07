package osutil_test

import (
	"fmt"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/palantir/pkg/osutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidRegexPath(t *testing.T) {
	currentAbs, err := filepath.Abs(".")
	require.NoError(t, err)

	fmtString := `here is a path: %s with some text after`
	myRegexDef := fmt.Sprintf(
		"^"+fmtString+"$",
		osutil.MakeValidRegexPath(currentAbs),
	)

	assert.Regexp(t, regexp.MustCompile(myRegexDef), fmt.Sprintf(fmtString, currentAbs))
}
