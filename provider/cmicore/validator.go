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

// Validator is a composable validator that chains multiple validation functions.
// Each function is called in order; the first error stops validation.
type Validator[T any] struct {
	functions []ValidateFunc[T]
}

// ValidateFunc is a single validation function that returns an error if invalid
type ValidateFunc[T any] func(t T) error

// NewValidator creates a new Validator with the given validation functions.
// Functions are applied in order during validation.
func NewValidator[T any](functions ...ValidateFunc[T]) *Validator[T] {
	return &Validator[T]{
		functions: functions,
	}
}

// Validate runs all validation functions in order.
// Returns the first error encountered, or nil if all pass.
func (v *Validator[T]) Validate(t T) error {
	if len(v.functions) == 0 {
		return nil
	}
	for _, fn := range v.functions {
		if err := fn(t); err != nil {
			return err
		}
	}
	return nil
}

// Add appends more validation functions to the validator
func (v *Validator[T]) Add(fn ...ValidateFunc[T]) *Validator[T] {
	v.functions = append(v.functions, fn...)
	return v
}
