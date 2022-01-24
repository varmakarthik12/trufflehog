package engine

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"

	"github.com/trufflesecurity/trufflehog/pkg/common"
	"github.com/trufflesecurity/trufflehog/pkg/pb/sourcespb"
	"github.com/trufflesecurity/trufflehog/pkg/sources/github"
)

func (e *Engine) ScanGitHub(ctx context.Context, endpoint string, repos, orgs []string, token string, filter *common.Filter, concurrency int) error {
	source := github.Source{}
	connection := sourcespb.GitHub{
		Endpoint:      endpoint,
		Organizations: orgs,
		Repositories:  repos,
	}
	if len(token) > 0 {
		connection.Credential = &sourcespb.GitHub_Token{
			Token: token,
		}
	} else {
		connection.Credential = &sourcespb.GitHub_Unauthenticated{}
	}
	var conn anypb.Any
	err := anypb.MarshalFrom(&conn, &connection, proto.MarshalOptions{})
	if err != nil {
		logrus.WithError(err).Error("failed to marshal github connection")
		return err
	}
	err = source.Init(ctx, "trufflehog - github", 0, 0, false, &conn, concurrency)
	if err != nil {
		logrus.WithError(err).Error("failed to initialize github source")
		return err
	}

	go func() {
		err := source.Chunks(ctx, e.ChunksChan())
		if err != nil {
			logrus.WithError(err).Fatal("could not scan github")
		}
		close(e.ChunksChan())
	}()
	return nil
}