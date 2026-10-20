package pdfxtest

import (
	"bytes"
	"context"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Permission sets for EncryptPerms (the /P entry; PDF 32000-1 table 22).
const (
	PermNone  = int(model.PermissionsNone)
	PermPrint = int(model.PermissionsPrint)
	PermAll   = int(model.PermissionsAll)
	// PermCopy allows copying and extraction only.
	PermCopy = int(model.PermissionsNone | model.PermissionExtract | model.PermissionExtractRev3)
)

// EncryptPerms is Encrypt with the user access permissions perms.
func EncryptPerms(t *testing.T, pdf []byte, user, owner string, perms int) []byte {
	t.Helper()
	conf := model.NewStatelessConfiguration()
	conf.UserPW, conf.OwnerPW = user, owner
	conf.EncryptUsingAES, conf.EncryptKeyLength = true, 256
	conf.Permissions = model.PermissionFlags(perms)
	var out bytes.Buffer
	if err := api.Encrypt(context.Background(), bytes.NewReader(pdf), &out, conf); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return out.Bytes()
}

// EncryptCipher encrypts pdf with pdfcpu using RC4 (aes false; bits 40 or
// 128; pdfcpu writes 128 as /V 4 with /CFM /V2) or AES (bits 128, or 256
// written as /V 5 /R 5).
func EncryptCipher(t *testing.T, pdf []byte, user, owner string, aes bool, bits int) []byte {
	t.Helper()
	conf := model.NewStatelessConfiguration()
	conf.UserPW, conf.OwnerPW = user, owner
	conf.EncryptUsingAES, conf.EncryptKeyLength = aes, bits
	var out bytes.Buffer
	if err := api.Encrypt(context.Background(), bytes.NewReader(pdf), &out, conf); err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return out.Bytes()
}
