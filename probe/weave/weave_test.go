package weave_test

import (
	"testing"
	"time"

	"github.com/weaveworks/scope/probe/docker"
	probe "github.com/weaveworks/scope/probe/weave"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/weave"
)

const (
	mockHostID               = "host1"
	mockContainerIPWithScope = ";" + weave.MockContainerIP
)

func TestWeaveTaggerWeaveTopology(t *testing.T) {
	w := probe.NewWeave(mockHostID, weave.MockClient{})
	defer w.Stop()

	// Wait until the reporter reports some nodes
	test.Poll(t, 300*time.Millisecond, 1, func() interface{} {
		have, _ := w.Report()
		return len(have.Weave.Nodes)
	})

	{
		// Weave node should include peer name and nickname
		have, err := w.Report()
		if err != nil {
			t.Fatal(err)
		}

		nodeID := report.MakeWeaveNodeID(weave.MockWeavePeerName)
		node, ok := have.Weave.Nodes[nodeID]
		if !ok {
			t.Errorf("Expected weave node %q, but not found", nodeID)
		}
		if peerName, ok := node.Latest.Lookup(probe.WeavePeerName); !ok || peerName != weave.MockWeavePeerName {
			t.Errorf("Expected weave peer name %q, got %q", weave.MockWeavePeerName, peerName)
		}
		if peerNick, ok := node.Latest.Lookup(probe.WeavePeerNickName); !ok || peerNick != weave.MockWeavePeerNickName {
			t.Errorf("Expected weave peer nickname %q, got %q", weave.MockWeavePeerNickName, peerNick)
		}
	}

	{
		// Container nodes should be tagged with their weave info
		nodeID := report.MakeContainerNodeID(weave.MockContainerID)
		have, err := w.Tag(report.Report{
			Container: report.MakeTopology().AddNode(nodeID, report.MakeNodeWith(map[string]string{
				docker.ContainerID: weave.MockContainerID,
			})),
		})
		if err != nil {
			t.Fatal(err)
		}

		node, ok := have.Container.Nodes[nodeID]
		if !ok {
			t.Errorf("Expected container node %q, but not found", nodeID)
		}

		// Should have Weave DNS Hostname
		if have, ok := node.Latest.Lookup(probe.WeaveDNSHostname); !ok || have != weave.MockHostname {
			t.Errorf("Expected weave dns hostname %q, got %q", weave.MockHostname, have)
		}
		// Should have Weave MAC Address
		if have, ok := node.Latest.Lookup(probe.WeaveMACAddress); !ok || have != weave.MockContainerMAC {
			t.Errorf("Expected weave mac address %q, got %q", weave.MockContainerMAC, have)
		}
		// Should have Weave container ip
		if have, ok := node.Sets.Lookup(docker.ContainerIPs); !ok || !have.Contains(weave.MockContainerIP) {
			t.Errorf("Expected container ips to include the weave IP %q, got %q", weave.MockContainerIP, have)
		}
		// Should have Weave container ip (with scope)
		if have, ok := node.Sets.Lookup(docker.ContainerIPsWithScopes); !ok || !have.Contains(mockContainerIPWithScope) {
			t.Errorf("Expected container ips to include the weave IP (with scope) %q, got %q", mockContainerIPWithScope, have)
		}
	}
}
