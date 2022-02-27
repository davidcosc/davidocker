package namespacecreator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFinalizeNamespaces(t *testing.T) {
	// given
	namespaceCreator := &NamespaceCreatorImpl{}
	// when
	err := namespaceCreator.FinalizeNamespaces()
	// then
	assert.Equal(t, nil, err, "should be equal")
}
