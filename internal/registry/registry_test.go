package registry

import (
	"strings"
	"testing"
)

const (
	srcImage = "111111111111.dkr.ecr.eu-west-1.amazonaws.com/my-app:tag-1"
	dstImage = "222222222222.dkr.ecr.eu-central-1.amazonaws.com/team/my-app-new:latest"
)

func TestOpen(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	r, err := Open(t.Context(), dstImage, "")
	if err != nil {
		t.Fatal(err)
	}
	if r.ECR.ID != "222222222222" || r.Cfg.Region != "eu-central-1" || r.Repository() != "team/my-app-new" {
		t.Errorf("got id %q region %q repository %q", r.ECR.ID, r.Cfg.Region, r.Repository())
	}
}

func TestOpenRejects(t *testing.T) {
	for _, image := range []string{
		"docker.io/library/alpine:3",
		"public.ecr.aws/nginx/nginx:latest",
		"111111111111.dkr.ecr.eu-west-1.amazonaws.com/UPPER:tag",
	} {
		if _, err := Open(t.Context(), image, ""); err == nil {
			t.Errorf("%s: expected error", image)
		}
	}
}

func TestNewKeychain(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	src, err := Open(t.Context(), srcImage, "")
	if err != nil {
		t.Fatal(err)
	}
	dst, err := Open(t.Context(), dstImage, "")
	if err != nil {
		t.Fatal(err)
	}
	k, err := NewKeychain(src, dst)
	if err != nil || len(k) != 2 {
		t.Fatalf("got %d registries, err %v", len(k), err)
	}

	other := *src
	other.Profile = "other"
	if _, err := NewKeychain(src, &other); err == nil || !strings.Contains(err.Error(), "one profile per registry") {
		t.Errorf("same host, two profiles: got %v", err)
	}
}
