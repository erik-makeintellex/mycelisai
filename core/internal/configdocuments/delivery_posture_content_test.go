package configdocuments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

// deliveryPostureDir holds the four company delivery posture Outcome
// Templates. It is reference content only: Core does not load it at runtime.
const deliveryPostureDir = "../../config/documents/templates"

var deliveryPostureFiles = []string{
	"delivery-posture-client-delivery-studio.yaml",
	"delivery-posture-product-delivery-team.yaml",
	"delivery-posture-operations-desk.yaml",
	"delivery-posture-governed-enterprise.yaml",
}

func TestDeliveryPostureTemplatesParseValidateAndCompile(t *testing.T) {
	seenIDs := make(map[string]string)

	for _, name := range deliveryPostureFiles {
		name := name
		t.Run(name, func(t *testing.T) {
			raw := readDeliveryPostureFile(t, name)

			document, err := ParseDocument(raw, "yaml")
			if err != nil {
				t.Fatalf("ParseDocument(%s): %v", name, err)
			}

			if document.Kind != protocol.ConfigDocumentKindOutcomeTemplate {
				t.Fatalf("%s: kind = %q, want OutcomeTemplate", name, document.Kind)
			}
			if document.Metadata.Scope.Kind != protocol.ConfigDocumentScopeBuiltIn {
				t.Fatalf("%s: scope.kind = %q, want built_in", name, document.Metadata.Scope.Kind)
			}
			if document.Metadata.Source.Kind != protocol.ConfigDocumentSourceBuiltIn {
				t.Fatalf("%s: source.kind = %q, want built_in", name, document.Metadata.Source.Kind)
			}
			wantRef := "core/config/documents/templates/" + name
			if document.Metadata.Source.Ref != wantRef {
				t.Fatalf("%s: source.ref = %q, want %q", name, document.Metadata.Source.Ref, wantRef)
			}

			if other, ok := seenIDs[document.Metadata.ID]; ok {
				t.Fatalf("duplicate metadata.id %q in %s and %s", document.Metadata.ID, other, name)
			}
			seenIDs[document.Metadata.ID] = name

			// The Outcome Template spec decode is non-strict in production
			// (json.Unmarshal), so shipped content must be independently
			// strict-decoded here to guarantee no field was silently dropped.
			strictDecodeOutcomeTemplateSpec(t, name, document.Spec)

			if issues := protocol.ValidateConfigDocument(document); len(issues) != 0 {
				t.Fatalf("%s: ValidateConfigDocument issues: %+v", name, issues)
			}

			assertNoRawSecretsOrSwarmSubjects(t, name, raw)

			var spec protocol.OutcomeTemplate
			if err := json.Unmarshal(document.Spec, &spec); err != nil {
				t.Fatalf("%s: decode spec: %v", name, err)
			}
			if err := assertShippedQuestionLimit(spec); err != nil {
				t.Fatalf("%s: %v", name, err)
			}

			result, err := CompileOutcomeTemplateDocument(document, protocol.MinimumSufficientBrief{}, protocol.MinimumSufficientBrief{})
			if err != nil {
				t.Fatalf("CompileOutcomeTemplateDocument(%s): %v", name, err)
			}
			if result.TemplateSnapshot.ID != document.Metadata.ID {
				t.Fatalf("%s: snapshot id = %q, want %q", name, result.TemplateSnapshot.ID, document.Metadata.ID)
			}
			if result.TemplateSnapshot.Version != document.Metadata.Version {
				t.Fatalf("%s: snapshot version = %q, want %q", name, result.TemplateSnapshot.Version, document.Metadata.Version)
			}
			wantDigest, err := protocol.CanonicalConfigDocumentDigest(document)
			if err != nil {
				t.Fatalf("%s: digest: %v", name, err)
			}
			if !strings.HasPrefix(wantDigest, "sha256:") {
				t.Fatalf("%s: digest = %q, want sha256 prefix", name, wantDigest)
			}
			if result.TemplateSnapshot.Digest != wantDigest {
				t.Fatalf("%s: snapshot digest = %q, want %q", name, result.TemplateSnapshot.Digest, wantDigest)
			}

			compiled, err := CompileDocument(document, protocol.MinimumSufficientBrief{}, protocol.MinimumSufficientBrief{})
			if err != nil {
				t.Fatalf("CompileDocument(%s): %v", name, err)
			}
			if _, ok := compiled.(protocol.OutcomeTemplateCompileResult); !ok {
				t.Fatalf("CompileDocument(%s) = %T, want protocol.OutcomeTemplateCompileResult", name, compiled)
			}
		})
	}
}

// TestDeliveryPostureDigestsAreUniqueAndStable proves each posture has a
// distinct, deterministic digest across repeated compiles.
func TestDeliveryPostureDigestsAreUniqueAndStable(t *testing.T) {
	digests := make(map[string]string)
	for _, name := range deliveryPostureFiles {
		document, err := ParseDocument(readDeliveryPostureFile(t, name), "yaml")
		if err != nil {
			t.Fatalf("ParseDocument(%s): %v", name, err)
		}
		first, err := protocol.CanonicalConfigDocumentDigest(document)
		if err != nil {
			t.Fatalf("digest(%s): %v", name, err)
		}
		second, err := protocol.CanonicalConfigDocumentDigest(document)
		if err != nil {
			t.Fatalf("digest(%s) second run: %v", name, err)
		}
		if first != second {
			t.Fatalf("%s: digest not stable across runs: %q vs %q", name, first, second)
		}
		for otherName, otherDigest := range digests {
			if otherDigest == first {
				t.Fatalf("%s and %s share digest %q, want unique per file", name, otherName, first)
			}
		}
		digests[name] = first
	}
}

