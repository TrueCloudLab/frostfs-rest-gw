package handlers

import (
	"testing"

	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi/operations"
	"github.com/TrueCloudLab/frostfs-rest-gw/internal/util"
	"github.com/stretchr/testify/require"
)

func TestPrepareOffset(t *testing.T) {
	for _, tc := range []struct {
		err            bool
		expectedOffset uint64
		expectedLength uint64
		params         operations.GetObjectInfoParams
		objSize        uint64
	}{
		{
			params: operations.GetObjectInfoParams{
				RangeLength: util.NewInteger(1),
				RangeOffset: util.NewInteger(0),
			},
			objSize:        1,
			expectedOffset: 0,
			expectedLength: 1,
		},
		{
			params: operations.GetObjectInfoParams{
				RangeLength: util.NewInteger(3),
				RangeOffset: util.NewInteger(1),
			},
			objSize:        5,
			expectedOffset: 1,
			expectedLength: 3,
		},
		{
			objSize:        1,
			expectedOffset: 0,
			expectedLength: 1,
		},
		{
			err: true,
			params: operations.GetObjectInfoParams{
				RangeLength: util.NewInteger(1),
				RangeOffset: nil,
			},
		},
		{
			err: true,
			params: operations.GetObjectInfoParams{
				RangeLength: nil,
				RangeOffset: util.NewInteger(1),
			},
		},
		{
			err: true,
			params: operations.GetObjectInfoParams{
				RangeLength: util.NewInteger(1),
				RangeOffset: util.NewInteger(0),
			},
			objSize: 0,
		},
	} {
		t.Run("", func(t *testing.T) {
			offset, length, err := prepareOffsetLength(tc.params, tc.objSize)
			if tc.err {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedOffset, offset)
			require.Equal(t, tc.expectedLength, length)
		})
	}
}
