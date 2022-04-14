package outputs

type ValidationResults struct {
	ClusterValidationResults []ClusterValidationResult `json:"clusters"` //data sprawdzenia
}

type ClusterValidationResult struct {
	ClusterPath       string                   `json:"cluster"`
	ValidationResults []PolicyValidationResult `json:"result"`
}

type PolicyValidationResult struct {
	PolicyGroup       string      `json:"policyGroup"`
	PolicyTitle       string      `json:"policyName"`
	PolicyDescription string      `json:"policyDescription"`
	IsValid           bool        `json:"isValid"`
	Violations        []Violation `json:"violations"`
}

type Violation struct {
	ErrorMessage string `json:"errorMessage"`
}
