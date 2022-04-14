package outputs

import (
	"github.com/google/gke-policy-automation/internal/policy"
)

type ConsoleResultCollector struct {
	out *Output
}

func NewConsoleResultCollector(output *Output) ValidationResultCollector {
	return &ConsoleResultCollector{
		out: output,
	}
}

func (p *ConsoleResultCollector) RegisterResult(results []*policy.PolicyEvaluationResult) error {
	for _, result := range results {
		p.out.ColorPrintf("[yellow][bold]GKE Cluster [%s]:", result.ClusterName)
		for _, group := range result.Groups() {
			p.out.ColorPrintf("\n[light_gray][bold]Group %q:\n\n", group)
			for _, policy := range result.Valid[group] {
				p.out.ColorPrintf("[bold][green][\u2713] %s: [reset][green]%s\n", policy.Title, policy.Description)
			}
			for _, policy := range result.Violated[group] {
				p.out.ColorPrintf("[bold][red][x] %s: [reset][red]%s. [bold]Violations:[reset][red] %s\n", policy.Title, policy.Description, policy.Violations[0])
			}
		}
		p.out.ColorPrintf("\n[bold][green]GKE cluster [%s]: Policies: %d valid, %d violated, %d errored.\n",
			result.ClusterName,
			result.ValidCount(),
			result.ViolatedCount(),
			result.ErroredCount())
	}

	return nil
}

func (p *ConsoleResultCollector) Close() error {
	return nil
}
