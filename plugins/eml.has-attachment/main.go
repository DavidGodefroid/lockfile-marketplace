// Command eml.has-attachment answers: does the e-mail message carry an
// attachment?
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"strconv"
	"strings"

	"github.com/DavidGodefroid/lockfile-marketplace/sdk"
)

func main() {
	sdk.Main(func(f *os.File) (bool, error) {
		return hasAttachment(f)
	})
}

// maxDepth is the deepest multipart nesting walked; the message itself is
// depth 0 and each multipart level adds one.
const maxDepth = 32

var errTooDeep = fmt.Errorf("multipart nesting deeper than %d", maxDepth)

// hasAttachment parses r as an RFC 5322 message and walks its MIME tree. The
// whole tree is walked even after a match, so a broken multipart is always an
// error, never an answer.
//
// Plugins may not import net/mail (importcheck denies net/...), so the
// message is wrapped as the only part of a synthetic multipart body whose
// boundary does not occur in it: mime/multipart then parses the top-level
// header with the same net/textproto rules net/mail uses, and the part body
// is exactly the message body.
func hasAttachment(r io.Reader) (bool, error) {
	msg, err := io.ReadAll(r)
	if err != nil {
		return false, err
	}
	if len(msg) == 0 {
		return false, errors.New("empty message")
	}
	boundary := "lockfile-eml"
	for i := 0; bytes.Contains(msg, []byte(boundary)); i++ {
		boundary = "lockfile-eml-" + strconv.Itoa(i)
	}
	wrapped := io.MultiReader(
		strings.NewReader("--"+boundary+"\r\n"),
		bytes.NewReader(msg),
		strings.NewReader("\r\n--"+boundary+"--"),
	)
	p, err := multipart.NewReader(wrapped, boundary).NextRawPart()
	if err != nil {
		return false, fmt.Errorf("message header: %w", err)
	}
	return walk(p.Header, p, 0)
}

// walk reports whether the part with header h and body is an attachment or
// contains one.
func walk(h map[string][]string, body io.Reader, depth int) (bool, error) {
	if depth > maxDepth {
		return false, errTooDeep
	}
	mediaType, params := "text/plain", map[string]string(nil)
	if v := get(h, "Content-Type"); v != "" {
		var err error
		if mediaType, params, err = mime.ParseMediaType(v); err != nil {
			return false, fmt.Errorf("content-type: %w", err)
		}
	}
	found := false
	if v := get(h, "Content-Disposition"); v != "" {
		disp, dparams, err := mime.ParseMediaType(v)
		if err != nil {
			return false, fmt.Errorf("content-disposition: %w", err)
		}
		_, named := dparams["filename"]
		found = disp == "attachment" ||
			named && !strings.HasPrefix(mediaType, "text/") && !strings.HasPrefix(mediaType, "multipart/")
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return found, nil
	}
	boundary := params["boundary"]
	if boundary == "" {
		return false, errors.New("multipart without a boundary")
	}
	mr := multipart.NewReader(body, boundary)
	for {
		p, err := mr.NextRawPart()
		// Only a bare io.EOF marks the close delimiter: NextRawPart wraps
		// the io.EOF of a body that ends before it.
		if err == io.EOF {
			return found, nil
		}
		if err != nil {
			return false, err
		}
		sub, err := walk(p.Header, p, depth+1)
		if err != nil {
			return false, err
		}
		found = found || sub
	}
}

// get returns the first value of the canonical header key k; part headers
// are keyed canonically by mime/multipart.
func get(h map[string][]string, k string) string {
	if v := h[k]; len(v) > 0 {
		return v[0]
	}
	return ""
}
