package organizations

import (
	"encoding/json"
	"net/http"
	"uuid"

	"github.com/duddy57/sperium/internal/plataform/json_utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  organizationService
	validate *validator.Validate
}

func NewOrganizationHandler(service organizationService, validate *validator.Validate) *Handler {
	return &Handler{service, validate}
}

var _ organizationService = (*Service)(nil)

func (h *Handler) RegisterRoutes(r chi.Router, auth func(http.Handler) http.Handler) {
	r.Route("/", func(r chi.Router) {
		r.Use(auth)
		r.Post("/create", h.create)
		r.Get("/details/{id}", h.get)
		r.Put("/update/{id}", h.update)
		r.Delete("/delete/{id}", h.delete)
		r.Get("/my-org", h.getDefault)
	})

}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrganizationReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		json_utils.WriteError(w, http.StatusUnprocessableEntity, "invalid json body")
		return
	}

	id, slug, err := h.service.Create(r.Context(), req)
	if err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}

	json_utils.WriteJSON(w, http.StatusCreated, createOrganizationResp{
		Message: "Organization created successfully!",
		ID:      id,
		Slug:    slug,
	})
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	orgId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid organization id")
	}

	organization, err := h.service.Details(r.Context(), orgId)
	if err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
	}

	json_utils.WriteJSON(w, http.StatusOK, organization)
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	orgId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid organization id")
	}

	var req UpdateOrganizationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		json_utils.WriteError(w, http.StatusUnprocessableEntity, "invalid json body")
		return
	}

	if err := h.service.Update(r.Context(), orgId, req); err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}

	json_utils.WriteJSON(w, http.StatusCreated, genericResp{
		Message: "Organization created successfully!",
	})

}
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	orgId, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid organization id")
	}

	if err := h.service.Delete(r.Context(), orgId); err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}

	json_utils.WriteJSON(w, http.StatusCreated, genericResp{
		Message: "Organization deleted successfully!",
	})
}
func (h *Handler) getDefault(w http.ResponseWriter, r *http.Request) {
	org, err := h.service.GetDefault(r.Context())
	if err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}

	json_utils.WriteJSON(w, http.StatusOK, GetOrganizationReq{
		Message:      "Organization created successfully!",
		Organization: org,
	})
}
