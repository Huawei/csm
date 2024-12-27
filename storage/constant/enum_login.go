/*
 Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

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

// Package constant is related with storage client constant
package constant

// AccountState is login account state
type AccountState float64

// login account state
const (
	LoginNormal                    AccountState = 1
	LoginPasswordExpired           AccountState = 3
	LoginInitialPassword           AccountState = 4
	LoginPasswordIsAboutToExpire   AccountState = 5
	NextLoginPasswordMustBeChanged AccountState = 6
	LoginPasswordNeverExpires      AccountState = 7
	LoginAuthenticateEmailAddress  AccountState = 8
	LoginPasswordNeedInitialized   AccountState = 9
	LoginAuthenticateRadius        AccountState = 10
	LoginChallengeRadiusResponse   AccountState = 11
)

var (
	// LoginAccountStateMap is login account state map
	LoginAccountStateMap = map[float64]string{
		1:  "normal",
		3:  "password expired",
		4:  "initial password, which must be reset",
		5:  "The password is about to expire",
		6:  "The password must be changed upon the next login",
		7:  "The password never expires",
		8:  "one-time password for authenticating the email address",
		9:  "The device is in the first login state and the password needs to be initialized",
		10: "RADIUS one-time password authentication is required",
		11: "RADIUS challenge response is required",
	}
)
