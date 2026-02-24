# Phase 19-B Fix Coordination State

## Status: COMPLETED

## Fix Alpha: Proposal Detection in Chat Handlers
**Status**: COMPLETED
**Files OWNED (only Alpha may modify)**:
- `core/pkg/protocol/envelopes.go` — Add ChatProposal struct + Proposal field to ChatResponsePayload
- `core/internal/server/cognitive.go` — Detect mutation tools in HandleChat/HandleCouncilChat, switch to ModeProposal

**What to do**:
1. Add `ChatProposal` struct to envelopes.go:
   ```go
   type ChatProposal struct {
       Intent        string   `json:"intent"`
       Tools         []string `json:"tools"`
       RiskLevel     string   `json:"risk_level"`
       ConfirmToken  string   `json:"confirm_token"`
       IntentProofID string   `json:"intent_proof_id"`
   }
   ```
2. Add `Proposal *ChatProposal` to ChatResponsePayload (after Brain field)
3. In cognitive.go, define mutation tools set at package level:
   ```go
   var mutationTools = map[string]bool{
       "generate_blueprint": true, "delegate": true,
       "write_file": true, "publish_signal": true,
       "broadcast": true,
   }
   ```
4. In HandleChat (lines ~195-256): after parsing agentResult, check if any tools_used are in mutationTools. If yes:
   - Build scope from tools_used
   - Call createIntentProof + generateConfirmToken (already available on *AdminServer)
   - Build ChatProposal and attach to chatPayload.Proposal
   - Switch auditEvent template to TemplateChatToProposal
   - Set envelope.TemplateID = TemplateChatToProposal, Mode = ModeProposal
5. Do the same in HandleCouncilChat (lines ~387-459)
6. Refactor: extract the mutation detection + proposal building into a helper function `buildChatProposal(agentResult, sourceNode)` to avoid duplication

**DO NOT TOUCH**: admin.go, brains.go, useCortexStore.ts, templates.go

---

## Fix Beta: Wire confirmProposal + Frontend Extraction
**Status**: COMPLETED
**Files OWNED (only Beta may modify)**:
- `interface/store/useCortexStore.ts` — Extract proposal from chat response, wire confirmProposal
- `core/internal/server/templates.go` — Add HandleConfirmAction handler
- `core/internal/server/admin.go` — Register new route (add ONE line only)

**What to do**:
1. In useCortexStore.ts, update the `CTSChatEnvelope` interface (line ~203-217) to add `proposal` to payload:
   ```typescript
   payload: {
       text: string;
       consultations?: string[];
       tools_used?: string[];
       artifacts?: ChatArtifactRef[];
       provenance?: AnswerProvenance;
       brain?: BrainProvenance;
       proposal?: {              // NEW
           intent: string;
           tools: string[];
           risk_level: string;
           confirm_token: string;
           intent_proof_id: string;
       };
   };
   ```

2. In sendMissionChat (line ~1600-1630), after building chatMsg, extract proposal data:
   ```typescript
   // Phase 19-B: Extract proposal from chat response
   proposal: envelope.payload?.proposal ? {
       intent: envelope.payload.proposal.intent,
       teams: 0,
       agents: 0,
       tools: envelope.payload.proposal.tools || [],
       risk_level: envelope.payload.proposal.risk_level || 'medium',
       confirm_token: envelope.payload.proposal.confirm_token,
       intent_proof_id: envelope.payload.proposal.intent_proof_id,
   } : undefined,
   ```
   Also set pendingProposal + activeConfirmToken in the set() call when proposal is present.

3. Rewrite confirmProposal (line ~2071-2076):
   ```typescript
   confirmProposal: async () => {
       const { activeConfirmToken, pendingProposal } = get();
       if (!activeConfirmToken || !pendingProposal) return;
       try {
           const res = await fetch('/api/v1/intent/confirm-action', {
               method: 'POST',
               headers: { 'Content-Type': 'application/json' },
               body: JSON.stringify({ confirm_token: activeConfirmToken }),
           });
           if (!res.ok) console.error('Confirm failed:', await res.text());
       } catch (err) {
           console.error('confirmProposal:', err);
       }
       set({ pendingProposal: null, activeConfirmToken: null });
   },
   ```

4. In templates.go, add HandleConfirmAction after the existing handleGetIntentProof:
   ```go
   // POST /api/v1/intent/confirm-action — confirm a chat-based proposal action.
   func (s *AdminServer) HandleConfirmAction(w http.ResponseWriter, r *http.Request) {
       var req struct { ConfirmToken string `json:"confirm_token"` }
       json.NewDecoder(r.Body).Decode(&req)
       // validate + consume token
       proofID, err := s.validateConfirmToken(req.ConfirmToken)
       if err != nil { respondAPIError(w, err.Error(), 400); return }
       // confirm the proof (no mission ID for chat proposals)
       s.confirmIntentProof(proofID, "")  // need to handle empty mission
       // audit
       auditID, _ := s.createAuditEvent(protocol.TemplateChatToProposal, "confirm", "Chat action confirmed", map[string]any{"proof_id": proofID})
       respondAPIJSON(w, 200, protocol.NewAPISuccess(map[string]any{"confirmed": true, "proof_id": proofID, "audit_event_id": auditID}))
   }
   ```
   Note: confirmIntentProof currently requires a UUID for missionID. For chat proposals, pass a nil UUID or update the function to handle empty strings.

5. In admin.go, add ONE route registration line (near the other intent routes ~line 152):
   ```go
   mux.HandleFunc("POST /api/v1/intent/confirm-action", s.HandleConfirmAction)
   ```

**DO NOT TOUCH**: cognitive.go, brains.go, envelopes.go

---

## Fix Gamma: Persist Brains Toggle to cognitive.yaml
**Status**: COMPLETED
**Files OWNED (only Gamma may modify)**:
- `core/internal/server/brains.go` — Add SaveConfig() calls

**What to do**:
1. In HandleToggleBrain (after line 104): add SaveConfig call:
   ```go
   if err := s.Cognitive.SaveConfig(); err != nil {
       log.Printf("Failed to persist brain toggle: %v", err)
       respondError(w, "Toggle applied but failed to persist: "+err.Error(), http.StatusInternalServerError)
       return
   }
   ```
2. In HandleUpdateBrainPolicy (after line 143): add SaveConfig call:
   ```go
   if err := s.Cognitive.SaveConfig(); err != nil {
       log.Printf("Failed to persist brain policy: %v", err)
       respondError(w, "Policy applied but failed to persist: "+err.Error(), http.StatusInternalServerError)
       return
   }
   ```

**DO NOT TOUCH**: cognitive.go, admin.go, useCortexStore.ts, templates.go

---

## File Ownership Matrix (NO OVERLAP)
| File | Alpha | Beta | Gamma |
|------|-------|------|-------|
| envelopes.go | WRITE | — | — |
| cognitive.go | WRITE | — | — |
| useCortexStore.ts | — | WRITE | — |
| templates.go | — | WRITE | — |
| admin.go | — | WRITE (1 line) | — |
| brains.go | — | — | WRITE |
