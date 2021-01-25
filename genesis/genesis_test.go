package genesis

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"terorie.dev/nimiq/accounts"
	"terorie.dev/nimiq/tree"
)

func TestGenesis(t *testing.T) {
	_, callerFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(callerFile)

	t.Run("Mainnet", func(t *testing.T) {
		inf, err := ReadInfo(filepath.Join(dir, "files/mainnet.toml"))
		require.NoError(t, err)
		accs := accounts.Accounts{Tree: &tree.PMTree{Store: tree.NewMemStore()}}
		require.NoError(t, inf.InitAccounts(&accs))
	})

	t.Run("Testnet", func(t *testing.T) {
		inf, err := ReadInfo(filepath.Join(dir, "files/testnet.toml"))
		require.NoError(t, err)
		accs := accounts.Accounts{Tree: &tree.PMTree{Store: tree.NewMemStore()}}
		require.NoError(t, inf.InitAccounts(&accs))
	})
}
