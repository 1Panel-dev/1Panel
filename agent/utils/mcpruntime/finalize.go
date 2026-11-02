package mcpruntime

import (
	"context"
	"errors"
	"strconv"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

func (c *Controller) installationCleanup(ctx context.Context, cli DockerClient, before *container.InspectResponse) (*CleanupReference, error) {
	if before == nil {
		return nil, nil
	}
	current, err := c.owned(ctx, cli, ContainerName)
	if err != nil || current == nil {
		return nil, errors.New("installed MCP container is unavailable")
	}
	if current.ID == before.ID {
		return nil, nil
	}
	previous, err := c.owned(ctx, cli, previousName)
	if err != nil || previous == nil || previous.ID != before.ID || previous.Image != before.Image || isRunning(previous) {
		return nil, errors.New("MCP previous container changed during installation")
	}
	return &CleanupReference{ContainerID: current.ID, PreviousContainerID: previous.ID, PreviousImageID: previous.Image}, nil
}

// finalize runs only after Core commits the installed settings. The reference
// binds cleanup to this replacement; fixed names alone cannot authorize it.
func (c *Controller) finalize(ctx context.Context, cli DockerClient, req Request) (*CleanupResult, error) {
	reference := req.Cleanup
	if reference == nil || reference.ContainerID == "" || reference.PreviousContainerID == "" ||
		reference.ContainerID == reference.PreviousContainerID || !digestPattern.MatchString(reference.PreviousImageID) || !digestPattern.MatchString(req.ImageID) {
		return nil, errors.New("invalid MCP installation cleanup reference")
	}
	current, err := c.owned(ctx, cli, ContainerName)
	if err != nil || current == nil || current.ID != reference.ContainerID || current.Image != req.ImageID {
		return nil, errors.New("MCP installed container changed before cleanup")
	}
	verified, err := c.verify(ctx, cli, current.Image)
	if err != nil || !matchesRuntimeWithCapabilities(ctx, cli, current, req.ImageID, req, verified.metadata.AdapterVersion) {
		return nil, errors.New("MCP installed container is unsafe or incompatible for cleanup")
	}
	if req.Enabled {
		if err = c.waitHealthy(ctx, cli, current.ID, req, verified.metadata.AdapterVersion); err != nil {
			return nil, errors.New("MCP installed container is not healthy; previous container was retained")
		}
	} else if current.State == nil || isRunning(current) {
		return nil, errors.New("MCP installed container is not in its expected stopped state")
	}
	previous, err := c.owned(ctx, cli, previousName)
	if err != nil {
		return nil, err
	}
	if previous == nil {
		// A retry after deletion has no inspected image provenance. Never use
		// a caller-provided image ID as a substitute for that missing instance.
		return &CleanupResult{ImageRetainedReason: "previous_missing"}, nil
	}
	if previous.ID != reference.PreviousContainerID || previous.Image != reference.PreviousImageID || previous.State == nil || isRunning(previous) {
		return nil, errors.New("MCP previous container changed before cleanup")
	}
	oldPort, err := strconv.Atoi(previous.Config.Labels[portLabel])
	oldRequest := Request{Image: previous.Config.Labels[imageLabel], ImageID: previous.Image, Port: oldPort}
	if err != nil || oldPort < 1024 || oldPort > 65535 || ValidateImage(oldRequest.Image) != nil {
		return nil, errors.New("MCP previous container configuration is invalid")
	}
	oldImage, err := c.resolve(ctx, cli, previous.Image)
	if err != nil || oldImage.ID != previous.Image {
		return nil, errors.New("MCP previous image is unavailable or changed")
	}
	oldVerified, err := c.verify(ctx, cli, previous.Image)
	if err != nil || !matchesRuntimeWithCapabilities(ctx, cli, previous, previous.Image, oldRequest, oldVerified.metadata.AdapterVersion) {
		return nil, errors.New("MCP previous container was modified; cleanup was rejected")
	}
	if err = cli.ContainerRemove(ctx, previous.ID, container.RemoveOptions{}); err != nil {
		return nil, errors.New("MCP previous container could not be removed")
	}
	result := &CleanupResult{ContainerRemoved: true}
	if !req.DeleteOldImage {
		result.ImageRetainedReason = "not_requested"
		return result, nil
	}
	if previous.Image == current.Image {
		result.ImageRetainedReason = "same_image"
		return result, nil
	}
	if len(oldImage.RepoTags) > 1 {
		result.ImageRetainedReason = "multiple_tags"
		return result, nil
	}
	_, err = cli.ImageRemove(ctx, previous.Image, image.RemoveOptions{})
	if errdefs.IsConflict(err) {
		result.ImageRetainedReason = "in_use"
		return result, nil
	}
	if err != nil && !errdefs.IsNotFound(err) {
		return nil, errors.New("MCP previous image could not be removed")
	}
	result.ImageRemoved = true
	return result, nil
}
