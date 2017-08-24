package app

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/10gen/stitch-cli/ui"
)

// Diff compares two apps and gives a human-readable string representing the
// differences between them.
func Diff(old, new App) (diff string) {
	var lines, d []string

	// app.Group and app.ID are immutable.

	// app.Name
	d = diffValue(old.Name, new.Name, "name")
	lines = append(lines, d...)

	// app.Services
	d = diffServices(old.Services, new.Services)
	lines = append(lines, d...)

	// app.Pipelines
	d = diffPipelines(old.Pipelines, new.Pipelines)
	lines = append(lines, d...)

	// app.Values
	d = diffValues(old.Values, new.Values)
	lines = append(lines, d...)

	// app.AuthProviders
	d = diffAuthProviders(old.AuthProviders, new.AuthProviders)
	lines = append(lines, d...)

	// prepend "* "
	if len(lines) > 1 {
		lines[0] = "* " + lines[0]
	}
	diff = strings.Join(lines, "\n* ")
	return
}

func diffValue(old, new, name string) []string {
	if new == old {
		return nil
	}
	return []string{
		fmt.Sprintf(`modified %s from "%s" to "%s"`,
			name,
			ui.Color(ui.DiffOld, old),
			ui.Color(ui.DiffNew, new),
		),
	}
}

func diffBool(old, new bool, name string) []string {
	if new == old {
		return nil
	}
	return []string{
		fmt.Sprintf(`modified %s from "%s" to "%s"`,
			name,
			ui.Color(ui.DiffOld, strconv.FormatBool(old)),
			ui.Color(ui.DiffNew, strconv.FormatBool(new)),
		),
	}
}

func diffStringSlice(old, new []string, singularName string) (d []string) {
	oldCopy := make([]string, len(old))
	newCopy := make([]string, len(new))
	copy(oldCopy, old)
	copy(newCopy, new)
	old, new = oldCopy, newCopy
	sort.Strings(old)
	sort.Strings(new)

	var created, deleted []string
	var i, j int
	for i < len(old) && j < len(new) {
		if old[i] == new[j] {
			i++
			j++
			continue
		}
		if old[i] < new[j] {
			deleted = append(deleted, old[i])
			i++
			continue
		}
		// new[j] < old[i]
		created = append(created, new[j])
		j++
	}
	for ; i < len(old); i++ {
		deleted = append(deleted, old[i])
	}
	for ; j < len(new); j++ {
		created = append(created, new[j])
	}

	for _, s := range deleted {
		d = append(d, fmt.Sprintf(`deleted %s: "%s"`, singularName, ui.Color(ui.DiffOld, s)))
	}
	for _, s := range created {
		d = append(d, fmt.Sprintf(`created %s: "%s"`, singularName, ui.Color(ui.DiffNew, s)))
	}
	return
}

func diffServices(old, new []Service) (d []string) {
	oldM := make(map[string]*Service) // by ID
	newM := make(map[string]*Service) // by ID
	for i, svc := range old {
		oldM[svc.ID] = &old[i]
	}
	for i, svc := range new {
		id := svc.ID
		if id == "" {
			id = uid()
			new[i].ID = id
		}
		newM[id] = &new[i]
	}
	deletedSvcs := make(map[string]bool)
	createdSvcs := make(map[string]bool)
	for id := range oldM {
		if _, ok := newM[id]; !ok {
			deletedSvcs[id] = true
		}
	}
	for id := range newM {
		if _, ok := oldM[id]; !ok {
			createdSvcs[id] = true
		}
	}
	for _, svc := range old { // iter over slice (not map) for determinism
		if _, ok := deletedSvcs[svc.ID]; ok {
			d = append(d, fmt.Sprintf(`deleted service "%s"`,
				ui.Color(ui.DiffOld, svc.Name)))
		}
	}
	for _, svc := range new { // iter over slice (not map) for determinism
		if _, ok := createdSvcs[svc.ID]; ok {
			d = append(d, fmt.Sprintf(`created service "%s"`,
				ui.Color(ui.DiffNew, svc.Name)))
		}
	}
	for _, newSvc := range new {
		if oldSvc, ok := oldM[newSvc.ID]; ok {
			d = append(d, diffService(*oldSvc, newSvc)...)
		}
	}
	return
}

