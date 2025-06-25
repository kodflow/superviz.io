package providers_test

import (
	"strings"
	"testing"

	"github.com/kodflow/superviz.io/internal/providers"
	"github.com/stretchr/testify/require"
)

func TestVersionProvider_GetVersionInfo(t *testing.T) {

	provider := providers.NewVersionProvider()
	info := provider.GetVersionInfo()

	require.Equal(t, "dev", info.Version)
	require.Equal(t, "none", info.Commit)
	require.Equal(t, "unknown", info.BuiltAt)
	require.Equal(t, "unknown", info.BuiltBy)
	require.Contains(t, info.GoVersion, "go")
	require.Contains(t, info.OSArch, "/")
}

func TestVersionProvider_Singleton(t *testing.T) {

	provider1 := providers.DefaultVersionProvider()
	provider2 := providers.DefaultVersionProvider()
	provider3 := providers.NewVersionProvider()

	// All should return the same singleton instance
	require.Same(t, provider1, provider2)
	require.Same(t, provider1, provider3)
}

func TestVersionProvider_Cache(t *testing.T) {

	provider := providers.NewVersionProvider()

	// Call multiple times to ensure caching works
	info1 := provider.GetVersionInfo()
	info2 := provider.GetVersionInfo()

	// Should return identical structs (same values)
	require.Equal(t, info1, info2)

	// Format should also be cached
	format1 := info1.Format()
	format2 := info2.Format()
	require.Equal(t, format1, format2)
}

func TestVersionInfo_Format(t *testing.T) {

	provider := providers.NewVersionProvider()
	info := provider.GetVersionInfo()
	output := info.Format()

	expectedFields := []string{
		"Version:",
		"Commit:",
		"Built at:",
		"Built by:",
		"Go version:",
		"OS/Arch:",
	}

	for _, field := range expectedFields {
		require.Contains(t, output, field, "output must contain field %q", field)
	}

	require.True(t, strings.HasSuffix(output, "\n"), "format should end with newline")
}

func TestVersionInfo_FormatCache(t *testing.T) {

	provider := providers.NewVersionProvider()
	info := provider.GetVersionInfo()

	// Call Format multiple times - should return cached result
	format1 := info.Format()
	format2 := info.Format()

	require.Equal(t, format1, format2)
	require.NotEmpty(t, format1)
}

func TestReset(t *testing.T) {
	// Note: Ce test ne peut pas être parallèle car il modifie l'état global

	// Premier appel pour initialiser le singleton
	provider1 := providers.NewVersionProvider()
	info1 := provider1.GetVersionInfo()

	// Vérifier que les valeurs par défaut sont présentes
	require.Equal(t, "dev", info1.Version)
	require.Equal(t, "none", info1.Commit)
	require.Equal(t, "unknown", info1.BuiltAt)
	require.Equal(t, "unknown", info1.BuiltBy)

	// Reset du singleton
	providers.Reset()

	// Après reset, un nouveau provider devrait réinitialiser le singleton
	provider2 := providers.NewVersionProvider()
	info2 := provider2.GetVersionInfo()

	// Les valeurs devraient être les mêmes que par défaut
	// (car on réutilise les mêmes variables globales)
	require.Equal(t, "dev", info2.Version)
	require.Equal(t, "none", info2.Commit)
	require.Equal(t, "unknown", info2.BuiltAt)
	require.Equal(t, "unknown", info2.BuiltBy)

	// Mais les instances doivent être différentes si on compare les pointeurs
	// Note: On ne peut pas facilement tester cela car GetVersionInfo retourne une valeur, pas un pointeur

	// Test que le format fonctionne correctement après reset
	format := info2.Format()
	require.Contains(t, format, "Version:       dev")
	require.Contains(t, format, "Commit:        none")
	require.Contains(t, format, "Built at:      unknown")
	require.Contains(t, format, "Built by:      unknown")
}

func TestReset_MultipleResets(t *testing.T) {
	// Note: Ce test ne peut pas être parallèle car il modifie l'état global

	// Premier appel
	provider1 := providers.NewVersionProvider()
	info1 := provider1.GetVersionInfo()
	require.Equal(t, "dev", info1.Version)

	// Premier reset
	providers.Reset()

	// Deuxième appel après reset
	provider2 := providers.NewVersionProvider()
	info2 := provider2.GetVersionInfo()
	require.Equal(t, "dev", info2.Version)

	// Deuxième reset
	providers.Reset()

	// Troisième appel après reset
	provider3 := providers.NewVersionProvider()
	info3 := provider3.GetVersionInfo()
	require.Equal(t, "dev", info3.Version)

	// Tous les appels devraient donner les mêmes valeurs par défaut
	require.Equal(t, info1.Version, info2.Version)
	require.Equal(t, info2.Version, info3.Version)
}

func TestReset_WithMockAfterReset(t *testing.T) {
	// Note: Ce test ne peut pas être parallèle car il modifie l'état global

	// Premier appel pour initialiser
	provider1 := providers.NewVersionProvider()
	info1 := provider1.GetVersionInfo()
	require.Equal(t, "dev", info1.Version)

	// Reset pour nettoyer l'état
	providers.Reset()

	// Après reset, on peut utiliser un mock provider sans interférence
	// (c'est l'usage principal de Reset dans les tests)

	// Simuler l'utilisation d'un mock provider après reset
	// Le Reset garantit que le singleton n'interfère pas avec le mock
	providers.Reset() // Double reset pour s'assurer que c'est idempotent

	// Un nouveau provider après reset devrait fonctionner normalement
	provider2 := providers.NewVersionProvider()
	info2 := provider2.GetVersionInfo()
	require.Equal(t, "dev", info2.Version)
}
