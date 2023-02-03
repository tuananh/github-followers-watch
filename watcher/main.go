package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"facette.io/natsort"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

var cmds = map[string]func(context.Context, *github.Client) error{
	"following": printFollowing,
	"followers": printFollowers,
}

func init() { flag.Parse() }

func main() {
	if flag.NArg() != 1 {
		panic(fmt.Sprintf("requires 1 positional argument (cmd name); got %d args", flag.NArg()))
	}
	cmd, ok := cmds[flag.Arg(0)]
	if !ok {
		panic("unknown command: " + flag.Arg(0))
	}

	var (
		patToken = "GH_PAT_TOKEN"
	)
	for _, v := range []*string{&patToken} {
		if vv := os.Getenv(*v); vv == "" {
			panic(*v + " env var not set")
		} else {
			*v = vv
		}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: patToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	if err := cmd(ctx, client); err != nil {
		panic(err)
	}
}

func selfID(ctx context.Context, c *github.Client) (*string, error) {
	me, _, err := c.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get self ID: %w", err)
	}
	return me.Login, nil
}

func printFollowing(ctx context.Context, c *github.Client) error {
	id, err := selfID(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to get self id: %w", err)
	}
	var following []string

	listOpts := &github.ListOptions{PerPage: 50}
	for {
		users, resp, err := c.Users.ListFollowing(ctx, *id, listOpts)
		if err != nil {
			return fmt.Errorf("failed to list following: %w", err)
		}

		for i := len(users) - 1; i >= 0; i-- {
			following = append(following, *users[i].Login)
		}

		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	natsort.Sort(following)
	for _, v := range following {
		fmt.Printf("%s\n", v)
	}
	return nil
}

func printFollowers(ctx context.Context, c *github.Client) error {
	id, err := selfID(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to get self id: %w", err)
	}

	var followers []string
	listOpts := &github.ListOptions{PerPage: 50}
	for {
		users, resp, err := c.Users.ListFollowers(ctx, *id, listOpts)
		if err != nil {
			return fmt.Errorf("failed to list followers: %w", err)
		}

		for i := len(users) - 1; i >= 0; i-- {
			followers = append(followers, *users[i].Login)
		}
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	natsort.Sort(followers)
	for _, v := range followers {
		fmt.Printf("%s\n", v)
	}
	return nil
}
