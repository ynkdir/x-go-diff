package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/hattya/go.diff"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const EXIT_NO_DIFFERENCE_WERE_FOUND = 0
const EXIT_DIFFERENCE_WERE_FOUND = 1
const EXIT_AN_ERROR_OCCURRED = 2
const CONTEXT_DEFAULT = 3
const NONEWLINE = "\\ No newline at end of file"

//http://pubs.opengroup.org/onlinepubs/9699919799/utilities/diff.html
var flag_b = flag.Bool("b", false, "Ignore changes in amount of white space.")
var flag_c = flag.Bool("c", false, "Context diff (three line context).")
var flag_C = flag.Int("C", 0, "Context diff (specified line context).")

//var flag_e = flag.Bool("e", false, "Produce output in a form suitable as input for the ed utility, which can then be used to convert file1 into file2.")

//var flag_f = flag.Bool("f", false, "Produce output in an alternative form, similar in format to -e, but not intended to be suitable as input for the ed utility, and in the opposite order.")
//var flag_r = flag.Bool("r", false, "Apply diff recursively to files and directories of the same name when file1 and file2 are both directories.")
var flag_u = flag.Bool("u", false, "Unified diff (three line context).")
var flag_U = flag.Int("U", 0, "Unified diff (specified line context).")

var flag_i = flag.Bool("i", false, "Ignore changes in case of text.")

var flag_utc = flag.Bool("utc", false, "Print time in UTC (for test)")

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(EXIT_AN_ERROR_OCCURRED)
	}

	difffound, err := run(flag.Arg(0), flag.Arg(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(EXIT_AN_ERROR_OCCURRED)
	}

	if difffound {
		os.Exit(EXIT_DIFFERENCE_WERE_FOUND)
	} else {
		os.Exit(EXIT_NO_DIFFERENCE_WERE_FOUND)
	}
}

func run(apath string, bpath string) (bool, error) {
	ainfo, err := os.Stat(apath)
	if err != nil {
		return false, err
	}

	binfo, err := os.Stat(bpath)
	if err != nil {
		return false, err
	}

	if ainfo.IsDir() && binfo.IsDir() {
		return diffdir(apath, bpath)
	} else if ainfo.IsDir() {
		return difffile(filepath.Join(apath, binfo.Name()), bpath)
	} else if binfo.IsDir() {
		return difffile(apath, filepath.Join(bpath, ainfo.Name()))
	} else {
		return difffile(apath, bpath)
	}
}

func diffdir(apath string, bpath string) (bool, error) {
	return false, fmt.Errorf("NOT IMPLEMENTED")
}

func difffile(apath string, bpath string) (bool, error) {
	al, err := readfile(apath)
	if err != nil {
		return false, err
	}

	bl, err := readfile(bpath)
	if err != nil {
		return false, err
	}

	cl := diff.Strings(cmpfilter(al), cmpfilter(bl))
	if len(cl) == 0 {
		return false, nil
	}

	if hasflag("C") {
		err := print_context_diff(cl, al, bl, apath, bpath, *flag_C)
		if err != nil {
			return false, err
		}
	} else if *flag_c {
		err := print_context_diff(cl, al, bl, apath, bpath, CONTEXT_DEFAULT)
		if err != nil {
			return false, err
		}
	} else if hasflag("U") {
		err := print_unified_diff(cl, al, bl, apath, bpath, *flag_U)
		if err != nil {
			return false, err
		}
	} else if *flag_u {
		err := print_unified_diff(cl, al, bl, apath, bpath, CONTEXT_DEFAULT)
		if err != nil {
			return false, err
		}
	} else {
		print_plain_diff(cl, al, bl)
	}

	return true, nil
}

func cmpfilter(lines []string) []string {
	alt := lines[:]
	for i, _ := range alt {
		if *flag_b {
			re1 := regexp.MustCompile("[ \t\r]*\n$")
			alt[i] = re1.ReplaceAllString(alt[i], "\n")
			re2 := regexp.MustCompile("[ \t\r]*$")
			alt[i] = re2.ReplaceAllString(alt[i], "")
			re3 := regexp.MustCompile("[ \t]+")
			alt[i] = re3.ReplaceAllString(alt[i], " ")
		}
		if *flag_i {
			alt[i] = strings.ToLower(alt[i])
		}
	}
	return alt
}

