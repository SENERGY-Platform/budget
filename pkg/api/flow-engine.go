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

	"github.com/SENERGY-Platform/budget/pkg/api/util"
	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/controller"
	"github.com/SENERGY-Platform/budget/pkg/log"
	"github.com/SENERGY-Platform/budget/pkg/models"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/gin-gonic/gin"
)

func init() {
	endpoints = append(endpoints, CheckFlowEngineEndpoints)
}

func CheckFlowEngineEndpoints(router *gin.Engine, conf configuration.Config, control *controller.Controller) {
	router.POST("/check/analytics/flow-engine", func(c *gin.Context) {
		parsed, err := util.ParseRequest(c.Request.Body, conf.Debug)
		if err != nil {
			log.Logger.Warn("parse flow-engine request failed", attributes.ErrorKey, err)
			_ = c.Error(errors.Join(models.ErrBadRequest, err))
			return
		}
		code, err := control.CheckFlowEngine(parsed)
		if err != nil {
			if code >= http.StatusInternalServerError {
				log.Logger.Error("flow-engine check failed", attributes.ErrorKey, err, "status_code", code)
			} else {
				log.Logger.Warn("flow-engine check rejected", attributes.ErrorKey, err, "status_code", code)
			}
			_ = c.Error(errors.Join(models.GetError(code), err))
		}
	})
}
