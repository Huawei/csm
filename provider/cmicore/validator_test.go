/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// Package cmicore provides a shared library for storage backend operations
package cmicore

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidator_Success(t *testing.T) {
	// arrange
	validateFunc := func(s string) error {
		if s == "" {
			return errors.New("empty")
		}
		return nil
	}

	// action
	gotValidator := NewValidator(validateFunc)

	// assert
	assert.NotNil(t, gotValidator)
}

func TestValidator_Validate_Success(t *testing.T) {
	// arrange
	validateFunc := func(s string) error {
		if s == "" {
			return errors.New("empty")
		}
		return nil
	}
	v := NewValidator(validateFunc)

	// action
	gotErr := v.Validate("valid")

	// assert
	assert.NoError(t, gotErr)
}

func TestValidator_Validate_EmptyInput(t *testing.T) {
	// arrange
	wantErr := errors.New("empty")
	validateFunc := func(s string) error {
		if s == "" {
			return wantErr
		}
		return nil
	}
	v := NewValidator(validateFunc)

	// action
	gotErr := v.Validate("")

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestValidator_Validate_ChainAllPass(t *testing.T) {
	// arrange
	validateFunc1 := func(s string) error {
		if len(s) < 3 {
			return errors.New("too short")
		}
		return nil
	}
	validateFunc2 := func(s string) error {
		if len(s) > 10 {
			return errors.New("too long")
		}
		return nil
	}
	v := NewValidator(validateFunc1, validateFunc2)

	// action
	gotErr := v.Validate("valid")

	// assert
	assert.NoError(t, gotErr)
}

func TestValidator_Validate_ChainFirstFail(t *testing.T) {
	// arrange
	wantErr := errors.New("too short")
	validateFunc1 := func(s string) error {
		return wantErr
	}
	validateFunc2 := func(s string) error {
		return errors.New("should not reach")
	}
	v := NewValidator(validateFunc1, validateFunc2)

	// action
	gotErr := v.Validate("ab")

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestValidator_Validate_ChainSecondFail(t *testing.T) {
	// arrange
	wantErr := errors.New("too long")
	validateFunc1 := func(s string) error {
		return nil
	}
	validateFunc2 := func(s string) error {
		return wantErr
	}
	v := NewValidator(validateFunc1, validateFunc2)

	// action
	gotErr := v.Validate("toolongstring123")

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestValidator_Validate_EmptyValidator(t *testing.T) {
	// arrange
	v := NewValidator[string]()

	// action
	gotErr := v.Validate("any")

	// assert
	assert.NoError(t, gotErr)
}

func TestValidator_Add_Success(t *testing.T) {
	// arrange
	wantErr := errors.New("empty")
	v := NewValidator[string]()
	addFunc := func(s string) error {
		if s == "" {
			return wantErr
		}
		return nil
	}

	// action
	v.Add(addFunc)
	gotErr := v.Validate("")

	// assert
	assert.ErrorIs(t, gotErr, wantErr)
}

func TestValidator_Add_MultipleFunctions(t *testing.T) {
	// arrange
	wantErr := errors.New("too short")
	v := NewValidator[string]()
	fn1 := func(s string) error {
		if len(s) < 3 {
			return wantErr
		}
		return nil
	}
	fn2 := func(s string) error {
		if len(s) > 10 {
			return errors.New("too long")
		}
		return nil
	}

	// action
	v.Add(fn1, fn2)

	// assert - first function triggers error
	gotErr := v.Validate("ab")
	assert.ErrorIs(t, gotErr, wantErr)

	// assert - all functions pass
	gotErr = v.Validate("valid")
	assert.NoError(t, gotErr)
}

func TestValidator_Validate_StopOnFirstError(t *testing.T) {
	// arrange
	callCount := 0
	validateFunc1 := func(s string) error {
		callCount++
		return errors.New("first error")
	}
	validateFunc2 := func(s string) error {
		callCount++
		return errors.New("second error")
	}
	v := NewValidator(validateFunc1, validateFunc2)

	// action
	_ = v.Validate("test")

	// assert - should stop after first error
	assert.Equal(t, 1, callCount)
}