func print_plain_diff(cl []diff.Change, al []string, bl []string) {
	for _, c := range cl {
		if c.Del == 0 {
			fmt.Printf("%sa%s\n", format_range_plain(c.A, c.Del), format_range_plain(c.B, c.Ins))
			for b := c.B; b < c.B+c.Ins; b++ {
				print_line(fmt.Sprintf("> %s", bl[b]))
			}
		} else if c.Ins == 0 {
			fmt.Printf("%sd%s\n", format_range_plain(c.A, c.Del), format_range_plain(c.B, c.Ins))
			for a := c.A; a < c.A+c.Del; a++ {
				print_line(fmt.Sprintf("< %s", al[a]))
			}
		} else {
			fmt.Printf("%sc%s\n", format_range_plain(c.A, c.Del), format_range_plain(c.B, c.Ins))
			for a := c.A; a < c.A+c.Del; a++ {
				print_line(fmt.Sprintf("< %s", al[a]))
			}
			fmt.Printf("---\n")
			for b := c.B; b < c.B+c.Ins; b++ {
				print_line(fmt.Sprintf("> %s", bl[b]))
			}
		}
	}
}

func format_range_plain(start int, count int) string {
	base := 1
	if count == 0 {
		return fmt.Sprintf("%d", start)
	} else if count == 1 {
		return fmt.Sprintf("%d", base+start)
	} else {
		return fmt.Sprintf("%d,%d", base+start, base+start+count-1)
	}
}

func print_context_diff(cl []diff.Change, al []string, bl []string, apath string, bpath string, context int) error {
	err := print_context_head(apath, bpath)
	if err != nil {
		return err
	}
	cstart := 0
	for cstart < len(cl) {
		cend, astart, acount, bstart, bcount := make_hunk(cl, cstart, len(al), len(bl), context)
		fmt.Printf("***************\n")
		fmt.Printf("*** %s ****\n", format_range_context(astart, acount))
		hasdel := false
		hasins := false
		for _, c := range cl[cstart : cend+1] {
			if c.Del != 0 {
				hasdel = true
			}
			if c.Ins != 0 {
				hasins = true
			}
		}
		if hasdel {
			a := astart
			for _, c := range cl[cstart : cend+1] {
				for ; a < c.A; a++ {
					print_line(fmt.Sprintf("  %s", al[a]))
				}
				for ; a < c.A+c.Del; a++ {
					if c.Ins == 0 {
						print_line(fmt.Sprintf("- %s", al[a]))
					} else {
						print_line(fmt.Sprintf("! %s", al[a]))
					}
				}
			}
			for ; a < astart+acount; a++ {
				print_line(fmt.Sprintf("  %s", al[a]))
			}
		}
		fmt.Printf("--- %s ----\n", format_range_context(bstart, bcount))
		if hasins {
			b := bstart
			for _, c := range cl[cstart : cend+1] {
				for ; b < c.B; b++ {
					print_line(fmt.Sprintf("  %s", bl[b]))
				}
				for ; b < c.B+c.Ins; b++ {
					if c.Del == 0 {
						print_line(fmt.Sprintf("+ %s", bl[b]))
					} else {
						print_line(fmt.Sprintf("! %s", bl[b]))
					}
				}
			}
			for ; b < bstart+bcount; b++ {
				print_line(fmt.Sprintf("  %s", bl[b]))
			}
		}
		cstart = cend + 1
	}
	return nil
}

func print_context_head(apath string, bpath string) error {
	as, err := os.Stat(apath)
	if err != nil {
		return err
	}
	bs, err := os.Stat(bpath)
	if err != nil {
		return err
	}
	if *flag_utc {
		fmt.Printf("*** %s\t%s\n", apath, as.ModTime().UTC().Format("Mon Jan _2 15:04:05 2006"))
		fmt.Printf("--- %s\t%s\n", bpath, bs.ModTime().UTC().Format("Mon Jan _2 15:04:05 2006"))
	} else {
		fmt.Printf("*** %s\t%s\n", apath, as.ModTime().Format("Mon Jan _2 15:04:05 2006"))
		fmt.Printf("--- %s\t%s\n", bpath, bs.ModTime().Format("Mon Jan _2 15:04:05 2006"))
	}
	return nil
}

