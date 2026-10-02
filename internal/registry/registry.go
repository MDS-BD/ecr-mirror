// Package registry resolves ECR registries to the AWS profile that can reach them.
package registry

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/awslabs/amazon-ecr-credential-helper/ecr-login/api"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
)

// Registry is one side of a mirror: an image reference and the AWS config able to reach its registry.
type Registry struct {
	Cfg     aws.Config
	ECR     *api.Registry
	Profile string
	Ref     name.Reference
}

// Open parses image (e.g. 111111111111.dkr.ecr.eu-west-1.amazonaws.com/app:tag) and loads the AWS config of profile,
// pinned to the registry region. An empty profile uses the default credential chain.
func Open(ctx context.Context, image, profile string) (*Registry, error) {
	ref, err := name.ParseReference(image, name.StrictValidation)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", image, err)
	}
	reg, err := api.ExtractRegistry(ref.Context().RegistryStr())
	if err != nil {
		return nil, fmt.Errorf("%q: %w", image, err)
	}
	if reg.Service != api.ServiceECR {
		return nil, fmt.Errorf("%q: only private ECR registries are supported", image)
	}

	opts := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(reg.Region)}
	if profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(profile))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config (profile %q): %w", profile, err)
	}
	return &Registry{Cfg: cfg, ECR: reg, Profile: profile, Ref: ref}, nil
}

// Repository is the ECR repository name, without registry host and tag.
func (r *Registry) Repository() string {
	return r.Ref.Context().RepositoryStr()
}

// CreateRepository creates the repository if it does not exist yet; ECR refuses pushes to unknown repositories.
// It reports whether the repository was created.
func (r *Registry) CreateRepository(ctx context.Context) (bool, error) {
	_, err := ecr.NewFromConfig(r.Cfg).CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RegistryId:     aws.String(r.ECR.ID),
		RepositoryName: aws.String(r.Repository()),
	})
	if err == nil {
		return true, nil
	}
	if _, ok := errors.AsType[*ecrtypes.RepositoryAlreadyExistsException](err); ok {
		return false, nil
	}
	return false, fmt.Errorf("ecr create-repository %s: %w", r.Repository(), err)
}

// Keychain authenticates every registry with the AWS config of the Registry pointing at it.
type Keychain map[string]*Registry

// NewKeychain maps each registry host to its Registry. Two images on the same host must share one profile.
func NewKeychain(regs ...*Registry) (Keychain, error) {
	k := Keychain{}
	for _, r := range regs {
		host := r.Ref.Context().RegistryStr()
		if prev, ok := k[host]; ok && prev.Profile != r.Profile {
			return nil, fmt.Errorf("%s is used with profiles %q and %q: use one profile per registry", host, prev.Profile, r.Profile)
		}
		k[host] = r
	}
	return k, nil
}

// Resolve returns ECR credentials for known registries and anonymous access otherwise.
func (k Keychain) Resolve(target authn.Resource) (authn.Authenticator, error) {
	r, ok := k[target.RegistryStr()]
	if !ok {
		return authn.Anonymous, nil
	}
	auth, err := api.DefaultClientFactory{}.NewClient(r.Cfg).GetCredentials(target.RegistryStr())
	if err != nil {
		return nil, fmt.Errorf("ecr credentials for %s: %w", target.RegistryStr(), err)
	}
	return authn.FromConfig(authn.AuthConfig{Username: auth.Username, Password: auth.Password}), nil
}
