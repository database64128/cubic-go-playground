// Domain set converter takes a plaintext domain set file in v2fly/dlc format
// and converts it to database64128/shadowsocks-go format.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	in  = flag.String("in", "", "Input file path")
	out = flag.String("out", "", "Output file path")
	tag = flag.String("tag", "", "Select lines with the specified tag. If empty, select all lines.")
)

func main() {
	flag.Parse()

	if *in == "" {
		fmt.Println("Missing input file path: -in <file path>.")
		flag.Usage()
		return
	}

	if *out == "" {
		*out = "ss-go-" + *in
	}

	fin, err := os.Open(*in)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fin.Close()

	scanner := bufio.NewScanner(fin)

	fout, err := os.Create(*out)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fout.Close()

	writer := bufio.NewWriter(fout)
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || strings.IndexByte(line, '#') == 0 {
			continue
		}

		end := strings.IndexByte(line, '@')
		if end == 0 {
			fmt.Println("Invalid line:", line)
			return
		}

		if *tag == "" { // select all lines
			if end == -1 {
				end = len(line)
			} else {
				end--
			}
		} else { // select matched tag
			if end == -1 || line[end+1:] != *tag { // no tag or different tag
				continue
			} else {
				end--
			}
		}

		switch {
		case strings.HasPrefix(line, "full:"):
			writer.WriteString("domain:")
			writer.WriteString(line[5:end])
			writer.WriteByte('\n')
		case strings.HasPrefix(line, "domain:"):
			writer.WriteString("suffix:")
			writer.WriteString(line[7:end])
			writer.WriteByte('\n')
		case strings.HasPrefix(line, "keyword:"):
			writer.WriteString("keyword:")
			writer.WriteString(line[8:end])
			writer.WriteByte('\n')
		case strings.HasPrefix(line, "regexp:"):
			writer.WriteString("regexp:")
			writer.WriteString(line[7:end])
			writer.WriteByte('\n')
		default:
			fmt.Println("Invalid line:", line)
			return
		}
	}
}
