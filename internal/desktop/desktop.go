// Package desktop is the application's boundary to native desktop services.
// These methods require a running MyGo application; blocking calls belong in
// a worker. Native-only concerns never leak into the files/config packages.
package desktop

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"github.com/ntnyq/rivu/locales"
)

var ErrCanceled = errors.New("dialog canceled")

type Service struct{ Parent *mygo.Window }

func (s Service) PickFolder() (string, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: s.Parent, Title: locales.FilesChooseDownloadFolder(), Directory: true})
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", ErrCanceled
	}
	return paths[0], nil
}

func (Service) Copy(text string) error { return mygo.Clipboard.Write(transfer.TextData(text)) }
func (Service) Paste() (string, error) {
	data, err := mygo.Clipboard.ReadFormat(transfer.Text)
	return string(data), err
}

// ValidateURL deliberately accepts browser destinations only, never file or
// executable custom schemes. Extend this policy explicitly for your product.
func ValidateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return fmt.Errorf("expected an http(s) URL without credentials")
	}
	return nil
}

func (Service) OpenURL(raw string) error {
	if err := ValidateURL(raw); err != nil {
		return err
	}
	return mygo.Shell.OpenExternal(raw)
}

func (Service) Reveal(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	mygo.Shell.ShowItemInFolder(path)
	return nil
}
