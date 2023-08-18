/*
 *    Copyright 2023 InfAI (CC SES)
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package models

type Budget struct {
	BudgetIdentifier string `json:"budget_identifier"`
	Role             string `json:"role,omitempty"`
	UserId           string `json:"user_id,omitempty"`
	Value            uint64 `json:"value"`
}

func (b *Budget) Valid() bool {
	if len(b.BudgetIdentifier) == 0 {
		return false
	}
	if len(b.Role) == 0 && len(b.UserId) == 0 {
		return false
	}
	if len(b.Role) > 0 && len(b.UserId) > 0 {
		return false
	}
	return true
}
