/*
 * UpdateHub
 * Copyright (C) 2017
 * O.S. Systems Sofware LTDA: contato@ossystems.com.br
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ubifsmock

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
)

type UbifsUtilsMock struct {
	mock.Mock
}

func (uum *UbifsUtilsMock) GetTargetDeviceFromUbiVolumeName(fsBackend afero.Fs, volume string) (string, error) {
	args := uum.Called(fsBackend, volume)
	return args.String(0), args.Error(1)
}
