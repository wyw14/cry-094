package script

import (
	"fmt"
	"strconv"
	"strings"
)

type SemanticVersion struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

func ParseVersion(raw string) (SemanticVersion, error) {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(raw), "v"), ".")
	if len(parts) < 1 || len(parts) > 3 {
		return SemanticVersion{}, fmt.Errorf("invalid semantic version %q", raw)
	}
	values := []int{0, 0, 0}
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return SemanticVersion{}, fmt.Errorf("invalid semantic version %q", raw)
		}
		values[i] = value
	}
	return SemanticVersion{Major: values[0], Minor: values[1], Patch: values[2]}, nil
}

func (v SemanticVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v SemanticVersion) Compare(other SemanticVersion) int {
	if v.Major != other.Major {
		return compareInt(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return compareInt(v.Minor, other.Minor)
	}
	return compareInt(v.Patch, other.Patch)
}

func compareInt(left, right int) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

type Constraint struct {
	Operator string          `json:"operator"`
	Version  SemanticVersion `json:"version"`
}

func ParseConstraint(raw string) (Constraint, error) {
	raw = strings.TrimSpace(raw)
	for _, op := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(raw, op) {
			version, err := ParseVersion(strings.TrimSpace(strings.TrimPrefix(raw, op)))
			return Constraint{Operator: op, Version: version}, err
		}
	}
	version, err := ParseVersion(raw)
	return Constraint{Operator: "=", Version: version}, err
}

func (c Constraint) SatisfiedBy(version SemanticVersion) bool {
	comparison := version.Compare(c.Version)
	switch c.Operator {
	case ">=":
		return comparison >= 0
	case "<=":
		return comparison <= 0
	case ">":
		return comparison > 0
	case "<":
		return comparison < 0
	case "=":
		return comparison == 0
	default:
		return false
	}
}