func diffService(old, new Service) (d []string) {
	d = append(d, diffValue(old.Name, new.Name,
		fmt.Sprintf("name of service %q", old.ID))...)
	d = append(d, diffValue(old.Type, new.Type,
		fmt.Sprintf("type of service %s", new.Name))...)
	d = append(d, diffValue(string(old.Config), string(new.Config),
		fmt.Sprintf("config of service %s", new.Name))...)
	d = append(d, diffWebhooks(old.Webhooks, new.Webhooks, new.Name)...)
	d = append(d, diffRules(old.Rules, new.Rules, new.Name)...)
	return
}

func diffWebhooks(old, new []Webhook, svc string) (d []string) {
	oldM := make(map[string]*Webhook)
	newM := make(map[string]*Webhook)
	for i, wh := range old {
		oldM[wh.ID] = &old[i]
	}
	for i, wh := range new {
		id := wh.ID
		if id == "" {
			id = uid()
			new[i].ID = id
		}
		newM[id] = &new[i]
	}
	deleted := make(map[string]bool)
	created := make(map[string]bool)
	for _, wh := range old {
		if _, ok := newM[wh.ID]; !ok {
			deleted[wh.ID] = true
		}
	}
	for _, wh := range new {
		if _, ok := oldM[wh.ID]; !ok {
			created[wh.ID] = true
		}
	}

	for _, wh := range old { // iter over slice (not map) for determinism
		if _, ok := deleted[wh.ID]; ok {
			d = append(d, fmt.Sprintf(`deleted webhook "%s" in service %s`,
				ui.Color(ui.DiffOld, wh.Name), svc))
		}
	}
	for _, wh := range new { // iter over slice (not map) for determinism
		if _, ok := created[wh.ID]; ok {
			d = append(d, fmt.Sprintf(`created webhook "%s" in service %s`,
				ui.Color(ui.DiffNew, wh.Name), svc))
		}
	}
	for _, newWh := range new {
		if oldWh, ok := oldM[newWh.ID]; ok {
			d = append(d, diffWebhook(*oldWh, newWh, svc)...)
		}
	}
	return
}

func diffWebhook(old, new Webhook, svc string) (d []string) {
	d = append(d, diffValue(old.Name, new.Name,
		fmt.Sprintf("name of webhook %q in service %s", old.ID, svc))...)
	d = append(d, diffValue(old.Output, new.Output,
		fmt.Sprintf("output of webhook %s in service %s", new.Name, svc))...)
	d = append(d, diffValue(string(old.Pipeline), string(new.Pipeline),
		fmt.Sprintf("pipeline of webhook %s in service %s", new.Name, svc))...)
	return
}

func diffRules(old, new []ServiceRule, svc string) (d []string) {
	oldM := make(map[string]*ServiceRule)
	newM := make(map[string]*ServiceRule)
	for i, r := range old {
		oldM[r.ID] = &old[i]
	}
	for i, r := range new {
		id := r.ID
		if id == "" {
			id = uid()
			new[i].ID = id
		}
		newM[id] = &new[i]
	}
	deleted := make(map[string]bool)
	created := make(map[string]bool)
	for _, r := range old {
		if _, ok := newM[r.ID]; !ok {
			deleted[r.ID] = true
		}
	}
	for _, r := range new {
		if _, ok := oldM[r.ID]; !ok {
			created[r.ID] = true
		}
	}

	for _, r := range old { // iter over slice (not map) for determinism
		if _, ok := deleted[r.ID]; ok {
			d = append(d, fmt.Sprintf(`deleted rule "%s" in service %s`,
				ui.Color(ui.DiffOld, r.Name), svc))
		}
	}
	for _, r := range new { // iter over slice (not map) for determinism
		if _, ok := created[r.ID]; ok {
			d = append(d, fmt.Sprintf(`created rule "%s" in service %s`,
				ui.Color(ui.DiffNew, r.Name), svc))
		}
	}
	for _, newRule := range new {
		if oldRule, ok := oldM[newRule.ID]; ok {
			d = append(d, diffRule(*oldRule, newRule, svc)...)
		}
	}
	return
}

