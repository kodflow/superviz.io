package pkgmanager_test

import (
	"testing"

	"github.com/kodflow/superviz.io/internal/infrastructure/pkgmanager"
	"github.com/stretchr/testify/assert"
)

func TestDetect_LiveEnvironment(t *testing.T) {
	mgr, err := pkgmanager.Detect()

	// soit on a détecté un gestionnaire valide
	if err == nil {
		assert.NotNil(t, mgr)
		assert.NotEmpty(t, mgr.Name())
		t.Logf("Detected package manager: %s", mgr.Name())
		return
	}

	// soit on est dans un environnement trop épuré (e.g. scratch)
	assert.Nil(t, mgr)
	t.Logf("No package manager detected: %v", err)
}
