package searchcap

import "testing"

func seedManagedSource(t *testing.T, svc *Service, input SourceInput) Source {
	t.Helper()
	source, err := normalizeSourceInput(input)
	if err != nil {
		t.Fatalf("normalize source: %v", err)
	}
	source.Managed = true
	svc.registryMu.Lock()
	defer svc.registryMu.Unlock()
	source.ID = svc.uniqueSourceIDLocked(source)
	svc.registry = append(svc.registry, source)
	return source
}
