package tuf

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/sigstore/scaffolding/pkg/certs"
	"knative.dev/pkg/logging"
)

func ProcessTufFiles(ctx context.Context, tufFiles []fs.DirEntry, dir string) map[string][]byte {
	trimDir := strings.TrimSuffix(dir, "/")
	files := map[string][]byte{}
	for _, file := range tufFiles {
		if !file.IsDir() {
			logging.FromContext(ctx).Infof("Got file %s", file.Name())
			// Kubernetes adds some extra files here that are prefixed with
			// .., for example '..data' so skip those.
			if strings.HasPrefix(file.Name(), "..") {
				logging.FromContext(ctx).Infof("Skipping .. file %s", file.Name())
				continue
			}
			fileName := fmt.Sprintf("%s/%s", trimDir, file.Name())
			fileBytes, err := os.ReadFile(fileName)
			if err != nil {
				logging.FromContext(ctx).Fatalf("failed to read file %s/%s: %v", fileName, err)
			}
			// If it's a TSA file, we need to split it into multiple TUF
			// targets.
			if strings.Contains(file.Name(), "tsa") {
				logging.FromContext(ctx).Infof("Splitting TSA certchain into individual certs")

				certFiles, err := certs.SplitCertChain(fileBytes, "tsa")
				if err != nil {
					logging.FromContext(ctx).Fatalf("failed to parse  %s/%s: %v", fileName, err)
				}
				for k, v := range certFiles {
					logging.FromContext(ctx).Infof("Got tsa cert file %s", k)
					files[k] = v
				}
			} else {
				files[file.Name()] = fileBytes
			}
		}
	}
	return files
}
