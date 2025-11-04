package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/parthpati1102/todo-gin-jwt/config"
	"github.com/parthpati1102/todo-gin-jwt/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ListTodosPage(c *gin.Context) {
	// identity set by middleware
	identity := c.GetString("identity") // our middleware will set this to the email
	coll := config.MongoClient.Database(config.DBName).Collection("todos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := coll.Find(ctx, bson.M{"user_email": identity})
	if err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}
	var todos []models.Todo
	if err := cursor.All(ctx, &todos); err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}

	c.HTML(http.StatusOK, "lists.html", gin.H{
		"todos": todos,
		"email": identity,
	})
}

func CreateTodoPage(c *gin.Context) {
	c.HTML(http.StatusOK, "create.html", gin.H{"error": ""})
}

func CreateTodoHandler(c *gin.Context) {
	type req struct {
		Title   string `form:"title" binding:"required"`
		Content string `form:"content"`
	}
	var r req
	if err := c.ShouldBind(&r); err != nil {
		c.HTML(http.StatusBadRequest, "create.html", gin.H{"error": "Title is required"})
		return
	}
	identity := c.GetString("identity")
	coll := config.MongoClient.Database(config.DBName).Collection("todos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	todo := models.Todo{
		UserEmail: identity,
		Title:     r.Title,
		Content:   r.Content,
		Done:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := coll.InsertOne(ctx, todo)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "create.html", gin.H{"error": "DB error"})
		return
	}
	c.Redirect(http.StatusSeeOther, "/lists")
}

func EditTodoPage(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	coll := config.MongoClient.Database(config.DBName).Collection("todos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var todo models.Todo
	if err := coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&todo); err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}
	c.HTML(http.StatusOK, "edit.html", gin.H{"todo": todo})
}

func UpdateTodoHandler(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	type req struct {
		Title   string `form:"title" binding:"required"`
		Content string `form:"content"`
		Done    string `form:"done"`
	}
	var r req
	if err := c.ShouldBind(&r); err != nil {
		c.String(http.StatusBadRequest, "Invalid input")
		return
	}
	done := false
	if r.Done == "on" {
		done = true
	}
	coll := config.MongoClient.Database(config.DBName).Collection("todos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{
		"title":      r.Title,
		"content":    r.Content,
		"done":       done,
		"updated_at": time.Now(),
	}})
	if err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}
	c.Redirect(http.StatusSeeOther, "/lists")
}

func DeleteTodoHandler(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID")
		return
	}
	coll := config.MongoClient.Database(config.DBName).Collection("todos")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = coll.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		c.String(http.StatusInternalServerError, "DB error")
		return
	}
	c.Redirect(http.StatusSeeOther, "/lists")
}
