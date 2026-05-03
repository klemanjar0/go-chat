package app

import "github.com/gofiber/fiber/v3"

type Router struct {
	app *fiber.App
}

func NewRouter(app *fiber.App) *Router {
	return &Router{app: app}
}

func (r *Router) Register() {
	api := r.app.Group("/api")

	users := api.Group("/users")
	users.Get("/", r.getUsers)
	//users.Post("/", r.createUser)

	//chats := api.Group("/chats")
	//chats.Get("/", r.getChats)
	//chats.Post("/", r.createChat)

	//messages := api.Group("/messages")
	//messages.Post("/", r.sendMessage)

	//r.app.Get("/ws", r.ws)
}

func (r *Router) getUsers(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"users": []string{},
	})
}
