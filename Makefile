

path = /nfs/kshedden/Teal_Furnholm/FixUniref/

raw_cluster_info_file = clusterinfo.dat.gz

rev_cluster_info_file = clusterinfo.json.gz

raw_uniref_file = uniref-all.tab.gz

rev_uniref_file = uniref-new.tsv.gz

.PHONY: clean_clustinfo clean_uniref clean uniref clustinfo all

clustinfo_rev = $(path)$(rev_cluster_info_file)
uniref_rev = $(path)$(rev_uniref_file)

clean_clustinfo:
	rm -f $(path)$(rev_cluster_info_file)

clean_uniref:
	rm -f $(path)$(rev_uniref_file)

clean: clean_clustinfo clean_uniref

uniref: $(uniref_rev)

clustinfo: $(clustinfo_rev)

$(clustinfo_rev):
	go run prep_clustinfo.go -cluster=$(path)$(raw_cluster_info_file) -output=$(path)$(rev_cluster_info_file)

$(uniref_rev): $(clustinfo_rev)
	go run fix_uniref.go -cluster=$(path)$(rev_cluster_info_file) -uniref=$(path)$(raw_uniref_file) -output=$(path)$(rev_uniref_file)

all: uniref
