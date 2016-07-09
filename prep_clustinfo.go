package main

// Restructure the "cluster information" file into a JSON objct for
// each unique cluster id containing three lists: PID's, Taxa, and
// Functions.
//
// Each row of the output file contains the cluster id, then a tab,
// then the JSON-encoded lists.

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

var (
	// File path for input cluster information file
	cluster_info_file = "/nfs/kshedden/Teal_Furnholm/7-7-2016/clusterinfo.dat.gz"

	// File path for output JSON file
	outfile = "/nfs/kshedden/Teal_Furnholm/7-7-2016/clusterinfo.json.gz"

	// If positive, process only this many lines of the
	// cluster_info_file
	truncate = -1

	// PID, taxa, and function data for each cluster
	clustinfo map[string]*clustrec
)

type clustrec struct {
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
	fmt.Printf(" ...done\n")
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

	process_clustinfo()

	fmt.Printf("Sorting...\n")
	for _, v := range clustinfo {
		v.Pid = unique(v.Pid)
		v.Tax = unique(v.Tax)
		v.Fnc = unique(v.Fnc)
	}

	fmt.Printf("Writing to disk...\n")
	fid, err := os.Create(outfile)
	if err != nil {
		panic(err)
	}
	defer fid.Close()
	wtr := gzip.NewWriter(fid)
	defer wtr.Close()

	ar := make([]string, len(clustinfo))
	jj := 0
	for k, _ := range clustinfo {
		ar[jj] = k
		jj++
	}
	sort.StringSlice(ar).Sort()

	enc := json.NewEncoder(wtr)
	for _, v := range ar {
		// Mixing calls to Encode and underlying writer seems
		// OK
		wtr.Write([]byte(v))
		wtr.Write([]byte("\t"))
		err = enc.Encode(clustinfo[v])
		if err != nil {
			panic(err)
		}
	}
}
