package handler

import (
	"encoding/json"
	"fmt"
	"go-graphql-user-svc/internal/model"
	"go-graphql-user-svc/internal/service"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
)

type UserHandler struct {
	Service service.IUserService
}

// NewUserHandler creates a new handler for user-related routes
func NewUserHandler(service service.IUserService) *UserHandler {
	return &UserHandler{
		Service: service,
	}
}

// GraphQL Object for User
var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"id":    &graphql.Field{Type: graphql.String},
		"name":  &graphql.Field{Type: graphql.String},
		"email": &graphql.Field{Type: graphql.String},
		"role":  &graphql.Field{Type: graphql.String},
	},
})

// ServeGraphQL handles GraphQL requests
func (h *UserHandler) ServeGraphQL(w http.ResponseWriter, r *http.Request) {

	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fields := graphql.Fields{
		"getUser": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := p.Args["id"].(string)
				return h.Service.GetUserByID(p.Context, id)
			},
		},
		"users": &graphql.Field{
			Type: graphql.NewList(userType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return h.Service.GetAllUser(p.Context), nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	mutationFields := graphql.Fields{
		"createUser": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"name":     &graphql.ArgumentConfig{Type: graphql.String},
				"email":    &graphql.ArgumentConfig{Type: graphql.String},
				"role":     &graphql.ArgumentConfig{Type: graphql.String},
				"password": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				name := p.Args["name"].(string)
				email := p.Args["email"].(string)
				role := p.Args["role"].(string)
				password := p.Args["password"].(string)

				user := model.User{Name: name, Email: email, Role: role, Password: password}
				return h.Service.CreateUser(p.Context, user)
			},
		},
		"updateUser": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"id":       &graphql.ArgumentConfig{Type: graphql.String},
				"name":     &graphql.ArgumentConfig{Type: graphql.String},
				"email":    &graphql.ArgumentConfig{Type: graphql.String},
				"role":     &graphql.ArgumentConfig{Type: graphql.String},
				"password": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := p.Args["id"].(string)
				name := p.Args["name"].(string)
				email := p.Args["email"].(string)
				role := p.Args["role"].(string)
				password := p.Args["password"].(string)

				user := model.User{Name: name, Email: email, Role: role, Password: password}
				return h.Service.UpdateUser(p.Context, id, user)
			},
		},
		"deleteUser": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id := p.Args["id"].(string)
				err := h.Service.DeleteUser(p.Context, id)
				if err != nil {
					return false, err
				}
				return true, nil
			},
		},
	}
	rootMutation := graphql.ObjectConfig{Name: "RootMutation", Fields: mutationFields}
	schemaConfig := graphql.SchemaConfig{
		Query:    graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(rootMutation),
	}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: params["query"].(string),
	})

	if len(result.Errors) > 0 {
		http.Error(w, fmt.Sprintf("Failed to execute GraphQL operation: %v", result.Errors), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
