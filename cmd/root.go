package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ohmayank/hako/internal/handlers"
	"github.com/ohmayank/hako/internal/store"
	"github.com/spf13/cobra"
)

func Main(args []string) {
	var addr string

	root := &cobra.Command{
		Use:           "hako",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetArgs(args)
	root.PersistentFlags().StringVar(&addr, "addr", "http://localhost:6060", "server address")

	client := &http.Client{}

	root.AddCommand(&cobra.Command{
		Use: "serve",
		Run: func(*cobra.Command, []string) {
			s := store.NewFileStore("./.data")
			mux := http.NewServeMux()

			mux.HandleFunc("PUT /objects/{bucket}/{key...}", handlers.PutObject(s))
			mux.HandleFunc("GET /objects/{bucket}/{key...}", handlers.GetObject(s))
			mux.HandleFunc("DELETE /objects/{bucket}/{key...}", handlers.DeleteObject(s))

			log.Println("listening on :6060")
			log.Fatal(http.ListenAndServe(":6060", mux))
		},
	})

	root.AddCommand(&cobra.Command{
		Use:  "put <bucket> <key> <file>",
		Args: cobra.ExactArgs(3),
		RunE: func(_ *cobra.Command, args []string) error {
			f, err := os.Open(args[2])
			if err != nil {
				return err
			}

			defer f.Close()

			req, err := http.NewRequest(http.MethodPut, objectURL(addr, args[0], args[1]), f)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				return err
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("put failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
			}
			return nil
		},
	})

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func objectURL(addr, bucket, key string) string {
	base := strings.TrimRight(addr, "/")
	return fmt.Sprintf("%s/objects/%s/%s", base, url.PathEscape(bucket), url.PathEscape(key))
}
