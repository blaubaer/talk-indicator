//go:build !windows

package credentials

func (this *Credentials) ReadFromStore() (supported bool, err error) {
	return false, nil
}

func (this *Credentials) WriteToStore() (supported bool, err error) {
	return false, nil
}
