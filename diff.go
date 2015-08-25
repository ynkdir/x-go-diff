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
//var flag_c = flag.Bool("c", false, "Produce output in a form that provides three lines of copied context.")
//var flag_C = flag.Int("C", 0, "Produce output in a form that provides n lines of copied context (where n shall be interpreted as a positive decimal integer).")
//var flag_e = flag.Bool("e", false, "Produce output in a form suitable as input for the ed utility, which can then be used to convert file1 into file2.")

//var flag_f = flag.Bool("f", false, "Produce output in an alternative form, similar in format to -e, but not intended to be suitable as input for the ed utility, and in the opposite order.")
//var flag_r = flag.Bool("r", false, "Apply diff recursively to files and directories of the same name when file1 and file2 are both directories.")
var flag_u = flag.Bool("u", false, "Produce output in a form that provides three lines of unified context.")
var flag_U = flag.Int("U", 0, "Produce output in a form that provides n lines of unified context (where n shall be interpreted as a non-negative decimal integer).")

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

	if *flag_U != 0 {
		print_unified(cl, a, b, apath, bpath, *flag_U)
	} else if *flag_u {
		print_unified(cl, a, b, apath, bpath, CONTEXT_DEFAULT)
	} else {
		print_default_diff(cl, a, b)
	}

	return nil
}

func print_default_diff(cl []diff.Change, a []string, b []string) {
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

func print_unified(cl []diff.Change, a []string, b []string, apath string, bpath string, context int) {
	print_unified_head(apath, bpath)
	i := 0
	for i < len(cl) {
		cl_start := i

		// unify change for context
		cl_end := cl_start
		for ; cl_end+1 < len(cl); cl_end++ {
			prev_end := cl[cl_end].A + cl[cl_end].Del
			next_start := cl[cl_end+1].A
			if next_start-prev_end > context*2 {
				break
			}
		}

		astart := cl[cl_start].A - context
		if astart < 0 {
			astart = 0
		}

		acount := cl[cl_end].A + cl[cl_end].Del - cl[cl_start].A + context
		if astart+acount > len(a) {
			acount = len(a) - astart
		}

		bstart := cl[cl_start].B - context
		if bstart < 0 {
			bstart = 0
		}

		bcount := cl[cl_end].B + cl[cl_end].Ins - cl[cl_start].B + context
		if bstart+bcount > len(b) {
			bcount = len(b) - astart
		}

		fmt.Printf("@@-%s +%s @@\n", format_range_unified(astart, acount), format_range_unified(bstart, bcount))

		lnum := astart
		for _, c := range cl[cl_start : cl_end+1] {
			for ; lnum < c.A; lnum++ {
				fmt.Printf("  %s\n", a[lnum])
			}
			for j := c.A; j < c.A+c.Del; j++ {
				fmt.Printf("- %s\n", a[j])
				lnum++
			}
			for j := c.B; j < c.B+c.Ins; j++ {
				fmt.Printf("+ %s\n", b[j])
			}
		}
		for ; lnum < astart+acount; lnum++ {
			fmt.Printf("  %s\n", a[lnum])
		}

		i = cl_end + 1
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
	fmt.Printf("--- %s\t%s\n", bpath, bs.ModTime().Format("2006-01-02 15:04:05.000000000 -0700"))
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
