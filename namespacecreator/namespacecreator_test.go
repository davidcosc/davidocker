package namespacecreator

import (
        "testing"
        "github.com/stretchr/testify/assert"
)

func testCreateNamespaces(t *testing.T) {
        //given
        //when
        actual := CreateNamespaces()
        //then
        expected := "Creating namespaces...[ok]"
        assert.Equal(t, expected, actual, "should be equal")
}
