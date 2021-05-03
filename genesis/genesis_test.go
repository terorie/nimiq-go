package genesis

import (
	"testing"

	"github.com/stretchr/testify/require"
	"terorie.dev/nimiq/accounts"
	"terorie.dev/nimiq/tree"
)

func TestOpenProfile(t *testing.T) {
	profiles := []string{ProfileMain, ProfileTest}
	for _, profile := range profiles {
		t.Run(profile, func(t *testing.T) {
			inf, err := OpenProfile(profile)
			require.NoError(t, err)
			accs := accounts.Accounts{Tree: &tree.PMTree{Store: tree.NewMemStore()}}
			require.NoError(t, inf.InitAccounts(&accs))
		})
	}
}
