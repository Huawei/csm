/*
 Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

// Package topology
package topology

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func Test_option_ValidateConfig_Success(t *testing.T) {
	// arrange
	o := &option{
		controllerWorkers: 4,
		supportResources:  []string{"Pod", "PersistentVolume"},
		resyncPeriod:      defaultResyncPeriod,
	}

	// act
	err := o.ValidateConfig()

	// assert
	if err != nil {
		t.Errorf("Test_option_ValidateConfig_Success failed: [%v]", err)
	}
}

func Test_option_ValidateConfig_ControllerWorkersLessThanOne_Failed(t *testing.T) {
	// arrange
	o := &option{
		controllerWorkers: 0,
		supportResources:  []string{"Pod", "PersistentVolume"},
	}
	want := fmt.Errorf("invalid controller workers count [%d]", o.controllerWorkers)

	// act
	got := o.ValidateConfig()

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Test_option_ValidateConfig_ControllerWorkersLessThanOne_Failed: want [%v], got [%v]", want, got)
	}
}

func Test_option_ValidateConfig_SupportResourcesLengthLessThanTwo_Failed(t *testing.T) {
	// arrange
	o := &option{
		controllerWorkers: 1,
		supportResources:  []string{},
	}
	want := errors.New("supported resources should be at least 2")

	// act
	got := o.ValidateConfig()

	// assert
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Test_option_ValidateConfig_SupportResourcesLengthLessThanTwo_Failed: "+
			"want [%v], got [%v]", want, got)
	}
}
