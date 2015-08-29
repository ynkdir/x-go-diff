package main

import (
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

func dotest(t *testing.T, args []string, okfile string) {
	cmd := exec.Command(CMDNAME, append([]string{"-utc"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatal(err)
		}
	}
	ok, err := ioutil.ReadFile(okfile)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, ok) {
		t.Errorf("error: result mismatch:\nRESULT:\n%s\nEXPECTED:\n%s", string(out), string(ok))
	}
}

func Test1(t *testing.T) {
	dotest(t, []string{"diff_test/test1_a", "diff_test/test1_b"}, "diff_test/test1_ok")
}
func Test2(t *testing.T) {
	dotest(t, []string{"diff_test/test2_a", "diff_test/test2_b"}, "diff_test/test2_ok")
}
func Test3(t *testing.T) {
	dotest(t, []string{"diff_test/test3_a", "diff_test/test3_b"}, "diff_test/test3_ok")
}
func Test4(t *testing.T) {
	dotest(t, []string{"diff_test/test4_a", "diff_test/test4_b"}, "diff_test/test4_ok")
}
func Test5(t *testing.T) {
	dotest(t, []string{"diff_test/test5_a", "diff_test/test5_b"}, "diff_test/test5_ok")
}
func Test6(t *testing.T) {
	dotest(t, []string{"diff_test/test6_a", "diff_test/test6_b"}, "diff_test/test6_ok")
}
func Test7(t *testing.T) {
	dotest(t, []string{"diff_test/test7_a", "diff_test/test7_b"}, "diff_test/test7_ok")
}
func Test8(t *testing.T) {
	dotest(t, []string{"diff_test/test8_a", "diff_test/test8_b"}, "diff_test/test8_ok")
}
func Test9(t *testing.T) {
	dotest(t, []string{"diff_test/test9_a", "diff_test/test9_b"}, "diff_test/test9_ok")
}
func Test10(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test10_a", "diff_test/test10_b"}, "diff_test/test10_ok")
}
func Test11(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test11_a", "diff_test/test11_b"}, "diff_test/test11_ok")
}
func Test12(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test12_a", "diff_test/test12_b"}, "diff_test/test12_ok")
}
func Test13(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test13_a", "diff_test/test13_b"}, "diff_test/test13_ok")
}
func Test14(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test14_a", "diff_test/test14_b"}, "diff_test/test14_ok")
}
func Test15(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test15_a", "diff_test/test15_b"}, "diff_test/test15_ok")
}
func Test16(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test16_a", "diff_test/test16_b"}, "diff_test/test16_ok")
}
func Test17(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test17_a", "diff_test/test17_b"}, "diff_test/test17_ok")
}
func Test18(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test18_a", "diff_test/test18_b"}, "diff_test/test18_ok")
}
func Test19(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test19_a", "diff_test/test19_b"}, "diff_test/test19_ok")
}
func Test20(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test20_a", "diff_test/test20_b"}, "diff_test/test20_ok")
}
func Test21(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test21_a", "diff_test/test21_b"}, "diff_test/test21_ok")
}
func Test22(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test22_a", "diff_test/test22_b"}, "diff_test/test22_ok")
}
func Test23(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test23_a", "diff_test/test23_b"}, "diff_test/test23_ok")
}
func Test24(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test24_a", "diff_test/test24_b"}, "diff_test/test24_ok")
}
func Test25(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test25_a", "diff_test/test25_b"}, "diff_test/test25_ok")
}
func Test26(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test26_a", "diff_test/test26_b"}, "diff_test/test26_ok")
}
func Test27(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test27_a", "diff_test/test27_b"}, "diff_test/test27_ok")
}
func Test28(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test28_a", "diff_test/test28_b"}, "diff_test/test28_ok")
}
func Test29(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test29_a", "diff_test/test29_b"}, "diff_test/test29_ok")
}
func Test30(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test30_a", "diff_test/test30_b"}, "diff_test/test30_ok")
}
func Test31(t *testing.T) {
	dotest(t, []string{"-i", "diff_test/test31_a", "diff_test/test31_b"}, "diff_test/test31_ok")
}
func Test32(t *testing.T) {
	dotest(t, []string{"diff_test/test32_a", "diff_test/test32_b"}, "diff_test/test32_ok")
}
func Test33(t *testing.T) {
	dotest(t, []string{"diff_test/test33_a", "diff_test/test33_b"}, "diff_test/test33_ok")
}
func Test34(t *testing.T) {
	dotest(t, []string{"diff_test/test34_a", "diff_test/test34_b"}, "diff_test/test34_ok")
}
func Test35(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test35_a", "diff_test/test35_b"}, "diff_test/test35_ok")
}
func Test36(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test36_a", "diff_test/test36_b"}, "diff_test/test36_ok")
}
func Test37(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test37_a", "diff_test/test37_b"}, "diff_test/test37_ok")
}
func Test38(t *testing.T) {
	dotest(t, []string{"-c", "diff_test/test38_a", "diff_test/test38_b"}, "diff_test/test38_ok")
}
func Test39(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test39_a", "diff_test/test39_b"}, "diff_test/test39_ok")
}
func Test40(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test40_a", "diff_test/test40_b"}, "diff_test/test40_ok")
}
func Test41(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test41_a", "diff_test/test41_b"}, "diff_test/test41_ok")
}
func Test42(t *testing.T) {
	dotest(t, []string{"-u", "diff_test/test42_a", "diff_test/test42_b"}, "diff_test/test42_ok")
}
func Test43(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test43_a", "diff_test/test43_b"}, "diff_test/test43_ok")
}
func Test44(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test44_a", "diff_test/test44_b"}, "diff_test/test44_ok")
}
func Test45(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test45_a", "diff_test/test45_b"}, "diff_test/test45_ok")
}
func Test46(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test46_a", "diff_test/test46_b"}, "diff_test/test46_ok")
}
func Test47(t *testing.T) {
	dotest(t, []string{"-b", "diff_test/test47_a", "diff_test/test47_b"}, "diff_test/test47_ok")
}
func Test48(t *testing.T) {
	dotest(t, []string{"diff_test/test48_a", "diff_test/test48_b"}, "diff_test/test48_ok")
}
func Test49(t *testing.T) {
	dotest(t, []string{"-r", "diff_test/test49_a", "diff_test/test49_b"}, "diff_test/test49_ok")
}
