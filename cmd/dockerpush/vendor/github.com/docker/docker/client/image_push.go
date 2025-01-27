package client

import (
	"errors"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/context"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/hinshun/opentracing-registry/cmd/dockerpush/xmetaheaders"
	opentracing "github.com/opentracing/opentracing-go"
)

// ImagePush requests the docker host to push an image to a remote registry.
// It executes the privileged function if the operation is unauthorized
// and it tries one more time.
// It's up to the caller to handle the io.ReadCloser and close it properly.
func (cli *Client) ImagePush(ctx context.Context, image string, options types.ImagePushOptions) (io.ReadCloser, error) {
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return nil, err
	}

	if _, isCanonical := ref.(reference.Canonical); isCanonical {
		return nil, errors.New("cannot push a digest reference")
	}

	tag := ""
	name := reference.FamiliarName(ref)

	if nameTaggedRef, isNamedTagged := ref.(reference.NamedTagged); isNamedTagged {
		tag = nameTaggedRef.Tag()
	}

	query := url.Values{}
	query.Set("tag", tag)

	resp, err := cli.tryImagePush(ctx, name, query, options.RegistryAuth)
	if resp.statusCode == http.StatusUnauthorized && options.PrivilegeFunc != nil {
		newAuthHeader, privilegeErr := options.PrivilegeFunc()
		if privilegeErr != nil {
			return nil, privilegeErr
		}
		resp, err = cli.tryImagePush(ctx, name, query, newAuthHeader)
	}
	if err != nil {
		return nil, err
	}
	return resp.body, nil
}

func (cli *Client) tryImagePush(ctx context.Context, imageID string, query url.Values, registryAuth string) (serverResponse, error) {
	headers := map[string][]string{"X-Registry-Auth": {registryAuth}}

	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		carrier := xmetaheaders.XMetaHeadersCarrier{
			TextMapWriter: opentracing.HTTPHeadersCarrier(headers),
		}
		opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	}

	return cli.post(ctx, "/images/"+imageID+"/push", query, nil, headers)
}
