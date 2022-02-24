package namespacecreator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateNamespaces(t *testing.T) {
	// given
	namespaceCreator := &NamespaceCreatorImpl{}
	// when
	actualPid, actualHostname, err := namespaceCreator.CreateNamespaces()
	// then
	expectedPid := 1
	expectedHostname := "container"
	assert.Equal(t, expectedPid, actualPid, "should be equal")
	assert.Equal(t, expectedHostname, actualHostname, "should be equal")
	assert.Equal(t, nil, err, "should be equal")
}
