package k8s

import (
	"sort"

	"github.com/owenrumney/lazytrivy/pkg/output"
)

type state struct {
	currentContext    string
	availableContexts []string
	selectedNamespace string
	selectedKind      string
	selectedResource  string
	currentReport     *output.Report
	currentResult     *output.Result
	// Tree structure for navigation
	namespaces map[string]map[string][]string // namespace -> kind -> []names
}

func (s *state) setCurrentContext(context string) {
	s.currentContext = context
}

func (s *state) setSelectedNamespace(namespace string) {
	s.selectedNamespace = namespace
	// Clear lower level selections when namespace changes
	s.selectedKind = ""
	s.selectedResource = ""
}

func (s *state) setSelectedKind(kind string) {
	s.selectedKind = kind
	// Clear resource selection when kind changes
	s.selectedResource = ""
}

func (s *state) setSelectedResource(resource string) {
	s.selectedResource = resource
}

func (s *state) buildResourceTree(report *output.Report) {
	s.namespaces = make(map[string]map[string][]string)

	if report == nil || report.Resources == nil {
		return
	}

	// Build a map of resources that have actual issues
	resourcesWithIssues := make(map[string]bool) // key: "namespace/kind/name"

	// First pass: identify which resources have issues
	for _, result := range report.Results {
		if len(result.Issues) > 0 {
			resourcesWithIssues[result.Target] = true
		}
	}

	// Use sets to prevent duplicates
	resourceSets := make(map[string]map[string]map[string]bool) // namespace -> kind -> resource -> bool

	// Second pass: only include resources that have issues
	for _, resource := range report.Resources {
		// Check if any of the results for this resource have issues
		hasIssues := false
		for _, result := range resource.Results {
			if len(result.Issues) > 0 {
				hasIssues = true
				break
			}
		}

		// Skip resources that don't have any issues
		if !hasIssues {
			continue
		}

		namespace := resource.Namespace
		if namespace == "" {
			namespace = "Global" // For cluster-scoped resources
		}

		// Initialize nested maps if needed
		if resourceSets[namespace] == nil {
			resourceSets[namespace] = make(map[string]map[string]bool)
		}
		if resourceSets[namespace][resource.Kind] == nil {
			resourceSets[namespace][resource.Kind] = make(map[string]bool)
		}

		// Add to set (automatically handles duplicates)
		resourceSets[namespace][resource.Kind][resource.Name] = true
	}

	// Convert sets to sorted slices
	for namespace, kinds := range resourceSets {
		if s.namespaces[namespace] == nil {
			s.namespaces[namespace] = make(map[string][]string)
		}

		for kind, resources := range kinds {
			// Convert set to slice and sort
			resourceList := make([]string, 0, len(resources))
			for resource := range resources {
				resourceList = append(resourceList, resource)
			}
			sort.Strings(resourceList)
			s.namespaces[namespace][kind] = resourceList
		}
	}
}

func (s *state) getNamespaces() []string {
	namespaces := make([]string, 0, len(s.namespaces))
	for ns := range s.namespaces {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)
	return namespaces
}

func (s *state) getKindsForNamespace(namespace string) []string {
	if nsMap, exists := s.namespaces[namespace]; exists {
		kinds := make([]string, 0, len(nsMap))
		for kind := range nsMap {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		return kinds
	}
	return []string{}
}

func (s *state) getResourcesForKind(namespace, kind string) []string {
	if nsMap, exists := s.namespaces[namespace]; exists {
		if resources, exists := nsMap[kind]; exists {
			// Already sorted in buildResourceTree
			return resources
		}
	}
	return []string{}
}
