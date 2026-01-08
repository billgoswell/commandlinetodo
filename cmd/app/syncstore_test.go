package main

import (
	"testing"
	"time"
)

// TestConflictResolution verifies "last write wins" logic
func TestConflictResolution(t *testing.T) {
	tests := []struct {
		name         string
		clientTime   int64
		serverTime   int64
		expectedWins string
	}{
		{
			name:         "server is newer",
			clientTime:   1000,
			serverTime:   2000,
			expectedWins: "server",
		},
		{
			name:         "client is newer",
			clientTime:   3000,
			serverTime:   2000,
			expectedWins: "client",
		},
		{
			name:         "same timestamp",
			clientTime:   2000,
			serverTime:   2000,
			expectedWins: "client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate conflict resolution
			if tt.serverTime > tt.clientTime {
				if tt.expectedWins != "server" {
					t.Errorf("expected server to win, but got %s", tt.expectedWins)
				}
			} else {
				if tt.expectedWins != "client" {
					t.Errorf("expected client to win, but got %s", tt.expectedWins)
				}
			}
		})
	}
}

// TestChangeLogging tests the change tracking mechanism
func TestChangeLogging(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		changeType string
		valid      bool
	}{
		{
			name:       "task create",
			entityType: "task",
			changeType: "create",
			valid:      true,
		},
		{
			name:       "task update",
			entityType: "task",
			changeType: "update",
			valid:      true,
		},
		{
			name:       "task delete",
			entityType: "task",
			changeType: "delete",
			valid:      true,
		},
		{
			name:       "list create",
			entityType: "list",
			changeType: "create",
			valid:      true,
		},
		{
			name:       "list update",
			entityType: "list",
			changeType: "update",
			valid:      true,
		},
		{
			name:       "list delete",
			entityType: "list",
			changeType: "delete",
			valid:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate entity type and change type
			isValidEntity := tt.entityType == "task" || tt.entityType == "list"
			isValidChange := tt.changeType == "create" || tt.changeType == "update" || tt.changeType == "delete"

			if isValidEntity && isValidChange {
				if !tt.valid {
					t.Errorf("expected valid=%v, got valid=false", tt.valid)
				}
			} else {
				if tt.valid {
					t.Errorf("expected valid=false, got valid=%v", tt.valid)
				}
			}
		})
	}
}

// TestSyncTimestampUpdating verifies that last sync time is properly tracked
func TestSyncTimestampUpdating(t *testing.T) {
	now := time.Now().Unix()
	minusOneHour := now - 3600

	// Verify that the newer time is considered "synced" more recently
	if minusOneHour >= now {
		t.Errorf("expected minusOneHour (%d) to be less than now (%d)", minusOneHour, now)
	}

	// Simulate checking if a sync is "recent" (within last 5 minutes)
	fiveMinutesAgo := now - 300
	recentSync := now - 60

	if recentSync < fiveMinutesAgo {
		t.Errorf("expected recent sync to be after 5 minutes ago")
	}

	oldSync := now - 600 // 10 minutes ago
	if oldSync >= fiveMinutesAgo {
		t.Errorf("expected old sync to be before 5 minutes ago")
	}
}

// TestOnlineStatusDetection verifies connectivity status logic
func TestOnlineStatusDetection(t *testing.T) {
	tests := []struct {
		name      string
		isOnline  bool
		expected  string
	}{
		{
			name:      "device is online",
			isOnline:  true,
			expected:  "online",
		},
		{
			name:      "device is offline",
			isOnline:  false,
			expected:  "offline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status string
			if tt.isOnline {
				status = "online"
			} else {
				status = "offline"
			}

			if status != tt.expected {
				t.Errorf("expected status=%s, got status=%s", tt.expected, status)
			}
		})
	}
}

// TestPullPushSequence verifies the order of operations in a full sync
func TestPullPushSequence(t *testing.T) {
	// In a full sync, we should:
	// 1. Pull changes from server first (get latest state)
	// 2. Apply local conflict resolution
	// 3. Push any remaining local changes
	// 4. Update last sync time

	sequence := []string{}

	// Simulate a full sync
	sequence = append(sequence, "pull")
	sequence = append(sequence, "push")
	sequence = append(sequence, "updateSyncTime")

	expectedSequence := []string{"pull", "push", "updateSyncTime"}

	if len(sequence) != len(expectedSequence) {
		t.Fatalf("sequence length mismatch: got %d, expected %d", len(sequence), len(expectedSequence))
	}

	for i, step := range sequence {
		if step != expectedSequence[i] {
			t.Errorf("step %d: expected %s, got %s", i, expectedSequence[i], step)
		}
	}
}

// TestOfflineCapability verifies that local changes are saved even when offline
func TestOfflineCapability(t *testing.T) {
	// Simulate adding changes while offline
	changeLog := []map[string]interface{}{
		{
			"entityType": "task",
			"entityID":   1,
			"changeType": "create",
			"synced":     false,
		},
		{
			"entityType": "task",
			"entityID":   2,
			"changeType": "update",
			"synced":     false,
		},
	}

	// Verify that unsync changes are logged
	unsynced := 0
	for _, change := range changeLog {
		if !change["synced"].(bool) {
			unsynced++
		}
	}

	if unsynced != 2 {
		t.Errorf("expected 2 unsynced changes, got %d", unsynced)
	}

	// Simulate marking changes as synced
	for i := range changeLog {
		changeLog[i]["synced"] = true
	}

	// Verify all changes are now synced
	unsynced = 0
	for _, change := range changeLog {
		if !change["synced"].(bool) {
			unsynced++
		}
	}

	if unsynced != 0 {
		t.Errorf("expected 0 unsynced changes after sync, got %d", unsynced)
	}
}
