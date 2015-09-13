package main

import (
	"bufio"
	"diff/histogramdiff"
	"diff/patiencediff"
	"flag"
	"fmt"
	"github.com/hattya/go.diff"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const EXIT_NO_DIFFERENCE_WERE_FOUND = 0
const EXIT_DIFFERENCE_WERE_FOUND = 1
const EXIT_AN_ERROR_OCCURRED = 2
const CONTEXT_DEFAULT = 3
const NONEWLINE = "No newline at end of file"

//http://pubs.opengroup.org/onlinepubs/9699919799/utilities/diff.html
var flag_b = flag.Bool("b", false, "Ignore changes in amount of white space.")
var flag_c = flag.Bool("c", false, "Context diff (three line context).")
var flag_C = flag.Int("C", 0, "Context diff (specified line context).")
var flag_e = flag.Bool("e", false, "Ed script diff.")
var flag_f = flag.Bool("f", false, "Alternative form of ed script diff.")
var flag_r = flag.Bool("r", false, "Compare directory recursively.")
var flag_u = flag.Bool("u", false, "Unified diff (three line context).")
var flag_U = flag.Int("U", 0, "Unified diff (specified line context).")

var flag_i = flag.Bool("i", false, "Ignore changes in case of text.")

var flag_patience = flag.Bool("patience", false, "Patience Diff.")
var flag_histogram = flag.Bool("histogram", false, "Histogram Diff.")

var flag_utc = flag.Bool("utc", false, "Print time in UTC (for test)")

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(EXIT_AN_ERROR_OCCURRED)
	}

	difffound, err := run(flag.Arg(0), flag.Arg(1))
	if err != nil {
		print_error(fmt.Sprintf("%s", err))
		os.Exit(EXIT_AN_ERROR_OCCURRED)
	}

	if difffound {
		os.Exit(EXIT_DIFFERENCE_WERE_FOUND)
	} else {
		os.Exit(EXIT_NO_DIFFERENCE_WERE_FOUND)
	}
}

func run(apath string, bpath string) (bool, error) {
	if apath == "-" && bpath == "-" {
		return false, nil
	}

	aisdir, err := isdir(apath)
	if err != nil {
		return false, err
	}

	bisdir, err := isdir(bpath)
	if err != nil {
		return false, err
	}

	if (apath == "-" && bisdir) || (bpath == "-" && aisdir) {
		return false, fmt.Errorf("%s", "cannot compare '-' to a directory")
	}

	if aisdir && bisdir {
		return diffdir(apath, bpath)
	} else if aisdir {
		return difffile(xjoinpath(apath, filepath.Base(bpath)), bpath, "")
	} else if bisdir {
		return difffile(apath, xjoinpath(bpath, filepath.Base(apath)), "")
	} else {
		return difffile(apath, bpath, "")
	}
}

func diffdir(adir string, bdir string) (bool, error) {
	afi, err := readdir(adir)
	if err != nil {
		return false, err
	}
	bfi, err := readdir(bdir)
	if err != nil {
		return false, err
	}
	difffound := false
	a := 0
	b := 0
	for a < len(afi) || b < len(bfi) {
		if a >= len(afi) {
			fmt.Printf("Only in %s: %s\n", bdir, bfi[b].Name())
			difffound = true
			b++
		} else if b >= len(afi) {
			fmt.Printf("Only in %s: %s\n", adir, afi[a].Name())
			difffound = true
			a++
		} else if afi[a].Name() < bfi[b].Name() {
			fmt.Printf("Only in %s: %s\n", adir, afi[a].Name())
			difffound = true
			a++
		} else if afi[a].Name() > bfi[b].Name() {
			fmt.Printf("Only in %s: %s\n", bdir, bfi[b].Name())
			difffound = true
			b++
		} else {
			apath := xjoinpath(adir, afi[a].Name())
			bpath := xjoinpath(bdir, bfi[b].Name())
			if afi[a].IsDir() && bfi[b].IsDir() {
				if *flag_r {
					df, err := diffdir(apath, bpath)
					if err != nil {
						return false, err
					}
					if df {
						difffound = true
					}
				} else {
					fmt.Printf("Common subdirectories: %s and %s\n", apath, bpath)
				}
			} else if afi[a].IsDir() {
				fmt.Printf("File %s is a directory while file %s is a regular file\n", apath, bpath)
				difffound = true
			} else if bfi[b].IsDir() {
				fmt.Printf("File %s is a regular file while file %s is a directory\n", apath, bpath)
				difffound = true
			} else {
				head := fmt.Sprintf("%s %s %s\n", reconstructargs(), apath, bpath)
				df, err := difffile(apath, bpath, head)
				if err != nil {
					return false, err
				}
				if df {
					difffound = true
				}
			}
			a++
			b++
		}
	}
	return difffound, nil
}

