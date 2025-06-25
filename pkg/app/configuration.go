package app

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
)

func (this *Configuration) loadFrom(r io.Reader) error {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	return dec.Decode(this)
}

func (this *Configuration) loadFromFile(fn string, ignoreNotFound bool) error {
	f, err := os.Open(fn)
	if os.IsNotExist(err) && ignoreNotFound {
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot open configuration file %q: %w", fn, err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := this.loadFrom(f); err != nil {
		return fmt.Errorf("cannot load configuration file %q: %w", fn, err)
	}

	return nil
}

func (this *Configuration) loadDefault(ignoreNotFound bool) error {
	return this.loadFromFile(defaultConfigurationFile(), ignoreNotFound)
}

func (this *Configuration) saveTo(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(this)
}

func (this *Configuration) saveToFile(fn string) error {
	_ = os.MkdirAll(filepath.Dir(fn), 0700)

	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot open configuration file %q: %w", fn, err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := this.saveTo(f); err != nil {
		return fmt.Errorf("cannot write file %q: %w", fn, err)
	}

	return nil
}

func (this *Configuration) saveDefault() error {
	return this.saveToFile(defaultConfigurationFile())
}
