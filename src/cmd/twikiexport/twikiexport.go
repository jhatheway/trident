///usr/bin/env go run $0 "$@"; exit

/*
 * Trident Wiki Export (twikiexport)
 *
 * This gathers all the relevant files of a FosWiki installation
 * and stores them in a .wiki file (a .tar.gz)
 */

package main

func main() {
	pf_cmd_wikiexport.WikiExport("twikiexport")
}
