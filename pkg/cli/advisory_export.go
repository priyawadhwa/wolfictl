package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/wolfi-dev/wolfictl/pkg/advisory"
	"github.com/wolfi-dev/wolfictl/pkg/configs"
	advisoryconfigs "github.com/wolfi-dev/wolfictl/pkg/configs/advisory"
	rwos "github.com/wolfi-dev/wolfictl/pkg/configs/rwfs/os"
	"github.com/wolfi-dev/wolfictl/pkg/distro"
)

func AdvisoryExport() *cobra.Command {
	p := &exportParams{}
	cmd := &cobra.Command{
		Use:           "export",
		Short:         "Export advisory data (experimental)",
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		Hidden:        true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(p.advisoriesRepoDirs) == 0 {
				if p.doNotDetectDistro {
					return fmt.Errorf("no advisories repo dir specified")
				}

				d, err := distro.Detect()
				if err != nil {
					return fmt.Errorf("no advisories repo dir specified, and distro auto-detection failed: %w", err)
				}

				p.advisoriesRepoDirs = append(p.advisoriesRepoDirs, d.AdvisoriesRepoDir)
				_, _ = fmt.Fprint(os.Stderr, renderDetectedDistro(d))
			}

			indices := make([]*configs.Index[advisoryconfigs.Document], 0, len(p.advisoriesRepoDirs))
			for _, dir := range p.advisoriesRepoDirs {
				advisoryFsys := rwos.DirFS(dir)
				index, err := advisoryconfigs.NewIndex(advisoryFsys)
				if err != nil {
					return fmt.Errorf("unable to index advisory configs for directory %q: %w", dir, err)
				}

				indices = append(indices, index)
			}

			opts := advisory.ExportOptions{
				AdvisoryCfgIndices: indices,
			}

			export, err := advisory.Export(opts)
			if err != nil {
				return fmt.Errorf("unable to export advisory data: %w", err)
			}

			var outputFile *os.File
			if p.outputLocation == "" {
				outputFile = os.Stdout
			} else {
				outputFile, err = os.Create(p.outputLocation)
				if err != nil {
					return fmt.Errorf("unable to create output file: %w", err)
				}
				defer outputFile.Close()
			}

			_, err = io.Copy(outputFile, export)
			if err != nil {
				return fmt.Errorf("unable to export data to specified location: %w", err)
			}

			return nil
		},
	}

	p.addFlagsTo(cmd)
	return cmd
}

type exportParams struct {
	doNotDetectDistro bool

	advisoriesRepoDirs []string

	outputLocation string
}

func (p *exportParams) addFlagsTo(cmd *cobra.Command) {
	addNoDistroDetectionFlag(&p.doNotDetectDistro, cmd)

	cmd.Flags().StringSliceVarP(&p.advisoriesRepoDirs, "advisories-repo-dir", "a", nil, "directory containing an advisories repository")

	cmd.Flags().StringVarP(&p.outputLocation, "output", "o", "", "output location (default: stdout)")
}
