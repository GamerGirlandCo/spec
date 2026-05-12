package spec_test

import (
	"reflect"
	"strings"
	"time"

	"github.com/oaswrap/spec"
	"github.com/oaswrap/spec/openapi"
	"github.com/oaswrap/spec/option"
)

type PetstorePet struct {
	ID        int64             `json:"id" example:"10"`
	Name      string            `json:"name" example:"doggie" required:"true"`
	Category  *PetstoreCategory `json:"category"`
	PhotoURLs []string          `json:"photoUrls" required:"true" xmlName:"photoUrl" xmlWrapped:"true"`
	Tags      []PetstoreTag     `json:"tags" xmlWrapped:"true"`
	Status    string            `json:"status" description:"pet status in the store" enum:"available,pending,sold"`
}

type PetstoreCategory struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Dogs"`
}

type PetstoreTag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type PetstoreOrder struct {
	ID       int64     `json:"id" example:"10"`
	PetID    int64     `json:"petId" example:"198772"`
	Quantity int       `json:"quantity" example:"7"`
	ShipDate time.Time `json:"shipDate"`
	Status   string    `json:"status" example:"approved" description:"Order Status" enum:"placed,approved,delivered"`
	Complete bool      `json:"complete"`
}

type PetstoreUser struct {
	ID         int64  `json:"id" example:"10"`
	Username   string `json:"username" example:"theUser"`
	FirstName  string `json:"firstName" example:"John"`
	LastName   string `json:"lastName" example:"James"`
	Email      string `json:"email" example:"john@email.com"`
	Password   string `json:"password" example:"12345"`
	Phone      string `json:"phone" example:"12345"`
	UserStatus int    `json:"userStatus" example:"1" description:"User Status"`
}

type PetstoreAPIResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type PetstoreFindPetsByStatusRequest struct {
	Status string `query:"status" required:"true" description:"Status values that need to be considered for filter" default:"available" enum:"available,pending,sold"`
}

type PetstoreFindPetsByTagsRequest struct {
	Tags []string `query:"tags" required:"true" description:"Tags to filter by"`
}

type PetstorePetByIDRequest struct {
	PetID int64 `path:"petId" required:"true" description:"ID of pet to return"`
}

type PetstoreUpdatePetWithFormRequest struct {
	PetID  int64  `path:"petId" required:"true" description:"ID of pet that needs to be updated"`
	Name   string `description:"Name of pet that needs to be updated" query:"name"`
	Status string `description:"Status of pet that needs to be updated" query:"status"`
}

type PetstoreDeletePetRequest struct {
	APIKey string `header:"api_key"`
	PetID  int64  `path:"petId" required:"true" description:"Pet id to delete"`
}

type PetstoreUploadImageRequest struct {
	PetID              int64  `path:"petId" required:"true" description:"ID of pet to update"`
	AdditionalMetadata string `description:"Additional Metadata" query:"additionalMetadata"`
}

type PetstoreOrderByIDRequest struct {
	OrderID int64 `path:"orderId" required:"true" description:"ID of order that needs to be fetched"`
}

type PetstoreDeleteOrderRequest struct {
	OrderID int64 `path:"orderId" required:"true" description:"ID of the order that needs to be deleted"`
}

type PetstoreUserByNameRequest struct {
	Username string `path:"username" required:"true" description:"The name that needs to be fetched. Use user1 for testing"`
}

type PetstoreLoginRequest struct {
	Username string `query:"username" description:"The user name for login"`
	Password string `query:"password" description:"The password for login in clear text"`
}

