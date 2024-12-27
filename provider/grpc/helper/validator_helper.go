/*
 *  Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

// Package helper is a package that helper function
package helper

// Validator are validate functions
type Validator[T any] struct {
	functions []validateFunc[T]
}

// validateFunc validate function format
type validateFunc[T any] func(t T) error

// NewValidator get a instance of Validator
func NewValidator[T any](functions ...validateFunc[T]) *Validator[T] {
	return &Validator[T]{
		functions: functions,
	}
}

// Validate validate object
func (receiver *Validator[T]) Validate(t T) error {
	if len(receiver.functions) == 0 {
		return nil
	}
	for _, function := range receiver.functions {
		if err := function(t); err != nil {
			return err
		}
	}
	return nil
}