func difffile(apath string, bpath string, head string) (bool, error) {
	al, err := readfile(apath)
	if err != nil {
		return false, err
	}

	bl, err := readfile(bpath)
	if err != nil {
		return false, err
	}

	acmp := cmpfilter(al)
	bcmp := cmpfilter(bl)

	var cl []diff.Change
	if *flag_histogram {
		cl = histogramdiff.Strings(acmp, bcmp)
	} else if *flag_patience {
		cl = patiencediff.Strings(acmp, bcmp)
	} else {
		cl = diff.Strings(acmp, bcmp)
	}
	cl = change_compact(cl, acmp, bcmp)

	if len(cl) != 0 {
		if head != "" {
			fmt.Print(head)
		}
	}

	if hasflag("C") {
		if len(cl) != 0 {
			err := print_context_diff(cl, al, bl, apath, bpath, *flag_C)
			if err != nil {
				return false, err
			}
		}
	} else if *flag_c {
		if len(cl) != 0 {
			err := print_context_diff(cl, al, bl, apath, bpath, CONTEXT_DEFAULT)
			if err != nil {
				return false, err
			}
		}
	} else if hasflag("U") {
		if len(cl) != 0 {
			err := print_unified_diff(cl, al, bl, apath, bpath, *flag_U)
			if err != nil {
				return false, err
			}
		}
	} else if *flag_u {
		if len(cl) != 0 {
			err := print_unified_diff(cl, al, bl, apath, bpath, CONTEXT_DEFAULT)
			if err != nil {
				return false, err
			}
		}
	} else if *flag_e {
		if len(cl) != 0 {
			print_ed_diff(cl, al, bl)
		}
		if len(al) != 0 && !strings.HasSuffix(al[len(al)-1], "\n") {
			print_error(fmt.Sprintf("%s: %s\n", apath, NONEWLINE))
		}
		if len(bl) != 0 && !strings.HasSuffix(bl[len(bl)-1], "\n") {
			print_error(fmt.Sprintf("%s: %s\n", bpath, NONEWLINE))
		}
	} else if *flag_f {
		if len(cl) != 0 {
			print_alt_ed_diff(cl, al, bl)
		}
		if len(al) != 0 && !strings.HasSuffix(al[len(al)-1], "\n") {
			print_error(fmt.Sprintf("%s: %s\n", apath, NONEWLINE))
		}
		if len(bl) != 0 && !strings.HasSuffix(bl[len(bl)-1], "\n") {
			print_error(fmt.Sprintf("%s: %s\n", bpath, NONEWLINE))
		}
	} else {
		if len(cl) != 0 {
			print_normal_diff(cl, al, bl)
		}
	}

	return len(cl) != 0, nil
}

func cmpfilter(lines []string) []string {
	alt := lines[:]
	for i, _ := range alt {
		if *flag_b {
			re1 := regexp.MustCompile("[ \t\r\n]*$")
			alt[i] = re1.ReplaceAllString(alt[i], "\n")
			re2 := regexp.MustCompile("[ \t\r]+")
			alt[i] = re2.ReplaceAllString(alt[i], " ")
		}
		if *flag_i {
			alt[i] = strings.ToLower(alt[i])
		}
	}
	return alt
}

