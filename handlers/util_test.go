package handlers

import (
	"math"
	"strconv"
	"testing"
	"time"

	objectv2 "github.com/TrueCloudLab/frostfs-api-go/v2/object"
	"github.com/stretchr/testify/require"
)

func TestPrepareExpirationHeader(t *testing.T) {
	tomorrow := time.Now().Add(24 * time.Hour)
	tomorrowUnix := tomorrow.Unix()
	tomorrowUnixNano := tomorrow.UnixNano()
	tomorrowUnixMilli := tomorrowUnixNano / 1e6

	epoch := "100"
	duration := "24h"
	timestampSec := strconv.FormatInt(tomorrowUnix, 10)
	timestampMilli := strconv.FormatInt(tomorrowUnixMilli, 10)
	timestampNano := strconv.FormatInt(tomorrowUnixNano, 10)

	defaultDurations := &epochDurations{
		currentEpoch:  10,
		msPerBlock:    1000,
		blockPerEpoch: 101,
	}

	msPerBlock := defaultDurations.blockPerEpoch * uint64(defaultDurations.msPerBlock)
	epochPerDay := uint64((24 * time.Hour).Milliseconds()) / msPerBlock
	if uint64((24*time.Hour).Milliseconds())%msPerBlock != 0 {
		epochPerDay++
	}

	defaultExpEpoch := strconv.FormatUint(defaultDurations.currentEpoch+epochPerDay, 10)

	for _, tc := range []struct {
		name      string
		headers   map[string]string
		durations *epochDurations
		err       bool
		expected  map[string]string
	}{
		{
			name:     "valid epoch",
			headers:  map[string]string{objectv2.SysAttributeExpEpoch: epoch},
			expected: map[string]string{objectv2.SysAttributeExpEpoch: epoch},
		},
		{
			name: "valid epoch, valid duration",
			headers: map[string]string{
				objectv2.SysAttributeExpEpoch: epoch,
				ExpirationDurationAttr:        duration,
			},
			durations: defaultDurations,
			expected:  map[string]string{objectv2.SysAttributeExpEpoch: epoch},
		},
		{
			name: "valid epoch, valid rfc3339",
			headers: map[string]string{
				objectv2.SysAttributeExpEpoch: epoch,
				ExpirationRFC3339Attr:         tomorrow.Format(time.RFC3339),
			},
			durations: defaultDurations,
			expected:  map[string]string{objectv2.SysAttributeExpEpoch: epoch},
		},
		{
			name: "valid epoch, valid timestamp sec",
			headers: map[string]string{
				objectv2.SysAttributeExpEpoch: epoch,
				ExpirationTimestampAttr:       timestampSec,
			},
			durations: defaultDurations,
			expected:  map[string]string{objectv2.SysAttributeExpEpoch: epoch},
		},
		{
			name: "valid epoch, valid timestamp milli",
			headers: map[string]string{
				objectv2.SysAttributeExpEpoch: epoch,
				ExpirationTimestampAttr:       timestampMilli,
			},
			durations: defaultDurations,
			expected:  map[string]string{objectv2.SysAttributeExpEpoch: epoch},
		},
		{
			name: "valid epoch, valid timestamp nano",
			headers: map[string]string{
				objectv2.SysAttributeExpEpoch: epoch,
				ExpirationTimestampAttr:       timestampNano,
			},
			durations: defaultDurations,
			expected:  map[string]string{objectv2.SysAttributeExpEpoch: epoch},
		},
		{
			name:      "valid timestamp sec",
			headers:   map[string]string{ExpirationTimestampAttr: timestampSec},
			durations: defaultDurations,
			expected:  map[string]string{objectv2.SysAttributeExpEpoch: defaultExpEpoch},
		},
		{
			name:      "valid duration",
			headers:   map[string]string{ExpirationDurationAttr: duration},
			durations: defaultDurations,
			expected:  map[string]string{objectv2.SysAttributeExpEpoch: defaultExpEpoch},
		},
		{
			name:      "valid rfc3339",
			headers:   map[string]string{ExpirationRFC3339Attr: tomorrow.Format(time.RFC3339)},
			durations: defaultDurations,
			expected:  map[string]string{objectv2.SysAttributeExpEpoch: defaultExpEpoch},
		},
		{
			name:    "valid max uint 64",
			headers: map[string]string{ExpirationRFC3339Attr: tomorrow.Format(time.RFC3339)},
			durations: &epochDurations{
				currentEpoch:  math.MaxUint64 - 1,
				msPerBlock:    defaultDurations.msPerBlock,
				blockPerEpoch: defaultDurations.blockPerEpoch,
			},
			expected: map[string]string{objectv2.SysAttributeExpEpoch: strconv.FormatUint(uint64(math.MaxUint64), 10)},
		},
		{
			name:    "invalid timestamp sec",
			headers: map[string]string{ExpirationTimestampAttr: "abc"},
			err:     true,
		},
		{
			name:    "invalid timestamp sec zero",
			headers: map[string]string{ExpirationTimestampAttr: "0"},
			err:     true,
		},
		{
			name:    "invalid duration",
			headers: map[string]string{ExpirationDurationAttr: "1d"},
			err:     true,
		},
		{
			name:    "invalid duration negative",
			headers: map[string]string{ExpirationDurationAttr: "-5h"},
			err:     true,
		},
		{
			name:    "invalid rfc3339",
			headers: map[string]string{ExpirationRFC3339Attr: "abc"},
			err:     true,
		},
		{
			name:    "invalid rfc3339 zero",
			headers: map[string]string{ExpirationRFC3339Attr: time.RFC3339},
			err:     true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := prepareExpirationHeader(tc.headers, tc.durations)
			if tc.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, tc.headers)
			}
		})
	}
}
