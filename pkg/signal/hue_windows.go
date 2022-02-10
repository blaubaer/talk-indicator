package signal

import (
	"fmt"
	"github.com/danieljoos/wincred"
	log "github.com/echocat/slf4g"
	"golang.org/x/sys/windows"
)

func (this *Hue) readCredentials() (HueCredentials, error) {
	c, err := wincred.GetGenericCredential(appName)
	if err == windows.ERROR_NOT_FOUND {
		return HueCredentials{}, nil
	}
	if err != nil {
		return HueCredentials{}, fmt.Errorf("cannot retrieve HUE crendtials from Windows Credentials store: %w", err)
	}
	var result HueCredentials
	if err := result.UnmarshalBinary(c.CredentialBlob); err != nil {
		log.WithError(err).
			Error("Cannot unmarshal credentials from Windows Credentials storage. Assume it was empty.")
	}

	return result, nil
}

func (this *Hue) storeCredentials(v HueCredentials) error {
	b, err := v.MarshalBinary()
	if err != nil {
		return fmt.Errorf("cannot marshal HUE crendtials to JSON: %w", err)
	}

	cred := wincred.NewGenericCredential(appName)
	cred.CredentialBlob = b
	if err := cred.Write(); err != nil {
		return fmt.Errorf("cannot store HUE crendtials to Windows Credentials store: %w", err)
	}

	return nil
}
