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
	"github.com/SENERGY-Platform/budget/pkg/configuration"
	"github.com/SENERGY-Platform/budget/pkg/database"
	importdeployapi "github.com/SENERGY-Platform/import-deploy/lib/api"
	importdeployclient "github.com/SENERGY-Platform/import-deploy/lib/client"
	importrepoapi "github.com/SENERGY-Platform/import-repository/lib/api"
	importrepoclient "github.com/SENERGY-Platform/import-repository/lib/client"
)

type Controller struct {
	config       configuration.Config
	db           *database.Mongo
	importDeploy importdeployapi.Controller
	importRepo   importrepoapi.Controller
}

func New(config configuration.Config, db *database.Mongo) *Controller {
	return &Controller{config: config, db: db, importDeploy: importdeployclient.NewClient(config.ImportDeployUrl), importRepo: importrepoclient.NewClient(config.ImportRepoUrl)}
}
