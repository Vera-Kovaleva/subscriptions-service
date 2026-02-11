package pointer_test

import (
	"testing"

	"github.com/Vera-Kovaleva/subscriptions-service/internal/infra/pointer"

	"github.com/stretchr/testify/require"
)

func TestRef(t *testing.T) {
	t.Parallel()

	value := "some value"

	require.Equal(t, value, *pointer.Ref(value))
}
