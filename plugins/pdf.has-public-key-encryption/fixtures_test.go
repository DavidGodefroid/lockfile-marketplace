package main

import (
	"testing"

	"github.com/DavidGodefroid/lockfile-marketplace/lib/pdfx/pdfxtest"
)

// pubSec is a synthetic public-key /Encrypt dictionary (the recipient is a
// placeholder, nothing is decrypted).
const pubSec = "<< /Filter /Adobe.PubSec /SubFilter /adbe.pkcs7.s5 /V 4 /R 4" +
	" /CF << /DefaultCryptFilter << /CFM /AESV2 /Recipients [<73616d706c65>] >> >>" +
	" /StmF /DefaultCryptFilter /StrF /DefaultCryptFilter >>"

// otherHandler names a third-party security handler.
const otherHandler = "<< /Filter /SampleHandler /V 2 /Length 128 >>"

// standardR6 is a /Standard handler with values pdfcpu does not write.
const standardR6 = "<< /Filter /Standard /V 5 /R 6 /Length 256 /P -4 /O (sample) /U (sample)" +
	" /CF << /StdCF << /CFM /AESV3 /Length 32 >> >> /StmF /StdCF /StrF /StdCF >>"

func writeFixtures(t *testing.T) {
	plain := pdfxtest.Write(pdfxtest.Pages("text"))
	stream := pdfxtest.Write(pdfxtest.Doc{Pages: []pdfxtest.Page{{}}, XRefStream: true})
	pdfxtest.Raw(t, "testdata", "plain.pdf", plain)
	pdfxtest.Raw(t, "testdata", "plain-xref-stream.pdf", stream)
	pdfxtest.Raw(t, "testdata", "standard-rc4-40.pdf", pdfxtest.EncryptCipher(t, plain, "sample-user", "sample-owner", false, 40))
	pdfxtest.Raw(t, "testdata", "standard-aes-128.pdf", pdfxtest.EncryptCipher(t, plain, "", "sample-owner", true, 128))
	pdfxtest.Raw(t, "testdata", "standard-aes-256.pdf", pdfxtest.EncryptCipher(t, plain, "sample-user", "sample-owner", true, 256))
	pdfxtest.Raw(t, "testdata", "standard-r6.pdf", pdfxtest.AppendUpdate(plain, []string{standardR6}, "/Encrypt {{obj 0}}"))
	pdfxtest.Raw(t, "testdata", "pubsec.pdf", pdfxtest.AppendUpdate(plain, []string{pubSec}, "/Encrypt {{obj 0}}"))
	pdfxtest.Raw(t, "testdata", "pubsec-direct.pdf", pdfxtest.AppendUpdate(plain, nil, "/Encrypt "+pubSec))
	pdfxtest.Raw(t, "testdata", "pubsec-xref-stream.pdf", pdfxtest.AppendUpdate(stream, []string{pubSec}, "/Encrypt {{obj 0}}"))
	pdfxtest.Raw(t, "testdata", "other-handler.pdf", pdfxtest.AppendUpdate(plain, []string{otherHandler}, "/Encrypt {{obj 0}}"))
}
