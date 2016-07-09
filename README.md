Download the two Go scripts and the Makefile, and put them into a
directory somewhere.

Edit the path and file name variables in the Makefile as neeed.  The
variables that can be edited are:

`path`, `raw_cluster_info_file`, `rev_cluster_info_file`,
`raw_uniref_file`, `rev_uniref_file`.

Run `make all` to build everything.

This requires Go version 1.5 or greater since we are using
`bufio.Scanner.Buffer`.  Use `go version` to check the version.

Details:

Run `make clean` to remove all constructed data files.  Run `make
clean_clustinfo` to remove only the constructed cluster information,
and `make clean_uniref` to remove only the constructed uniref file.

Run `make clustinfo` to build only the cluster information file, and
run `make uniref` to build only the revised uniref file.

The Go scripts can be run directly without using the Makefile:

`> go run prep_clustinfo.go -cluster=/nfs/kshedden/Teal_Furnholm/FixUniref/clusterinfo.dat.gz -output=/nfs/kshedden/Teal_Furnholm/FixUniref/clusterinfo.json.gz`

`> go run fix_uniref.go -cluster=/nfs/kshedden/Teal_Furnholm/FixUniref/clusterinfo.json.gz -uniref=/nfs/kshedden/Teal_Furnholm/FixUniref/uniref-all.tab.gz -output=/nfs/kshedden/Teal_Furnholm/FixUniref/uniref-new.tsv.gz`
