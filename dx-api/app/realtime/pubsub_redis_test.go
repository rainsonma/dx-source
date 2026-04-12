package realtime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newRedisForPubSub(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, mr
}

func TestRedisPubSub_PublishReceive(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ps := NewRedisPubSub(ctx, client)
	t.Cleanup(func() { _ = ps.Close() })

	ch, unsub := ps.Subscribe("test:topic")
	defer unsub()

	time.Sleep(50 * time.Millisecond)

	if err := ps.Publish(ctx, "test:topic", Event{Type: "hello", Data: "world"}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case ev := <-ch:
		if ev.Type != "hello" {
			t.Errorf("wrong type: %s", ev.Type)
		}
		if s, ok := ev.Data.(string); !ok || s != "world" {
			t.Errorf("wrong data: %#v", ev.Data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestRedisPubSub_MultipleSubscribersOnSameTopic(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ps := NewRedisPubSub(ctx, client)
	t.Cleanup(func() { _ = ps.Close() })

	ch1, unsub1 := ps.Subscribe("test:topic")
	defer unsub1()
	ch2, unsub2 := ps.Subscribe("test:topic")
	defer unsub2()

	time.Sleep(50 * time.Millisecond)

	_ = ps.Publish(ctx, "test:topic", Event{Type: "ping", Data: nil})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		select {
		case ev := <-ch1:
			if ev.Type != "ping" {
				t.Errorf("ch1 wrong type: %s", ev.Type)
			}
		case <-time.After(2 * time.Second):
			t.Error("ch1 timeout")
		}
	}()
	go func() {
		defer wg.Done()
		select {
		case ev := <-ch2:
			if ev.Type != "ping" {
				t.Errorf("ch2 wrong type: %s", ev.Type)
			}
		case <-time.After(2 * time.Second):
			t.Error("ch2 timeout")
		}
	}()
	wg.Wait()
}

func TestRedisPubSub_UnsubscribeStopsDelivery(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ps := NewRedisPubSub(ctx, client)
	t.Cleanup(func() { _ = ps.Close() })

	ch, unsub := ps.Subscribe("test:topic")

	time.Sleep(50 * time.Millisecond)

	unsub()

	time.Sleep(50 * time.Millisecond)

	_ = ps.Publish(ctx, "test:topic", Event{Type: "shouldnt-arrive"})

	select {
	case ev, ok := <-ch:
		if ok {
			t.Errorf("unexpected event after unsubscribe: %+v", ev)
		}
	case <-time.After(300 * time.Millisecond):
	}
}

func TestRedisPubSub_PublishContextCanceled(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ps := NewRedisPubSub(context.Background(), client)
	t.Cleanup(func() { _ = ps.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ps.Publish(ctx, "test:topic", Event{Type: "nope"})
	if err == nil {
		t.Error("expected error from canceled context")
	}
}

func TestRedisPubSub_CloseTerminatesLoop(t *testing.T) {
	client, _ := newRedisForPubSub(t)
	ctx := context.Background()
	ps := NewRedisPubSub(ctx, client)

	if err := ps.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
	_ = ps.Close()
}

func TestRedisPubSub_ConcurrentUnsubscribeAndDispatch(t *testing.T) {
	// This test exercises the race between loop() dispatching and
	// unsubscribe() closing the channel. Before the fix (close outside
	// lock), it would reliably panic with "send on closed channel"
	// under -race. After the fix, it must run cleanly.
	client, _ := newRedisForPubSub(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ps := NewRedisPubSub(ctx, client)
	t.Cleanup(func() { _ = ps.Close() })

	const topic = "test:race"

	// Initial subscribe to force the Redis SUBSCRIBE registration.
	_, initialUnsub := ps.Subscribe(topic)
	defer initialUnsub()
	time.Sleep(50 * time.Millisecond)

	var wg sync.WaitGroup

	// Goroutine A: publisher
	wg.Go(func() {
		for i := range 200 {
			_ = ps.Publish(ctx, topic, Event{Type: "burst", Data: i})
			time.Sleep(100 * time.Microsecond)
		}
	})

	// Goroutine B: subscribe/unsubscribe churn
	wg.Go(func() {
		for range 200 {
			_, unsub := ps.Subscribe(topic)
			// Immediately unsubscribe — creates close-vs-send windows
			unsub()
		}
	})

	wg.Wait()
	// If we reach here without a panic under -race, the fix works.
}
