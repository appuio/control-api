package countries

import (
	"os"

	"sigs.k8s.io/yaml"
)

type Country struct {
	ID   int    `yaml:"id"`
	Code string `yaml:"code"`
	Name string `yaml:"name"`
}

func LoadCountryIDs(path string) (map[string]int, error) {
	r, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var countries []Country
	if err := yaml.UnmarshalStrict(r, &countries); err != nil {
		return nil, err
	}

	countryIDs := make(map[string]int, len(countries))
	for _, c := range countries {
		countryIDs[c.Name] = c.ID
	}

	return countryIDs, nil
}
