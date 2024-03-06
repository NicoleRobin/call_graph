package call_graph

import (
	"context"
	"fmt"
	"github.com/nicolerobin/call_graph/parser"
	"go.uber.org/zap"
	"os"

	"github.com/nicolerobin/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "call_graph",
	Short: "call_graph is a tool to generate call graph for a go project",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		if !isExist(ctx, dir) {
			fmt.Printf("dir:%s is not exist, please check and try again\n", dir)
			return
		}

		err := parser.ParseDir(ctx, dir)
		if err != nil {
			log.Error("parser.ParseDir() failed", zap.Error(err))
			return
		}
	},
}

func isExist(ctx context.Context, dir string) bool {
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

var dir string

func init() {
	rootCmd.PersistentFlags().StringVarP(&dir, "dir", "d", "", "go project directory")
	rootCmd.MarkFlagRequired("dir")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error("rootCmd.Execute() failed, err:%s", err)
		os.Exit(1)
	}
}
