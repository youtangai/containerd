/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package tasks

import (
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cmd/ctr/commands"
	"github.com/containerd/containerd/runtime/linux/runctypes"
	"github.com/containerd/containerd/runtime/v2/runc/options"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var checkpointCommand = cli.Command{
	Name:      "checkpoint",
	Usage:     "checkpoint a container",
	ArgsUsage: "[flags] CONTAINER",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "exit",
			Usage: "stop the container after the checkpoint",
		},
		cli.BoolFlag{
			Name:  "track-mem",
			Usage: "for tracking memory change",
		},
		cli.StringFlag{
			Name:  "parent-checkpoint",
			Value: "",
			Usage: "set parent-checkpoint image's digest",
		},
	},
	Action: func(context *cli.Context) error {
		id := context.Args().First()
		if id == "" {
			return errors.New("container id must be provided")
		}
		client, ctx, cancel, err := commands.NewClient(context, containerd.WithDefaultRuntime(context.String("runtime")))
		if err != nil {
			return err
		}
		defer cancel()
		container, err := client.LoadContainer(ctx, id)
		if err != nil {
			return err
		}
		task, err := container.Task(ctx, nil)
		if err != nil {
			return err
		}
		info, err := container.Info(ctx)
		if err != nil {
			return err
		}
		var opts []containerd.CheckpointTaskOpts
		if context.Bool("exit") {
			opts = append(opts, withExit(info.Runtime.Name))
		}
		if context.Bool("track-mem") {
			opts = append(opts, withTrackMem(info.Runtime.Name))
		}
		if digest := context.String("parent-checkpoint"); digest != "" {
			opts = append(opts, setParentCheckpoint(digest))
		}

		checkpoint, err := task.Checkpoint(ctx, opts...)
		if err != nil {
			return err
		}
		fmt.Println(checkpoint.Name())
		return nil
	},
}

func withExit(rt string) containerd.CheckpointTaskOpts {
	return func(r *containerd.CheckpointTaskInfo) error {
		switch rt {
		case "io.containerd.runc.v1":
			if r.Options == nil {
				r.Options = &options.CheckpointOptions{
					Exit: true,
				}
			} else {
				opts, _ := r.Options.(*options.CheckpointOptions)
				opts.Exit = true
			}
		default:
			if r.Options == nil {
				r.Options = &runctypes.CheckpointOptions{
					Exit: true,
				}
			} else {
				opts, _ := r.Options.(*runctypes.CheckpointOptions)
				opts.Exit = true
			}
		}
		return nil
	}
}

func withTrackMem(rt string) containerd.CheckpointTaskOpts {
	return func(r *containerd.CheckpointTaskInfo) error {
		switch rt {
		case "io.containerd.runc.v1":
			if r.Options == nil {
				r.Options = &options.CheckpointOptions{
					TrackMem: true,
				}
			} else {
				opts, _ := r.Options.(*options.CheckpointOptions)
				opts.TrackMem = true
			}
		default:
			if r.Options == nil {
				r.Options = &runctypes.CheckpointOptions{
					TrackMem: true,
				}
			} else {
				opts, _ := r.Options.(*options.CheckpointOptions)
				opts.TrackMem = true
			}
		}
		return nil
	}
}

func setParentCheckpoint(checkpoint string) containerd.CheckpointTaskOpts {
	digest := digest.Digest(checkpoint)
	return func(r *containerd.CheckpointTaskInfo) error {
		r.ParentCheckpoint = digest
		return nil
	}
}
