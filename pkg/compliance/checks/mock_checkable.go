// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package checks

import (
	"github.com/DataDog/datadog-agent/pkg/compliance/checks/env"

	"github.com/stretchr/testify/mock"
)

type mockCheckable struct {
	mock.Mock
}

func (m *mockCheckable) check(env env.Env) (*report, error) {
	args := m.Called(env)
	return args.Get(0).(*report), args.Error(1)
}