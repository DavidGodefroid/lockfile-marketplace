package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

const (
	pw    = " /O (sample) /U (sample) /OE (sample) /UE (sample) /Perms (sample) /P -4"
	aesv3 = " /CF << /StdCF << /CFM /AESV3 /Length 32 >> >>"
	both  = " /StmF /StdCF /StrF /StdCF >>"
	r6    = "<< /Filter /Standard /V 5 /R 6 /Length 256" + pw + aesv3 + both
	r7    = "<< /Filter /Standard /V 5 /R 7 /Length 256" + pw + aesv3 + both
	r5    = "<< /Filter /Standard /V 5 /R 5 /Length 256" + pw + aesv3 + both
	// AESV3 for streams, strings left in clear.
	identityStr = "<< /Filter /Standard /V 5 /R 6" + pw + aesv3 + " /StmF /StdCF /StrF /Identity >>"
	// AESV3 for streams, StrF absent (defaults to Identity).
	noStrF = "<< /Filter /Standard /V 5 /R 6" + pw + aesv3 + " /StmF /StdCF >>"
	// AESV3 for streams, AES-128 for strings.
	mixed = "<< /Filter /Standard /V 5 /R 6" + pw +
		" /CF << /StdCF << /CFM /AESV3 >> /OldCF << /CFM /AESV2 >> >> /StmF /StdCF /StrF /OldCF >>"
	// The crypt filter dictionary is an indirect object.
	indirectCF = "<< /Filter /Standard /V 5 /R 6" + pw + " /CF << /StdCF {{obj 1}} >>" + both
	cfObj      = "<< /CFM /AESV3 /Length 32 >>"
)

func writeFixtures(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	stream := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, XRefStream: true})
	user, owner := "sample-user", "sample-owner"
	for name, b := range map[string][]byte{
		"plain.pdf":   plain,
		"rc4-128.pdf": pdfxtest.EncryptCipher(t, plain, user, owner, false, 128),
		"aes-128.pdf": pdfxtest.EncryptCipher(t, plain, user, owner, true, 128),
		// pdfcpu writes AES-256 as /V 5 /R 5.
		"aes-256-r5-pdfcpu.pdf": pdfxtest.EncryptCipher(t, plain, user, owner, true, 256),
		"aes-256-r5.pdf":        pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+r5),
		"aes-256-r6.pdf":        pdfxtest.AppendUpdate(plain, []string{r6}, "/Encrypt {{obj 0}}"),
		"aes-256-r7.pdf":        pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+r7),
		"r6-xref-stream.pdf":    pdfxtest.AppendUpdate(stream, []string{r6}, "/Encrypt {{obj 0}}"),
		"r6-indirect-cf.pdf":    pdfxtest.AppendUpdate(plain, []string{indirectCF, cfObj}, "/Encrypt {{obj 0}}"),
		"r6-identity-str.pdf":   pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+identityStr),
		"r6-no-strf.pdf":        pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+noStrF),
		"r6-mixed-filters.pdf":  pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+mixed),
	} {
		pdfxtest.Raw(t, "testdata", name, b)
	}
}
