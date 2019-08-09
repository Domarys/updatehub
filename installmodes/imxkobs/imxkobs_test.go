/*
 * UpdateHub
 * Copyright (C) 2017
 * O.S. Systems Sofware LTDA: contato@ossystems.com.br
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package imxkobs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/UpdateHub/updatehub/installmodes"
	"github.com/UpdateHub/updatehub/testsmocks/cmdlinemock"
	"github.com/UpdateHub/updatehub/utils"

	"github.com/stretchr/testify/assert"
)

func TestImxKobsInit(t *testing.T) {
	val, err := installmodes.GetObject("imxkobs")
	assert.NoError(t, err)

	ik1, ok := val.(*ImxKobsObject)
	if !ok {
		t.Error("Failed to cast return value of \"installmodes.GetObject()\" to ImxKobsObject")
	}

	ik2, ok := getObject().(*ImxKobsObject)
	if !ok {
		t.Error("Failed to cast return value of \"getObject()\" to ImxKobsObject")
	}

	assert.Equal(t, ik2, ik1)
}

func TestImxKobsGetObject(t *testing.T) {
	ik, ok := getObject().(*ImxKobsObject)

	if !ok {
		t.Error("Failed to cast return value of \"getObject()\" to ImxKobsObject")
	}

	cmd := ik.CmdLineExecuter
	_, ok = cmd.(*utils.CmdLine)

	if !ok {
		t.Error("Failed to cast default implementation of \"CmdLineExecuter\" to CmdLine")
	}
}

func TestImxKobsCheckRequirementsWithKobsNGBinaryNotFound(t *testing.T) {
	// setup a temp dir on PATH
	testPath, err := ioutil.TempDir("", "imxkobs-test")
	assert.Nil(t, err)
	defer os.RemoveAll(testPath)
	err = os.Setenv("PATH", testPath)
	assert.NoError(t, err)

	// test the call
	err = checkRequirements()

	assert.EqualError(t, err, "exec: \"kobs-ng\": executable file not found in $PATH")
}

func TestImxKobsCheckRequirementsWithKobsNGBinaryFound(t *testing.T) {
	// setup a temp dir on PATH
	testPath, err := ioutil.TempDir("", "imxkobs-test")
	assert.Nil(t, err)
	defer os.RemoveAll(testPath)
	err = os.Setenv("PATH", testPath)
	assert.NoError(t, err)

	// setup the "kobs-ng" binary on PATH
	kobsngPath := path.Join(testPath, "kobs-ng")
	kobsngFile, err := os.Create(kobsngPath)
	assert.Nil(t, err)
	err = os.Chmod(kobsngPath, 0777)
	assert.Nil(t, err)
	defer kobsngFile.Close()

	// test the call
	err = checkRequirements()

	assert.NoError(t, err)
}

func TestImxKobsSetupNil(t *testing.T) {
	ik := ImxKobsObject{}
	assert.Nil(t, ik.Setup())
}

func TestImxKobsCleanupNil(t *testing.T) {
	ik := ImxKobsObject{}
	assert.Nil(t, ik.Cleanup())
}

func TestImxKobsInstallSuccessCases(t *testing.T) {
	downloadDir := "/dummy-download-dir"
	sha256sum := "a562ce06ed7398848eb910bb60c8c6f68ff36c33701afc30705a96d8eab12123"
	sourcePath := path.Join(downloadDir, sha256sum)

	// FIXME: populate these fields with a json sample?
	testCases := []struct {
		Name                    string
		Add1KPadding            bool
		SearchExponent          int
		Chip0DevicePath         string
		Chip1DevicePath         string
		ExpectedCmdLineExecuter string
	}{
		{
			"SuccessWithAllFields",
			true,
			1,
			"/dev/mtd0",
			"/dev/mtd1",
			fmt.Sprintf("kobs-ng init -x %s --search_exponent=1 --chip_0_device_path=/dev/mtd0 --chip_1_device_path=/dev/mtd1 -v", sourcePath),
		},
		{
			"SuccessWithoutAdd1kPadding",
			false,
			1,
			"/dev/mtd0",
			"/dev/mtd1",
			fmt.Sprintf("kobs-ng init %s --search_exponent=1 --chip_0_device_path=/dev/mtd0 --chip_1_device_path=/dev/mtd1 -v", sourcePath),
		},
		{
			"SuccessWithoutSearchExponent",
			true,
			0,
			"/dev/mtd0",
			"/dev/mtd1",
			fmt.Sprintf("kobs-ng init -x %s --chip_0_device_path=/dev/mtd0 --chip_1_device_path=/dev/mtd1 -v", sourcePath),
		},
		{
			"SuccessWithoutChip0DevicePath",
			true,
			1,
			"",
			"/dev/mtd1",
			fmt.Sprintf("kobs-ng init -x %s --search_exponent=1 --chip_1_device_path=/dev/mtd1 -v", sourcePath),
		},
		{
			"SuccessWithoutChip1DevicePath",
			true,
			1,
			"/dev/mtd0",
			"",
			fmt.Sprintf("kobs-ng init -x %s --search_exponent=1 --chip_0_device_path=/dev/mtd0 -v", sourcePath),
		},
		{
			"SuccessWithNegativeSearchExponent",
			true,
			-1,
			"/dev/mtd0",
			"/dev/mtd1",
			fmt.Sprintf("kobs-ng init -x %s --chip_0_device_path=/dev/mtd0 --chip_1_device_path=/dev/mtd1 -v", sourcePath),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			clm := &cmdlinemock.CmdLineExecuterMock{}
			clm.On("Execute", tc.ExpectedCmdLineExecuter).Return([]byte("combinedOutput"), nil)

			ik := ImxKobsObject{CmdLineExecuter: clm}

			ik.Mode = "imxkobs"
			ik.Sha256sum = "a562ce06ed7398848eb910bb60c8c6f68ff36c33701afc30705a96d8eab12123"
			ik.Add1KPadding = tc.Add1KPadding
			ik.SearchExponent = tc.SearchExponent
			ik.Chip0DevicePath = tc.Chip0DevicePath
			ik.Chip1DevicePath = tc.Chip1DevicePath

			err := ik.Install(downloadDir)
			assert.NoError(t, err)

			clm.AssertExpectations(t)

			assert.Equal(t, "/dev/mtd0ro", ik.GetTarget())
		})
	}
}

func TestImxKobsInstallFailure(t *testing.T) {
	clm := &cmdlinemock.CmdLineExecuterMock{}

	downloadDir := "/dummy-download-dir"
	sha256sum := "a562ce06ed7398848eb910bb60c8c6f68ff36c33701afc30705a96d8eab12123"
	sourcePath := path.Join(downloadDir, sha256sum)

	expectedCmdline := fmt.Sprintf("kobs-ng init -x %s --search_exponent=1 --chip_0_device_path=/dev/mtd0 --chip_1_device_path=/dev/mtd1 -v", sourcePath)
	combinedOutput := "combinedOutput"
	clm.On("Execute", expectedCmdline).Return([]byte(combinedOutput), fmt.Errorf("Error executing 'kobs-ng'. Output: "+combinedOutput))

	ik := ImxKobsObject{CmdLineExecuter: clm}

	ik.Mode = "imxkobs"
	ik.Sha256sum = "a562ce06ed7398848eb910bb60c8c6f68ff36c33701afc30705a96d8eab12123"
	ik.Add1KPadding = true
	ik.SearchExponent = 1
	ik.Chip0DevicePath = "/dev/mtd0"
	ik.Chip1DevicePath = "/dev/mtd1"

	err := ik.Install(downloadDir)
	assert.EqualError(t, err, "Error executing 'kobs-ng'. Output: combinedOutput")

	clm.AssertExpectations(t)

	assert.Equal(t, "/dev/mtd0ro", ik.GetTarget())
}

func TestImxKobsGetTarget(t *testing.T) {
	clm := &cmdlinemock.CmdLineExecuterMock{}

	ik := ImxKobsObject{CmdLineExecuter: clm}

	// default value
	assert.Equal(t, "/dev/mtd0ro", ik.GetTarget())

	// chip0 value
	ik.Chip0DevicePath = "/dev/mtd7"
	assert.Equal(t, "/dev/mtd7ro", ik.GetTarget())
}