func diffRule(old, new ServiceRule, svc string) (d []string) {
	d = append(d, diffValue(old.Name, new.Name,
		fmt.Sprintf("name of webhook %q in service %s", old.ID, svc))...)
	d = append(d, diffValue(string(old.Rule), string(new.Rule),
		fmt.Sprintf("pipeline of webhook %s in service %s", new.Name, svc))...)
	return
}

func diffPipelines(old, new []Pipeline) (d []string) {
	oldM := make(map[string]*Pipeline) // by ID
	newM := make(map[string]*Pipeline) // by ID
	for i, p := range old {
		oldM[p.ID] = &old[i]
	}
	for i, p := range new {
		id := p.ID
		if id == "" {
			id = uid()
			new[i].ID = id
		}
		newM[id] = &new[i]
	}
	deleted := make(map[string]bool)
	created := make(map[string]bool)
	for id := range oldM {
		if _, ok := newM[id]; !ok {
			deleted[id] = true
		}
	}
	for id := range newM {
		if _, ok := oldM[id]; !ok {
			created[id] = true
		}
	}
	for _, p := range old { // iter over slice (not map) for determinism
		if _, ok := deleted[p.ID]; ok {
			d = append(d, fmt.Sprintf(`deleted pipeline "%s"`,
				ui.Color(ui.DiffOld, p.Name)))
		}
	}
	for _, p := range new { // iter over slice (not map) for determinism
		if _, ok := created[p.ID]; ok {
			d = append(d, fmt.Sprintf(`created pipeline "%s"`,
				ui.Color(ui.DiffNew, p.Name)))
		}
	}
	for _, newP := range new {
		if oldP, ok := oldM[newP.ID]; ok {
			d = append(d, diffPipeline(*oldP, newP)...)
		}
	}
	return
}

func diffPipeline(old, new Pipeline) (d []string) {
	d = append(d, diffValue(old.Name, new.Name,
		fmt.Sprintf("name of pipeline %q", old.ID))...)
	d = append(d, diffValue(old.Output, new.Output,
		fmt.Sprintf("output of pipeline %s", new.Name))...)
	d = append(d, diffBool(old.Private, new.Private,
		fmt.Sprintf("private of pipeline %s", new.Name))...)
	d = append(d, diffBool(old.SkipRules, new.SkipRules,
		fmt.Sprintf("skip-rules of pipeline %s", new.Name))...)
	d = append(d, diffValue(string(old.CanEvaluate), string(new.CanEvaluate),
		fmt.Sprintf("can-evaluate of pipeline %s", new.Name))...)
	d = append(d, diffValue(string(old.Pipeline), string(new.Pipeline),
		fmt.Sprintf("pipeline of pipeline %s", new.Name))...)
	d = append(d, diffPipelineParameters(old.Parameters, new.Parameters, new.Name)...)
	return
}

func diffPipelineParameters(old, new []PipelineParameter, pipeline string) (d []string) {
	oldM := make(map[string]bool) // by name
	newM := make(map[string]bool) // by name
	for _, p := range old {
		oldM[p.Name] = p.Required
	}
	for _, p := range new {
		newM[p.Name] = p.Required
	}
	created := make(map[string]bool)
	deleted := make(map[string]bool)
	for name := range oldM {
		if _, ok := newM[name]; !ok {
			deleted[name] = true
		}
	}
	for name := range newM {
		if _, ok := oldM[name]; !ok {
			created[name] = true
		}
	}
	for _, p := range old { // iter over slice (not map) for determinism
		if _, ok := deleted[p.Name]; ok {
			d = append(d, fmt.Sprintf(`deleted parameter "%s" in pipeline %s`,
				ui.Color(ui.DiffOld, p.Name), pipeline))
		}
	}
	for _, p := range new { // iter over slice (not map) for determinism
		if _, ok := created[p.Name]; ok {
			d = append(d, fmt.Sprintf(`created parameter "%s" in pipeline %s`,
				ui.Color(ui.DiffOld, p.Name), pipeline))
		}
	}
	for _, newP := range new {
		if oldRequired, ok := oldM[newP.Name]; ok {
			d = append(d, diffBool(oldRequired, newP.Required,
				fmt.Sprintf("required of pipeline parameter %s in pipeline %s", newP.Name, pipeline))...)
		}
	}
	return
}

