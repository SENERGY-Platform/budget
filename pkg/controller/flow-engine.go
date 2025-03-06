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
	"github.com/SENERGY-Platform/analytics-flow-engine/pkg/lib"
	"github.com/SENERGY-Platform/budget/pkg/models"
	"log"
	"net/http"
	"slices"
	"strings"
)

func (c *Controller) CheckFlowEngine(request *models.ParsedRequest) (int, error) {
	if c.config.AdminAllowAlways && slices.Contains(request.Roles, "admin") {
		if c.config.Debug {
			log.Println("Allowed: User is admin")
		}
		return http.StatusOK, nil
	}

	var requiredBudget uint = 0
	var nowFreeBudget uint = 0

	switch request.TargetMethod {
	default:
		// Most methods do not create a new import flow.
		if c.config.Debug {
			log.Println("Allowed: Unsupervised method")
		}
		return http.StatusOK, nil
	case http.MethodPut:
		var flow lib.PipelineRequest
		err := json.Unmarshal(request.BodyData, &flow)
		if err != nil {
			return http.StatusBadRequest, errors.New("invalid body")
		}
		pipeline, err, code := c.pipelineClient.GetPipeline(request.AuthToken, request.UserId, flow.Id)
		if err != nil {
			return code, err
		}
		for _, operator := range pipeline.Operators {
			nowFreeBudget += operator.Cost
		}
		requiredBudget, err = c.getFlowCost(request.AuthToken, flow.FlowId, request.UserId)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	case http.MethodPost:
		var flow lib.PipelineRequest
		if strings.Split(request.TargetUri, "/")[len(strings.Split(request.TargetUri, "/"))-1] == "pipelines" {
			if c.config.Debug {
				log.Println("Allowed: Unsupervised method")
			}
			return http.StatusOK, nil
		}
		err := json.Unmarshal(request.BodyData, &flow)
		if err != nil {
			return http.StatusBadRequest, errors.New("invalid body")
		}

		requiredBudget, err = c.getFlowCost(request.AuthToken, flow.FlowId, request.UserId)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	totalBudget, err := c.CheckBudgets(request.Roles, request.UserId, models.BudgeIdentifierFlowEngine)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	usedBudget, err := c.GetCurrentlyUsedFlowEngineBudget(request.AuthToken, request.UserId)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if usedBudget > nowFreeBudget { // check for underflow
		usedBudget -= nowFreeBudget
	} else {
		usedBudget = 0
	}

	var availableBudget uint64 = 0
	if totalBudget > uint64(usedBudget) { // check for underflow
		availableBudget = totalBudget - uint64(usedBudget)
	}
	if uint64(requiredBudget) > availableBudget {
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

func (c *Controller) GetCurrentlyUsedFlowEngineBudget(token string, userId string) (uint, error) {
	limit := 10000
	offset := 0
	var budget uint = 0
	for {
		pipelineResp, err, _ := c.pipelineClient.GetPipelines(token, userId, limit, offset, "", false)
		if err != nil {
			return 0, err
		}
		pipelines := pipelineResp.Data
		for _, pipeline := range pipelines {
			for _, operator := range pipeline.Operators {
				budget += operator.Cost
			}
		}
		if len(pipelines) < limit {
			break
		}
		offset += len(pipelines)
	}
	return budget, nil
}

func (c *Controller) getFlowCost(token string, flowId string, userId string) (uint, error) {
	flow, err := c.parsingApi.GetPipeline(flowId, userId, token)
	if err != nil {
		return 0, err
	}
	var cost uint = 0
	for _, operator := range flow.Operators {
		cost += operator.Cost
	}
	return cost, nil
}
