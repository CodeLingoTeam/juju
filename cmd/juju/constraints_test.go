package main

import (
	"bytes"
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/cmd"
	"launchpad.net/juju-core/juju/testing"
	"launchpad.net/juju-core/state"
	coretesting "launchpad.net/juju-core/testing"
)

type ConstraintsCommandsSuite struct {
	testing.JujuConnSuite
}

var _ = Suite(&ConstraintsCommandsSuite{})

func runCmdLine(c *C, com cmd.Command, args ...string) (code int, stdout, stderr string) {
	ctx := coretesting.Context(c)
	code = cmd.Main(com, ctx, args)
	stdout = ctx.Stdout.(*bytes.Buffer).String()
	stderr = ctx.Stderr.(*bytes.Buffer).String()
	c.Logf("args:   %#v\ncode:   %d\nstdout: %q\nstderr: %q", args, code, stdout, stderr)
	return
}

func uint64p(val uint64) *uint64 {
	return &val
}

func assertSet(c *C, args ...string) {
	rcode, rstdout, rstderr := runCmdLine(c, &SetConstraintsCommand{}, args...)
	c.Assert(rcode, Equals, 0)
	c.Assert(rstdout, Equals, "")
	c.Assert(rstderr, Equals, "")
}

func (s *ConstraintsCommandsSuite) TestSetEnviron(c *C) {
	// Set constraints.
	assertSet(c, "mem=4G", "cpu-power=250")
	cons, err := s.State.EnvironConstraints()
	c.Assert(err, IsNil)
	c.Assert(cons, DeepEquals, state.Constraints{
		CpuPower: uint64p(250),
		Mem:      uint64p(4096),
	})

	// Clear constraints.
	assertSet(c)
	cons, err = s.State.EnvironConstraints()
	c.Assert(err, IsNil)
	c.Assert(cons, DeepEquals, state.Constraints{})
}

func (s *ConstraintsCommandsSuite) TestSetService(c *C) {
	svc, err := s.State.AddService("svc", s.AddTestingCharm(c, "dummy"))
	c.Assert(err, IsNil)

	// Set constraints.
	assertSet(c, "-s", "svc", "mem=4G", "cpu-power=250")
	cons, err := svc.Constraints()
	c.Assert(err, IsNil)
	c.Assert(cons, DeepEquals, state.Constraints{
		CpuPower: uint64p(250),
		Mem:      uint64p(4096),
	})

	// Clear constraints.
	assertSet(c, "-s", "svc")
	cons, err = svc.Constraints()
	c.Assert(err, IsNil)
	c.Assert(cons, DeepEquals, state.Constraints{})
}

func assertSetError(c *C, code int, stderr string, args ...string) {
	rcode, rstdout, rstderr := runCmdLine(c, &SetConstraintsCommand{}, args...)
	c.Assert(rcode, Equals, code)
	c.Assert(rstdout, Equals, "")
	c.Assert(rstderr, Matches, "error: "+stderr+"\n")
}

func (s *ConstraintsCommandsSuite) TestSetErrors(c *C) {
	assertSetError(c, 2, `invalid service name "badname-0"`, "-s", "badname-0")
	assertSetError(c, 2, `malformed constraint "="`, "=")
	assertSetError(c, 2, `malformed constraint "="`, "-s", "s", "=")
	assertSetError(c, 1, `service "missing" not found`, "-s", "missing")
}

func assertGet(c *C, stdout string, args ...string) {
	rcode, rstdout, rstderr := runCmdLine(c, &GetConstraintsCommand{}, args...)
	c.Assert(rcode, Equals, 0)
	c.Assert(rstdout, Equals, stdout)
	c.Assert(rstderr, Equals, "")
}

func (s *ConstraintsCommandsSuite) TestGetEnvironEmpty(c *C) {
	assertGet(c, "")
}

func (s *ConstraintsCommandsSuite) TestGetEnvironValues(c *C) {
	cons := state.Constraints{CpuCores: uint64p(64)}
	err := s.State.SetEnvironConstraints(cons)
	c.Assert(err, IsNil)
	assertGet(c, "cpu-cores=64\n")
}

func (s *ConstraintsCommandsSuite) TestGetServiceEmpty(c *C) {
	_, err := s.State.AddService("svc", s.AddTestingCharm(c, "dummy"))
	c.Assert(err, IsNil)
	assertGet(c, "", "svc")
}

func (s *ConstraintsCommandsSuite) TestGetServiceValues(c *C) {
	svc, err := s.State.AddService("svc", s.AddTestingCharm(c, "dummy"))
	c.Assert(err, IsNil)
	err = svc.SetConstraints(state.Constraints{CpuCores: uint64p(64)})
	c.Assert(err, IsNil)
	assertGet(c, "cpu-cores=64\n", "svc")
}

func (s *ConstraintsCommandsSuite) TestGetFormats(c *C) {
	cons := state.Constraints{CpuCores: uint64p(64), CpuPower: uint64p(0)}
	err := s.State.SetEnvironConstraints(cons)
	c.Assert(err, IsNil)
	assertGet(c, "cpu-cores=64 cpu-power=\n", "--format", "constraints")
	assertGet(c, "cpu-cores: 64\ncpu-power: 0\n", "--format", "yaml")
	assertGet(c, `{"cpu-cores":64,"cpu-power":0}`+"\n", "--format", "json")
}

func assertGetError(c *C, code int, stderr string, args ...string) {
	rcode, rstdout, rstderr := runCmdLine(c, &GetConstraintsCommand{}, args...)
	c.Assert(rcode, Equals, code)
	c.Assert(rstdout, Equals, "")
	c.Assert(rstderr, Matches, "error: "+stderr+"\n")
}

func (s *ConstraintsCommandsSuite) TestGetErrors(c *C) {
	assertGetError(c, 2, `invalid service name "badname-0"`, "badname-0")
	assertGetError(c, 2, `unrecognized args: \["blether"\]`, "goodname", "blether")
	assertGetError(c, 1, `service "missing" not found`, "missing")
}
