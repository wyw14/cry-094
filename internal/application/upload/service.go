package upload

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/wyw14/cry-094/internal/application/ports"
	"github.com/wyw14/cry-094/internal/domain/script"
	"github.com/wyw14/cry-094/internal/domain/team"
	"github.com/wyw14/cry-094/internal/platform/files"
)

type IDGenerator interface{ NewID() string }

type Request struct {
	TeamID         string
	LibraryID      string
	ActorID        string
	Filename       string
	MIME           string
	Content        []byte
	Version        int
	IdempotencyKey string
}

type Service struct {
	teams     ports.TeamRepository
	artifacts ports.ArtifactRepository
	store     ports.ObjectStore
	clock     ports.Clock
	ids       IDGenerator
}

func New(teams ports.TeamRepository, artifacts ports.ArtifactRepository, store ports.ObjectStore, clock ports.Clock, ids IDGenerator) *Service {
	return &Service{teams: teams, artifacts: artifacts, store: store, clock: clock, ids: ids}
}

func (s *Service) CreateLibrary(ctx context.Context, teamID, actorID, name string, tags map[string]string) (*script.Library, error) {
	owner, err := s.teams.Get(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if !owner.Can(actorID, team.PermissionUploadScript) {
		return nil, fmt.Errorf("upload permission required")
	}
	value := &script.Library{ID: s.ids.NewID(), TeamID: teamID, Name: strings.TrimSpace(name), Tags: tags, Version: 1, Created: s.clock.Now()}
	if value.Name == "" {
		return nil, fmt.Errorf("library name required")
	}
	if err := s.artifacts.CreateLibrary(ctx, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (s *Service) Upload(ctx context.Context, request Request) (*script.Artifact, error) {
	owner, err := s.teams.Get(ctx, request.TeamID)
	if err != nil {
		return nil, err
	}
	if !owner.Can(request.ActorID, team.PermissionUploadScript) {
		return nil, fmt.Errorf("upload permission required")
	}
	library, err := s.artifacts.GetLibrary(ctx, request.LibraryID)
	if err != nil {
		return nil, err
	}
	if library.TeamID != request.TeamID {
		return nil, fmt.Errorf("library is outside actor team")
	}
	if err := validateMIME(request.Filename, request.MIME); err != nil {
		return nil, err
	}
	encoding, normalized, err := normalizeEncoding(request.Content)
	if err != nil {
		return nil, err
	}
	shellType, err := shellFor(request.Filename)
	if err != nil {
		return nil, err
	}
	digest := files.Digest(normalized)
	id := s.ids.NewID()
	storageKey := fmt.Sprintf("teams/%s/libraries/%s/%s", request.TeamID, request.LibraryID, digest)
	value, err := script.NewArtifact(id, request.LibraryID, request.Filename, shellType, encoding, digest, storageKey, request.ActorID, int64(len(normalized)), request.Version, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.store.Put(ctx, storageKey, normalized); err != nil {
		return nil, fmt.Errorf("store immutable script: %w", err)
	}
	if err := s.artifacts.CreateArtifact(ctx, value); err != nil {
		return nil, fmt.Errorf("record immutable script: %w", err)
	}
	return value, nil
}

func (s *Service) List(ctx context.Context, teamID, actorID, libraryID, cursor string, limit int) ([]script.Artifact, string, error) {
	owner, err := s.teams.Get(ctx, teamID)
	if err != nil {
		return nil, "", err
	}
	if !owner.Can(actorID, team.PermissionViewLibrary) {
		return nil, "", fmt.Errorf("library view permission required")
	}
	if strings.TrimSpace(libraryID) == "" {
		return s.artifacts.ListArtifacts(ctx, "", cursor, limit)
	}
	return s.artifacts.ListArtifacts(ctx, libraryID, cursor, limit)
}
func (s *Service) Download(ctx context.Context, teamID, actorID, artifactID string) ([]byte, string, error) {
	owner, err := s.teams.Get(ctx, teamID)
	if err != nil {
		return nil, "", err
	}
	if !owner.Can(actorID, team.PermissionViewLibrary) {
		return nil, "", fmt.Errorf("library view permission required")
	}
	artifact, err := s.artifacts.GetArtifact(ctx, artifactID)
	if err != nil {
		return nil, "", err
	}
	persistedContent, err := s.store.Get(ctx, artifact.StorageKey)
	if err != nil {
		return nil, "", err
	}
	if files.Digest(persistedContent) != artifact.Digest {
		return nil, "", fmt.Errorf("stored artifact digest mismatch")
	}
	return persistedContent, artifact.Filename, nil
}

func validateMIME(filename, mediaType string) error {
	base, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return fmt.Errorf("invalid mime type: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if (ext != ".sh" && ext != ".ps1") || (base != "text/plain" && base != "application/x-sh" && base != "application/octet-stream") {
		return fmt.Errorf("file extension and MIME type are not allowed")
	}
	return nil
}

func shellFor(filename string) (script.Shell, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".sh":
		return script.ShellPOSIX, nil
	case ".ps1":
		return script.ShellPowerShell, nil
	default:
		return "", fmt.Errorf("unsupported script type")
	}
}

func normalizeEncoding(content []byte) (script.Encoding, []byte, error) {
	if bytes.HasPrefix(content, []byte{0xff, 0xfe}) {
		if len(content)%2 != 0 {
			return "", nil, fmt.Errorf("invalid UTF-16LE byte length")
		}
		units := make([]uint16, 0, len(content)/2-1)
		for index := 2; index < len(content); index += 2 {
			units = append(units, binary.LittleEndian.Uint16(content[index:index+2]))
		}
		return script.EncodingUTF16LE, []byte(string(utf16.Decode(units))), nil
	}
	if bytes.IndexByte(content, 0) >= 0 {
		return "", nil, fmt.Errorf("binary script content is not accepted")
	}
	return script.EncodingUTF8, bytes.TrimPrefix(content, []byte{0xef, 0xbb, 0xbf}), nil
}
