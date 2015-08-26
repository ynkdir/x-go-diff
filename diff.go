package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/hattya/go.diff"
	"log"
	"os"
)

const CONTEXT_DEFAULT = 3

//var flag_b = flag.Bool("b", false, "Cause any amount of white space at the end of a line to be treated as a single <newline> (that is, the white-space characters preceding the <newline> are ignored) and other strings of white-space characters, not including <newline> characters, to compare equal.")
var flag_c = flag.Bool("c", false, "Produce output in a form that provides three lines of copied context.")
var flag_C = flag.Int("C", -1, "Produce output in a form that provides n lines of copied context (where n shall be interpreted as a positive decimal integer).")

//var flag_e = flag.Bool("e", false, "Produce output in a form suitable as input for the ed utility, which can then be used to convert file1 into file2.")

//var flag_f = flag.Bool("f", false, "Produce output in an alternative form, similar in format to -e, but not intended to be suitable as input for the ed utility, and in the opposite order.")
//var flag_r = flag.Bool("r", false, "Apply diff recursively to files and directories of the same name when file1 and file2 are both directories.")
var flag_u = flag.Bool("u", false, "Produce output in a form that provides three lines of unified context.")
var flag_U = flag.Int("U", -1, "Produce output in a form that provides n lines of unified context (where n shall be interpreted as a non-negative decimal integer).")

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	err := difffile(flag.Arg(0), flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}
}

func difffile(apath string, bpath string) error {
	a, err := readfile(apath)
	if err != nil {
		return err
	}

	b, err := readfile(bpath)
	if err != nil {
		return err
	}

	cl := diff.Strings(a, b)

	if *flag_C >= 0 {
		print_context_diff(cl, a, b, apath, bpath, *flag_C)
	} else if *flag_c {
		print_context_diff(cl, a, b, apath, bpath, CONTEXT_DEFAULT)
	} else if *flag_U >= 0 {
		print_unified_diff(cl, a, b, apath, bpath, *flag_U)
	} else if *flag_u {
		print_unified_diff(cl, a, b, apath, bpath, CONTEXT_DEFAULT)
	} else {
		print_plain_diff(cl, a, b)
	}

	return nil
}

func print_plain_diff(cl []diff.Change, a []string, b []string) {
	for _, c := range cl {
		if c.Del == 0 {
			fmt.Printf("%sa%s\n", format_range_ed(c.A, c.Del), format_range_ed(c.B, c.Ins))
			for i := 0; i < c.Ins; i++ {
				fmt.Printf("> %s\n", b[c.B+i])
			}
		} else if c.Ins == 0 {
			fmt.Printf("%sd%s\n", format_range_ed(c.A, c.Del), format_range_ed(c.B, c.Ins))
			for i := 0; i < c.Del; i++ {
				fmt.Printf("< %s\n", a[c.A+i])
			}
		} else {
			fmt.Printf("%sc%s\n", format_range_ed(c.A, c.Del), format_range_ed(c.B, c.Ins))
			for i := 0; i < c.Del; i++ {
				fmt.Printf("< %s\n", a[c.A+i])
			}
			fmt.Printf("---\n")
			for i := 0; i < c.Ins; i++ {
				fmt.Printf("> %s\n", b[c.B+i])
			}
		}
	}
}

func format_range_ed(start int, count int) string {
	base := 1
	if count == 0 {
		return fmt.Sprintf("%d", count)
	} else if count == 1 {
		return fmt.Sprintf("%d", base+start)
	} else {
		return fmt.Sprintf("%d,%d", base+start, base+start+count-1)
	}
}

func print_context_diff(cl []diff.Change, a []string, b []string, apath string, bpath string, context int) {
	print_context_head(apath, bpath)
	i := 0
	for i < len(cl) {
		cstart := i
		cend, astart, acount, bstart, bcount := make_hunk(cl, cstart, len(a), len(b), context)
		fmt.Printf("***************\n")
		fmt.Printf("*** %s ****\n", format_range_context(astart, acount))
		lnum := astart
		for _, c := range cl[cstart : cend+1] {
			for ; lnum < c.A; lnum++ {
				fmt.Printf("  %s\n", a[lnum])
			}
			for j := c.A; j < c.A+c.Del; j++ {
				if c.Ins == 0 {
					fmt.Printf("- %s\n", a[j])
				} else {
					fmt.Printf("! %s\n", a[j])
				}
				lnum++
			}
		}
		for ; lnum < astart+acount; lnum++ {
			fmt.Printf("  %s\n", a[lnum])
		}
		fmt.Printf("--- %s ----\n", format_range_context(bstart, bcount))
		lnum = bstart
		for _, c := range cl[cstart : cend+1] {
			for ; lnum < c.B; lnum++ {
				fmt.Printf("  %s\n", b[lnum])
			}
			for j := c.B; j < c.B+c.Ins; j++ {
				if c.Del == 0 {
					fmt.Printf("+ %s\n", b[j])
				} else {
					fmt.Printf("! %s\n", b[j])
				}
				lnum++
			}
		}
		for ; lnum < bstart+bcount; lnum++ {
			fmt.Printf("  %s\n", b[lnum])
		}
		i = cend + 1
	}
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
	fmt.Printf("*** %s\t%s\n", apath, as.ModTime().Format("Mon Jan 02 15:04:05 2006"))
	fmt.Printf("--- %s\t%s\n", bpath, bs.ModTime().Format("Mon Jan 02 15:04:05 2006"))
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

func print_unified_diff(cl []diff.Change, a []string, b []string, apath string, bpath string, context int) {
	print_unified_head(apath, bpath)
	i := 0
	for i < len(cl) {
		cstart := i
		cend, astart, acount, bstart, bcount := make_hunk(cl, cstart, len(a), len(b), context)
		fmt.Printf("@@ -%s +%s @@\n", format_range_unified(astart, acount), format_range_unified(bstart, bcount))
		lnum := astart
		for _, c := range cl[cstart : cend+1] {
			for ; lnum < c.A; lnum++ {
				fmt.Printf(" %s\n", a[lnum])
			}
			for j := c.A; j < c.A+c.Del; j++ {
				fmt.Printf("-%s\n", a[j])
				lnum++
			}
			for j := c.B; j < c.B+c.Ins; j++ {
				fmt.Printf("+%s\n", b[j])
			}
		}
		for ; lnum < astart+acount; lnum++ {
			fmt.Printf(" %s\n", a[lnum])
		}
		i = cend + 1
	}
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
	fmt.Printf("--- %s\t%s\n", apath, as.ModTime().Format("2006-01-02 15:04:05.000000000 -0700"))
	fmt.Printf("+++ %s\t%s\n", bpath, bs.ModTime().Format("2006-01-02 15:04:05.000000000 -0700"))
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

func readfile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
