package main

// Takes a uniref file as input and returns a modified file in which
// the functions, taxa, and PID elements are extended to include all
// elements in the same cluster. [TODO: clearer description]

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type empty_t struct{}

var (
	// A structured version of the cluster information, produced
	// by the prep_custinfo script.
	cluster_info_file string

	// The raw UNIREF data
	uniref_file string

	// The output path for the processed UNIREF data
	outfile string

	// If trunacte is positive, only this many lines are read from the uniref file
	uniref_truncate int = 0

	// If positive, only this many lines are read from the clustinfo file
	clustinfo_truncate int = 0

	empty     struct{}
	clustinfo map[string]*clustrec
)

type clustrecid struct {
	Id  string
	Pid []string
	Tax []string
	Fnc []string
}

type clustrec struct {
	Pid []string
	Tax []string
	Fnc []string
}

func extend_map(m map[string]empty_t, q []string) {
	for _, v := range q {
		m[v] = empty
	}
}

func read_clustinfo() {

	clustinfo = make(map[string]*clustrec)

	fmt.Printf("Reading cluster information...")
	defer func() { fmt.Printf(" done\n") }()

	fid, err := os.Open(cluster_info_file)
	if err != nil {
		panic(err)
	}
	defer fid.Close()
	rdr, err := gzip.NewReader(fid)
	if err != nil {
		panic(err)
	}
	defer rdr.Close()

	// Get the file size so we can write progress reports
	finfo, err := fid.Stat()
	if err != nil {
		panic(err)
	}
	fsize := finfo.Size()

	dec := json.NewDecoder(rdr)

	lines_read := 0
	for {
		var idrec clustrecid
		err = dec.Decode(&idrec)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		id := idrec.Id
		rec := &clustrec{Pid: idrec.Pid, Fnc: idrec.Fnc, Tax: idrec.Tax}
		clustinfo[id] = rec

		lines_read++
		if (clustinfo_truncate > 0) && (lines_read > clustinfo_truncate) {
			return
		}

		// Progress report
		if lines_read%1000000 == 0 {
			pos, err := fid.Seek(0, 1)
			if err != nil {
				panic(err)
			}
			pread := 100 * float64(pos) / float64(fsize)
			fmt.Printf(" %.1f%% ", pread)
		}
	}
}

func main() {

	// Read flags
	flag.StringVar(&cluster_info_file, "cluster", "", "restructured cluster information input file path")
	flag.StringVar(&uniref_file, "uniref", "", "raw uniref input file path")
	flag.StringVar(&outfile, "output", "", "output file path for revised uniref file")
	flag.Parse()

	read_clustinfo()

	// Open the uniref file
	fid, err := os.Open(uniref_file)
	if err != nil {
		panic(err)
	}
	defer fid.Close()
	rdr, err := gzip.NewReader(fid)
	if err != nil {
		panic(err)
	}
	defer rdr.Close()
	scanner := bufio.NewScanner(rdr)
	var tks int = 1e7
	scanner.Buffer(make([]byte, tks), tks)
	scanner.Scan()

	// Create an output file
	gid, err := os.Create(outfile)
	if err != nil {
		panic(err)
	}
	defer gid.Close()
	out := gzip.NewWriter(gid)
	defer out.Close()

	// Get the file size so we can write progress reports
	finfo, err := fid.Stat()
	if err != nil {
		panic(err)
	}
	fsize := finfo.Size()

	// Create a header
	out.Write([]byte("Cluster ID\tPIDS\tFunc\tTax\n"))

	// Read through the uniref file one line at a time
	lines_read := 0
	fmt.Printf("Processing uniref... ")
	for scanner.Scan() {

		line := scanner.Text()
		lines_read++

		if (uniref_truncate > 0) && (lines_read > uniref_truncate) {
			break
		}

		fields := strings.Split(line, "\t")

		// Progress report
		if lines_read%1000000 == 0 {
			pos, err := fid.Seek(0, 1)
			if err != nil {
				panic(err)
			}
			pread := 100 * float64(pos) / float64(fsize)
			fmt.Printf(" %.1f%% ", pread)
		}

		all_pid := make(map[string]empty_t)
		all_tax := make(map[string]empty_t)
		all_fnc := make(map[string]empty_t)

		tax := fields[9]
		tax = strings.Replace(tax, " ", "", -1)
		taxlist := strings.Split(tax, ";")
		for _, v := range taxlist {
			all_tax[v] = empty
		}

		ac := fields[4]
		ac = strings.Replace(ac, " ", "", -1)
		aclist := strings.Split(ac, ";")

		for _, acid := range aclist {
			m, ok := clustinfo[acid]
			if !ok {
				// no cluster information
				continue
			}

			extend_map(all_pid, m.Pid)
			extend_map(all_tax, m.Tax)
			extend_map(all_fnc, m.Fnc)
		}

		// Cluster ids
		out.Write([]byte(fields[0]))
		out.Write([]byte("\t"))

		// PIDS
		var s []string
		for k, _ := range all_pid {
			s = append(s, k)
		}
		sort.StringSlice(s).Sort()
		out.Write([]byte(strings.Join(s, ";")))
		out.Write([]byte("\t"))

		// Functions
		s = s[:0]
		for k, _ := range all_fnc {
			s = append(s, k)
		}
		sort.StringSlice(s).Sort()
		out.Write([]byte(strings.Join(s, "@")))
		out.Write([]byte("\t"))

		// Taxa
		s = s[:0]
		for k, _ := range all_tax {
			s = append(s, k)
		}
		sort.StringSlice(s).Sort()
		out.Write([]byte(strings.Join(s, ";")))
		out.Write([]byte("\n"))
	}

	err = scanner.Err()
	if err != nil {
		panic(err)
	}

	fmt.Printf("  done\n\n")
}
