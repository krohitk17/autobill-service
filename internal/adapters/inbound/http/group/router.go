package GroupAdapter

import (
	Middlewares "autobill-service/internal/adapters/inbound/http/middleware"
	JWTUtil "autobill-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

type GroupRouter struct {
	App     *fiber.App
	handler GroupHandler
	util    JWTUtil.JWTUtil
}

func CreateGroupRouter(app *fiber.App, handler GroupHandler, util JWTUtil.JWTUtil) GroupRouter {
	return GroupRouter{
		App:     app,
		handler: handler,
		util:    util,
	}
}

func (r GroupRouter) RegisterRoutes() {
	r.App.Use(Middlewares.AuthMiddleware(r.util))

	r.App.Post("/", r.handler.CreateGroupHandler).Name("createGroup")
	r.App.Get("/", r.handler.GetGroupsHandler).Name("getGroups")
	r.App.Get("/:groupId", r.handler.GetGroupHandler).Name("getGroup")
	r.App.Patch("/:groupId", r.handler.UpdateGroupHandler).Name("updateGroup")
	r.App.Delete("/:groupId", r.handler.DeleteGroupHandler).Name("deleteGroup")
	r.App.Post("/:groupId/leave", r.handler.LeaveGroupHandler).Name("leaveGroup")

	r.App.Post("/:groupId/members", r.handler.AddMemberHandler).Name("addMember")
	r.App.Patch("/:groupId/members/:userId/role", r.handler.UpdateMemberRoleHandler).Name("updateMemberRole")
	r.App.Delete("/:groupId/members/:userId", r.handler.RemoveMemberHandler).Name("removeMember")
}