func format_range_context(start int, count int) string {
	base := 1
	if count == 0 {
		return fmt.Sprintf("%d", start)
	} else if count == 1 {
		return fmt.Sprintf("%d", base+start)
	} else {
		return fmt.Sprintf("%d,%d", base+start, base+start+count-1)
	}
}

func print_unified_diff(cl []diff.Change, al []string, bl []string, apath string, bpath string, context int) error {
	err := print_unified_head(apath, bpath)
	if err != nil {
		return err
	}
	cstart := 0
	for cstart < len(cl) {
		cend, astart, acount, bstart, bcount := make_hunk(cl, cstart, len(al), len(bl), context)
		fmt.Printf("@@ -%s +%s @@\n", format_range_unified(astart, acount), format_range_unified(bstart, bcount))
		a := astart
		for _, c := range cl[cstart : cend+1] {
			for ; a < c.A; a++ {
				print_line(fmt.Sprintf(" %s", al[a]))
			}
			for ; a < c.A+c.Del; a++ {
				print_line(fmt.Sprintf("-%s", al[a]))
			}
			for b := c.B; b < c.B+c.Ins; b++ {
				print_line(fmt.Sprintf("+%s", bl[b]))
			}
		}
		for ; a < astart+acount; a++ {
			print_line(fmt.Sprintf(" %s", al[a]))
		}
		cstart = cend + 1
	}
	return nil
}

func print_unified_head(apath string, bpath string) error {
	as, err := os.Stat(apath)
	if err != nil {
		return err
	}
	bs, err := os.Stat(bpath)
	if err != nil {
		return err
	}
	if *flag_utc {
		fmt.Printf("--- %s\t%s\n", apath, as.ModTime().UTC().Format("2006-01-02 15:04:05.000000000 -0700"))
		fmt.Printf("+++ %s\t%s\n", bpath, bs.ModTime().UTC().Format("2006-01-02 15:04:05.000000000 -0700"))
	} else {
		fmt.Printf("--- %s\t%s\n", apath, as.ModTime().Format("2006-01-02 15:04:05.000000000 -0700"))
		fmt.Printf("+++ %s\t%s\n", bpath, bs.ModTime().Format("2006-01-02 15:04:05.000000000 -0700"))
	}
	return nil
}

func format_range_unified(start int, count int) string {
	base := 1
	if start == 0 && count == 0 {
		return "0,0"
	} else if count == 1 {
		return fmt.Sprintf("%d", base+start)
	} else {
		return fmt.Sprintf("%d,%d", base+start, count)
	}
}

func print_line(line string) {
	fmt.Print(line)
	if !strings.HasSuffix(line, "\n") {
		fmt.Printf("\n%s\n", NONEWLINE)
	}
}

func hasflag(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func make_hunk(cl []diff.Change, cstart int, alen int, blen int, context int) (cend, astart, acount, bstart, bcount int) {
	cend = cstart
	for ; cend+1 < len(cl); cend++ {
		prev_end := cl[cend].A + cl[cend].Del
		next_start := cl[cend+1].A
		if next_start-prev_end > context*2 {
			break
		}
	}

	astart = cl[cstart].A - context
	if astart < 0 {
		astart = 0
	}

	acount = cl[cend].A + cl[cend].Del - astart + context
	if astart+acount > alen {
		acount = alen - astart
	}

	bstart = cl[cstart].B - context
	if bstart < 0 {
		bstart = 0
	}

	bcount = cl[cend].B + cl[cend].Ins - bstart + context
	if bstart+bcount > blen {
		bcount = blen - astart
	}

	return
}

func readfile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			if line != "" {
				lines = append(lines, line)
			}
			break
		}
		lines = append(lines, line)
	}
	return lines, nil
}
