package main

// Restructure the "cluster information" file into a struct for each
// unique cluster id containing the cluster id and three lists: PID's,
// Taxa, and Functions.
//
// The output file is a sequence of Gob-encoded structs as defined
// above.

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

var (
	// File path for input cluster information file
	cluster_info_file string

	// File path for output JSON file
	outfile string

	// If positive, process only this many lines of the
	// cluster_info_file
	truncate = -1

	// PID, taxa, and function data for each cluster
	clustinfo map[string]*clustrec
)

type clustrec struct {
	Id  string
	Pid []string
	Tax []string
	Fnc []string
}

func process_clustinfo() {

	// Set up to read the file
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
	scanner := bufio.NewScanner(rdr)

	// Get the file size so we can write progress reports
	finfo, err := fid.Stat()
	if err != nil {
		panic(err)
	}
	fsize := finfo.Size()

	// Scan through the file
	clustinfo = make(map[string]*clustrec)
	lines_read := 0
	fmt.Printf("Processing cluster information...")
	for scanner.Scan() {
		line := scanner.Text()
		lines_read++
		fields := strings.Split(line, "\t")
		id := fields[1]

		rec, ok := clustinfo[id]
		if !ok {
			// First time seeing this cluster
			rec = new(clustrec)
			rec.Id = id
			clustinfo[id] = rec
		}

		rec.Pid = append(rec.Pid, fields[0])
		rec.Tax = append(rec.Tax, fields[2])

		fncs := strings.Split(fields[3], "@")
		for _, f := range fncs {
			rec.Fnc = append(rec.Fnc, f)
		}

		if (truncate > 0) && (lines_read > truncate) {
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
	fmt.Printf(" done\n")
}

// Return the unique elements of an array of strings.
func unique(x []string) []string {

	sort.StringSlice(x).Sort()

	if len(x) == 1 {
		return x
	}

	i := 1
	for j := 1; j < len(x); j++ {
		if x[j] != x[i-1] {
			x[i] = x[j]
			i++
		}
	}

	return x[0:i]
}

func main() {

	// Read flags
	flag.StringVar(&cluster_info_file, "cluster", "", "raw cluster information input file path")
	flag.StringVar(&outfile, "output", "", "output file path for restructured cluster information")
	flag.Parse()

	process_clustinfo()

	fmt.Printf("Dropping duplicates...\n")
	for _, v := range clustinfo {
		v.Pid = unique(v.Pid)
		v.Tax = unique(v.Tax)
		v.Fnc = unique(v.Fnc)
	}

	fmt.Printf("Writing to disk... ")
	fid, err := os.Create(outfile)
	if err != nil {
		panic(err)
	}
	defer fid.Close()
	wtr := gzip.NewWriter(fid)
	defer wtr.Close()

	enc := json.NewEncoder(wtr)
	jj := 0
	for _, v := range clustinfo {
		err = enc.Encode(v)
		if err != nil {
			panic(err)
		}

		// Progress message
		if jj%1000000 == 0 {
			pd := 100 * float64(jj) / float64(len(clustinfo))
			fmt.Printf("%.0f%% ", pd)
		}
		jj++
	}
	fmt.Printf(" done\n")
}
