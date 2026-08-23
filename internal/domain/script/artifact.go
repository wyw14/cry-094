package script

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/wyw14/cry-094/internal/domain/common"
)

type Shell string

const (
	ShellPOSIX      Shell = "shell"
	ShellPowerShell Shell = "powershell"
)

type Encoding string

const (
	EncodingUTF8    Encoding = "utf-8"
	EncodingUTF16LE Encoding = "utf-16le"
)

type Library struct {
	ID      string            `json:"id"`
	TeamID  string            `json:"team_id"`
	Name    string            `json:"name"`
	Tags    map[string]string `json:"tags"`
	Version int64             `json:"version"`
	Created time.Time         `json:"created_at"`
}

type Artifact struct {
	ID          string    `json:"id"`
	LibraryID   string    `json:"library_id"`
	Number      int       `json:"number"`
	Filename    string    `json:"filename"`
	Shell       Shell     `json:"shell"`
	Encoding    Encoding  `json:"encoding"`
	Digest      string    `json:"digest"`
	StorageKey  string    `json:"storage_key"`
	Size        int64     `json:"size"`
	UploadedBy  string    `json:"uploaded_by"`
	UploadedAt  time.Time `json:"uploaded_at"`
	ParserBuild string    `json:"parser_build,omitempty"`
	Version     int64     `json:"version"`
}

func NewArtifact(id, libraryID, filename string, shell Shell, encoding Encoding, digest, storageKey, userID string, size int64, number int, now time.Time) (*Artifact, error) {
	if id == "" || libraryID == "" || userID == "" || storageKey == "" {
		return nil, common.NewCoded("ARTIFACT_INVALID", "artifact ownership and storage are required", nil)
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".sh") && !strings.HasSuffix(strings.ToLower(filename), ".ps1") {
		return nil, common.NewCoded("FILE_EXTENSION", "only .sh and .ps1 scripts are accepted", nil)
	}
	if size <= 0 || size > 1024*1024 {
		return nil, common.NewCoded("FILE_SIZE", "script size must be between 1 byte and 1 MiB", nil)
	}
	digestBytes, err := hex.DecodeString(digest)
	if err != nil || len(digestBytes) != 32 {
		return nil, common.NewCoded("DIGEST_INVALID", "a SHA-256 content digest is required", err)
	}
	if number < 1 {
		return nil, common.NewCoded("VERSION_INVALID", "version number must be positive", nil)
	}
	return &Artifact{ID: id, LibraryID: libraryID, Number: number, Filename: filename, Shell: shell,
		Encoding: encoding, Digest: strings.ToLower(digest), StorageKey: storageKey, Size: size,
		UploadedBy: userID, UploadedAt: now.UTC(), Version: 1}, nil
}
