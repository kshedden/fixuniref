
Download the two scripts and edit the path variables:

In `prep_clustinfo.go`, edit: `cluster_info_file`, `outfile`

In `fixuniref.go`, edit: `cluster_info_file`, `uniref_file`, `outfile`

Note that `outfile` in `prep_clustinfo.go` must be the same as
`cluster_info_file` in `fixuniref.go`.

Then run `prep_clustinfo`, then run `fix_uniref`.