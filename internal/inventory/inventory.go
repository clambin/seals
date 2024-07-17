package inventory

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type Inventory struct {
	SecretsDir     string   `yaml:"secrets_dir"`
	DestinationDir string   `yaml:"destination_dir"`
	Secrets        []Secret `yaml:"secrets"`
}

type Secret struct {
	Source      string `yaml:"source"`
	Destination string `yaml:"destination"`
	Namespace   string `yaml:"namespace"`
}

func Read(r io.Reader) (Inventory, error) {
	var inv Inventory
	if err := yaml.NewDecoder(r).Decode(&inv); err != nil {
		return inv, err
	}
	return inv, nil
}

func ReadFromFile(path string) (Inventory, error) {
	var inv Inventory
	f, err := os.Open(path)
	if err == nil {
		inv, err = Read(f)
	}
	return inv, err
}

func (i *Inventory) Write(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	defer func() { _ = enc.Close() }()
	enc.SetIndent(2)
	return enc.Encode(i)
}

func (i *Inventory) WriteToFile(filename string) error {
	f, err := os.Create(filename)
	if err == nil {
		err = i.Write(f)
		_ = f.Close()
	}
	return err
}

func (i *Inventory) Add(secret Secret) {
	i.Delete(secret.Source)
	i.Secrets = append(i.Secrets, secret)
}

func (i *Inventory) Delete(source string) bool {
	if len(i.Secrets) == 0 {
		return false
	}
	oldSecrets := i.Secrets
	i.Secrets = make([]Secret, 0, len(oldSecrets)-1)
	for _, secret := range oldSecrets {
		if secret.Source != source {
			i.Secrets = append(i.Secrets, secret)
		}
	}
	return len(i.Secrets) < len(oldSecrets)
}
