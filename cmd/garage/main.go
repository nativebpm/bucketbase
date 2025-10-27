package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	markerFile := "/var/lib/garage/.initialized"

	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		fmt.Println("Initializing Garage...")

		// Start garage server in background
		cmd := exec.Command("/usr/local/bin/garage", "server")
		err := cmd.Start()
		if err != nil {
			panic(err)
		}
		defer cmd.Process.Kill()
		defer cmd.Wait()

		// Wait for server to be ready
		fmt.Println("Waiting for Garage server to start...")
		for {
			// Check garage server status (shell command: garage status)
			statusCmd := exec.Command("/usr/local/bin/garage", "status")
			if err := statusCmd.Run(); err == nil {
				break
			}
			time.Sleep(time.Second)
		}
		fmt.Println("Garage server is ready.")

		// Set up cluster layout
		fmt.Println("Setting up cluster layout...")
		// Get the current node ID (shell command: garage node id)
		nodeIdCmd := exec.Command("/usr/local/bin/garage", "node", "id")
		nodeIdOutput, err := nodeIdCmd.Output()
		if err != nil {
			fmt.Println("Error getting node ID:", err)
			cmd.Process.Kill()
			cmd.Wait()
			panic(err)
		}
		nodeIdLine := strings.TrimSpace(string(nodeIdOutput))
		nodeIdLines := strings.Split(nodeIdLine, "\n")
		if len(nodeIdLines) == 0 {
			fmt.Println("No node ID found")
			cmd.Process.Kill()
			cmd.Wait()
			panic("No node ID found")
		}
		nodeId := strings.Fields(nodeIdLines[len(nodeIdLines)-1])[0] // Take the last line and first field
		// Extract just the hex part before @
		if strings.Contains(nodeId, "@") {
			nodeId = strings.Split(nodeId, "@")[0]
		}
		fmt.Println("Node ID:", nodeId)

		// Assign role to node (shell command: garage layout assign -z <zone> -c <capacity> <nodeId>)
		assignCmd := exec.Command("/usr/local/bin/garage", "layout", "assign", "-z", os.Getenv("GARAGE_ZONE"), "-c", os.Getenv("GARAGE_CAPACITY"), nodeId)
		err = assignCmd.Run()
		if err != nil {
			fmt.Println("Error assigning role:", err)
			cmd.Process.Kill()
			cmd.Wait()
			// Print help on error (shell command: garage layout assign --help)
			assignHelpCmd := exec.Command("/usr/local/bin/garage", "layout", "assign", "--help")
			if assignHelpOutput, helpErr := assignHelpCmd.Output(); helpErr == nil {
				fmt.Println("Garage layout assign --help:")
				fmt.Println(string(assignHelpOutput))
			}
			panic(err)
		}

		// Apply layout (shell command: garage layout apply --version <version>)
		applyCmd := exec.Command("/usr/local/bin/garage", "layout", "apply", "--version", os.Getenv("GARAGE_LAYOUT_VERSION"))
		err = applyCmd.Run()
		if err != nil {
			fmt.Println("Error applying layout:", err)
			// Check current layout status (shell command: garage layout show)
			showCmd := exec.Command("/usr/local/bin/garage", "layout", "show")
			if showOutput, showErr := showCmd.Output(); showErr == nil {
				fmt.Println("Current layout status:")
				fmt.Println(string(showOutput))
			}
			fmt.Println("Continuing with initialization despite layout apply failure...")
		} else {
			fmt.Println("Cluster layout configured.")
		}

		// Import key
		fmt.Println("Importing key...")
		accessKey := os.Getenv("GARAGE_ACCESS_KEY")
		secretKey := os.Getenv("GARAGE_SECRET_KEY")
		fmt.Println("Access key:", accessKey, "Secret key length:", len(secretKey))

		// Check if key already exists
		fmt.Println("Checking garage key list...")
		// List existing keys (shell command: garage key list)
		listCmd := exec.Command("/usr/local/bin/garage", "key", "list")
		output, err := listCmd.Output()
		if err != nil {
			fmt.Println("Error listing keys:", err)
			cmd.Process.Kill()
			cmd.Wait()
			// Print help on error (shell command: garage key --help)
			keyHelpCmd := exec.Command("/usr/local/bin/garage", "key", "--help")
			if keyHelpOutput, helpErr := keyHelpCmd.Output(); helpErr == nil {
				fmt.Println("Garage key --help:")
				fmt.Println(string(keyHelpOutput))
			}
			panic(err)
		}
		fmt.Println("Key list output:", string(output))
		keyExists := false
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, accessKey) {
				keyExists = true
				break
			}
		}
		if !keyExists {
			fmt.Println("Importing key with garage key import...")
			// Import access key (shell command: garage key import [--yes] <accessKey> <secretKey>)
			importArgs := []string{"key", "import"}
			if os.Getenv("GARAGE_KEY_IMPORT_YES") == "true" {
				importArgs = append(importArgs, "--yes")
			}
			importArgs = append(importArgs, accessKey, secretKey)
			importCmd := exec.Command("/usr/local/bin/garage", importArgs...)
			err = importCmd.Run()
			if err != nil {
				fmt.Println("Error importing key:", err)
				// Print help on error (shell command: garage key import --help)
				importHelpCmd := exec.Command("/usr/local/bin/garage", "key", "import", "--help")
				if importHelpOutput, helpErr := importHelpCmd.Output(); helpErr == nil {
					fmt.Println("Garage key import --help:")
					fmt.Println(string(importHelpOutput))
				}
				fmt.Println("Import command failed, but checking if key was imported...")
				// Check again (shell command: garage key list)
				listCmd2 := exec.Command("/usr/local/bin/garage", "key", "list")
				output2, err2 := listCmd2.Output()
				if err2 == nil {
					lines2 := strings.Split(string(output2), "\n")
					for _, line := range lines2 {
						if strings.Contains(line, accessKey) {
							keyExists = true
							fmt.Println("Key was imported successfully despite error")
							break
						}
					}
				}
				if !keyExists {
					cmd.Process.Kill()
					cmd.Wait()
					panic(err)
				}
			}
		} else {
			fmt.Println("Key already exists, skipping import")
		}
		// Create buckets
		fmt.Println("Creating buckets with garage bucket create...")
		buckets := []string{os.Getenv("S3_BUCKET"), os.Getenv("LITESTREAM_BUCKET")}
		for _, bucket := range buckets {
			// Check if bucket exists
			fmt.Println("Checking if bucket", bucket, "exists...")
			// List existing buckets (shell command: garage bucket list)
			listBucketCmd := exec.Command("/usr/local/bin/garage", "bucket", "list")
			output, err := listBucketCmd.Output()
			bucketExists := false
			if err == nil {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if strings.Contains(line, bucket) {
						bucketExists = true
						break
					}
				}
			} else {
				fmt.Println("Error listing buckets:", err)
				cmd.Process.Kill()
				cmd.Wait()
				// Print help on error (shell command: garage bucket --help)
				bucketHelpCmd := exec.Command("/usr/local/bin/garage", "bucket", "--help")
				if bucketHelpOutput, helpErr := bucketHelpCmd.Output(); helpErr == nil {
					fmt.Println("Garage bucket --help:")
					fmt.Println(string(bucketHelpOutput))
				}
				panic(err)
			}
			if !bucketExists {
				fmt.Println("Creating bucket", bucket, "with garage bucket create...")
				// Create bucket (shell command: garage bucket create <bucket>)
				createCmd := exec.Command("/usr/local/bin/garage", "bucket", "create", bucket)
				err = createCmd.Run()
				if err != nil {
					fmt.Println("Error creating bucket", bucket, ":", err)
					cmd.Process.Kill()
					cmd.Wait()
					// Print help on error (shell command: garage bucket create --help)
					createHelpCmd := exec.Command("/usr/local/bin/garage", "bucket", "create", "--help")
					if createHelpOutput, helpErr := createHelpCmd.Output(); helpErr == nil {
						fmt.Println("Garage bucket create --help:")
						fmt.Println(string(createHelpOutput))
					}
					panic(err)
				}
			} else {
				fmt.Println("Bucket", bucket, "already exists, skipping create")
			}
		}

		// Allow access
		fmt.Println("Granting access with garage bucket allow...")
		for _, bucket := range buckets {
			// Grant bucket permissions (shell command: garage bucket allow <bucket> --key <accessKey> [--read] [--write] [--owner])
			allowArgs := []string{"bucket", "allow", bucket, "--key", accessKey}
			if os.Getenv("GARAGE_BUCKET_ALLOW_READ") == "true" {
				allowArgs = append(allowArgs, "--read")
			}
			if os.Getenv("GARAGE_BUCKET_ALLOW_WRITE") == "true" {
				allowArgs = append(allowArgs, "--write")
			}
			if os.Getenv("GARAGE_BUCKET_ALLOW_OWNER") == "true" {
				allowArgs = append(allowArgs, "--owner")
			}
			allowCmd := exec.Command("/usr/local/bin/garage", allowArgs...)
			err = allowCmd.Run()
			if err != nil {
				fmt.Println("Error allowing access to bucket", bucket, ":", err)
				// Kill background server before panicking
				cmd.Process.Kill()
				cmd.Wait()
				// Print help on error (shell command: garage bucket allow --help)
				allowHelpCmd := exec.Command("/usr/local/bin/garage", "bucket", "allow", "--help")
				if allowHelpOutput, helpErr := allowHelpCmd.Output(); helpErr == nil {
					fmt.Println("Garage bucket allow --help:")
					fmt.Println(string(allowHelpOutput))
				}
				panic(err)
			}
		}

		// Create marker
		file, err := os.Create(markerFile)
		if err != nil {
			cmd.Process.Kill()
			cmd.Wait()
			panic(err)
		}
		file.Close()
		fmt.Println("Initialization complete.")

		// Kill the background server before starting the final one
		cmd.Process.Kill()
		cmd.Wait()
	}

	// Start garage server
	fmt.Println("Starting Garage server...")
	// Start the garage server (shell command: garage server)
	serverCmd := exec.Command("/usr/local/bin/garage", "server")
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr
	err := serverCmd.Run()
	if err != nil {
		panic(err)
	}
}
