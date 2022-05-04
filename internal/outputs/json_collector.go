// Copyright 2022 Google LLC
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

package outputs

import (
	"encoding/json"
	"os"
	"time"

	"github.com/google/gke-policy-automation/internal/policy"
)

type JSONResultCollector struct {
	fileWriter        FileWriter
	filename          string
	validationResults ValidationResults
}

type FileWriter interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

type OSFileWriter struct {
}

// WriteFile implements the Writer interface that's been created so that ioutil.WriteFile can be mocked
func (w OSFileWriter) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func NewJSONResultToFileCollector(filename string) ValidationResultCollector {
	return &JSONResultCollector{
		filename:   filename,
		fileWriter: OSFileWriter{},
	}
}

func NewJSONResultToCustomWriterCollector(filename string, writer FileWriter) ValidationResultCollector {
	return &JSONResultCollector{
		filename:   filename,
		fileWriter: writer,
	}
}

func (p *JSONResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {

	for _, r := range results {
		p.validationResults.ClusterValidationResults = append(p.validationResults.ClusterValidationResults, MapClusterToJson(r))
	}
	return nil
}

func (p *JSONResultCollector) Close() error {

	p.validationResults.ValidationDate = time.Now()

	res, err := json.Marshal(p.validationResults)
	if err != nil {
		return err
	}

	d1 := []byte(res)
	err = p.fileWriter.WriteFile(p.filename, d1, 0644)
	if err != nil {
		return err
	}

	return nil
}
