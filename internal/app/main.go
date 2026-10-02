package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"

	"github.com/MDS-BD/ecr-mirror/internal/registry"
)

const progressInterval = 2 * time.Second

type CLI struct {
	CreateRepo bool             `default:"true" help:"Create the target repository when missing." negatable:""`
	Jobs       int              `default:"4" env:"ECR_MIRROR_JOBS" help:"Concurrent layer uploads."`
	NoClobber  bool             `help:"Fail when the target tag already exists." name:"no-clobber"`
	Platform   string           `help:"Copy only this platform (e.g. linux/amd64) instead of the whole multi-arch index."`
	Source     string           `arg:"" help:"Source image, e.g. 111111111111.dkr.ecr.eu-west-1.amazonaws.com/app:tag."`
	SrcProfile string           `env:"ECR_MIRROR_SRC_PROFILE" help:"AWS profile able to pull the source (default: AWS credential chain)." name:"src-profile"`
	Target     string           `arg:"" help:"Target image, e.g. 222222222222.dkr.ecr.eu-west-1.amazonaws.com/app-new:latest."`
	DstProfile string           `env:"ECR_MIRROR_DST_PROFILE" help:"AWS profile able to push the target (default: AWS credential chain)." name:"dst-profile"`
	Verbose    bool             `help:"Verbose logging." short:"v"`
	Version    kong.VersionFlag `help:"Print the version and exit."`
}

func Main(version string) {
	var cli CLI
	kong.Parse(&cli,
		kong.Name("ecr-mirror"),
		kong.Description("Copy an image (multi-arch included) between ECR registries reachable through different AWS profiles."),
		kong.Vars{"version": version},
	)

	level := zerolog.InfoLevel
	if cli.Verbose {
		level = zerolog.DebugLevel
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		// The ECR credential helper logs through logrus; its errors are returned to us anyway.
		logrus.SetOutput(io.Discard)
	}
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.TimeOnly}).
		Level(level).With().Timestamp().Logger()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx = log.WithContext(ctx)

	if err := cli.Run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(130)
		}
		log.Error().Err(err).Msg("ecr-mirror failed")
		os.Exit(1)
	}
}

func (c *CLI) Run(ctx context.Context) error {
	log := zerolog.Ctx(ctx)

	src, err := registry.Open(ctx, c.Source, c.SrcProfile)
	if err != nil {
		return err
	}
	dst, err := registry.Open(ctx, c.Target, c.DstProfile)
	if err != nil {
		return err
	}
	keychain, err := registry.NewKeychain(src, dst)
	if err != nil {
		return err
	}

	opts := []remote.Option{remote.WithContext(ctx), remote.WithAuthFromKeychain(keychain), remote.WithJobs(c.Jobs)}
	var platform *v1.Platform
	if c.Platform != "" {
		if platform, err = v1.ParsePlatform(c.Platform); err != nil {
			return fmt.Errorf("parse platform %q: %w", c.Platform, err)
		}
		opts = append(opts, remote.WithPlatform(*platform))
	}

	puller, err := remote.NewPuller(opts...)
	if err != nil {
		return err
	}
	desc, err := puller.Get(ctx, src.Ref)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", src.Ref, err)
	}
	log.Info().Str("digest", desc.Digest.String()).Str("mediaType", string(desc.MediaType)).Msg("source resolved")

	if c.NoClobber {
		if head, err := puller.Head(ctx, dst.Ref); err == nil {
			return fmt.Errorf("%s already exists (%s), refusing to overwrite", dst.Ref, head.Digest)
		}
	}

	if c.CreateRepo {
		created, err := dst.CreateRepository(ctx)
		if err != nil {
			return err
		}
		if created {
			log.Info().Str("repository", dst.Repository()).Msg("target repository created")
		}
	}

	var target remote.Taggable = desc
	if platform != nil {
		// With an explicit platform push only that image, not the whole index.
		if target, err = desc.Image(); err != nil {
			return fmt.Errorf("select platform %s: %w", c.Platform, err)
		}
	}

	// remote.Push closes updates when done; the Pusher.Push method does not, which would block logProgress forever.
	updates := make(chan v1.Update, 16)
	done := make(chan struct{})
	go func() {
		defer close(done)
		logProgress(ctx, updates)
	}()
	err = remote.Push(dst.Ref, target, append(opts, remote.WithProgress(updates))...)
	<-done
	if err != nil {
		return fmt.Errorf("push %s: %w", dst.Ref, err)
	}

	log.Info().Str("from", src.Ref.String()).Str("to", dst.Ref.String()).Msg("image mirrored")
	return nil
}

// logProgress reports upload progress at most every progressInterval until the pusher closes updates.
func logProgress(ctx context.Context, updates <-chan v1.Update) {
	log := zerolog.Ctx(ctx)
	var last time.Time
	for u := range updates {
		if u.Error != nil || time.Since(last) < progressInterval {
			continue
		}
		last = time.Now()
		pct := 0.0
		if u.Total > 0 {
			pct = float64(u.Complete) * 100 / float64(u.Total)
		}
		log.Info().Str("progress", fmt.Sprintf("%.0f%%", pct)).Int64("bytes", u.Complete).Int64("total", u.Total).Msg("uploading")
	}
}
