package topotests

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"

	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/topo/zk2topo"

	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
)

// This file contains tests for the tablet.go file.  Note we use a
// zk2topo, because memorytopo doesn't support all topo server
// methods yet.

// TestCreateTablet tests all the logic in the topo.CreateTablet method.
func TestCreateTablet(t *testing.T) {
	cell := "cell1"
	keyspace := "ks1"
	shard := "shard1"
	ctx := context.Background()
	ts := zk2topo.NewFakeServer(cell)

	// Create a tablet.
	alias := &topodatapb.TabletAlias{
		Cell: cell,
		Uid:  1,
	}
	tablet := &topodatapb.Tablet{
		Keyspace: keyspace,
		Shard:    shard,
		Alias:    alias,
	}
	if err := ts.CreateTablet(ctx, tablet); err != nil {
		t.Fatalf("CreateTablet failed: %v", err)
	}

	// Get the tablet, make sure it's good. Also check ShardReplication.
	ti, err := ts.GetTablet(ctx, alias)
	if err != nil || !proto.Equal(ti.Tablet, tablet) {
		t.Fatalf("Created Tablet doesn't match: %v %v", ti, err)
	}
	sri, err := ts.GetShardReplication(ctx, cell, keyspace, shard)
	if err != nil || len(sri.Nodes) != 1 || !proto.Equal(sri.Nodes[0].TabletAlias, alias) {
		t.Fatalf("Created ShardReplication doesn't match: %v %v", sri, err)
	}

	// Create the same tablet again, make sure it fails with ErrNodeExists.
	if err := ts.CreateTablet(ctx, tablet); err != topo.ErrNodeExists {
		t.Fatalf("CreateTablet(again) returned: %v", err)
	}

	// Remove the ShardReplication record, try to create the
	// tablets again, make sure it's fixed.
	if err := topo.RemoveShardReplicationRecord(ctx, ts, cell, keyspace, shard, alias); err != nil {
		t.Fatalf("RemoveShardReplicationRecord failed: %v", err)
	}
	sri, err = ts.GetShardReplication(ctx, cell, keyspace, shard)
	if err != nil || len(sri.Nodes) != 0 {
		t.Fatalf("Modifed ShardReplication doesn't match: %v %v", sri, err)
	}
	if err := ts.CreateTablet(ctx, tablet); err != topo.ErrNodeExists {
		t.Fatalf("CreateTablet(again and again) returned: %v", err)
	}
	sri, err = ts.GetShardReplication(ctx, cell, keyspace, shard)
	if err != nil || len(sri.Nodes) != 1 || !proto.Equal(sri.Nodes[0].TabletAlias, alias) {
		t.Fatalf("Created ShardReplication doesn't match: %v %v", sri, err)
	}
}
