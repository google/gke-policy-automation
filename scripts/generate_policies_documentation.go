package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

type Metadata struct {
	Title       string
	Description string
	FileName    string
}

const REPO_POLICIES_URL = "https://github.com/google/gke-policy-automation/blob/main/gke-policies/policy/"
const REGO_TEST_FILE_SUFFIX = "_test.rego"

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Missing folder path and output file full name.")
		return
	}

	var directory = os.Args[1]
	var output_file = os.Args[2]

	// Find all policies
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		fmt.Println(err)
		return
	}

	var metadata = make(map[string][]*Metadata)

	// Read policies contents
	for _, file := range files {

		// Skip all policy tests
		if strings.HasSuffix(file.Name(), REGO_TEST_FILE_SUFFIX) {
			continue
		}

		data, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, file.Name()))
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Parse policies modules
		module, err := ast.ParseModuleWithOpts(file.Name(), string(data),
			ast.ParserOptions{ProcessAnnotation: true})
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Extract annotations from parsed modules
		for _, annotation := range module.Annotations {

			// Metadata to group policies together
			group := annotation.Custom["group"]

			if group == nil {
				continue
			}

			metadataList, exists := metadata[group.(string)]

			if !exists {
				metadataList = make([]*Metadata, 0)
			}

			metadataList = append(metadataList, &Metadata{
				Title:       annotation.Title,
				Description: annotation.Description,
				FileName:    file.Name(),
			})

			metadata[group.(string)] = metadataList
		}
	}

	// Sort groups alphabetically
	groups := make([]string, 0, len(metadata))

	for key := range metadata {
		groups = append(groups, key)
	}

	sort.Strings(groups)

	// Ready to create the file
	f, err := os.Create(output_file)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	// Finally, write down the policy list
	f.WriteString("# Available Policies List\n")

	for _, group := range groups {

		policies := metadata[group]

		f.WriteString(fmt.Sprintf("\n## **%s** (%d policies)\n", group, len(policies)))

		for _, policy := range policies {
			f.WriteString(fmt.Sprintf("**Title**: %s ([link](%s%s))<br/>", policy.Title, REPO_POLICIES_URL, policy.FileName))
			f.WriteString(fmt.Sprintf("**Description**: %s\n\n", policy.Description))
		}

		f.WriteString("\n")
	}
}
