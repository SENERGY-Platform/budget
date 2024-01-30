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

package validator

import (
	"github.com/SENERGY-Platform/budget/pkg/controller"
	"github.com/SENERGY-Platform/budget/pkg/models"
)

func init() {
	validations = append(validations, validateImportDeploy)
}

func validateImportDeploy(controller *controller.Controller, token string, userid string, roles []string, adminToken string) (budgetIdentifier string, maxBudget uint64, actualBudget uint64, err error) {
	budgetIdentifier = models.BudgeIdentifierImportDeploy
	maxBudget, err = controller.CheckBudgets(roles, userid, budgetIdentifier)
	if err != nil {
		return
	}
	actualBudget, err = controller.GetCurrentlyUsedImportDeployBudget(userid, adminToken, userid)
	return
}
