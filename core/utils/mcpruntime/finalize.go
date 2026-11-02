package mcpruntime

import (
	"context"
	"errors"
	"regexp"
)

var dockerContainerIDPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func validCleanupReference(reference *cleanupReference) bool {
	return reference != nil && dockerContainerIDPattern.MatchString(reference.ContainerID) &&
		dockerContainerIDPattern.MatchString(reference.PreviousContainerID) &&
		reference.ContainerID != reference.PreviousContainerID && ValidDockerImageID(reference.PreviousImageID)
}

// Finalize is called only after the new runtime settings have been committed.
// It uses the exact containers returned by this installation, never a tag or a scan.
func (m *DockerManager) Finalize(ctx context.Context, deleteOldImage bool) error {
	m.installMu.Lock()
	if m.installCleanup == nil {
		m.installMu.Unlock()
		return nil
	}
	reference, imageID, enabled := *m.installCleanup, m.installImageID, m.installEnabled
	m.installMu.Unlock()
	if !validCleanupReference(&reference) || !ValidDockerImageID(imageID) {
		return errors.New("MCP installation cleanup reference is unavailable")
	}
	ReportProgress(ctx, Progress{Status: "cleanup"})
	response, err := m.send(ctx, dockerRequest{
		Operation: "finalize", Image: m.image, Port: m.port, Enabled: enabled, ImageID: imageID,
		Cleanup: &reference, DeleteOldImage: deleteOldImage,
	})
	if err != nil {
		return err
	}
	if !validCleanupResult(response.CleanupResult, deleteOldImage, reference.PreviousImageID == imageID) {
		return errors.New("MCP Docker agent returned an invalid cleanup result")
	}
	result := response.CleanupResult
	if result.ContainerRemoved {
		ReportProgress(ctx, Progress{Status: "cleanup_complete"})
	}
	switch {
	case result.ImageRemoved:
		ReportProgress(ctx, Progress{Status: "old_image_removed"})
	case result.ImageRetainedReason == "not_requested":
		ReportProgress(ctx, Progress{Status: "old_image_kept"})
	case result.ImageRetainedReason == "previous_missing":
		ReportProgress(ctx, Progress{Status: "cleanup_skipped"})
	default:
		ReportProgress(ctx, Progress{Status: "old_image_retained"})
	}
	m.installMu.Lock()
	if m.installCleanup != nil && *m.installCleanup == reference {
		m.installCleanup = nil
	}
	m.installMu.Unlock()
	return nil
}

func validCleanupResult(result *cleanupResult, deleteOldImage, sameImage bool) bool {
	if result == nil {
		return false
	}
	if result.ImageRetainedReason == "previous_missing" {
		return !result.ContainerRemoved && !result.ImageRemoved
	}
	if !result.ContainerRemoved {
		return false
	}
	if result.ImageRemoved {
		return deleteOldImage && !sameImage && result.ImageRetainedReason == ""
	}
	switch result.ImageRetainedReason {
	case "not_requested":
		return !deleteOldImage
	case "same_image":
		return deleteOldImage && sameImage
	case "in_use", "multiple_tags":
		return deleteOldImage && !sameImage
	default:
		return false
	}
}