func diffValues(old, new []Value) (d []string) {
	oldM := make(map[string]*json.RawMessage) // by name
	newM := make(map[string]*json.RawMessage) // by name
	for i, v := range old {
		oldM[v.Name] = &old[i].Value
	}
	for i, v := range new {
		newM[v.Name] = &new[i].Value
	}
	created := make(map[string]bool)
	deleted := make(map[string]bool)
	for name := range oldM {
		if _, ok := newM[name]; !ok {
			deleted[name] = true
		}
	}
	for name := range newM {
		if _, ok := oldM[name]; !ok {
			created[name] = true
		}
	}
	for _, v := range old { // iter over slice (not map) for determinism
		if _, ok := deleted[v.Name]; ok {
			d = append(d, fmt.Sprintf(`deleted value "%s"`, ui.Color(ui.DiffOld, v.Name)))
		}
	}
	for _, v := range new { // iter over slice (not map) for determinism
		if _, ok := created[v.Name]; ok {
			d = append(d, fmt.Sprintf(`created value "%s"`, ui.Color(ui.DiffNew, v.Name)))
		}
	}
	for _, n := range new {
		if oj, ok := oldM[n.Name]; ok {
			d = append(d, diffValue(string(*oj), string(n.Value), fmt.Sprintf("value %s", n.Name))...)
		}
	}
	return
}

func diffAuthProviders(old, new []AuthProvider) (d []string) {
	oldM := make(map[string]*AuthProvider) // by ID
	newM := make(map[string]*AuthProvider) // by ID
	for i, p := range old {
		oldM[p.ID] = &old[i]
	}
	for i, p := range new {
		id := p.ID
		if id == "" {
			id = uid()
			new[i].ID = id
		}
		newM[id] = &new[i]
	}
	deleted := make(map[string]bool)
	created := make(map[string]bool)
	for id := range oldM {
		if _, ok := newM[id]; !ok {
			deleted[id] = true
		}
	}
	for id := range newM {
		if _, ok := oldM[id]; !ok {
			created[id] = true
		}
	}
	for _, p := range old { // iter over slice (not map) for determinism
		if _, ok := deleted[p.ID]; ok {
			d = append(d, fmt.Sprintf(`deleted pipeline "%s"`,
				ui.Color(ui.DiffOld, p.Name)))
		}
	}
	for _, p := range new { // iter over slice (not map) for determinism
		if _, ok := created[p.ID]; ok {
			d = append(d, fmt.Sprintf(`created pipeline "%s"`,
				ui.Color(ui.DiffNew, p.Name)))
		}
	}
	for _, newP := range new {
		if oldP, ok := oldM[newP.ID]; ok {
			d = append(d, diffAuthProvider(*oldP, newP)...)
		}
	}
	return
}

func diffAuthProvider(old, new AuthProvider) (d []string) {
	d = append(d, diffValue(old.Name, new.Name,
		fmt.Sprintf("name of auth provider %q", old.ID))...)
	d = append(d, diffValue(old.Type, new.Type,
		fmt.Sprintf("type of auth provider %s", new.Name))...)
	d = append(d, diffBool(old.Enabled, new.Enabled,
		fmt.Sprintf("enablement auth provider %s", new.Name))...)
	d = append(d, diffStringSlice(old.Metadata, new.Metadata,
		fmt.Sprintf("metadata field of auth provider %s", new.Name))...)
	d = append(d, diffStringSlice(old.DomainRestrictions, new.DomainRestrictions,
		fmt.Sprintf("domain-restriction of auth provider %s", new.Name))...)
	d = append(d, diffStringSlice(old.RedirectURIs, new.RedirectURIs,
		fmt.Sprintf("redirect-URI of auth provider %s", new.Name))...)
	d = append(d, diffValue(string(old.Config), string(new.Config),
		fmt.Sprintf("config of auth provider %s", new.Name))...)
	return
}

// temporary local uid for new items which have unspecified ID fields
var uidc uint

func uid() string {
	uidc++
	return fmt.Sprintf("__uid_%d", uidc)
}