func print_normal_diff(cl []diff.Change, al []string, bl []string) {
	for _, c := range cl {
		if c.Del == 0 {
			fmt.Printf("%sa%s\n", format_range_normal(c.A, c.Del), format_range_normal(c.B, c.Ins))
			for b := c.B; b < c.B+c.Ins; b++ {
				print_line(fmt.Sprintf("> %s", bl[b]))
			}
		} else if c.Ins == 0 {
			fmt.Printf("%sd%s\n", format_range_normal(c.A, c.Del), format_range_normal(c.B, c.Ins))
			for a := c.A; a < c.A+c.Del; a++ {
				print_line(fmt.Sprintf("< %s", al[a]))
			}
		} else {
			fmt.Printf("%sc%s\n", format_range_normal(c.A, c.Del), format_range_normal(c.B, c.Ins))
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

func format_range_normal(start int, count int) string {
	base := 1
	if count == 0 {
		return fmt.Sprintf("%d", start)
	} else if count == 1 {
		return fmt.Sprintf("%d", base+start)
	} else {
		return fmt.Sprintf("%d,%d", base+start, base+start+count-1)
	}
}

func print_ed_diff(cl []diff.Change, al []string, bl []string) {
	for i := len(cl) - 1; i >= 0; i-- {
		c := cl[i]
		if c.Del == 0 {
			fmt.Printf("%sa\n", format_range_ed(c.A, c.Del))
			for b := c.B; b < c.B+c.Ins; b++ {
				fmt.Printf("%s", bl[b])
				if !strings.HasSuffix(bl[b], "\n") {
					fmt.Printf("\n")
				}
			}
			fmt.Printf(".\n")
		} else if c.Ins == 0 {
			fmt.Printf("%sd\n", format_range_ed(c.A, c.Del))
		} else {
			fmt.Printf("%sc\n", format_range_ed(c.A, c.Del))
			for b := c.B; b < c.B+c.Ins; b++ {
				fmt.Printf("%s", bl[b])
				if !strings.HasSuffix(bl[b], "\n") {
					fmt.Printf("\n")
				}
			}
			fmt.Printf(".\n")
		}
	}
}

func format_range_ed(start int, count int) string {
	base := 1
	if count == 0 {
		return fmt.Sprintf("%d", start)
	} else if count == 1 {
		return fmt.Sprintf("%d", base+start)
	} else {
		return fmt.Sprintf("%d,%d", base+start, base+start+count-1)
	}
}

func print_alt_ed_diff(cl []diff.Change, al []string, bl []string) {
	for _, c := range cl {
		if c.Del == 0 {
			fmt.Printf("a%s\n", format_range_alt_ed(c.A, c.Del))
			for b := c.B; b < c.B+c.Ins; b++ {
				fmt.Printf("%s", bl[b])
				if !strings.HasSuffix(bl[b], "\n") {
					fmt.Printf("\n")
				}
			}
			fmt.Printf(".\n")
		} else if c.Ins == 0 {
			fmt.Printf("d%s\n", format_range_alt_ed(c.A, c.Del))
		} else {
			fmt.Printf("c%s\n", format_range_alt_ed(c.A, c.Del))
			for b := c.B; b < c.B+c.Ins; b++ {
				fmt.Printf("%s", bl[b])
				if !strings.HasSuffix(bl[b], "\n") {
					fmt.Printf("\n")
				}
			}
			fmt.Printf(".\n")
		}
	}
}

func format_range_alt_ed(start int, count int) string {
	base := 1
	if count == 0 {
		return fmt.Sprintf("%d", start)
	} else if count == 1 {
		return fmt.Sprintf("%d", base+start)
	} else {
		return fmt.Sprintf("%d %d", base+start, base+start+count-1)
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
	amodtime, err := fmodtime(apath)
	if err != nil {
		return err
	}
	bmodtime, err := fmodtime(bpath)
	if err != nil {
		return err
	}
	if *flag_utc {
		fmt.Printf("*** %s\t%s\n", apath, amodtime.UTC().Format("Mon Jan _2 15:04:05 2006"))
		fmt.Printf("--- %s\t%s\n", bpath, bmodtime.UTC().Format("Mon Jan _2 15:04:05 2006"))
	} else {
		fmt.Printf("*** %s\t%s\n", apath, amodtime.Format("Mon Jan _2 15:04:05 2006"))
		fmt.Printf("--- %s\t%s\n", bpath, bmodtime.Format("Mon Jan _2 15:04:05 2006"))
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
	amodtime, err := fmodtime(apath)
	if err != nil {
		return err
	}
	bmodtime, err := fmodtime(bpath)
	if err != nil {
		return err
	}
	if *flag_utc {
		fmt.Printf("--- %s\t%s\n", apath, amodtime.UTC().Format("2006-01-02 15:04:05.000000000 -0700"))
		fmt.Printf("+++ %s\t%s\n", bpath, bmodtime.UTC().Format("2006-01-02 15:04:05.000000000 -0700"))
	} else {
		fmt.Printf("--- %s\t%s\n", apath, amodtime.Format("2006-01-02 15:04:05.000000000 -0700"))
		fmt.Printf("+++ %s\t%s\n", bpath, bmodtime.Format("2006-01-02 15:04:05.000000000 -0700"))
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
		fmt.Printf("\n\\ %s\n", NONEWLINE)
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
	var fin *os.File
	if path == "-" {
		fin = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		fin = f
	}
	var lines []string
	r := bufio.NewReader(fin)
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

func readdir(dir string) ([]os.FileInfo, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdir(0)
}

func reconstructargs() string {
	args := []string{cmdname()}
	i := 1
	for i < len(os.Args) {
		if (os.Args[i] == "-C" || os.Args[i] == "-U") && !strings.Contains(os.Args[i], "=") {
			args = append(args, os.Args[i], os.Args[i+1])
			i += 2
		} else if strings.HasPrefix(os.Args[i], "-") {
			args = append(args, os.Args[i])
			i++
		} else {
			i++
		}
	}
	return strings.Join(args, " ")
}

func cmdname() string {
	name := filepath.Base(os.Args[0])
	ext := filepath.Ext(name)
	if ext != "" {
		name = name[0 : len(name)-len(ext)]
	}
	return name
}

func xjoinpath(dir string, file string) string {
	dir = strings.TrimRight(dir, string(os.PathSeparator)+"/")
	return dir + "/" + file
}

func isdir(path string) (bool, error) {
	if path == "-" {
		return false, nil
	}
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

func fmodtime(path string) (time.Time, error) {
	if path == "-" {
		return time.Now(), nil
	}
	fi, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

func print_error(s string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", cmdname(), s)
}

// Move back and forward change groups for a consistent and pretty diff output.
func change_compact(cl []diff.Change, al []string, bl []string) []diff.Change {
	ad, bd := change_to_diff(cl, al, bl)
	ad = change_compact_sub(ad, al)
	bd = change_compact_sub(bd, bl)
	return diff_to_change(ad, bd)
}

func change_to_diff(cl []diff.Change, al []string, bl []string) ([]int, []int) {
	ad := []int{}
	bd := []int{}
	a := 0
	b := 0
	for _, c := range cl {
		for a < c.A {
			ad = append(ad, 0)
			a++
		}
		for a < c.A + c.Del {
			ad = append(ad, -1)
			a++
		}
		for b < c.B {
			bd = append(bd, 0)
			b++
		}
		for b < c.B + c.Ins {
			bd = append(bd, 1)
			b++
		}
	}
	for a < len(al) {
		ad = append(ad, 0)
		a++
	}
	for b < len(bl) {
		bd = append(bd, 0)
		b++
	}
	return ad, bd
}

func diff_to_change(ad []int, bd []int) []diff.Change {
	cl := []diff.Change{}
	a := 0
	b := 0
	for a < len(ad) && b < len(bd) {
		if ad[a] == 0 && bd[b] == 0 {
			a++
			b++
		} else {
			c := diff.Change{}
			c.A = a
			c.B = b
			for a < len(ad) && ad[a] != 0 {
				a++
			}
			for b < len(bd) && bd[b] != 0 {
				b++
			}
			c.Del = a - c.A
			c.Ins = b - c.B
			cl = append(cl, c)
		}
	}
	if a < len(ad) {
		cl = append(cl, diff.Change{A: a, B: b, Del: len(ad) - a, Ins: 0})
	}
	if b < len(bd) {
		cl = append(cl, diff.Change{A: a, B: b, Del: 0, Ins: len(bd) - b})
	}
	return cl
}

func change_compact_sub(df []int, lines []string) []int {
	i := 0
	for i < len(df) {
		for i < len(df) && df[i] == 0 {
			i++
		}
		s := i
		for i < len(df) && df[i] != 0 {
			i++
		}
		e := i
		if s == e {
			break
		}
		start := s
		end := e
		for 0 < s && lines[s - 1] == lines[e - 1] {
			df[s - 1] = df[e - 1]
			df[e - 1] = 0
			e--
			for 0 < s && df[s - 1] != 0 {
				s--
			}
		}
		for e < len(lines) && lines[s] == lines[e] {
			df[e] = df[s]
			df[s] = 0
			s++
			for e < len(df) && df[e] != 0 {
				e++
			}
		}
		if start != s || end != e {
			i = s
		} else {
			i = e
		}
	}
	return df
}
