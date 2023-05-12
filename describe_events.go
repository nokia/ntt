package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
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
)

func init() {
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

	switch Format() {
	case "json":
		return OutputJSON(events)
	default:
		return OutputText(events)

	}
}

func OutputJSON(events []Event) error {
	b, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return err
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
