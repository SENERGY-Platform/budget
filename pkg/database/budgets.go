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

package database

import (
	"context"
	"github.com/SENERGY-Platform/budget/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const budgetIdentifierFieldName = "BudgetIdentifier"
const roleFieldName = "Role"
const userIdFieldName = "UserId"

var budgetIdentifierKey string
var roleKey string
var userIdKey string

func init() {
	var err error
	budgetIdentifierKey, err = getBsonFieldName(models.Budget{}, budgetIdentifierFieldName)
	if err != nil {
		log.Fatal(err)
	}
	roleKey, err = getBsonFieldName(models.Budget{}, roleFieldName)
	if err != nil {
		log.Fatal(err)
	}
	userIdKey, err = getBsonFieldName(models.Budget{}, userIdFieldName)
	if err != nil {
		log.Fatal(err)
	}

	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoBudgetCollection)
		err = db.ensureCompoundIndex(collection, "identifierRoleUser", true, true, budgetIdentifierKey, roleKey, userIdKey)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) budgetCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoBudgetCollection)
}

// Search Logic:
// if userId is set, only budgets with either the set userId or empty userId are returned. If userId is not set, budgets will not be filtered by userId
// if roles is not empty, only budgets with one of provided roles or empty role are returned. If roles is empty, budgets will not be filtered by role
// if budgetIdentifier is set, only budgets with the excact budgetIdentifier are returned. If budgetIdentifier is not set, budgets will not be filtered by budgetIdentifier
func (this *Mongo) ListBudgets(limit int, offset int, budgetIdentifier string, userId string, roles []string) (result []models.Budget, err error) {
	opt := options.Find()
	opt.SetLimit(int64(limit))
	opt.SetSkip(int64(offset))

	filter := bson.M{}
	if userId != "" {
		filter[userIdKey] = bson.M{"$in": []string{"", userId}}
	}
	if budgetIdentifier != "" {
		filter[budgetIdentifierKey] = budgetIdentifier
	}
	if len(roles) > 0 {
		filter[roleKey] = bson.M{"$in": append([]string{""}, roles...)}
	}
	ctx, cancel := context.WithTimeout(this.ctx, this.timeout)
	defer cancel()
	cursor, err := this.budgetCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	result = []models.Budget{}
	for cursor.Next(ctx) {
		instance := models.Budget{}
		err = cursor.Decode(&instance)
		if err != nil {
			return nil, err
		}
		result = append(result, instance)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) SetBudget(budget models.Budget) error {
	ctx, cancel := context.WithTimeout(this.ctx, this.timeout)
	defer cancel()
	_, err := this.budgetCollection().ReplaceOne(ctx, bson.M{budgetIdentifierKey: budget.BudgetIdentifier, userIdKey: budget.UserId, roleKey: budget.Role}, budget, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveBudget(budgetIdentifier string, userId string, role string) error {
	ctx, cancel := context.WithTimeout(this.ctx, this.timeout)
	defer cancel()
	result, err := this.budgetCollection().DeleteOne(ctx, bson.M{budgetIdentifierKey: budgetIdentifier, userIdKey: userId, roleKey: role})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return models.ErrorNotFound
	}
	return nil
}
