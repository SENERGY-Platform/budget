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
	"context"
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

// BudgetEndpoints registers budget management routes.
func BudgetEndpoints(router *gin.Engine, _ configuration.Config, control *controller.Controller) {
	router.GET("/budgets", getBudgetsHandler(control))
	router.PUT("/budgets", setBudgetHandler(control))
	router.DELETE("/budgets", deleteBudgetHandler(control))
}

// getBudgetsHandler godoc
// @Summary List budgets
// @Description Returns a paginated list of configured budgets.
// @Tags budgets
// @Produce json
// @Param limit query int false "Maximum number of items to return" default(100)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {array} models.Budget
// @Failure 400 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Router /budgets [get]
func getBudgetsHandler(control *controller.Controller) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := requestContext(c)
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

		budgets, err := control.GetBudgets(ctx, limitInt, offsetInt, []string{}, "", "")
		if err != nil {
			log.Logger.Error("get budgets failed", attributes.ErrorKey, err)
			_ = c.Error(errors.Join(models.ErrInternalServerError, err))
			return
		}
		c.JSON(http.StatusOK, budgets)
	}
}

// setBudgetHandler godoc
// @Summary Create or update budget
// @Description Creates a budget or overwrites an existing one for the same identifier and subject.
// @Tags budgets
// @Accept json
// @Produce json
// @Param budget body models.Budget true "Budget to create or update"
// @Success 200 {string} string
// @Failure 400 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Router /budgets [put]
func setBudgetHandler(control *controller.Controller) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := requestContext(c)
		var budget models.Budget
		err := c.ShouldBindJSON(&budget)
		if err != nil {
			_ = c.Error(errors.Join(models.ErrBadRequest, err))
			return
		}
		err = control.SetBudget(ctx, budget)
		if err != nil {
			_ = c.Error(errors.Join(models.ErrInternalServerError, err))
			return
		}
	}
}

// deleteBudgetHandler godoc
// @Summary Delete budget
// @Description Deletes a budget by identifier and either user ID or role.
// @Tags budgets
// @Produce json
// @Param budget_identifier query string true "Budget identifier" Enums(flow-engine, import-deploy)
// @Param user_id query string false "User ID owning the budget"
// @Param role query string false "Role owning the budget"
// @Success 200 {string} string
// @Failure 404 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Router /budgets [delete]
func deleteBudgetHandler(control *controller.Controller) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := control.DeleteBudget(requestContext(c), c.Query("budget_identifier"), c.Query("user_id"), c.Query("role"))
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
	}
}

func requestContext(c *gin.Context) context.Context {
	if c != nil && c.Request != nil {
		return c.Request.Context()
	}
	return context.Background()
}
