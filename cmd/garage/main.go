package main

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Zone             string
	Capacity         string
	LayoutVersion    string
	AccessKey        string
	SecretKey        string
	KeyImportYes     bool
	S3Bucket         string
	LitestreamBucket string
	BucketAllowRead  bool
	BucketAllowWrite bool
	BucketAllowOwner bool
}

func getConfig() Config {
	return Config{
		Zone:             os.Getenv("GARAGE_ZONE"),
		Capacity:         os.Getenv("GARAGE_CAPACITY"),
		LayoutVersion:    os.Getenv("GARAGE_LAYOUT_VERSION"),
		AccessKey:        os.Getenv("GARAGE_ACCESS_KEY"),
		SecretKey:        os.Getenv("GARAGE_SECRET_KEY"),
		KeyImportYes:     os.Getenv("GARAGE_KEY_IMPORT_YES") == "true",
		S3Bucket:         os.Getenv("S3_BUCKET"),
		LitestreamBucket: os.Getenv("LITESTREAM_BUCKET"),
		BucketAllowRead:  os.Getenv("GARAGE_BUCKET_ALLOW_READ") == "true",
		BucketAllowWrite: os.Getenv("GARAGE_BUCKET_ALLOW_WRITE") == "true",
		BucketAllowOwner: os.Getenv("GARAGE_BUCKET_ALLOW_OWNER") == "true",
	}
}

