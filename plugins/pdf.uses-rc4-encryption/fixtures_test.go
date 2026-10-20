package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	pw = " /O (sample) /U (sample) /P -4"
	v2 = "<< /Filter /Standard /V 2 /R 3 /Length 128" + pw + " >>"
	// AES for streams, RC4 for strings.
	mixed = "<< /Filter /Standard /V 4 /R 4" + pw +
		" /CF << /AesCF << /CFM /AESV2 >> /Rc4CF << /CFM /V2 >> >> /StmF /AesCF /StrF /Rc4CF >>"
	pubSecAES = "<< /Filter /Adobe.PubSec /SubFilter /adbe.pkcs7.s5 /V 4 /R 4" +
		" /CF << /DefaultCryptFilter << /CFM /AESV2 /Recipients [(sample)] >> >>" +
		" /StmF /DefaultCryptFilter /StrF /DefaultCryptFilter >>"
	identity = "<< /Filter /Standard /V 4 /R 4" + pw + " /StmF /Identity /StrF /Identity >>"
	noV      = "<< /Filter /Standard /R 2" + pw + " >>"
)

func writeFixtures(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	stream := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, XRefStream: true})
	user, owner := "sample-user", "sample-owner"
	for name, b := range map[string][]byte{
		"plain.pdf":            plain,
		"rc4-40.pdf":           pdfxtest.EncryptCipher(t, plain, user, owner, false, 40),
		"rc4-128.pdf":          pdfxtest.EncryptCipher(t, plain, user, owner, false, 128),
		"rc4-owner-only.pdf":   pdfxtest.EncryptCipher(t, plain, "", owner, false, 128),
		"aes-128.pdf":          pdfxtest.EncryptCipher(t, plain, user, owner, true, 128),
		"aes-256.pdf":          pdfxtest.EncryptCipher(t, plain, user, owner, true, 256),
		"v2-xref-stream.pdf":   pdfxtest.AppendUpdate(stream, []string{v2}, "/Encrypt {{obj 0}}"),
		"mixed-filters.pdf":    pdfxtest.AppendUpdate(plain, []string{mixed}, "/Encrypt {{obj 0}}"),
		"pubsec-aes.pdf":       pdfxtest.AppendUpdate(plain, []string{pubSecAES}, "/Encrypt {{obj 0}}"),
		"identity-filters.pdf": pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+identity),
		"no-v.pdf":             pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+noV),
	} {
		pdfxtest.Raw(t, "testdata", name, b)
	}
}
