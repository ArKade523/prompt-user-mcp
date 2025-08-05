package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"prompt-mcp/server"
)

var (
	port    int
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "prompt-mcp",
	Short: "MCP server for user input and feedback",
	Long:  `An MCP server that enables LLM agents to request user feedback and approval without stopping their workflow.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("prompt-mcp server - use --help for available commands")
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long:  `Start the MCP server to handle user input requests from LLM agents.`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Fprintf(os.Stderr, "Starting MCP server...\n")
		}
		
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		// Handle shutdown signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			if verbose {
				fmt.Fprintf(os.Stderr, "Shutting down server...\n")
			}
			cancel()
		}()
		
		srv := server.NewMCPServer()
		if err := srv.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on (future use)")
	serveCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}