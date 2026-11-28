package server

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

type jwkDocument struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Use string `json:"use,omitempty"`
	Kid string `json:"kid,omitempty"`
	Alg string `json:"alg,omitempty"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
}

type jwksCacheEntry struct {
	keys      map[string]*rsa.PublicKey
	allKeys   []*rsa.PublicKey
	fetchedAt time.Time
}

var (
	jwksCacheMu  sync.RWMutex
	jwksCache    = map[string]jwksCacheEntry{}
	jwksCacheTTL = 5 * time.Minute
	jwksClient   = &http.Client{Timeout: 5 * time.Second}
)

func parseRSAJWKS(raw []byte) (map[string]*rsa.PublicKey, []*rsa.PublicKey) {
	var doc jwkDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, nil
	}
	byKid := make(map[string]*rsa.PublicKey, len(doc.Keys))
	all := make([]*rsa.PublicKey, 0, len(doc.Keys))
	for _, k := range doc.Keys {
		if !strings.EqualFold(k.Kty, "RSA") {
			continue
		}
		if k.Use != "" && !strings.EqualFold(k.Use, "sig") {
			continue
		}
		pub := decodeRSAPublicKey(k.N, k.E)
		if pub == nil {
			continue
		}
		all = append(all, pub)
		if kid := strings.TrimSpace(k.Kid); kid != "" {
			byKid[kid] = pub
		}
	}
	return byKid, all
}

func decodeRSAPublicKey(nB64, eB64 string) *rsa.PublicKey {
	nBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(nB64))
	if err != nil || len(nBytes) == 0 {
		return nil
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(eB64))
	if err != nil || len(eBytes) == 0 || len(eBytes) > 8 {
		return nil
	}
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	if !e.IsInt64() {
		return nil
	}
	exp := int(e.Int64())
	if exp <= 1 || n.BitLen() < 1024 {
		return nil
	}
	return &rsa.PublicKey{N: n, E: exp}
}

func fetchJWKSFromURL(jwksURL string, forceRefresh bool) (map[string]*rsa.PublicKey, []*rsa.PublicKey) {
	jwksURL = strings.TrimSpace(jwksURL)
	if jwksURL == "" {
		return nil, nil
	}
	if !forceRefresh {
		jwksCacheMu.RLock()
		entry, ok := jwksCache[jwksURL]
		jwksCacheMu.RUnlock()
		if ok && time.Since(entry.fetchedAt) < jwksCacheTTL {
			return entry.keys, entry.allKeys
		}
	}

	resp, err := jwksClient.Get(jwksURL)
	if err != nil || resp == nil {
		return nil, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	var doc jwkDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, nil
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, nil
	}
	byKid, all := parseRSAJWKS(raw)
	if len(all) == 0 {
		return nil, nil
	}

	jwksCacheMu.Lock()
	jwksCache[jwksURL] = jwksCacheEntry{keys: byKid, allKeys: all, fetchedAt: time.Now()}
	jwksCacheMu.Unlock()
	return byKid, all
}

func verifyRS256WithCandidates(candidates []*rsa.PublicKey, signingInput string, sigBytes []byte) bool {
	digest := sha256.Sum256([]byte(signingInput))
	for _, pub := range candidates {
		if pub != nil && rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], sigBytes) == nil {
			return true
		}
	}
	return false
}

func (c oidcAdapterConfig) verifyRS256Signature(kid, signingInput string, sigBytes []byte) bool {
	kid = strings.TrimSpace(kid)
	if strings.TrimSpace(c.JWKSJSON) != "" {
		byKid, all := parseRSAJWKS([]byte(c.JWKSJSON))
		if kid != "" {
			if pub := byKid[kid]; pub != nil {
				return verifyRS256WithCandidates([]*rsa.PublicKey{pub}, signingInput, sigBytes)
			}
			return false
		}
		return verifyRS256WithCandidates(all, signingInput, sigBytes)
	}
	if strings.TrimSpace(c.JWKSURL) == "" {
		return false
	}

	byKid, all := fetchJWKSFromURL(c.JWKSURL, false)
	if kid != "" {
		if pub := byKid[kid]; pub != nil {
			return verifyRS256WithCandidates([]*rsa.PublicKey{pub}, signingInput, sigBytes)
		}
		// Key rotation fallback: refresh JWKS once on unknown kid
		byKid, _ = fetchJWKSFromURL(c.JWKSURL, true)
		if pub := byKid[kid]; pub != nil {
			return verifyRS256WithCandidates([]*rsa.PublicKey{pub}, signingInput, sigBytes)
		}
		return false
	}
	return verifyRS256WithCandidates(all, signingInput, sigBytes)
}
