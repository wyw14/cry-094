package bootstrap

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-094/internal/adapter/local"
	"github.com/wyw14/cry-094/internal/application/analysisservice"
	"github.com/wyw14/cry-094/internal/application/authservice"
	"github.com/wyw14/cry-094/internal/application/precheckservice"
	"github.com/wyw14/cry-094/internal/application/teamservice"
	"github.com/wyw14/cry-094/internal/application/upload"
	"github.com/wyw14/cry-094/internal/domain/identity"
	"github.com/wyw14/cry-094/internal/middleware"
	"github.com/wyw14/cry-094/internal/parser"
	"github.com/wyw14/cry-094/internal/platform/clock"
	"github.com/wyw14/cry-094/internal/platform/files"
	"github.com/wyw14/cry-094/internal/platform/ids"
	"github.com/wyw14/cry-094/internal/repository/memory"
	httptransport "github.com/wyw14/cry-094/internal/transport/http"
)

type Container struct {
	Router     *gin.Engine
	Teams      *memory.TeamRepo
	Artifacts  *memory.ArtifactRepo
	Analyses   *memory.AnalysisRepo
	Plans      *memory.PlanRepo
	Audits     *memory.AuditRepo
	Identities *memory.IdentityRepo
}

func Build(cfg Config) (*Container, error) {
	teams := memory.NewTeamRepo()
	artifacts := memory.NewArtifactRepo()
	analyses := memory.NewAnalysisRepo()
	plans := memory.NewPlanRepo()
	audits := memory.NewAuditRepo()
	objects := files.New()
	systemClock := clock.System{}
	generator := ids.UUID{}
	demoUser, err := identity.Register("11111111-1111-4111-8111-111111111111", "admin@example.test", "scriptscope-demo", systemClock.Now())
	if err != nil {
		return nil, fmt.Errorf("seed demo identity: %w", err)
	}
	identities := memory.NewIdentityRepo(demoUser)
	signer, err := local.NewSigner(cfg.SigningKeyID, []byte(cfg.SigningSecret))
	if err != nil {
		return nil, fmt.Errorf("configure signer: %w", err)
	}
	teamApp := teamservice.New(teams, audits, systemClock, generator)
	uploadApp := upload.New(teams, artifacts, objects, systemClock, generator)
	analysisApp := analysisservice.New(teams, artifacts, analyses, objects, parser.Factory{BuildID: cfg.ParserBuild}, systemClock, generator)
	planApp := precheckservice.New(analyses, plans, teams, signer, systemClock, generator)
	authApp := authservice.New(identities, systemClock, generator, []byte(cfg.AuthSecret), 15*time.Minute, 7*24*time.Hour)
	handler := httptransport.NewHandler(teamApp, uploadApp, analysisApp, planApp, authApp)
	return &Container{Router: httptransport.NewRouter(handler, middleware.NewAuth([]byte(cfg.AuthSecret)).Require()), Teams: teams, Artifacts: artifacts, Analyses: analyses, Plans: plans, Audits: audits, Identities: identities}, nil
}
