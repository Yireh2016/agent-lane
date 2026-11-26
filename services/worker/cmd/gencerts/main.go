// gencerts writes the certificates the local test sites serve:
// a local CA, a valid leaf for goodsite, and an expired self-signed leaf for badsite.
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/agent-lane/worker/internal/certgen"
)

func main() {
	out := flag.String("out", "/certs", "output directory")
	flag.Parse()

	now := time.Now()
	ca, caKey, err := certgen.NewCA("agent-lane local CA", now.Add(-time.Hour), now.AddDate(1, 0, 0))
	must(err)
	good, err := certgen.Leaf([]string{"goodsite"}, now.Add(-time.Hour), now.AddDate(0, 6, 0), ca.Cert, caKey)
	must(err)
	bad, err := certgen.Leaf([]string{"badsite"}, now.AddDate(-2, 0, 0), now.AddDate(-1, 0, 0), nil, nil)
	must(err)

	must(os.MkdirAll(*out, 0o755))
	write(*out, "ca.pem", ca.CertPEM)
	write(*out, "goodsite.pem", good.CertPEM)
	write(*out, "goodsite.key", good.KeyPEM)
	write(*out, "badsite.pem", bad.CertPEM)
	write(*out, "badsite.key", bad.KeyPEM)
	log.Printf("wrote certificates to %s", *out)
}

func write(dir, name string, data []byte) {
	must(os.WriteFile(filepath.Join(dir, name), data, 0o644))
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