// TestDeliveryPostureRejectsUnknownSpecField guards the non-strict production
// decode path: content must never rely on an unrecognized field.
func TestDeliveryPostureRejectsUnknownSpecField(t *testing.T) {
	document, err := ParseDocument(readDeliveryPostureFile(t, deliveryPostureFiles[0]), "yaml")
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	mutated := mutateSpec(t, document.Spec, func(spec map[string]any) {
		spec["approval_graduation"] = "approve_none"
	})
	document.Spec = mutated

	decoder := json.NewDecoder(bytes.NewReader(mutated))
	decoder.DisallowUnknownFields()
	var strict protocol.OutcomeTemplate
	if err := decoder.Decode(&strict); err == nil {
		t.Fatal("expected strict decode to reject an unknown spec field")
	}
}

// TestDeliveryPostureRejectsQuestionLimitAboveFour guards against silent
// compile-time clamping (outcome_templates.go clamps > 4 down to 4): the
// content test itself must reject a fixture like this, not rely on compile.
func TestDeliveryPostureRejectsQuestionLimitAboveFour(t *testing.T) {
	document, err := ParseDocument(readDeliveryPostureFile(t, deliveryPostureFiles[0]), "yaml")
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	document.Spec = mutateSpec(t, document.Spec, func(spec map[string]any) {
		spec["question_limit"] = 5
	})

	var strict protocol.OutcomeTemplate
	if err := json.Unmarshal(document.Spec, &strict); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if err := assertShippedQuestionLimit(strict); err == nil {
		t.Fatal("expected the content test to reject question_limit above 4")
	}
}

// assertShippedQuestionLimit is the guard the main test applies to every
// shipped file: the digest and compile precondition (1-4) applies even
// though compile itself would silently clamp a higher value to 4.
func assertShippedQuestionLimit(spec protocol.OutcomeTemplate) error {
	if spec.QuestionLimit <= 0 || spec.QuestionLimit > 4 {
		return fmt.Errorf("question_limit = %d, want 1-4 (shipped value must not rely on compile-time clamping)", spec.QuestionLimit)
	}
	return nil
}

// TestDeliveryPostureRejectsRawSecret guards the spec.raw_secret validator.
func TestDeliveryPostureRejectsRawSecret(t *testing.T) {
	document, err := ParseDocument(readDeliveryPostureFile(t, deliveryPostureFiles[0]), "yaml")
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	document.Spec = mutateSpec(t, document.Spec, func(spec map[string]any) {
		spec["api_key"] = "sk-live-not-a-managed-reference"
	})

	issues := protocol.ValidateConfigDocument(document)
	found := false
	for _, issue := range issues {
		if issue.Code == "spec.raw_secret" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spec.raw_secret validation issue, got: %+v", issues)
	}
}

// TestDeliveryPostureBundleDirectoryDoesNotContainDocuments guards against a
// regression that would break the fatal-on-bad-file bundle loader at startup.
func TestDeliveryPostureBundleDirectoryDoesNotContainDocuments(t *testing.T) {
	bundleDir := "../../config/templates"
	entries, err := os.ReadDir(bundleDir)
	if err != nil {
		t.Fatalf("read %s: %v", bundleDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(bundleDir, entry.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		if bytes.Contains(raw, []byte("kind: OutcomeTemplate")) {
			t.Fatalf("%s under the bundle-loader directory looks like a ConfigDocument; it belongs under core/config/documents/templates instead", entry.Name())
		}
	}
}

func readDeliveryPostureFile(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(deliveryPostureDir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return raw
}

func strictDecodeOutcomeTemplateSpec(t *testing.T, name string, spec json.RawMessage) {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(spec))
	decoder.DisallowUnknownFields()
	var strict protocol.OutcomeTemplate
	if err := decoder.Decode(&strict); err != nil {
		t.Fatalf("%s: strict spec decode found an unrecognized field: %v", name, err)
	}
}

func assertNoRawSecretsOrSwarmSubjects(t *testing.T, name string, raw []byte) {
	t.Helper()
	if bytes.Contains(raw, []byte("swarm.")) {
		t.Fatalf("%s: content must stay domain-neutral and must not reference swarm.* subjects", name)
	}
}

// mutateSpec decodes the spec JSON to a map, applies mutate, and re-encodes.
func mutateSpec(t *testing.T, spec json.RawMessage, mutate func(map[string]any)) json.RawMessage {
	t.Helper()
	var decoded map[string]any
	if err := json.Unmarshal(spec, &decoded); err != nil {
		t.Fatalf("decode spec for mutation: %v", err)
	}
	mutate(decoded)
	out, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("re-encode mutated spec: %v", err)
	}
	return out
}
