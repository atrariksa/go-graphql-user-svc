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

type AuthHandler struct {
	Service service.IUserService
}

func NewAuthHandler(svc service.IUserService) *AuthHandler {
	return &AuthHandler{
		Service: svc,
	}
}

var loginType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Login",
	Fields: graphql.Fields{
		"token": &graphql.Field{Type: graphql.String},
	},
})

func (h *AuthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fields := graphql.Fields{
		"login": &graphql.Field{
			Type: loginType,
			Args: graphql.FieldConfigArgument{
				"email":    &graphql.ArgumentConfig{Type: graphql.String},
				"password": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				email := p.Args["email"].(string)
				password := p.Args["password"].(string)
				user := model.User{
					Email:    email,
					Password: password,
				}

				token, err := h.Service.Login(p.Context, user)
				if err != nil {
					return "", err
				}

				resp := make(map[string]interface{})
				resp["token"] = token
				return resp, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	mutationFields := graphql.Fields{
		"register": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"name":  &graphql.ArgumentConfig{Type: graphql.String},
				"email": &graphql.ArgumentConfig{Type: graphql.String},
				"role":  &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				name := p.Args["name"].(string)
				email := p.Args["email"].(string)
				role := p.Args["role"].(string)

				user := model.User{Name: name, Email: email, Role: role}
				return h.Service.CreateUser(p.Context, user)
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
