package handlers

import (
	"testing"

	sessionv2 "github.com/TrueCloudLab/frostfs-api-go/v2/session"
	"github.com/stretchr/testify/require"
)

func TestCheckContainerName(t *testing.T) {
	name64 := "container-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	name256 := "container-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	for _, tc := range []struct {
		name  string
		valid bool
	}{
		{name: "container", valid: true},
		{name: "container-name", valid: true},
		{name: "container.name", valid: true},
		{name: "container2", valid: true},
		{name: "2container.name", valid: true},
		{name: "containerName", valid: false},
		{name: "-container", valid: false},
		{name: "container-", valid: false},
		{name: "container name", valid: false},
		{name: "c", valid: false},
		{name: name64 + ".name", valid: false},
		{name: name256, valid: false},
	} {
		err := checkNNSContainerName(tc.name)
		if tc.valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}

func TestPrepareSessionToken(t *testing.T) {
	st := &SessionToken{
		BearerToken: BearerToken{
			Token:     "ChASxCTiXwREjLAG7nkxjDHVEhsKGTVxfQ56a0uQeFmOO63mqykBS1HNpw1rxSgaBgjIAhjkASIhAnLj82Qmdlcg7JtoyhDjJ1OsRFjtmxdXbzrwVkwxWAdWMgQIAxAB",
			Signature: "2ebdc1f2fea2bba397d1be6f982a6fe1b2bc9f46a348b700108fe2eba4e6531a1bb585febf9a40a3fa2e085fca5e2a75ca57f61166117c6d3e04a95ef9a2d2196f52648546784853e17c0b7ba762eae1",
			Key:       "03bd9108c0b49f657e9eee50d1399022bd1e436118e5b7529a1b7cd606652f578f",
		},
		Verb: sessionv2.ContainerVerbSetEACL,
	}

	_, err := prepareSessionToken(st, true)
	require.NoError(t, err)
}
