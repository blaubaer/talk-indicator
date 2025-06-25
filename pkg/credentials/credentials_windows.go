//go:build windows

package credentials

import (
	"fmt"
	"github.com/danieljoos/wincred"
	log "github.com/echocat/slf4g"
	"golang.org/x/sys/windows"
)

func (this *Credentials) ReadFromStore() (supported bool, err error) {
	c, err := wincred.GetGenericCredential(appName)
	if err == windows.ERROR_NOT_FOUND {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("cannot retrieve crendtials from Windows Credentials store: %w", err)
	}
	var buf Credentials
	if err := buf.UnmarshalBinary(c.CredentialBlob); err != nil {
		log.WithError(err).
			Error("Cannot unmarshal credentials from Windows Credentials storage. Assume it was empty.")
	}

	*this = buf
	return true, nil
}

func (this *Credentials) WriteToStore() (supported bool, err error) {
	b, err := this.MarshalBinary()
	if err != nil {
		return false, fmt.Errorf("cannot marshal crendtials to JSON: %w", err)
	}

	cred := wincred.NewGenericCredential(appName)
	cred.CredentialBlob = b
	if err := cred.Write(); err != nil {
		return false, fmt.Errorf("cannot store crendtials to Windows Credentials store: %w", err)
	}

	return true, nil
}
