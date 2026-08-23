package analysis

import "sort"

type Diff struct {
	AddedNodes    []string    `json:"added_nodes"`
	RemovedNodes  []string    `json:"removed_nodes"`
	AddedIssues   []IssueKind `json:"added_issues"`
	RemovedIssues []IssueKind `json:"removed_issues"`
	OrderChanged  bool        `json:"order_changed"`
}

func Compare(before, after Result) Diff {
	diff := Diff{OrderChanged: !equalStrings(before.Order, after.Order)}
	for id := range after.Graph.Nodes {
		if _, ok := before.Graph.Nodes[id]; !ok {
			diff.AddedNodes = append(diff.AddedNodes, id)
		}
	}
	for id := range before.Graph.Nodes {
		if _, ok := after.Graph.Nodes[id]; !ok {
			diff.RemovedNodes = append(diff.RemovedNodes, id)
		}
	}
	beforeIssues := issueSet(before.Issues)
	afterIssues := issueSet(after.Issues)
	for kind := range afterIssues {
		if !beforeIssues[kind] {
			diff.AddedIssues = append(diff.AddedIssues, kind)
		}
	}
	for kind := range beforeIssues {
		if !afterIssues[kind] {
			diff.RemovedIssues = append(diff.RemovedIssues, kind)
		}
	}
	sort.Strings(diff.AddedNodes)
	sort.Strings(diff.RemovedNodes)
	sort.Slice(diff.AddedIssues, func(i, j int) bool { return diff.AddedIssues[i] < diff.AddedIssues[j] })
	sort.Slice(diff.RemovedIssues, func(i, j int) bool { return diff.RemovedIssues[i] < diff.RemovedIssues[j] })
	return diff
}
func issueSet(values []Issue) map[IssueKind]bool {
	result := map[IssueKind]bool{}
	for _, value := range values {
		result[value.Kind] = true
	}
	return result
}
func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