func newPetstoreRouter(opts ...option.OpenAPIOption) spec.Generator {
	baseOpts := []option.OpenAPIOption{
		option.WithTitle("Swagger Petstore - OpenAPI 3.0"),
		option.WithVersion("1.0.27"),
		option.WithDescription(
			"This is a sample Pet Store Server based on the OpenAPI 3.0 specification.  You can find out more about\nSwagger at [https://swagger.io](https://swagger.io). In the third iteration of the pet store, we've switched to the design first approach!\nYou can now help us improve the API whether it's by making changes to the definition itself or to the code.\nThat way, with time, we can improve the API in general, and expose some of the new features in OAS3.\n\nSome useful links:\n- [The Pet Store repository](https://github.com/swagger-api/swagger-petstore)\n- [The source API definition for the Pet Store](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml)",
		),
		option.WithTermsOfService("https://swagger.io/terms/"),
		option.WithContact(openapi.Contact{Email: "apiteam@swagger.io"}),
		option.WithLicense(
			openapi.License{Name: "Apache 2.0", URL: "https://www.apache.org/licenses/LICENSE-2.0.html"},
		),
		option.WithExternalDocs("https://swagger.io", "Find out more about Swagger"),
		option.WithServer("https://petstore3.swagger.io/api/v3"),
		option.WithTags(
			openapi.Tag{
				Name:         "pet",
				Description:  "Everything about your Pets",
				ExternalDocs: &openapi.ExternalDocs{URL: "https://swagger.io", Description: "Find out more"},
			},
			openapi.Tag{
				Name:        "store",
				Description: "Access to Petstore orders",
				ExternalDocs: &openapi.ExternalDocs{
					URL:         "https://swagger.io",
					Description: "Find out more about our store",
				},
			},
			openapi.Tag{Name: "user", Description: "Operations about user"},
		),
		option.WithSecurity("petstore_auth", option.SecurityOAuth2(openapi.OAuthFlows{
			Implicit: &openapi.OAuthFlow{
				AuthorizationURL: "https://petstore3.swagger.io/oauth/authorize",
				Scopes: map[string]string{
					"write:pets": "modify pets in your account",
					"read:pets":  "read your pets",
				},
			},
		})),
		option.WithSecurity("api_key", option.SecurityAPIKey("api_key", "header")),
		option.WithReflectorConfig(
			option.InterceptDefName(func(_ reflect.Type, defaultName string) string {
				name := strings.TrimPrefix(defaultName, "SpecTest")
				return strings.TrimPrefix(name, "Petstore")
			}),
		),
		option.WithDocument(func(doc *openapi.Document) {
			if doc.Components.Schemas == nil {
				return
			}
			if s, ok := doc.Components.Schemas["Pet"]; ok {
				s.XML = &openapi.XML{Name: "pet"}
			}
			if s, ok := doc.Components.Schemas["Category"]; ok {
				s.XML = &openapi.XML{Name: "category"}
			}
			if s, ok := doc.Components.Schemas["Tag"]; ok {
				s.XML = &openapi.XML{Name: "tag"}
			}
			if s, ok := doc.Components.Schemas["Order"]; ok {
				s.XML = &openapi.XML{Name: "order"}
			}
			if s, ok := doc.Components.Schemas["User"]; ok {
				s.XML = &openapi.XML{Name: "user"}
			}
		}),
	}
	r := spec.NewRouter(append(baseOpts, opts...)...)

	// Pet
	pet := r.Group("/pet", option.GroupTags("pet"))

	pet.Put("/",
		option.OperationID("updatePet"),
		option.Summary("Update an existing pet."),
		option.Description("Update an existing pet by Id."),
		option.Security("petstore_auth", "write:pets", "read:pets"),
		option.Request(new(PetstorePet), option.ContentDescription("Update an existent pet in the store")),
		option.Response(200, new(PetstorePet), option.ContentDescription("Successful operation")),
		option.Response(400, nil, option.ContentDescription("Invalid ID supplied")),
		option.Response(404, nil, option.ContentDescription("Pet not found")),
		option.Response(422, nil, option.ContentDescription("Validation exception")),
	)

	pet.Post("/",
		option.OperationID("addPet"),
		option.Summary("Add a new pet to the store."),
		option.Description("Add a new pet to the store."),
		option.Security("petstore_auth", "write:pets", "read:pets"),
		option.Request(new(PetstorePet), option.ContentDescription("Create a new pet in the store")),
		option.Response(200, new(PetstorePet), option.ContentDescription("Successful operation")),
		option.Response(405, nil, option.ContentDescription("Invalid input")),
		option.Response(422, nil, option.ContentDescription("Validation exception")),
	)

	pet.Get("/findByStatus",
		option.OperationID("findPetsByStatus"),
		option.Summary("Finds Pets by status."),
		option.Description("Multiple status values can be provided with comma separated strings."),
		option.Security("petstore_auth", "write:pets", "read:pets"),
		option.Request(new(PetstoreFindPetsByStatusRequest)),
		option.Response(200, new([]PetstorePet), option.ContentDescription("successful operation")),
		option.Response(400, nil, option.ContentDescription("Invalid status value")),
	)

	pet.Get(
		"/findByTags",
		option.OperationID("findPetsByTags"),
		option.Summary("Finds Pets by tags."),
		option.Description(
			"Multiple tags can be provided with comma separated strings. Use tag1, tag2, tag3 for testing.",
		),
		option.Security("petstore_auth", "write:pets", "read:pets"),
		option.Request(new(PetstoreFindPetsByTagsRequest)),
		option.Response(200, new([]PetstorePet), option.ContentDescription("successful operation")),
		option.Response(400, nil, option.ContentDescription("Invalid tag value")),
	)

	pet.Get("/{petId}",
		option.OperationID("getPetById"),
		option.Summary("Find pet by ID."),
		option.Description("Returns a single pet."),
		option.Security("api_key"),
		option.Security("petstore_auth", "write:pets", "read:pets"),
		option.Request(new(PetstorePetByIDRequest)),
		option.Response(200, new(PetstorePet), option.ContentDescription("successful operation")),
		option.Response(400, nil, option.ContentDescription("Invalid ID supplied")),
		option.Response(404, nil, option.ContentDescription("Pet not found")),
	)

	pet.Post("/{petId}",
		option.OperationID("updatePetWithForm"),
		option.Summary("Updates a pet in the store with form data."),
		option.Description("Updates a pet resource based on the form data."),
		option.Security("petstore_auth", "write:pets", "read:pets"),
		option.Request(new(PetstoreUpdatePetWithFormRequest)),
		option.Response(200, new(PetstorePet), option.ContentDescription("successful operation")),
		option.Response(405, nil, option.ContentDescription("Invalid input")),
	)

	pet.Delete("/{petId}",
		option.OperationID("deletePet"),
		option.Summary("Deletes a pet."),
		option.Description("Delete a pet."),
		option.Security("petstore_auth", "write:pets", "read:pets"),
		option.Request(new(PetstoreDeletePetRequest)),
		option.Response(200, nil, option.ContentDescription("Pet deleted")),
		option.Response(400, nil, option.ContentDescription("Invalid pet value")),
	)

	pet.Post("/{petId}/uploadImage",
		option.OperationID("uploadFile"),
		option.Summary("Uploads an image."),
		option.Description("Upload image of the pet."),
		option.Security("petstore_auth", "write:pets", "read:pets"),
		option.Request(new(PetstoreUploadImageRequest)),
		option.Request(nil, option.ContentType("application/octet-stream"), option.ContentFormat("binary")),
		option.Response(200, new(PetstoreAPIResponse), option.ContentDescription("successful operation")),
	)

	// Store
	store := r.Group("/store", option.GroupTags("store"))

	store.Get("/inventory",
		option.OperationID("getInventory"),
		option.Summary("Returns pet inventories by status."),
		option.Description("Returns a map of status codes to quantities."),
		option.Security("api_key"),
		option.Response(200, new(map[string]int32), option.ContentDescription("successful operation")),
	)

	store.Post("/order",
		option.OperationID("placeOrder"),
		option.Summary("Place an order for a pet."),
		option.Description("Place a new order in the store."),
		option.Request(new(PetstoreOrder)),
		option.Response(200, new(PetstoreOrder), option.ContentDescription("successful operation")),
		option.Response(400, nil, option.ContentDescription("Invalid input")),
		option.Response(422, nil, option.ContentDescription("Validation exception")),
	)

	store.Get(
		"/order/{orderId}",
		option.OperationID("getOrderById"),
		option.Summary("Find purchase order by ID."),
		option.Description(
			"For valid response try integer IDs with value <= 5 or > 10. Other values will generate exceptions.",
		),
		option.Request(new(PetstoreOrderByIDRequest)),
		option.Response(200, new(PetstoreOrder), option.ContentDescription("successful operation")),
		option.Response(400, nil, option.ContentDescription("Invalid ID supplied")),
		option.Response(404, nil, option.ContentDescription("Order not found")),
	)

	store.Delete(
		"/order/{orderId}",
		option.OperationID("deleteOrder"),
		option.Summary("Delete purchase order by identifier."),
		option.Description(
			"For valid response try integer IDs with value < 1000. Anything above 1000 or non-integers will generate API errors.",
		),
		option.Request(new(PetstoreDeleteOrderRequest)),
		option.Response(200, nil, option.ContentDescription("order deleted")),
		option.Response(400, nil, option.ContentDescription("Invalid ID supplied")),
		option.Response(404, nil, option.ContentDescription("Order not found")),
	)

	// User
	user := r.Group("/user", option.GroupTags("user"))

	user.Post("/",
		option.OperationID("createUser"),
		option.Summary("Create user."),
		option.Description("This can only be done by the logged in user."),
		option.Request(new(PetstoreUser), option.ContentDescription("Created user object")),
		option.Response(200, new(PetstoreUser), option.ContentDescription("successful operation")),
	)

	user.Post("/createWithList",
		option.OperationID("createUsersWithListInput"),
		option.Summary("Creates list of users with given input array."),
		option.Description("Creates list of users with given input array."),
		option.Request(new([]PetstoreUser)),
		option.Response(200, new(PetstoreUser), option.ContentDescription("Successful operation")),
	)

	user.Get("/login",
		option.OperationID("loginUser"),
		option.Summary("Logs user into the system."),
		option.Description("Log into the system."),
		option.Request(new(PetstoreLoginRequest)),
		option.Response(200, "string", option.ContentDescription("successful operation")),
		option.Response(400, nil, option.ContentDescription("Invalid username/password supplied")),
		option.CustomizeOperation(func(op *openapi.Operation) {
			resp := op.Responses["200"]
			if resp.Headers == nil {
				resp.Headers = map[string]*openapi.Header{}
			}
			resp.Headers["X-Rate-Limit"] = &openapi.Header{
				Description: "calls per hour allowed by the user",
				Schema:      &openapi.Schema{Type: "integer", Format: "int32"},
			}
			resp.Headers["X-Expires-After"] = &openapi.Header{
				Description: "date in UTC when token expires",
				Schema:      &openapi.Schema{Type: "string", Format: "date-time"},
			}
		}),
	)

	user.Get("/logout",
		option.OperationID("logoutUser"),
		option.Summary("Logs out current logged in user session."),
		option.Description("Log user out of the system."),
		option.Response(200, nil, option.ContentDescription("successful operation")),
	)

	user.Get("/{username}",
		option.OperationID("getUserByName"),
		option.Summary("Get user by user name."),
		option.Description("Get user detail based on username."),
		option.Request(new(PetstoreUserByNameRequest)),
		option.Response(200, new(PetstoreUser), option.ContentDescription("successful operation")),
		option.Response(400, nil, option.ContentDescription("Invalid username supplied")),
		option.Response(404, nil, option.ContentDescription("User not found")),
	)

	user.Put("/{username}",
		option.OperationID("updateUser"),
		option.Summary("Update user resource."),
		option.Description("This can only be done by the logged in user."),
		option.Request(new(PetstoreUserByNameRequest)),
		option.Request(new(PetstoreUser), option.ContentDescription("Update an existent user in the store")),
		option.Response(200, nil, option.ContentDescription("successful operation")),
		option.Response(400, nil, option.ContentDescription("bad request")),
		option.Response(404, nil, option.ContentDescription("user not found")),
	)

	user.Delete("/{username}",
		option.OperationID("deleteUser"),
		option.Summary("Delete user resource."),
		option.Description("This can only be done by the logged in user."),
		option.Request(new(PetstoreUserByNameRequest)),
		option.Response(200, nil, option.ContentDescription("User deleted")),
		option.Response(400, nil, option.ContentDescription("Invalid username supplied")),
		option.Response(404, nil, option.ContentDescription("User not found")),
	)

	return r
}
