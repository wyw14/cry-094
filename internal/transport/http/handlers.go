package http

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw14/cry-094/internal/application/analysisservice"
	"github.com/wyw14/cry-094/internal/application/authservice"
	"github.com/wyw14/cry-094/internal/application/precheckservice"
	"github.com/wyw14/cry-094/internal/application/teamservice"
	"github.com/wyw14/cry-094/internal/application/upload"
	"github.com/wyw14/cry-094/internal/domain/host"
)

type Handler struct {
	teams    *teamservice.Service
	uploads  *upload.Service
	analyses *analysisservice.Service
	plans    *precheckservice.Service
	auth     *authservice.Service
	validate *validator.Validate
}

func NewHandler(teams *teamservice.Service, uploads *upload.Service, analyses *analysisservice.Service, plans *precheckservice.Service, auth *authservice.Service) *Handler {
	return &Handler{teams: teams, uploads: uploads, analyses: analyses, plans: plans, auth: auth, validate: validator.New()}
}

type createTeamRequest struct {
	Name string `json:"name" validate:"required,min=3"`
}

func (h *Handler) CreateTeam(c *gin.Context) {
	var req createTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, 400, "INVALID_JSON", "request body is invalid", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(c, 400, "VALIDATION_ERROR", "request fields are invalid", validationFields(err))
		return
	}
	value, err := h.teams.Create(c.Request.Context(), req.Name, actorID(c), "http")
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": value})
}

type createLibraryRequest struct {
	TeamID string            `json:"team_id" validate:"required"`
	Name   string            `json:"name" validate:"required,min=2"`
	Tags   map[string]string `json:"tags"`
}

func (h *Handler) CreateLibrary(c *gin.Context) {
	var req createLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, 400, "INVALID_JSON", "request body is invalid", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(c, 400, "VALIDATION_ERROR", "request fields are invalid", validationFields(err))
		return
	}
	value, err := h.uploads.CreateLibrary(c.Request.Context(), req.TeamID, actorID(c), req.Name, req.Tags)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": value})
}

type uploadRequest struct {
	TeamID    string `json:"team_id" validate:"required"`
	LibraryID string `json:"library_id" validate:"required"`
	Filename  string `json:"filename" validate:"required"`
	MIME      string `json:"mime" validate:"required"`
	Content   string `json:"content" validate:"required"`
	Version   int    `json:"version" validate:"min=1"`
}

func (h *Handler) Upload(c *gin.Context) {
	var req uploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, 400, "INVALID_JSON", "request body is invalid", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(c, 400, "VALIDATION_ERROR", "request fields are invalid", validationFields(err))
		return
	}
	value, err := h.uploads.Upload(c.Request.Context(), upload.Request{TeamID: req.TeamID, LibraryID: req.LibraryID, ActorID: actorID(c), Filename: req.Filename, MIME: req.MIME, Content: []byte(req.Content), Version: req.Version, IdempotencyKey: c.GetHeader("Idempotency-Key")})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": value})
}

type analysisRequest struct {
	TeamID      string         `json:"team_id" validate:"required"`
	LibraryID   string         `json:"library_id" validate:"required"`
	ArtifactIDs []string       `json:"artifact_ids" validate:"required,min=1"`
	Inventory   host.Inventory `json:"inventory"`
}

func (h *Handler) Analyze(c *gin.Context) {
	var req analysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, 400, "INVALID_JSON", "request body is invalid", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(c, 400, "VALIDATION_ERROR", "request fields are invalid", validationFields(err))
		return
	}
	value, err := h.analyses.Analyze(c.Request.Context(), req.TeamID, actorID(c), req.LibraryID, req.ArtifactIDs, req.Inventory)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": value})
}

type planRequest struct {
	AnalysisID string `json:"analysis_id" validate:"required"`
	TeamID     string `json:"team_id" validate:"required"`
}

func (h *Handler) GeneratePlan(c *gin.Context) {
	var req planRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, 400, "INVALID_JSON", "request body is invalid", nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeError(c, 400, "VALIDATION_ERROR", "request fields are invalid", validationFields(err))
		return
	}
	value, err := h.plans.Generate(c.Request.Context(), req.AnalysisID, req.TeamID, actorID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": value})
}
func (h *Handler) SubmitPlan(c *gin.Context) {
	id := c.Param("id")
	if err := h.plans.Submit(c.Request.Context(), id, actorID(c)); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) ReviewPlan(c *gin.Context) {
	approve := c.Query("decision") == "approve"
	if err := h.plans.Review(c.Request.Context(), c.Param("id"), actorID(c), c.Query("reason"), approve); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) SignPlan(c *gin.Context) {
	if err := h.plans.Sign(c.Request.Context(), c.Param("id"), actorID(c)); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) ListArtifacts(c *gin.Context) {
	libraryID := c.Query("library_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, next, err := h.uploads.List(c.Request.Context(), c.Query("team_id"), actorID(c), libraryID, c.Query("cursor"), limit)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"items": items, "next_cursor": next, "library_id": libraryID, "limit": limit}})
}

func (h *Handler) DownloadArtifact(c *gin.Context) {
	content, filename, err := h.uploads.Download(c.Request.Context(), c.Query("team_id"), actorID(c), c.Param("id"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+filepath.Base(filename)+`"`)
	c.Data(http.StatusOK, "text/plain; charset=utf-8", content)
}
func actorID(c *gin.Context) string { return c.GetString("actor_id") }
