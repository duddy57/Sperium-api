package users

import (
	"encoding/json"
	"net/http"

	"github.com/duddy57/sperium/internal/plataform/json_utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  userService
	validate *validator.Validate
}

func NewUserHandler(service userService, validate *validator.Validate) *Handler {
	return &Handler{
		service:  service,
		validate: validate,
	}
}

var _ userService = (*Service)(nil)

func (h *Handler) RegisterRoutes(r chi.Router, auth func(http.Handler) http.Handler) {
	r.Post("/create", h.create)
	r.Post("/login", h.login)

	r.Route("/", func(r chi.Router) {
		r.Use(auth)
		r.Get("/me", h.details)
		r.Get("/logout", h.logout)
		r.Delete("/delete", h.delete)
		r.Put("/update", h.update)
	})

}
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateUserReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		json_utils.WriteError(w, http.StatusUnprocessableEntity, "invalid json body")
		return
	}

	id, err := h.service.Create(r.Context(), req)
	if err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}

	json_utils.WriteJSON(w, http.StatusCreated, createUserResp{
		Message: "Welcome to sperium!",
		ID:      id,
	})
}
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginUserReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.service.Login(r.Context(), req); err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}

	json_utils.WriteJSON(w, http.StatusCreated, genericResp{
		Message: "Welcome to sperium!",
	})
}
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Logout(r.Context()); err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	json_utils.WriteJSON(w, http.StatusCreated, genericResp{
		Message: "Bye bye!",
	})
}
func (h *Handler) details(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.Details(r.Context())
	if err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	json_utils.WriteJSON(w, http.StatusOK, Users{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,

		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context()); err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	json_utils.WriteJSON(w, http.StatusOK, genericResp{
		Message: "Bye bye!",
	})
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var req UpdateUserReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		json_utils.WriteError(w, http.StatusUnprocessableEntity, "invalid json body")
		return
	}

	if err := h.service.Update(r.Context(), req); err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}

	json_utils.WriteJSON(w, http.StatusOK, genericResp{
		Message: "Update success!",
	})
}
