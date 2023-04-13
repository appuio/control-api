package countries

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadCountryIDs(t *testing.T) {
	const countriesYAML = `
- id: 1
  code: BE
  name: Belgium
- id: 5
  code: FR
  name: France
- id: 6
  code: false # Seen in the export of the Odoo database
  name: Kabott
`

	td := t.TempDir()
	p := td + "/countries.yaml"
	require.NoError(t, os.WriteFile(p, []byte(countriesYAML), 0644))

	countryIDs, err := LoadCountryIDs(p)
	require.NoError(t, err)
	require.Equal(t, map[string]int{
		"Belgium": 1,
		"France":  5,
		"Kabott":  6,
	}, countryIDs)
}
