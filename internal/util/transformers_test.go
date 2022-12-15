package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	apistatus "github.com/TrueCloudLab/frostfs-sdk-go/client/status"
	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	apiErr := fmt.Errorf("some context: %w", apistatus.ContainerNotFound{})

	resp := NewErrorResponse(apiErr)
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	require.Equal(t, `{"code":3072,"message":"some context: status: code = 3072","type":"API"}`, string(data))

	gwErr := fmt.Errorf("some context: %w", errors.New("sanity check error"))

	resp = NewErrorResponse(gwErr)
	data, err = json.Marshal(resp)
	require.NoError(t, err)
	require.Equal(t, `{"message":"some context: sanity check error","type":"GW"}`, string(data))
}
