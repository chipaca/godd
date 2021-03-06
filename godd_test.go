package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type GoddTestSuite struct {
	cwd string
}

var _ = Suite(&GoddTestSuite{})

func (g *GoddTestSuite) SetUpTests(c *C) {
	var err error
	g.cwd, err = os.Getwd()
	c.Assert(err, IsNil)
	tempdir := c.MkDir()
	os.Chdir(tempdir)

}

func (g *GoddTestSuite) TearDownTests(c *C) {
	os.Chdir(g.cwd)
}

func (g *GoddTestSuite) TestSimple(c *C) {
	// ensure we test with tiny buffer
	defaultBufSize = 2

	canary := []byte("foo bar")
	err := ioutil.WriteFile("src", canary, 0644)
	c.Assert(err, IsNil)

	err = dd("src", "dst", 0)
	c.Assert(err, IsNil)

	read, err := ioutil.ReadFile("dst")
	c.Assert(err, IsNil)
	c.Assert(read, DeepEquals, canary)
}

func (g *GoddTestSuite) TestParseTrivial(c *C) {
	opts, err := parseArgs([]string{"src", "dst"})
	c.Assert(err, IsNil)
	c.Check(opts.src, Equals, "src")
	c.Check(opts.dst, Equals, "dst")
}

func (g *GoddTestSuite) TestParseIfOf(c *C) {
	opts, err := parseArgs([]string{"if=src", "of=dst"})
	c.Assert(err, IsNil)
	c.Check(opts.src, Equals, "src")
	c.Check(opts.dst, Equals, "dst")
}

func (g *GoddTestSuite) TestParseError(c *C) {
	opts, err := parseArgs([]string{"if=src", "invalid=command"})
	c.Assert(err, ErrorMatches, `unknown argument "invalid=command"`)
	c.Assert(opts, IsNil)
}

func (g *GoddTestSuite) TestParseBs(c *C) {
	opts, err := parseArgs([]string{"if=src", "of=dst", "bs=5"})
	c.Assert(err, IsNil)
	c.Assert(opts, DeepEquals, &ddOpts{
		src: "src",
		dst: "dst",
		bs:  5,
	})
}

func (g *GoddTestSuite) TestParseBsWithString(c *C) {
	opts, err := parseArgs([]string{"if=src", "of=dst", "bs=5M"})
	c.Assert(err, IsNil)
	c.Assert(opts, DeepEquals, &ddOpts{
		src: "src",
		dst: "dst",
		bs:  int64(5 * 1024 * 1024),
	})
}

func (g *GoddTestSuite) TestParseDD(c *C) {
	n, err := ddAtoi("5M")
	c.Assert(err, IsNil)
	c.Assert(n, Equals, int64(5*1024*1024))
}

func (g *GoddTestSuite) TestParseDDTwo(c *C) {
	n, err := ddAtoi("5kB")
	c.Assert(err, IsNil)
	c.Assert(n, Equals, int64(5*1000))
}

func makeMountInfo(c *C, mountSrc, mountPath string) {
	// write a example mountinfo
	cwd, err := os.Getwd()
	c.Assert(err, IsNil)
	mountinfoPath = filepath.Join(cwd, "mountinfo")
	err = ioutil.WriteFile(mountinfoPath, []byte(fmt.Sprintf(`425 22 8:50 / %s rw,nosuid,nodev,relatime shared:442 - ext4 %s rw,data=ordered`, mountPath, mountSrc)), 0644)
	c.Assert(err, IsNil)
}

func (g *GoddTestSuite) TestSanityCheckDstOk(c *C) {
	makeMountInfo(c, "/dev/sdd2", "/media/ubuntu/a")
	err := sanityCheckDst("/some/path")
	c.Assert(err, IsNil)
}

func (g *GoddTestSuite) TestSanityCheckDstMounted(c *C) {
	makeMountInfo(c, "/dev/sdd2", "/media/ubuntu/a")
	err := sanityCheckDst("/dev/sdd2")
	c.Assert(err, ErrorMatches, "/dev/sdd2 is mounted on /media/ubuntu/a")
}

func (g *GoddTestSuite) TestSanityCheckDstParentMounted(c *C) {
	makeMountInfo(c, "/dev/sdd2", "/media/ubuntu/a")
	err := sanityCheckDst("/dev/sdd")
	c.Assert(err, ErrorMatches, "/dev/sdd2 is mounted on /media/ubuntu/a")
}

func (g *GoddTestSuite) TestSanityCheckDst(c *C) {
	makeMountInfo(c, "/dev/sdd2", "/media/ubuntu/a")
	err := sanityCheckDst("/dev/sdd1")
	c.Assert(err, IsNil)
}
