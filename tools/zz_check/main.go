package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/terraform-routeros/terraform-provider-routeros/routeros"
)

func load(v string) map[string]any {
	f, _ := os.Open("tools/schema_changes/ros-" + v + ".json.gz")
	gz, _ := gzip.NewReader(f)
	var m map[string]any
	json.NewDecoder(gz).Decode(&m)
	return m
}

func node(m map[string]any, p string) map[string]any {
	cur := m
	for _, s := range strings.Split(strings.Trim(p, "/"), "/") {
		n, ok := cur[s].(map[string]any)
		if !ok {
			return nil
		}
		cur = n
	}
	return cur
}

func args(m map[string]any, p string) map[string]bool {
	n := node(m, p)
	if n == nil {
		return nil
	}
	res := map[string]bool{}
	for _, c := range []string{"add", "set"} {
		if cm, ok := n[c].(map[string]any); ok {
			for k, v := range cm {
				if k == "_type" {
					continue
				}
				if vm, ok := v.(map[string]any); ok && vm["_type"] == "arg" {
					res[k] = true
				}
			}
		}
	}
	return res
}

func main() {
	ver := os.Args[1]
	m := load(ver)
	names := []string{}
	pm := routeros.Provider().ResourcesMap
	for k := range pm {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, name := range names {
		r := pm[name]
		ps, ok := r.Schema[routeros.MetaResourcePath]
		if !ok {
			continue
		}
		p := ps.Default.(string)
		a := args(m, p)
		if a == nil {
			fmt.Printf("%s %s: PATH NOT FOUND\n", name, p)
			continue
		}
		known := map[string]bool{}
		for k := range r.Schema {
			known[strings.ReplaceAll(k, "_", "-")] = true
		}
		for _, mk := range []string{routeros.MetaTransformSet, routeros.MetaSkipFields} {
			if s, ok := r.Schema[mk]; ok {
				for _, t := range strings.Split(s.Default.(string), ",") {
					t = strings.Trim(t, "\" ")
					for _, x := range strings.Split(t, ":") {
						known[strings.ReplaceAll(x, "_", "-")] = true
					}
				}
			}
		}
		var missing, extra []string
		for k := range a {
			if !known[k] {
				missing = append(missing, k)
			}
		}
		for k, s := range r.Schema {
			if strings.HasPrefix(k, "___") || (!s.Optional && !s.Required) {
				continue
			}
			kk := strings.ReplaceAll(k, "_", "-")
			if !a[kk] {
				extra = append(extra, kk)
			}
		}
		sort.Strings(missing)
		sort.Strings(extra)
		if len(missing)+len(extra) > 0 {
			fmt.Printf("%s %s\n  missing: %v\n  extra: %v\n", name, p, missing, extra)
		}
	}
}
