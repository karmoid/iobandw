// Package principal de DashMon
// Cross Compilation on Raspberry
// > set GOOS=linux
// > set GOARCH=arm
// > go build dashmon
//
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/efarrer/iothrottler"
)

func copyFileContents(src, dst string, bwlimit uint64) (written int64, err error) {
	fmt.Println("copying", src, "to", dst)

	pool := iothrottler.NewIOThrottlerPool(iothrottler.BytesPerSecond * iothrottler.Bandwidth(bwlimit))
	defer pool.ReleasePool()

	file, err := os.Open(src)
	if err != nil {
		fmt.Println("Error:", err) // handle error
		return 0, err
	}
	defer file.Close()

	throttledFile, err := pool.AddReader(file)
	if err != nil {
		fmt.Println("Error:", err) // handle error
		// handle error
		return 0, err
	}

	out, err := os.Create(dst)
	if err != nil {
		fmt.Println("Error:", err) // handle error
		return 0, err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	bytesw, err := io.Copy(out, throttledFile)
	if err != nil {
		return 0, err
	}
	err = out.Sync()
	return bytesw, err
}

// Check if src is a wildcard expression
// if True, we must have a Path in dst
// Else dst could be Path or File
func genericCopy(src, dst string, bwlimit uint64) (written int64, myerr error) {
	wildcard := strings.Contains(src, "*") || strings.Contains(src, "?")
	// fmt.Printf("Dir from dst(%s) <=> (%s)\nwild:%v\nBase: (%s)\nisAbs:%v\n", dst, filepath.Dir(dst), wildcard, filepath.Base(dst), filepath.IsAbs(dst))
	// basepath := filepath.Base(dst)
	// if wildcard && !strings.HasSuffix(basepath, "\\") {
	// 	fmt.Printf("HasSuffix \\ is %v", strings.HasSuffix(basepath, "\\"))
	// 	return 0, fmt.Errorf("Wildcard copy need a target path. not %s", basepath)
	// }
	var wholesize int64
	if wildcard {
		pattern := filepath.Base(src)
		files, err := ioutil.ReadDir(filepath.Dir(src))
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			if res, err := filepath.Match(pattern, file.Name()); res {
				if err != nil {
					return 0, err
				}
				// fmt.Println(file.Name(), "must be processed!")
				bytesco, err := copyFileContents(filepath.Join(filepath.Dir(src), file.Name()), filepath.Join(dst, "\\", file.Name()), bwlimit)
				if err != nil {
					return 0, err
				}
				wholesize += bytesco
			}
		}
		return wholesize, nil
	}
	// not in wildcard
	// fmt.Println(src, "will be be processed!")
	bytesco, err := copyFileContents(src, dst, bwlimit)
	if err != nil {
		return 0, err
	}
	return bytesco, nil
}

// V 1.0 - Initial release - 2017 05 17
func main() {
	fmt.Println("iobandw - IO with BandWith control - C.m. 2017 - V1.0")
	required := []string{"src", "dst"}

	srcPtr := flag.String("src", "", "Source file specification")
	dstPtr := flag.String("dst", "", "Target path")
	bandwithLimitPtr := flag.String("limit", "32k", "Bytes per second limit (default 32KB/s)")
	verbosePtr := flag.Bool("verbose", false, "Verbose mode")

	flag.Parse()

	bandwithLimit, err := humanize.ParseBytes(*bandwithLimitPtr)
	if err != nil {
		fmt.Println("Limit value - Error:", err) // handle error
		return
	}

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			// or possibly use `log.Fatalf` instead of:
			fmt.Fprintf(os.Stderr, "missing required -%s argument/flag\n", req)
			os.Exit(2) // the same exit code flag.Parse uses
		}
	}

	if *verbosePtr {
		fmt.Println("limit is", humanize.Bytes(uint64(bandwithLimit)), "by second")
		fmt.Printf("approx. %sit/s.\n  ", strings.ToLower(humanize.Bytes(uint64(bandwithLimit*9))))
	}

	copied, err := genericCopy(*srcPtr, *dstPtr, bandwithLimit)
	if err != nil {
		fmt.Println("Error:", err) // handle error
		os.Exit(1)
	}
	fmt.Printf("%d (%s) bytes copied.\n", copied, humanize.Bytes(uint64(copied)))
	os.Exit(0)
}
