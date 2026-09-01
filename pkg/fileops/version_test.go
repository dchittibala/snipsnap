package fileops_test

import (
	"testing"

	"github.com/dchittibala/snipsnap/pkg/fileops"
)

func TestVersionDefaults(t *testing.T) {
	if fileops.Version == "" {
		t.Error("expected Version to have a default value, got empty string")
	}

	if fileops.Commit == "" {
		t.Error("expected Commit to have a default value, got empty string")
	}

	if fileops.BuildDate == "" {
		t.Error("expected BuildDate to have a default value, got empty string")
	}
}