func main() {
	const markerFile = "/tmp/.initialized"

	config := getConfig()

	if _, err := os.Stat(markerFile); os.IsNotExist(err) {

		cmd := exec.Command("/usr/local/bin/garage", "server")
		err := cmd.Start()
		if err != nil {
			panic(err)
		}
		defer cmd.Process.Kill()
		defer cmd.Wait()

		for {
			// Check garage server status (shell command: garage status)
			statusCmd := exec.Command("/usr/local/bin/garage", "status")
			if err := statusCmd.Run(); err == nil {
				break
			}
			time.Sleep(time.Second)
		}

		// Get the current node ID (shell command: garage node id)
		nodeIdCmd := exec.Command("/usr/local/bin/garage", "node", "id")
		nodeIdOutput, err := nodeIdCmd.Output()
		if err != nil {
			cmd.Process.Kill()
			cmd.Wait()
			panic(err)
		}
		nodeIdLine := strings.TrimSpace(string(nodeIdOutput))
		nodeIdLines := strings.Split(nodeIdLine, "\n")
		if len(nodeIdLines) == 0 {
			cmd.Process.Kill()
			cmd.Wait()
			panic("No node ID found")
		}
		nodeId := strings.Fields(nodeIdLines[len(nodeIdLines)-1])[0]
		if strings.Contains(nodeId, "@") {
			nodeId = strings.Split(nodeId, "@")[0]
		}

		// Assign role to node (shell command: garage layout assign -z <zone> -c <capacity> <nodeId>)
		assignCmd := exec.Command("/usr/local/bin/garage", "layout", "assign", "-z", config.Zone, "-c", config.Capacity, nodeId)
		err = assignCmd.Run()
		if err != nil {
			cmd.Process.Kill()
			cmd.Wait()
			// Print help on error (shell command: garage layout assign --help)
			assignHelpCmd := exec.Command("/usr/local/bin/garage", "layout", "assign", "--help")
			if assignHelpOutput, helpErr := assignHelpCmd.Output(); helpErr == nil {
				slog.Info("Garage layout assign help", "command", strings.Join(assignHelpCmd.Args, " "), "output", string(assignHelpOutput))
			}
			slog.Error("Error executing command", "command", strings.Join(assignCmd.Args, " "), "error", err)
			panic(err)
		}

		// Apply layout (shell command: garage layout apply --version <version>)
		applyCmd := exec.Command("/usr/local/bin/garage", "layout", "apply", "--version", config.LayoutVersion)
		err = applyCmd.Run()
		if err != nil {
			// Check current layout status (shell command: garage layout show)
			showCmd := exec.Command("/usr/local/bin/garage", "layout", "show")
			if showOutput, showErr := showCmd.Output(); showErr == nil {
				slog.Info("Current layout status", "output", string(showOutput))
			}
		}

		accessKey := config.AccessKey
		secretKey := config.SecretKey

		// List existing keys (shell command: garage key list)
		listCmd := exec.Command("/usr/local/bin/garage", "key", "list")
		output, err := listCmd.Output()
		if err != nil {
			cmd.Process.Kill()
			cmd.Wait()
			// Print help on error (shell command: garage key --help)
			keyHelpCmd := exec.Command("/usr/local/bin/garage", "key", "--help")
			if keyHelpOutput, helpErr := keyHelpCmd.Output(); helpErr == nil {
				slog.Info("Garage key help", "command", strings.Join(keyHelpCmd.Args, " "), "output", string(keyHelpOutput))
			}
			slog.Error("Error executing command", "command", strings.Join(listCmd.Args, " "), "error", err)
			panic(err)
		}
		keyExists := false
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, accessKey) {
				keyExists = true
				break
			}
		}
		if !keyExists {
			// Import access key (shell command: garage key import [--yes] <accessKey> <secretKey>)
			importArgs := []string{"key", "import"}
			if config.KeyImportYes {
				importArgs = append(importArgs, "--yes")
			}
			importArgs = append(importArgs, accessKey, secretKey)
			importCmd := exec.Command("/usr/local/bin/garage", importArgs...)
			err = importCmd.Run()
			if err != nil {
				// Print help on error (shell command: garage key import --help)
				importHelpCmd := exec.Command("/usr/local/bin/garage", "key", "import", "--help")
				if importHelpOutput, helpErr := importHelpCmd.Output(); helpErr == nil {
					slog.Info("Garage key import help", "command", strings.Join(importHelpCmd.Args, " "), "output", string(importHelpOutput))
				}
				slog.Error("Error executing command", "command", strings.Join(importCmd.Args, " "), "error", err)
				// Check again (shell command: garage key list)
				listCmd2 := exec.Command("/usr/local/bin/garage", "key", "list")
				output2, err2 := listCmd2.Output()
				if err2 == nil {
					lines2 := strings.Split(string(output2), "\n")
					for _, line := range lines2 {
						if strings.Contains(line, accessKey) {
							keyExists = true
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
		}
		buckets := []string{config.S3Bucket, config.LitestreamBucket}
		for _, bucket := range buckets {
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
				cmd.Process.Kill()
				cmd.Wait()
				// Print help on error (shell command: garage bucket --help)
				bucketHelpCmd := exec.Command("/usr/local/bin/garage", "bucket", "--help")
				if bucketHelpOutput, helpErr := bucketHelpCmd.Output(); helpErr == nil {
					slog.Info("Garage bucket help", "command", strings.Join(bucketHelpCmd.Args, " "), "output", string(bucketHelpOutput))
				}
				slog.Error("Error executing command", "command", strings.Join(listBucketCmd.Args, " "), "error", err)
				panic(err)
			}
			if !bucketExists {
				// Create bucket (shell command: garage bucket create <bucket>)
				createCmd := exec.Command("/usr/local/bin/garage", "bucket", "create", bucket)
				err = createCmd.Run()
				if err != nil {
					cmd.Process.Kill()
					cmd.Wait()
					// Print help on error (shell command: garage bucket create --help)
					createHelpCmd := exec.Command("/usr/local/bin/garage", "bucket", "create", "--help")
					if createHelpOutput, helpErr := createHelpCmd.Output(); helpErr == nil {
						slog.Info("Garage bucket create help", "command", strings.Join(createHelpCmd.Args, " "), "output", string(createHelpOutput))
					}
					slog.Error("Error executing command", "command", strings.Join(createCmd.Args, " "), "error", err)
					panic(err)
				}
			}
		}

		for _, bucket := range buckets {
			// Grant bucket permissions (shell command: garage bucket allow <bucket> --key <accessKey> [--read] [--write] [--owner])
			allowArgs := []string{"bucket", "allow", bucket, "--key", accessKey}
			if config.BucketAllowRead {
				allowArgs = append(allowArgs, "--read")
			}
			if config.BucketAllowWrite {
				allowArgs = append(allowArgs, "--write")
			}
			if config.BucketAllowOwner {
				allowArgs = append(allowArgs, "--owner")
			}
			allowCmd := exec.Command("/usr/local/bin/garage", allowArgs...)
			err = allowCmd.Run()
			if err != nil {
				cmd.Process.Kill()
				cmd.Wait()
				// Print help on error (shell command: garage bucket allow --help)
				allowHelpCmd := exec.Command("/usr/local/bin/garage", "bucket", "allow", "--help")
				if allowHelpOutput, helpErr := allowHelpCmd.Output(); helpErr == nil {
					slog.Info("Garage bucket allow help", "command", strings.Join(allowHelpCmd.Args, " "), "output", string(allowHelpOutput))
				}
				slog.Error("Error executing command", "command", strings.Join(allowCmd.Args, " "), "error", err)
				panic(err)
			}
		}

		file, err := os.Create(markerFile)
		if err != nil {
			cmd.Process.Kill()
			cmd.Wait()
			panic(err)
		}
		file.Close()

		cmd.Process.Kill()
		cmd.Wait()
	}

	err := syscall.Exec("/usr/local/bin/garage", []string{"garage", "server"}, os.Environ())
	if err != nil {
		panic(err)
	}
}
