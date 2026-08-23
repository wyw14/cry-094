package script

import "testing"

func TestConstraintSatisfiedBy(t *testing.T) {
	tests := []struct {
		constraint, version string
		want                bool
	}{{">=1.7", "1.7.1", true}, {">=4.40", "4.39", false}, {"<3", "2.15.0", true}, {"=1.0.0", "1", true}}
	for _, tt := range tests {
		t.Run(tt.constraint+"_"+tt.version, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatal(err)
			}
			version, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatal(err)
			}
			if got := constraint.SatisfiedBy(version); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
