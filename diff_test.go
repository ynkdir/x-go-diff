package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

const CMDNAME = "./diff"

func init() {
	filepath.Walk("diff_test", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			t := time.Date(2015, 1, 2, 3, 4, 5, 67890000, time.UTC)
			e := os.Chtimes(path, t, t)
			if e != nil {
				return e
			}
		}
		return nil
	})
	exec.Command("go", "build", "diff.go").CombinedOutput()
}

func dotest(t *testing.T, args []string, okfile string, exitcode bool) {
	cmd := exec.Command(CMDNAME, append([]string{"-utc"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatal(err)
		}
	}
	cmdexitcode := (err == nil)
	ok, err := ioutil.ReadFile(okfile)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, ok) {
		t.Errorf("error: result mismatch:\nRESULT:\n%s\nEXPECTED:\n%s", string(out), string(ok))
	} else if cmdexitcode != exitcode {
		t.Errorf("error: exitcode mismatch:\nRESULT:\n%v\nEXPECTED:\n%v", cmdexitcode, exitcode)
	}
}

func dotestin(t *testing.T, args []string, infile string, okfile string, exitcode bool) {
	cmd := exec.Command(CMDNAME, append([]string{"-utc"}, args...)...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	fin, err := os.Open(infile)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(stdin, fin)
	if err != nil {
		t.Fatal(err)
	}
	stdin.Close()
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatal(err)
		}
	}
	cmdexitcode := (err == nil)
	ok, err := ioutil.ReadFile(okfile)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, ok) {
		t.Errorf("error: result mismatch:\nRESULT:\n%s\nEXPECTED:\n%s", string(out), string(ok))
	} else if cmdexitcode != exitcode {
		t.Errorf("error: exitcode mismatch:\nRESULT:\n%v\nEXPECTED:\n%v", cmdexitcode, exitcode)
	}
}

