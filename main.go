package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/buildkite/agent-stack-k8s/v2/api"
	"github.com/urfave/cli/v3"
)

var (
	endpoint        string
	queue           string
	token           string
	reserveJobUUIDs []string

	tokenIdentity *api.AgentTokenIdentity
	apiClient     *api.AgentClient
)

func main() {

	cmd := &cli.Command{
		Name:  "toystack",
		Usage: "A simple CLI tool for managing toy stacks",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "token",
				Usage:       "API token for authentication",
				Required:    true,
				Destination: &token,
			},
			&cli.StringFlag{
				Name:        "endpoint",
				Usage:       "API endpoint for the agent",
				Value:       "https://agent.buildkite.com/v3",
				Destination: &endpoint,
			},
			&cli.StringFlag{
				Name:        "queue",
				Usage:       "Queue to use for job scheduling",
				Value:       "_default",
				Destination: &queue,
			},
		},
		Before: populateGlobals,
		Commands: []*cli.Command{
			{
				Name: "reserve-jobs",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:        "job-uuids",
						Usage:       "List of job UUIDs to reserve",
						Destination: &reserveJobUUIDs,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if len(reserveJobUUIDs) == 0 {
						return fmt.Errorf("at least one job UUID must be provided")
					}

					resp, _, err := apiClient.ReserveJobs(ctx, reserveJobUUIDs)
					if err != nil {
						return fmt.Errorf("failed to reserve jobs: %w", err)
					}

					for _, uuid := range resp.ReservedJobUUIDs {
						fmt.Printf("Reserved job with UUID: %s\n", uuid)
					}
					for _, uuid := range resp.NotReservedJobUUIDs {
						fmt.Printf("Failed to reserve job with UUID: %s\n", uuid)
					}

					return nil
				},
			},
			{
				Name: "get-scheduled-jobs",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					resp, _, err := apiClient.GetScheduledJobs(ctx, "", 1000)
					if err != nil {
						return fmt.Errorf("failed to get scheduled jobs: %w", err)
					}

					if len(resp.Jobs) == 0 {
						fmt.Println("No scheduled jobs found.")
						return nil
					}

					fmt.Printf("Scheduled jobs in cluster %s:\n", tokenIdentity.ClusterName)
					for _, job := range resp.Jobs {
						fmt.Printf("  UUID: %s\n", job.ID)
					}

					return nil
				},
			},
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal("Error running command: ", err)
	}
}

func populateGlobals(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if token == "" {
		return ctx, fmt.Errorf("token is required")
	}

	var err error
	tokenIdentity, err = getTokenIdentity(ctx, token, endpoint)
	if err != nil {
		return ctx, fmt.Errorf("failed to discover cluster UUID: %w", err)
	}

	apiClient, err = api.NewAgentClient(token, endpoint, tokenIdentity.ClusterUUID, queue, nil, api.WithReservation(true))
	if err != nil {
		return ctx, fmt.Errorf("failed to create agent client: %w", err)
	}

	return ctx, nil
}

func getTokenIdentity(ctx context.Context, token, endpoint string) (*api.AgentTokenIdentity, error) {
	tokenClient, err := api.NewAgentTokenClient(token, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create token client: %w", err)
	}

	identity, _, err := tokenClient.GetTokenIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token identity: %w", err)
	}

	return identity, nil
}
