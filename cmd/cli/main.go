package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var (
	apiURL   string
	tenantID string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gpu-cli",
		Short: "GPU Scheduler CLI",
		Long:  "Command-line interface for GPU Scheduler",
	}

	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080", "API server URL")
	rootCmd.PersistentFlags().StringVar(&tenantID, "tenant-id", "default", "Tenant ID")

	rootCmd.AddCommand(
		submitJobCmd(),
		listJobsCmd(),
		getJobCmd(),
		cancelJobCmd(),
		clusterStatusCmd(),
		createTenantCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func submitJobCmd() *cobra.Command {
	var (
		name        string
		priority    int
		gpuCount    int
		cpuCores    int
		image       string
		script      string
	)

	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a new job",
		Run: func(cmd *cobra.Command, args []string) {
			job := map[string]interface{}{
				"tenant_id":    tenantID,
				"name":         name,
				"priority":     priority,
				"gpu_count":    gpuCount,
				"gpu_memory_mb": 16000,
				"cpu_cores":    cpuCores,
				"memory_mb":    32000,
				"image":        image,
				"script":       script,
			}

			resp, err := postJSON(apiURL+"/api/v1/jobs", job)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Job submitted successfully!\n")
			fmt.Printf("Job ID: %s\n", resp["job_id"])
			fmt.Printf("Status: %s\n", resp["status"])
		},
	}

	cmd.Flags().StringVar(&name, "name", "my-job", "Job name")
	cmd.Flags().IntVar(&priority, "priority", 100, "Job priority")
	cmd.Flags().IntVar(&gpuCount, "gpus", 1, "Number of GPUs")
	cmd.Flags().IntVar(&cpuCores, "cpus", 4, "Number of CPU cores")
	cmd.Flags().StringVar(&image, "image", "nvidia/cuda:12.0-base", "Container image")
	cmd.Flags().StringVar(&script, "script", "nvidia-smi", "Script to run")

	return cmd
}

func listJobsCmd() *cobra.Command {
	var state string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs",
		Run: func(cmd *cobra.Command, args []string) {
			url := fmt.Sprintf("%s/api/v1/jobs?tenant_id=%s", apiURL, tenantID)
			if state != "" {
				url += "&state=" + state
			}

			var result struct {
				Jobs  []map[string]interface{} `json:"jobs"`
				Total int                      `json:"total"`
			}

			if err := getJSON(url, &result); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "JOB ID\tNAME\tSTATE\tGPUs\tPRIORITY\tSUBMITTED")

			for _, job := range result.Jobs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%.0f\t%.0f\t%s\n",
					job["id"],
					job["name"],
					job["state"],
					job["gpu_count"],
					job["priority"],
					formatTime(job["submitted_at"]),
				)
			}

			w.Flush()
			fmt.Printf("\nTotal: %d jobs\n", result.Total)
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "Filter by state (pending, running, completed)")

	return cmd
}

func getJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [job-id]",
		Short: "Get job status",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jobID := args[0]
			url := fmt.Sprintf("%s/api/v1/jobs/%s", apiURL, jobID)

			var status map[string]interface{}
			if err := getJSON(url, &status); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Job ID: %s\n", status["job_id"])
			fmt.Printf("State: %s\n", status["state"])
			if status["node_name"] != nil {
				fmt.Printf("Node: %s\n", status["node_name"])
			}
			if status["queue_position"] != nil && status["queue_position"].(float64) > 0 {
				fmt.Printf("Queue Position: %.0f\n", status["queue_position"])
			}
		},
	}
}

func cancelJobCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cancel [job-id]",
		Short: "Cancel a job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jobID := args[0]
			url := fmt.Sprintf("%s/api/v1/jobs/%s", apiURL, jobID)

			req, _ := http.NewRequest("DELETE", url, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				fmt.Println("Job cancelled successfully")
			} else {
				fmt.Fprintf(os.Stderr, "Failed to cancel job: HTTP %d\n", resp.StatusCode)
				os.Exit(1)
			}
		},
	}
}

func clusterStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Get cluster status",
		Run: func(cmd *cobra.Command, args []string) {
			url := fmt.Sprintf("%s/api/v1/cluster/status", apiURL)

			var status map[string]interface{}
			if err := getJSON(url, &status); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Cluster Status:")
			fmt.Printf("  GPUs: %.0f total, %.0f available\n", status["total_gpus"], status["available_gpus"])
			fmt.Printf("  Nodes: %.0f total, %.0f online\n", status["total_nodes"], status["online_nodes"])
			fmt.Printf("  Jobs: %.0f total, %.0f running, %.0f pending\n", 
				status["total_jobs"], status["running_jobs"], status["pending_jobs"])
		},
	}
}

func createTenantCmd() *cobra.Command {
	var (
		name     string
		maxGPUs  int
		priority string
	)

	cmd := &cobra.Command{
		Use:   "create-tenant",
		Short: "Create a new tenant",
		Run: func(cmd *cobra.Command, args []string) {
			tenant := map[string]interface{}{
				"name":           name,
				"max_gpus":       maxGPUs,
				"max_cpu_cores":  64,
				"max_memory_mb":  256000,
				"priority_tier":  priority,
			}

			resp, err := postJSON(apiURL+"/api/v1/tenants", tenant)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Tenant created successfully!\n")
			fmt.Printf("Tenant ID: %s\n", resp["id"])
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Tenant name")
	cmd.Flags().IntVar(&maxGPUs, "max-gpus", 10, "Maximum GPUs")
	cmd.Flags().StringVar(&priority, "priority", "medium", "Priority tier")
	cmd.MarkFlagRequired("name")

	return cmd
}

func getJSON(url string, result interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

func postJSON(url string, data interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func formatTime(t interface{}) string {
	if t == nil {
		return "-"
	}
	if str, ok := t.(string); ok {
		if parsed, err := time.Parse(time.RFC3339, str); err == nil {
			return parsed.Format("2006-01-02 15:04:05")
		}
		return str
	}
	return "-"
}
