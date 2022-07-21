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
	"os"

	"github.com/google/gke-policy-automation/internal/log"
	"github.com/google/gke-policy-automation/internal/policy"
)

type JSONResultCollector struct {
	fileWriter   FileWriter
	filename     string
	reportMapper ValidationReportMapper
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
		filename:     filename,
		fileWriter:   OSFileWriter{},
		reportMapper: NewValidationReportMapper(),
	}
}

func NewJSONResultToCustomWriterCollector(filename string, writer FileWriter) ValidationResultCollector {
	return &JSONResultCollector{
		filename:     filename,
		fileWriter:   writer,
		reportMapper: NewValidationReportMapper(),
	}
}

func (p *JSONResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	p.reportMapper.AddResults(results)
	return nil
}

func (p *JSONResultCollector) Close() error {
	reportData, err := p.reportMapper.GetJsonReport()
	if err != nil {
		return err
	}
	if err = p.fileWriter.WriteFile(p.filename, reportData, 0644); err != nil {
		return err
	}
	log.Infof("Validation results written to the [%s] file", p.filename)
	return nil
}

func (p *JSONResultCollector) Name() string {
	return p.filename + " file"
}
