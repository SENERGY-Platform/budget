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

package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/controller"
	"github.com/SENERGY-Platform/budget/pkg/log"
	"github.com/SENERGY-Platform/budget/pkg/models"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/gin-gonic/gin"
)

func init() {
	endpoints = append(endpoints, BudgetEndpoints)
}

func BudgetEndpoints(router *gin.Engine, _ configuration.Config, control *controller.Controller) {
	router.GET("/budgets", func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "100")
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			log.Logger.Warn("invalid limit query parameter", attributes.ErrorKey, err, "limit", limit)
			_ = c.Error(errors.Join(models.ErrBadRequest, err))
			return
		}
		offset := c.DefaultQuery("offset", "0")
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			log.Logger.Warn("invalid offset query parameter", attributes.ErrorKey, err, "offset", offset)
			_ = c.Error(errors.Join(models.ErrBadRequest, err))
			return
		}

		budgets, err := control.GetBudgets(limitInt, offsetInt, []string{}, "", "")
		if err != nil {
			log.Logger.Error("get budgets failed", attributes.ErrorKey, err)
			_ = c.Error(errors.Join(models.ErrInternalServerError, err))
			return
		}
		c.JSON(http.StatusOK, budgets)
	})

	router.PUT("/budgets", func(c *gin.Context) {
		var budget models.Budget
		err := c.ShouldBindJSON(&budget)
		if err != nil {
			_ = c.Error(errors.Join(models.ErrBadRequest, err))
			return
		}
		err = control.SetBudget(budget)
		if err != nil {
			_ = c.Error(errors.Join(models.ErrInternalServerError, err))
			return
		}
	})

	router.DELETE("/budgets", func(c *gin.Context) {
		err := control.DeleteBudget(c.Query("budget_identifier"), c.Query("user_id"), c.Query("role"))
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				log.Logger.Warn("delete budget failed", attributes.ErrorKey, err)
				_ = c.Error(errors.Join(models.ErrNotFound, err))
			} else {
				log.Logger.Error("delete budget failed", attributes.ErrorKey, err)
				_ = c.Error(errors.Join(models.ErrInternalServerError, err))
			}
			return
		}
	})
}
