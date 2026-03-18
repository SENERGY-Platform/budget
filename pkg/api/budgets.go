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
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/controller"
	"github.com/SENERGY-Platform/budget/pkg/log"
	"github.com/SENERGY-Platform/budget/pkg/models"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/julienschmidt/httprouter"
)

func init() {
	endpoints = append(endpoints, BudgetEndpoints)
}

func BudgetEndpoints(router *httprouter.Router, _ configuration.Config, control *controller.Controller) {
	router.GET("/budgets", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		limit := request.URL.Query().Get("limit")
		if limit == "" {
			limit = "100"
		}
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			log.Logger.Warn("invalid limit query parameter", attributes.ErrorKey, err, "limit", limit)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		offset := request.URL.Query().Get("offset")
		if offset == "" {
			offset = "0"
		}
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			log.Logger.Warn("invalid offset query parameter", attributes.ErrorKey, err, "offset", offset)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		budgets, err := control.GetBudgets(limitInt, offsetInt, []string{}, "", "")
		if err != nil {
			log.Logger.Error("get budgets failed", attributes.ErrorKey, err)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(budgets)
		if err != nil {
			log.Logger.Error("unable to encode response", attributes.ErrorKey, err)
		}
	})

	router.PUT("/budgets", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var budget models.Budget
		err := json.NewDecoder(request.Body).Decode(&budget)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err = control.SetBudget(budget)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.DELETE("/budgets", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		err := control.DeleteBudget(request.URL.Query().Get("budget_identifier"), request.URL.Query().Get("user_id"), request.URL.Query().Get("role"))
		if err != nil {
			if errors.Is(err, models.ErrorNotFound) {
				log.Logger.Warn("delete budget failed", attributes.ErrorKey, err)
				http.Error(writer, err.Error(), http.StatusNotFound)
			} else {
				log.Logger.Error("delete budget failed", attributes.ErrorKey, err)
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	})
}
