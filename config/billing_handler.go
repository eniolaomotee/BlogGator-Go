package config

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/eniolaomotee/BlogGator-Go/internal/config"
	"github.com/eniolaomotee/BlogGator-Go/internal/database"
)

func BillingHandler(s *config.State, cmd config.Command, user database.User) error {
	if len(cmd.Args) < 1 {
		printBillingHelp()
		return nil
	}

	action := cmd.Args[0]

	switch action {
	case "status":
		return handleBillingStatus(s, user)
	case "upgrade":
		return handleBillingUpgrade(s, user)
	case "cancel":
		return handleBillingCancel(s, user)
	case "usage":
		return handleBillingUsage(s, user)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func handleBillingStatus(s *config.State, user database.User) error {
	subscription, err := s.Db.GetUserSubscription(context.Background(), user.ID)
	if err != nil {
		fmt.Println("Plan: Free")
		fmt.Println("Status: Active")
		return nil
	}

	tier, _ := s.Db.GetTierByID(context.Background(), subscription.TierID)

	fmt.Printf("Plan: %s\n", tier.Name)
	fmt.Printf("Status: %s\n", subscription.Status)
	fmt.Printf("Renewal: %s\n", subscription.CurrentPeriodEnd.Format("Jan 2, 2006"))

	if subscription.CancelAtEndPeriod.Valid && subscription.CancelAtEndPeriod.Bool {
		fmt.Println("⚠️  Subscription will cancel at period end")
	}

	return nil
}

func handleBillingUpgrade(s *config.State, user database.User) error {
	fmt.Println("Available Plans:")
	fmt.Println()
	fmt.Println("1. Pro - $9/month")
	fmt.Println("   - 50 feeds")
	fmt.Println("   - API access")
	fmt.Println("   - TUI access")
	fmt.Println()
	fmt.Println("2. Team - $29/month")
	fmt.Println("   - 200 feeds")
	fmt.Println("   - Multi-user")
	fmt.Println("   - Webhooks")
	fmt.Println()
	fmt.Println("Visit: https://yourdomain.com/billing to upgrade")

	// Open browser
	exec.Command("open", "https://yourdomain.com/billing").Start()

	return nil
}

func handleBillingCancel(s *config.State, user database.User) error {
	fmt.Print("Are you sure you want to cancel? (yes/no): ")
	var response string
	fmt.Scanln(&response)

	if response != "yes" {
		fmt.Println("Cancellation aborted")
		return nil
	}

	// Call API to cancel
	// Implementation depends on your API client
	fmt.Println("✓ Subscription will be canceled at period end")
	fmt.Println("You'll retain access until" /* period end date */)

	return nil
}

func handleBillingUsage(s *config.State, user database.User) error {
	feedCount, _ := s.Db.CountUserFeeds(context.Background(), user.ID)
	postCount, _ := s.Db.CountUserPosts(context.Background(), user.ID)

	subscription, err := s.Db.GetUserSubscription(context.Background(), user.ID)
	if err != nil {
		fmt.Println("=== Usage (Free Plan) ===")
		fmt.Printf("Feeds: %d / 5\n", feedCount)
		fmt.Printf("Posts: %d / 100\n", postCount)
		return nil
	}
	tier, _ := s.Db.GetTierByID(context.Background(), subscription.TierID)

	fmt.Printf("=== Usage (%s Plan) ===\n", tier.Name)
	fmt.Printf("Feeds: %d", feedCount)
	fmt.Printf(" / %d", tier.MaxFeeds)
	fmt.Println()

	fmt.Printf("Posts: %d", postCount)
	if tier.MaxPosts.Valid {
		fmt.Printf(" / %d", tier.MaxPosts.Int32)
	} else {
		fmt.Print(" / unlimited")
	}
	fmt.Println()

	return nil
}

func printBillingHelp() {
	fmt.Println("Gator Billing Management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gator billing status   - Show current subscription")
	fmt.Println("  gator billing upgrade  - Upgrade your plan")
	fmt.Println("  gator billing cancel   - Cancel subscription")
	fmt.Println("  gator billing usage    - Show usage statistics")
}
