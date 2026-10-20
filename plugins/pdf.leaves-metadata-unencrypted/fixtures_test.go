package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// encrypt returns a synthetic standard-handler /Encrypt dictionary holding
// entry.
func encrypt(entry string) string {
	return "<< /Filter /Standard /V 4 /R 4 /Length 128 /P -4 /O (sample) /U (sample)" +
		" /CF << /StdCF << /CFM /AESV2 /Length 16 >> >> /StmF /StdCF /StrF /StdCF " + entry + " >>"
}

func writeFixtures(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	stream := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, XRefStream: true})
	for name, b := range map[string][]byte{
		"plain.pdf":          plain,
		"pdfcpu-aes256.pdf":  pdfxtest.EncryptCipher(t, plain, "sample-user", "sample-owner", true, 256),
		"pdfcpu-rc4-128.pdf": pdfxtest.EncryptCipher(t, plain, "", "sample-owner", false, 128),
		"key-absent.pdf":     pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+encrypt("")),
		"metadata-true.pdf":  pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+encrypt("/EncryptMetadata true")),
		"metadata-false.pdf": pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+encrypt("/EncryptMetadata false")),
		"metadata-false-indirect.pdf": pdfxtest.AppendUpdate(plain,
			[]string{encrypt("/EncryptMetadata {{obj 1}}"), "false"}, "/Encrypt {{obj 0}}"),
		"metadata-false-xref-stream.pdf": pdfxtest.AppendUpdate(stream,
			[]string{encrypt("/EncryptMetadata false")}, "/Encrypt {{obj 0}}"),
	} {
		pdfxtest.Raw(t, "testdata", name, b)
	}
}