func Test1(t *testing.T) {
	dotest(t, []string{"diff_test/test1_a", "diff_test/test1_b"}, "diff_test/test1_ok", true)
}
func Test2(t *testing.T) {
	dotest(t, []string{"diff_test/test2_a", "diff_test/test2_b"}, "diff_test/test2_ok", true)
}
func Test3(t *testing.T) {
	dotest(t, []string{"diff_test/test3_a", "diff_test/test3_b"}, "diff_test/test3_ok", false)
}
func Test4(t *testing.T) {
	dotest(t, []string{"diff_test/test4_a", "diff_test/test4_b"}, "diff_test/test4_ok", false)
}
func Test5(t *testing.T) {
	dotest(t, []string{"diff_test/test5_a", "diff_test/test5_b"}, "diff_test/test5_ok", false)
}
func Test6(t *testing.T) {
	dotest(t, []string{"diff_test/test6_a", "diff_test/test6_b"}, "diff_test/test6_ok", false)
}
func Test7(t *testing.T) {
	dotest(t, []string{"diff_test/test7_a", "diff_test/test7_b"}, "diff_test/test7_ok", false)
}
func Test8(t *testing.T) {
	dotest(t, []string{"diff_test/test8_a", "diff_test/test8_b"}, "diff_test/test8_ok", false)
}
func Test9(t *testing.T) {
	dotest(t, []string{"diff_test/test9_a", "diff_test/test9_b"}, "diff_test/test9_ok", false)
}
func Test10(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test10_a", "diff_test/test10_b"}, "diff_test/test10_ok", false)
}
func Test11(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test11_a", "diff_test/test11_b"}, "diff_test/test11_ok", false)
}
func Test12(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test12_a", "diff_test/test12_b"}, "diff_test/test12_ok", false)
}
func Test13(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test13_a", "diff_test/test13_b"}, "diff_test/test13_ok", false)
}
func Test14(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test14_a", "diff_test/test14_b"}, "diff_test/test14_ok", false)
}
func Test15(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test15_a", "diff_test/test15_b"}, "diff_test/test15_ok", false)
}
func Test16(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test16_a", "diff_test/test16_b"}, "diff_test/test16_ok", false)
}
func Test17(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test17_a", "diff_test/test17_b"}, "diff_test/test17_ok", false)
}
func Test18(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test18_a", "diff_test/test18_b"}, "diff_test/test18_ok", false)
}
func Test19(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test19_a", "diff_test/test19_b"}, "diff_test/test19_ok", false)
}
func Test20(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test20_a", "diff_test/test20_b"}, "diff_test/test20_ok", false)
}
func Test21(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test21_a", "diff_test/test21_b"}, "diff_test/test21_ok", false)
}
func Test22(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test22_a", "diff_test/test22_b"}, "diff_test/test22_ok", false)
}
func Test23(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test23_a", "diff_test/test23_b"}, "diff_test/test23_ok", false)
}
func Test24(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test24_a", "diff_test/test24_b"}, "diff_test/test24_ok", false)
}
func Test25(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test25_a", "diff_test/test25_b"}, "diff_test/test25_ok", false)
}
func Test26(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test26_a", "diff_test/test26_b"}, "diff_test/test26_ok", false)
}
func Test27(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test27_a", "diff_test/test27_b"}, "diff_test/test27_ok", false)
}
func Test28(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test28_a", "diff_test/test28_b"}, "diff_test/test28_ok", false)
}
func Test29(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test29_a", "diff_test/test29_b"}, "diff_test/test29_ok", false)
}
func Test30(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test30_a", "diff_test/test30_b"}, "diff_test/test30_ok", false)
}
func Test31(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test31_a", "diff_test/test31_b"}, "diff_test/test31_ok", false)
}
func Test32(t *testing.T) {
	dotest(t, []string{"diff_test/test32_a", "diff_test/test32_b"}, "diff_test/test32_ok", false)
}
func Test33(t *testing.T) {
	dotest(t, []string{"diff_test/test33_a", "diff_test/test33_b"}, "diff_test/test33_ok", false)
}
func Test34(t *testing.T) {
	dotest(t, []string{"diff_test/test34_a", "diff_test/test34_b"}, "diff_test/test34_ok", false)
}
func Test35(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test35_a", "diff_test/test35_b"}, "diff_test/test35_ok", false)
}
func Test36(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test36_a", "diff_test/test36_b"}, "diff_test/test36_ok", false)
}
func Test37(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test37_a", "diff_test/test37_b"}, "diff_test/test37_ok", false)
}
func Test38(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test38_a", "diff_test/test38_b"}, "diff_test/test38_ok", false)
}
func Test39(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test39_a", "diff_test/test39_b"}, "diff_test/test39_ok", false)
}
func Test40(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test40_a", "diff_test/test40_b"}, "diff_test/test40_ok", false)
}
func Test41(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test41_a", "diff_test/test41_b"}, "diff_test/test41_ok", false)
}
func Test42(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test42_a", "diff_test/test42_b"}, "diff_test/test42_ok", false)
}
func Test43(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test43_a", "diff_test/test43_b"}, "diff_test/test43_ok", true)
}
func Test44(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test44_a", "diff_test/test44_b"}, "diff_test/test44_ok", true)
}
func Test45(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test45_a", "diff_test/test45_b"}, "diff_test/test45_ok", true)
}
func Test46(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test46_a", "diff_test/test46_b"}, "diff_test/test46_ok", true)
}
func Test47(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test47_a", "diff_test/test47_b"}, "diff_test/test47_ok", true)
}
func Test48(t *testing.T) {
	dotest(t, []string{"diff_test/test48_a", "diff_test/test48_b"}, "diff_test/test48_ok", false)
}
func Test49(t *testing.T) {
	dotest(t, []string{"-r", "diff_test/test49_a", "diff_test/test49_b"}, "diff_test/test49_ok", false)
}
func Test50(t *testing.T) {
	dotest(t, []string{"-e", "diff_test/test50_a", "diff_test/test50_b"}, "diff_test/test50_ok", false)
}
func Test51(t *testing.T) {
	dotest(t, []string{"-e", "diff_test/test51_a", "diff_test/test51_b"}, "diff_test/test51_ok", false)
}
func Test52(t *testing.T) {
	dotest(t, []string{"-e", "diff_test/test52_a", "diff_test/test52_b"}, "diff_test/test52_ok", true)
}
func Test53(t *testing.T) {
	dotest(t, []string{"-f", "diff_test/test53_a", "diff_test/test53_b"}, "diff_test/test53_ok", false)
}
func Test54(t *testing.T) {
	dotest(t, []string{"-f", "diff_test/test54_a", "diff_test/test54_b"}, "diff_test/test54_ok", false)
}
func Test55(t *testing.T) {
	dotest(t, []string{"-f", "diff_test/test55_a", "diff_test/test55_b"}, "diff_test/test55_ok", true)
}
func Test56(t *testing.T) {
	dotestin(t, []string{"-", "diff_test/test56_b"}, "diff_test/test56_a", "diff_test/test56_ok", true)
}
func Test57(t *testing.T) {
	dotestin(t, []string{"diff_test/test57_a", "-"}, "diff_test/test57_b", "diff_test/test57_ok", true)
}
func Test58(t *testing.T) {
	dotestin(t, []string{"-", "-"}, "diff_test/test58_a", "diff_test/test58_ok", true)
}
func Test59(t *testing.T) {
	dotestin(t, []string{"-", "diff_test/test59_b"}, "diff_test/test59_a", "diff_test/test59_ok", false)
}
func Test60(t *testing.T) {
	dotestin(t, []string{"diff_test/test60_a", "-"}, "diff_test/test60_b", "diff_test/test60_ok", false)
}
