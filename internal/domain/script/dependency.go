package script

type DependencyKind string

const (
	DependencyTool        DependencyKind = "tool"
	DependencyLibrary     DependencyKind = "library"
	DependencyEnvironment DependencyKind = "environment"
	DependencyInput       DependencyKind = "input"
	DependencyOutput      DependencyKind = "output"
)

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Dependency struct {
	Kind       DependencyKind `json:"kind"`
	Name       string         `json:"name"`
	Constraint *Constraint    `json:"constraint,omitempty"`
	Optional   bool           `json:"optional"`
	Implicit   bool           `json:"implicit"`
	Location   Location       `json:"location"`
	Evidence   string         `json:"evidence"`
}

type DataContract struct {
	Inputs  []string `json:"inputs"`
	Outputs []string `json:"outputs"`
}

type ParseConclusion struct {
	ArtifactID   string       `json:"artifact_id"`
	Digest       string       `json:"digest"`
	ParserName   string       `json:"parser_name"`
	ParserBuild  string       `json:"parser_build"`
	Dependencies []Dependency `json:"dependencies"`
	Contract     DataContract `json:"contract"`
	Warnings     []string     `json:"warnings"`
}

func (c ParseConclusion) BoundTo(artifact Artifact) bool {
	return c.ArtifactID == artifact.ID && c.Digest == artifact.Digest && c.ParserBuild != ""
}
