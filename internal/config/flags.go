package config

import (
	"fmt"
	"strconv"
	"strings"
)


// Handles sorting,filtering and Paginatio 
func ParseBrowseFlags(args []string)(*BrowseFlags, error){
	flags := &BrowseFlags{
		Limit: 2,
		SortBy: "published_at",
		Order: "desc",
		FeedFilter: "",
		Page:10,
	}

	// Parsing args for flags and positional flags
	for i := 0; i< len(args); i++{
		arg := args[i]
		// Handle limit or -l
		if strings.HasPrefix(arg, "--limit="){
			val, err := strconv.Atoi(strings.TrimPrefix(arg, "--limit="))
			if err != nil{
				return nil, fmt.Errorf("invalud limit value %w", err)
			}
			flags.Limit = val
		} else if arg == "--limit" || arg == "-l" {
			if i+1 >= len(args){
				return nil, fmt.Errorf("--Limit requires a value")
			}
			val, err := strconv.Atoi(args[i+1])
			if err != nil{
				return nil, fmt.Errorf("invaluid limit value: %w",err)
			}
			flags.Limit = val
			i++
			// sorting
		} else if strings.HasPrefix(arg, "--page="){
			val,err := strconv.Atoi(strings.TrimPrefix(arg, "--page="))
			if err != nil{
				return nil, fmt.Errorf("invalid page value %w", err)
			}
			flags.Page = val
		}else if arg == "--page" || arg == "-p"{
			if i+1 >= len(args){
				return nil, fmt.Errorf("--page requires a value")
			}
			val, err := strconv.Atoi(args[i+1])
			if err != nil{
				return nil, fmt.Errorf("invalid page value :%w", err)
			}
			if val < 1{
				return nil, fmt.Errorf("page must be >= 1")
			}
			flags.Page = val
			i++
		}else if strings.HasPrefix(arg, "--sort="){
			flags.SortBy = strings.TrimPrefix(arg,"--sort=")
		}else if arg == "--sort" || arg == "-s"{
			if i+1 >= len(args){
				return nil, fmt.Errorf("--sort requires a value")
			}
			flags.SortBy = args[i+1]
			i++
			// Handle order
		} else if strings.HasPrefix(arg, "--order="){
			flags.Order = strings.TrimPrefix(arg, "--order=")
		}else if arg == "--order" || arg == "-o"{
			if i+1 >= len(args){
				return nil, fmt.Errorf("--order requires a value")
			}
			flags.Order = args[i+1]
			i++
			// Handle feed
		} else if strings.HasPrefix(arg, "--feed="){
			flags.FeedFilter = strings.TrimPrefix(arg, "--feed=")
		} else if arg == "--feed" || arg == "-f"{
			if i+1 >= len(args){
				return nil, fmt.Errorf("--feed required a value")
			}
			flags.FeedFilter = args[i+1]
			i++
		}else if !strings.HasPrefix(arg, "-"){
			val, err := strconv.Atoi(arg)
			if err != nil{
				return nil, fmt.Errorf("invalid limit %w",err)
			}
			flags.Limit = val
		}else{
			return nil, fmt.Errorf("unknown flag: ")
		}
	}

	validSorts := map[string]bool{
		"published_at": true,
		"title": true,
		"created_at": true,
	}

	if !validSorts[flags.SortBy]{
		return nil,fmt.Errorf("invalid sorting option: %s (valid: published_at, title, created_at )", flags.SortBy)
	}

	// Validate and normalize order
	flags.Order = strings.ToLower(flags.Order)
	if flags.Order != "asc" && flags.Order != "desc"{
		return nil, fmt.Errorf("invalid order: %s (valid:asc,desc)", flags.Order)
	}

	return flags, nil
}