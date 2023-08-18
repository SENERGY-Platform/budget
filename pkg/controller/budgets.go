/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"errors"
	"github.com/SENERGY-Platform/budget/pkg/models"
)

func (c *Controller) GetBudgets(limit int, offset int, roles []string, userId, budgetIdentifier string) (budgets []models.Budget, err error) {
	return c.db.ListBudgets(limit, offset, budgetIdentifier, userId, roles)
}

func (c *Controller) SetBudget(budget models.Budget) (err error) {
	if !budget.Valid() {
		return errors.New("invalid budget")
	}
	return c.db.SetBudget(budget)
}

func (c *Controller) DeleteBudget(budgetIdentifier string, userId string, role string) (err error) {
	return c.db.RemoveBudget(budgetIdentifier, userId, role)
}

func (c *Controller) CheckBudgets(roles []string, userId, budgetIdentifier string) (available uint64, err error) {
	limit := 10000
	offset := 0
	available = 0
	for {
		returned, err := c.db.ListBudgets(limit, offset, budgetIdentifier, userId, roles)
		if err != nil {
			return 0, err
		}
		for i := range returned {
			if returned[i].Value > available {
				available = returned[i].Value
			}
		}
		if len(returned) < limit {
			break
		}
		offset += len(returned)
	}
	return
}
