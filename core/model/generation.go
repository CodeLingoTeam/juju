// Copyright 2018 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package model

import (
	"github.com/juju/errors"
)

// TODO (manadart 2019-04-21) Change the nomenclature here to indicate "branch"
// instead of "generation", and remove Current/Next.

// GenerationMaster is used to indicate the main model configuration,
// i.e. that not dealing with in-flight branches.
const GenerationMaster = "master"

// ValidateBranchName returns an error if the input name is not suitable for
// identifying a new in-flight branch.
func ValidateBranchName(name string) error {
	if name == "" {
		return errors.NotValidf("empty branch name")
	}
	if name == GenerationMaster {
		return errors.NotValidf("branch name %q", GenerationMaster)
	}
	return nil
}

// GenerationUnits indicates which units from an application are and are not
// tracking a model branch.
type GenerationUnits struct {
	// UnitsTracking is the names of application units that have been set to
	// track the branch.
	UnitsTracking []string `yaml:"tracking,omitempty"`

	// UnitsPending is the names of application units that are still tracking
	// the master generation.
	UnitsPending []string `yaml:"incomplete,omitempty"`
}

// GenerationApplication represents changes to an application
// made under a generation.
type GenerationApplication struct {
	// ApplicationsName is the name of the application.
	ApplicationName string `yaml:"application"`

	// UnitProgress is summary information about units tracking the branch.
	UnitProgress string `yaml:"progress"`

	// UnitDetail specifies which units are and are not tracking the branch.
	UnitDetail *GenerationUnits `yaml:"units,omitempty"`

	// Config changes are the differing configuration values between this
	// generation and the current.
	// TODO (manadart 2018-02-22) This data-type will evolve as more aspects
	// of the application are made generational.
	ConfigChanges map[string]interface{} `yaml:"config"`
}

// Generation represents detail of a model generation including config changes.
type Generation struct {
	// Created is the formatted time at generation creation.
	Created string `yaml:"created"`

	// Created is the user who created the generation.
	CreatedBy string `yaml:"created-by"`

	// Applications is a collection of applications with changes in this
	// generation including advanced units and modified configuration.
	Applications []GenerationApplication `yaml:"applications"`
}

// GenerationSummaries is a type alias for a representation
// of changes-by-generation.
type GenerationSummaries = map[string]Generation
