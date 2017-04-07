// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package application

import (
	"fmt"

	"github.com/juju/cmd"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils/series"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v6-unstable"

	"github.com/juju/juju/cmd/cmdtesting"
	jujutesting "github.com/juju/juju/juju/testing"
	"github.com/juju/juju/state"
	"github.com/juju/juju/testcharms"
	"github.com/juju/juju/testing"
)

type RemoveUnitSuite struct {
	jujutesting.RepoSuite
	testing.CmdBlockHelper
}

func (s *RemoveUnitSuite) SetUpTest(c *gc.C) {
	s.RepoSuite.SetUpTest(c)
	s.CmdBlockHelper = testing.NewCmdBlockHelper(s.APIState)
	c.Assert(s.CmdBlockHelper, gc.NotNil)
	s.AddCleanup(func(*gc.C) { s.CmdBlockHelper.Close() })
}

var _ = gc.Suite(&RemoveUnitSuite{})

func runRemoveUnit(c *gc.C, args ...string) (*cmd.Context, error) {
	return cmdtesting.RunCommand(c, NewRemoveUnitCommand(), args...)
}

func (s *RemoveUnitSuite) setupUnitForRemove(c *gc.C) *state.Application {
	ch := testcharms.Repo.CharmArchivePath(s.CharmsPath, "dummy")
	err := runDeploy(c, "-n", "2", ch, "dummy", "--series", series.LatestLts())
	c.Assert(err, jc.ErrorIsNil)
	curl := charm.MustParseURL(fmt.Sprintf("local:%s/dummy-1", series.LatestLts()))
	svc, _ := s.AssertService(c, "dummy", curl, 2, 0)
	return svc
}

func (s *RemoveUnitSuite) TestRemoveUnit(c *gc.C) {
	svc := s.setupUnitForRemove(c)

	ctx, err := runRemoveUnit(c, "dummy/0", "dummy/1", "dummy/2", "sillybilly/17")
	c.Assert(err, gc.Equals, cmd.ErrSilent)
	stderr := cmdtesting.Stderr(ctx)
	c.Assert(stderr, gc.Equals, `
removing unit dummy/0
removing unit dummy/1
removing unit dummy/2 failed: unit "dummy/2" does not exist
removing unit sillybilly/17 failed: unit "sillybilly/17" does not exist
`[1:])

	units, err := svc.AllUnits()
	c.Assert(err, jc.ErrorIsNil)
	for _, u := range units {
		c.Assert(u.Life(), gc.Equals, state.Dying)
	}
}

func (s *RemoveUnitSuite) TestRemoveUnitInformsStorageRemoval(c *gc.C) {
	ch := testcharms.Repo.CharmArchivePath(s.CharmsPath, "storage-filesystem")
	err := runDeploy(c, ch, "storage-filesystem", "--series", "quantal", "--storage", "data=rootfs,2")
	c.Assert(err, jc.ErrorIsNil)

	ctx, err := runRemoveUnit(c, "storage-filesystem/0")
	c.Assert(err, jc.ErrorIsNil)
	stderr := cmdtesting.Stderr(ctx)
	c.Assert(stderr, gc.Equals, `
removing unit storage-filesystem/0
- will remove storage data/0
- will remove storage data/1
`[1:])
}

func (s *RemoveUnitSuite) TestBlockRemoveUnit(c *gc.C) {
	svc := s.setupUnitForRemove(c)

	// block operation
	s.BlockRemoveObject(c, "TestBlockRemoveUnit")
	_, err := runRemoveUnit(c, "dummy/0", "dummy/1")
	s.AssertBlocked(c, err, ".*TestBlockRemoveUnit.*")
	c.Assert(svc.Life(), gc.Equals, state.Alive)
}
