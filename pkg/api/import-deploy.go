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
	endpoints = append(endpoints, CheckImportDeployEndpoints)
}

func CheckImportDeployEndpoints(router *gin.Engine, conf configuration.Config, control *controller.Controller) {
	router.POST("/check/import/deploy", checkImportDeployHandler(conf, control))
}

// checkImportDeployHandler godoc
// @Summary Check import-deploy budget
// @Description Validates whether the requested import-deploy operation is within the caller budget.
// @Tags checks
// @Accept json
// @Produce json
// @Param request body object true "Original proxied request payload"
// @Success 200 {string} string
// @Failure 400 {string} ErrorResponse
// @Failure 402 {string} ErrorResponse
// @Failure 403 {string} ErrorResponse
// @Failure 404 {string} ErrorResponse
// @Failure 500 {string} ErrorResponse
// @Router /check/import/deploy [post]
func checkImportDeployHandler(conf configuration.Config, control *controller.Controller) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := requestContext(c)
		parsed, err := util.ParseRequest(c.Request.Body, conf.Debug)
		if err != nil {
			log.Logger.Warn("parse import-deploy request failed", attributes.ErrorKey, err)
			_ = c.Error(errors.Join(models.ErrBadRequest, err))
			return
		}
		code, err := control.CheckImportDeploy(ctx, parsed)
		if err != nil {
			if code >= http.StatusInternalServerError {
				log.Logger.Error("import-deploy check failed", attributes.ErrorKey, err, "status_code", code)
			} else {
				log.Logger.Warn("import-deploy check rejected", attributes.ErrorKey, err, "status_code", code)
			}
			_ = c.Error(errors.Join(models.GetError(code), err))
		}
	}
}
