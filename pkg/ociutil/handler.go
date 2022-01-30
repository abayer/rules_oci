package ociutil

import (
	"context"

	"github.com/DataDog/rules_oci/internal/set"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// CopyContentHandler copies the parent descriptor from the provider to the
// ingestor
func CopyContentHandler(handler images.HandlerFunc, from content.Provider, to content.Ingester) images.HandlerFunc {
	return func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		err := CopyContent(ctx, from, to, desc)
		if err != nil {
			return nil, err
		}

		return handler(ctx, desc)
	}
}

// CopyContentHandler copies the parent descriptor from the provider to the
// ingestor
func CopyContentFromHandler(ctx context.Context, handler images.HandlerFunc, from content.Provider, to content.Ingester, desc ocispec.Descriptor) error {
	children, err := handler.Handle(ctx, desc)
	if err != nil {
		return err
	}

	for _, child := range children {
		err = CopyContentFromHandler(ctx, handler, from, to, child)
		if err != nil {
			return err
		}
	}

	err = CopyContent(ctx, from, to, desc)
	if err != nil {
		return err
	}

	return nil
}

// ContentTypesFilterHandler filters the children of the handler to only include
// the listed content types
func ContentTypesFilterHandler(handler images.HandlerFunc, contentTypes ...string) images.HandlerFunc {
	set := make(set.String)
	set.Add(contentTypes...)
	return func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		children, err := handler(ctx, desc)
		if err != nil {
			return nil, err
		}

		var rtChildren []ocispec.Descriptor
		for _, c := range children {
			if set.Contains(c.MediaType) {
				rtChildren = append(rtChildren, c)
			}
		}

		return rtChildren, nil
	}
}
