package lxd

import (
	"context"
	"fmt"
	"time"

	lxdapi "github.com/canonical/lxd/shared/api"
)

// SnapshotInfo represents a snapshot of an instance
type SnapshotInfo struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Size      int64     `json:"size"`
}

// ListSnapshots returns all snapshots for an instance
func (c *Client) ListSnapshots(ctx context.Context, instanceName string) ([]SnapshotInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	snapshots, err := c.instanceServer.GetInstanceSnapshots(instanceName)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	result := make([]SnapshotInfo, 0, len(snapshots))
	for _, snap := range snapshots {
		result = append(result, SnapshotInfo{
			Name:      snap.Name,
			CreatedAt: snap.CreatedAt,
			Size:      snap.Size,
		})
	}

	return result, nil
}

// CreateSnapshot creates a snapshot of an instance
func (c *Client) CreateSnapshot(ctx context.Context, instanceName, snapshotName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	req := lxdapi.InstanceSnapshotsPost{
		Name: snapshotName,
	}

	op, err := c.instanceServer.CreateInstanceSnapshot(instanceName, req)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	return op.WaitContext(ctx)
}

// DeleteSnapshot deletes a snapshot
func (c *Client) DeleteSnapshot(ctx context.Context, instanceName, snapshotName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	op, err := c.instanceServer.DeleteInstanceSnapshot(instanceName, snapshotName, "")
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return op.WaitContext(ctx)
}

// RestoreSnapshot restores an instance from a snapshot
func (c *Client) RestoreSnapshot(ctx context.Context, instanceName, snapshotName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	// Use UpdateInstance to restore from snapshot
	inst, etag, err := c.instanceServer.GetInstance(instanceName)
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	// Set the restore field
	instPut := inst.Writable()
	instPut.Restore = snapshotName

	op, err := c.instanceServer.UpdateInstance(instanceName, instPut, etag)
	if err != nil {
		return fmt.Errorf("failed to restore snapshot: %w", err)
	}

	return op.WaitContext(ctx)
}

// RenameSnapshot renames a snapshot
func (c *Client) RenameSnapshot(ctx context.Context, instanceName, oldName, newName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to LXD")
	}

	req := lxdapi.InstanceSnapshotPost{
		Name: newName,
	}

	op, err := c.instanceServer.RenameInstanceSnapshot(instanceName, oldName, req)
	if err != nil {
		return fmt.Errorf("failed to rename snapshot: %w", err)
	}

	return op.WaitContext(ctx)
}

// GetSnapshot gets a specific snapshot
func (c *Client) GetSnapshot(ctx context.Context, instanceName, snapshotName string) (*SnapshotInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to LXD")
	}

	snap, _, err := c.instanceServer.GetInstanceSnapshot(instanceName, snapshotName)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return &SnapshotInfo{
		Name:      snap.Name,
		CreatedAt: snap.CreatedAt,
		Size:      snap.Size,
	}, nil
}
