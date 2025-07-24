# Toy Buildkite Stack

This repo contains a simple implementation of a buildkite stack that can pretend to orchestrate buildkite agents, and knows how to query for scheduled jobs and reserve them. The intent is that this stack can be used to test the buildkite agent API and the buildkite stack API without needing to run a full buildkite agent or stack.

## Usage

At time of writing there are two commands available:
- `get-scheduled-jobs`: This command will query the buildkite API for scheduled jobs and print their UUIDs to stdout
- `reserve-jobs`: This command will take a list of job UUIDs, and reserve them in the buildkite API, then print their UUIDs to stdout
