package cluster

import (
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//func TestPublishRaceCondition(t *testing.T) {
//	actorSystem := actor.NewActorSystem()
//	c := New(actorSystem, Configure("mycluster", nil, nil, remote.Configure("127.0.0.1", 0)))
//	NewMemberList(c)
//	rounds := 1000
//
//	var wg sync.WaitGroup
//	wg.Add(2 * rounds)
//
//	go func() {
//		for i := 0; i < rounds; i++ {
//			actorSystem.EventStream.Publish(TopologyEvent(Members{{}, {}}))
//			actorSystem.EventStream.Publish(TopologyEvent(Members{{}}))
//			wg.Done()
//		}
//	}()
//
//	go func() {
//		for i := 0; i < rounds; i++ {
//			s := actorSystem.EventStream.Subscribe(func(evt interface{}) {})
//			actorSystem.EventStream.Unsubscribe(s)
//			wg.Done()
//		}
//	}()
//
//	if waitTimeout(&wg, 2*time.Second) {
//		t.Error("Should not run into a timeout")
//	}
//}

// https://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func TestMemberList_UpdateClusterToplogy(t *testing.T) {
	t.Skipf("Maintaining")
	c := newClusterForTest("test-UpdateClusterToplogy", nil)
	obj := NewMemberList(c)
	dumpMembers := func(list Members) {
		t.Logf("membersByMemberId=%d", len(list))
		for _, m := range list {
			t.Logf("\t%s", m.Address())
		}
	}
	empty := []*Member{}
	_ = dumpMembers
	_sorted := func(tpl *ClusterTopology) {
		_sortMembers := func(list Members) {
			sort.Slice(list, func(i, j int) bool {
				return (list)[i].Port < (list)[j].Port
			})
		}
		// dumpMembers(tpl.MemberSet)
		_sortMembers(tpl.Members)
		// dumpMembers(tpl.MemberSet)
		_sortMembers(tpl.Left)
		_sortMembers(tpl.Joined)
	}

	t.Run("init", func(t *testing.T) {
		assert := assert.New(t)
		members := _newTopologyEventForTest(2)
		changes, _, _, _, _ := obj.getTopologyChanges(members)
		_sorted(changes)
		expected := &ClusterTopology{TopologyHash: TopologyHash(members), Members: members, Joined: members, Left: empty}
		assert.Equalf(expected, changes, "%s\n%s", expected, changes)
	})

	t.Run("join", func(t *testing.T) {
		assert := assert.New(t)
		members := _newTopologyEventForTest(4)
		changes, _, _, _, _ := obj.getTopologyChanges(members)
		_sorted(changes)
		expected := &ClusterTopology{TopologyHash: TopologyHash(members), Members: members, Joined: members[2:4], Left: empty}
		assert.Equalf(expected, changes, "%s\n%s", expected, changes)
	})

	t.Run("left", func(t *testing.T) {
		assert := assert.New(t)
		members := _newTopologyEventForTest(4)
		changes, _, _, _, _ := obj.getTopologyChanges(members[2:4])
		_sorted(changes)
		expected := &ClusterTopology{TopologyHash: TopologyHash(members), Members: members[2:4], Joined: empty, Left: members[0:2]}
		assert.Equal(expected, changes)
	})
}

func _newTopologyEventForTest(membersCount int, kinds ...string) Members {
	if len(kinds) <= 0 {
		kinds = append(kinds, "kind")
	}
	members := make(Members, membersCount)
	for i := 0; i < membersCount; i++ {
		memberId := fmt.Sprintf("memberId-%d", i)
		members[i] = &Member{
			Id:    memberId,
			Host:  "127.0.0.1",
			Port:  int32(i),
			Kinds: kinds,
		}
	}
	return members
}

func TestMemberList_getPartitionMember(t *testing.T) {
	c := newClusterForTest("test-memberlist", nil)
	obj := NewMemberList(c)

	for _, v := range []int{1, 2, 10, 100, 1000} {
		members := _newTopologyEventForTest(v)
		obj.UpdateClusterTopology(members)

		testName := fmt.Sprintf("member*%d", v)
		t.Run(testName, func(t *testing.T) {
			//assert := assert.New(t)
			//
			//identity := NewClusterIdentity("name", "kind")
			////	address := obj.getPartitionMemberV2(identity)
			////	assert.NotEmpty(address)
			//
			//identity = NewClusterIdentity("name", "nonkind")
			////		address = obj.getPartitionMemberV2(identity)
			////	assert.Empty(address)
		})
	}
}

//func BenchmarkMemberList_getPartitionMemberV2(b *testing.B) {
//	SetLogLevel(log.ErrorLevel)
//	actorSystem := actor.NewActorSystem()
//	c := New(actorSystem, Configure("mycluster", nil, nil, remote.Configure("127.0.0.1", 0)))
//	obj := NewMemberList(c)
//	for i, v := range []int{1, 2, 3, 5, 10, 100, 1000, 2000} {
//		members := _newTopologyEventForTest(v)
//		obj.UpdateClusterTopology(members)
//		testName := fmt.Sprintf("member*%d", v)
//		runtime.GC()
//
//		identity := &ClusterIdentity{Identity: fmt.Sprintf("name-%d", rand.Int()), Kind: "kind"}
//		b.Run(testName, func(b *testing.B) {
//			for i := 0; i < b.N; i++ {
//				address := obj.getPartitionMemberV2(identity)
//				if address == "" {
//					b.Fatalf("empty address membersByMemberId=%d", v)
//				}
//			}
//		})
//	}
//}

//func TestMemberList_getPartitionMemberV2(t *testing.T) {
//	assert := assert.New(t)
//
//	tplg := _newTopologyEventForTest(10)
//	c := _newClusterForTest("test-memberlist")
//	obj := NewMemberList(c)
//	obj.UpdateClusterTopology(tplg, 1)
//
//	assert.Contains(obj.memberStrategyByKind, "kind")
//	addr := obj.getPartitionMemberV2(&ClusterIdentity{Kind: "kind", Identity: "name"})
//	assert.NotEmpty(addr)
//
//	// consistent
//	for i := 0; i < 10; i++ {
//		addr2 := obj.getPartitionMemberV2(&ClusterIdentity{Kind: "kind", Identity: "name"})
//		assert.NotEmpty(addr2)
//		assert.Equal(addr, addr2)
//	}
//}

func TestMemberList_newMemberStrategies(t *testing.T) {
	assert := assert.New(t)

	c := newClusterForTest("test-memberlist", nil)
	obj := NewMemberList(c)
	for _, v := range []int{1, 10, 100, 1000} {
		members := _newTopologyEventForTest(v, "kind1", "kind2")
		obj.UpdateClusterTopology(members)
		assert.Equal(2, len(obj.memberStrategyByKind))
		assert.Contains(obj.memberStrategyByKind, "kind1")

		assert.Equal(v, len(obj.memberStrategyByKind["kind1"].GetAllMembers()))
		assert.Equal(v, len(obj.memberStrategyByKind["kind2"].GetAllMembers()))
	}
}
