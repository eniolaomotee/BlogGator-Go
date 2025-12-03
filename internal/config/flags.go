package config

import (
	"fmt"
	"strconv"
	"strings"
)

// parseFlagValue is a helper to parse flag values in both formats: --flag=value or --flag value
func parseFlagValue(args []string, i int, flagName string, shortFlag string) (string, int, error) {
	arg := args[i]

	// Check for --flag=value or -f=value format
	if strings.HasPrefix(arg, flagName+"=") {
		return strings.TrimPrefix(arg, flagName+"="), i, nil
	}
	if shortFlag != "" && strings.HasPrefix(arg, shortFlag+"=") {
		return strings.TrimPrefix(arg, shortFlag+"="), i, nil
	}

	// Check for --flag value or -f value format
	if arg == flagName || (shortFlag != "" && arg == shortFlag) {
		if i+1 >= len(args) {
			return "", i, fmt.Errorf("%s requires a value", flagName)
		}
		return args[i+1], i + 1, nil
	}

	return "", i, nil
}

// parseIntFlag parses an integer flag
func parseIntFlag(args []string, i int, flagName string, shortFlag string) (int, int, error) {
	strVal, newIndex, err := parseFlagValue(args, i, flagName, shortFlag)
	if err != nil {
		return 0, i, err
	}
	if strVal == "" {
		return 0, i, nil
	}

	val, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, i, fmt.Errorf("invalid %s value: %w", flagName, err)
	}

	return val, newIndex, nil
}

// Handles sorting,filtering and Paginatio
func ParseBrowseFlags(args []string) (*BrowseFlags, error) {
	flags := &BrowseFlags{
		Limit:      5,
		SortBy:     "published_at",
		Order:      "desc",
		FeedFilter: "",
		Page:       1,
	}

	// Parsing args for flags and positional flags
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Handle limit or -l
		if strings.HasPrefix(arg, "--limit=") {
			val, newIndex, err := parseIntFlag(args, i, "--limit", "-l")
			if err != nil {
				return nil, err
			}
			if val > 0 {
				flags.Limit = val
				i = newIndex
				continue
			}
		}
		// Handle --page or -p
		if strings.HasPrefix(arg, "--page") || arg == "-p" {
			val, newIndex, err := parseIntFlag(args, i, "--page", "-p")
			if err != nil {
				return nil, err
			}
			if val > 0 {
				if val < 1 {
					return nil, fmt.Errorf("page must be >= 1")
				}
				flags.Page = val
				i = newIndex
				continue
			}
		}

		// Handle --sort or -s
		if strings.HasPrefix(arg, "--sort") || arg == "-s" {
			val, newIndex, err := parseFlagValue(args, i, "--sort", "-s")
			if err != nil {
				return nil, err
			}
			if val != "" {
				flags.SortBy = val
				i = newIndex
				continue
			}
		}

		// Handle --order or -o
		if strings.HasPrefix(arg, "--order") || arg == "-o" {
			val, newIndex, err := parseFlagValue(args, i, "--order", "-o")
			if err != nil {
				return nil, err
			}
			if val != "" {
				flags.Order = val
				i = newIndex
				continue
			}
		}

		// Handle --feed or -f
		if strings.HasPrefix(arg, "--feed") || arg == "-f" {
			val, newIndex, err := parseFlagValue(args, i, "--feed", "-f")
			if err != nil {
				return nil, err
			}
			if val != "" {
				flags.FeedFilter = val
				i = newIndex
				continue
			}
		}

		// Positional argument (backward compatibility for limit)
		if !strings.HasPrefix(arg, "-") {
			val, err := strconv.Atoi(arg)
			if err != nil {
				return nil, fmt.Errorf("invalid limit: %w", err)
			}
			flags.Limit = val
			continue
		}
		return nil, fmt.Errorf("unknown flag : %s", arg)
		// First non-flag argument is query
	}

	// validate sort options
	validSorts := map[string]bool{
		"published_at": true,
		"title":        true,
		"created_at":   true,
	}

	if !validSorts[flags.SortBy] {
		return nil, fmt.Errorf("invalid sorting option: %s (valid: published_at, title, created_at )", flags.SortBy)
	}

	// Validate and normalize order
	flags.Order = strings.ToLower(flags.Order)
	if flags.Order != "asc" && flags.Order != "desc" {
		return nil, fmt.Errorf("invalid order: %s (valid:asc,desc)", flags.Order)
	}

	return flags, nil

}

// Parse Search Flags
func ParseSearchFlags(args []string) (*SearchFlags, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("search query is required")
	}

	flags := &SearchFlags{
		Query: "",
		Limit: 10,
		Field: "all",
	}

	querySet := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Handle --limit or -l
		if strings.HasPrefix(arg, "--limit") || arg == "-l" {
			val, newIndex, err := parseIntFlag(args, i, "--limit", "-l")
			if err != nil {
				return nil, err
			}
			if val > 0 {
				flags.Limit = val
				i = newIndex
				continue
			}
		}

		// Handle --field or -f
		if strings.HasPrefix(arg, "--field") || arg == "-f" {
			val, newIndex, err := parseFlagValue(args, i, "--field", "-f")
			if err != nil {
				return nil, err
			}
			if val != "" {
				flags.Field = val
				i = newIndex
				continue
			}
		}

		// First non-flag argument is the query
		if !strings.HasPrefix(arg, "-") && !querySet {
			flags.Query = arg
			querySet = true
			continue
		}

		// Additional non-flag words are part of the query
		if !strings.HasPrefix(arg, "-") {
			flags.Query += " " + arg
			continue
		}

		return nil, fmt.Errorf("unknown flag: %s", arg)
	}

	if flags.Query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Validate field
	validFields := map[string]bool{
		"all":         true,
		"title":       true,
		"description": true,
		"feed":        true,
	}
	if !validFields[flags.Field] {
		return nil, fmt.Errorf("invalid field: %s (valid: all, title, description, feed)", flags.Field)
	}

	return flags, nil

}
