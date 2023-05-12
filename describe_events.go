package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/nokia/ntt/internal/lnav"
	"github.com/nokia/ntt/k3/log"
	"github.com/spf13/cobra"
)

var (
	DescribeEventsCommand = &cobra.Command{
		Use:    "describe-events",
		Short:  "Outputs a description of k3 runtime events to stdin",
		Hidden: true,
		RunE:   describe,
	}
	formatLnav bool
)

func init() {
	DescribeEventsCommand.Flags().BoolVar(&formatLnav, "lnav", false, "Output in lnav format")
	RootCommand.AddCommand(DescribeEventsCommand)
}

type Event struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Fields      []string `json:"fields,omitempty"`

	log.Category `json:"-"`
}

func describe(cmd *cobra.Command, args []string) error {
	var events []Event
	for _, c := range log.Categories {
		fields := strings.Split(c.String(), "|")
		events = append(events, Event{
			Name:        fields[0],
			Description: fields[1],
			Fields:      fields[2:],
			Category:    c,
		})
	}
	sort.Slice(events, func(i, j int) bool {
		return strings.ToLower(events[i].Name) < strings.ToLower(events[j].Name)
	})

	switch {
	case formatLnav:
		return OutputLnavSpec(events)
	case Format() == "json":
		return OutputJSON(events)
	default:
		return OutputText(events)

	}
}

func OutputLnavSpec(events []Event) error {

	regexes := make(map[string]lnav.Regex)
	for _, e := range events {
		pattern := fmt.Sprintf(`^(?<timestamp>\d{8}T\d{6}\.\d{6})\|(?<eventtype>%s)\|(?<component>(\?|\w+))(=(?<location>[^\|]*))?`, e.Name)
		for range e.Fields {
			pattern += fmt.Sprintf(`\|.*`)
		}
		pattern += "$"
		regexes[e.Name] = lnav.Regex{
			Pattern: pattern,
		}
	}

	formats := map[string]lnav.Format{
		"ttcn3_log": {
			Title:           "TTCN3 Log Format",
			Description:     "Format describing TTCN3 log files.",
			URL:             []string{"https://pkg.go.dev/github.com/nokia/ntt/k3/log#Category"},
			TimestampFormat: []string{"%Y%m%dT%H%M%S.%f"},
			OrderedByTime:   true,
			Regex:           regexes,
			OPidField:       "eventtype",
			LevelField:      "verdict",
			Level: map[string]string{
				"error":   "error|[A-Z]{4}|DIV0|UTF8|inconc",
				"trace":   "none",
				"notice":  "pass",
				"warning": "fail",
			},
			Value: map[string]lnav.Value{
				"eventtype": {
					Kind:        "string",
					Identifier:  true,
					Hidden:      false,
					Description: "The k3r event type",
				},
				"component": {Kind: "string", Identifier: true},
				"location":  {Kind: "string", Identifier: false},
				"timestamp": {Kind: "string", Identifier: false},
			},
		},
	}

	spec := lnav.Spec{
		Schema:  "https://lnav.org/schemas/format-v1.schema.json",
		Formats: formats,
	}

	b, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func OutputJSON(events []Event) error {
	b, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}

func OutputText(events []Event) error {
	for _, e := range events {
		c := color.New(color.Bold)
		if e.IsError() {
			c = color.New(color.Bold, color.FgRed)
		}
		c.Printf("%s", e.Name)
		fmt.Printf(": %s\n", e.Description)
		for i, f := range e.Fields {
			fmt.Printf("  %d: %s\n", 4+i, f)
		}
		fmt.Println()
	}
	return nil
}
