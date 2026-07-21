// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

// WriteTestServiceAccountFile writes a valid temporary service account JSON credentials
// file with a freshly generated in-memory RSA key to path.
func WriteTestServiceAccountFile(path string) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return err
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	})
	data, err := json.Marshal(map[string]string{
		"type":         "service_account",
		"client_email": "test@example.com",
		"private_key":  string(pemBytes),
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// CreateTempSACredentialsFile creates a temporary service account JSON credentials
// file with a unique filename and returns its path.
func CreateTempSACredentialsFile(t testing.TB) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "sa_creds_*.json")
	if err != nil {
		t.Fatalf("failed to create temp service account file: %v", err)
	}
	path := f.Name()
	f.Close()
	if err := WriteTestServiceAccountFile(path); err != nil {
		t.Fatalf("failed to write temp service account credentials file: %v", err)
	}
	return path
}
