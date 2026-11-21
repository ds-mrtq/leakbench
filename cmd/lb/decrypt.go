package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ds-mrtq/leakbench/internal/sealed"
)

func runDecrypt(args []string, out io.Writer) error {
	f := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	keyPath := f.String("key", "", "runner private key")
	dest := f.String("out", "", "new plaintext archive")
	var paths []string
	for len(args) > 0 {
		if args[0] == "--" {
			paths = append(paths, args[1:]...)
			break
		}
		if err := f.Parse(args); err != nil {
			return err
		}
		remaining := f.Args()
		if len(remaining) == 0 {
			break
		}
		paths = append(paths, remaining[0])
		args = remaining[1:]
	}
	if *keyPath == "" || *dest == "" || len(paths) != 1 {
		return fmt.Errorf("usage: lb decrypt --key PRIVATE.pem INPUT.bin --out ARCHIVE.tar.gz")
	}
	key, err := os.ReadFile(*keyPath)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		return err
	}
	plain, err := sealed.Decrypt(key, data)
	if err != nil {
		return err
	}
	fout, err := os.OpenFile(*dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := fout.Write(plain)
	closeErr := fout.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(*dest)
		if writeErr != nil {
			return writeErr
		}
		return closeErr
	}
	_, err = fmt.Fprintln(out, *dest)
	return err
}
