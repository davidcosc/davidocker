package containercreator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateContainer(t *testing.T) {
	// given
	containerCreator := &ContainerCreatorImpl{}
	// when
	err := containerCreator.CreateContainerNamespaces([]string{"/proc/self/exe", "-run", "CreateContainerHelper"})
	// then
	assert.Equal(t, nil, err, "should be equal")
}

func CreateContainerHelper() {
	containerCreator := &ContainerCreatorImpl{}
	containerCreator.FinalizeContainer([]string{"echo"})
}
