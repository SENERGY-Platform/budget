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
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/budget/pkg/models"
	"github.com/SENERGY-Platform/import-deploy/lib/auth"
	"github.com/SENERGY-Platform/import-deploy/lib/model"
	auth2 "github.com/SENERGY-Platform/import-repository/lib/auth"
	"log"
	"net/http"
	"slices"
	"strings"
)

func (c *Controller) CheckImportDeploy(request *models.ParsedRequest) (int, error) {
	if c.config.AdminAllowAlways && slices.Contains(request.Roles, "admin") {
		if c.config.Debug {
			log.Println("Allowed: User is admin")
		}
		return http.StatusOK, nil
	}

	switch request.TargetMethod {
	default:
		// Most methods do not create a new import instance. A change of import type in PUT is not prohibited and is enforced by import-deploy
		if c.config.Debug {
			log.Println("Allowed: Unsupervised method")
		}
		return http.StatusOK, nil
	case http.MethodPost:
		var instance model.Instance
		err := json.Unmarshal(request.BodyData, &instance)
		if err != nil {
			return http.StatusBadRequest, errors.New("invalid body")
		}

		totalBudget, err := c.CheckBudgets(request.Roles, request.UserId, models.BudgeIdentifierImportDeploy)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		requiredBudget, err := c.getImportTypeCost(request.UserId, request.AuthToken, instance.ImportTypeId)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		usedBudget, err := c.GetCurrentlyUsedImportDeployBudget(request.UserId, request.AuthToken, "")
		if err != nil {
			return http.StatusInternalServerError, err
		}
		var availableBudget uint64 = 0
		if totalBudget > usedBudget { // check for underflow
			availableBudget = totalBudget - usedBudget
		}
		if requiredBudget > availableBudget {
			if c.config.Debug {
				log.Printf("Forbidden: Budget exceeded, required %d, available %d, total %d\n", requiredBudget, availableBudget, totalBudget)
			}
			return http.StatusPaymentRequired, errors.New("budget exceeded")
		}
		if c.config.Debug {
			log.Printf("Allowed: Budget ok, required %d, available %d, total %d\n", requiredBudget, availableBudget, totalBudget)
		}
		return http.StatusOK, nil
	}
}

func (c *Controller) GetCurrentlyUsedImportDeployBudget(userId string, token string, overwriteTokenUserId string) (uint64, error) {
	token = strings.TrimPrefix(token, "bearer ")
	token = strings.TrimPrefix(token, "Bearer ")

	limit := 10000
	var offset int64 = 0
	var budget uint64 = 0
	for {
		instances, err, _ := c.importDeploy.ListInstances(auth.Token{Token: token}, int64(limit), offset, "", false, "", true, overwriteTokenUserId)
		if err != nil {
			return 0, err
		}
		for i := range instances {
			cost, err := c.getImportTypeCost(userId, token, instances[i].ImportTypeId)
			if err != nil {
				return 0, err
			}
			budget += cost
		}
		if len(instances) < limit {
			break
		}
		offset += int64(len(instances))
	}
	return budget, nil
}

func (c *Controller) getImportTypeCost(userId string, token string, importTypeId string) (uint64, error) {
	token = strings.TrimPrefix(token, "bearer ")
	token = strings.TrimPrefix(token, "Bearer ")
	importType, err, code := c.importRepo.ReadImportType(importTypeId, auth2.Token{Token: token})
	if err != nil {
		if code == http.StatusNotFound {
			msg := "WARNING: Import Type " + importTypeId + " no longer exists, but still in use by " + userId
			log.Println(msg + ", assuming 0 cost")
			err = c.SendSlackMessage(msg)
			return 0, nil
		} else if code == http.StatusForbidden {
			msg := "WARNING: Import Type " + importTypeId + " is in by " + userId + ", but has no access (403)"
			log.Println(msg + ", assuming 0 cost")
			err = c.SendSlackMessage(msg)
			return 0, nil
		}
	}
	// TODO cache
	return importType.Cost, nil
}
