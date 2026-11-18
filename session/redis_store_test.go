package session

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// startServer launches a throwaway server (the first of bins found in PATH)
// on a free port. The test is skipped when none is installed.
func startServer(t *testing.T, bins ...string) *redis.Client {
	t.Helper()
	var bin string
	for _, name := range bins {
		if p, err := exec.LookPath(name); err == nil {
			bin = p
			break
		}
	}
	if bin == "" {
		t.Skipf("none of %v installed", bins)
	}
	t.Logf("using %s", bin)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	cmd := exec.Command(bin, "--port", fmt.Sprint(port), "--bind", "127.0.0.1",
		"--save", "", "--appendonly", "no", "--dir", t.TempDir())
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })

	client := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("127.0.0.1:%d", port)})
	t.Cleanup(func() { _ = client.Close() })
	for i := 0; i < 50; i++ {
		if client.Ping(context.Background()).Err() == nil {
			return client
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("server did not start")
	return nil
}

// kvBackend describes a key-value store implementation to test.
type kvBackend struct {
	name     string
	servers  []string // server binaries, in order of preference
	newStore func() Store
}

var kvBackends = []kvBackend{
	{"redis", []string{"redis-server"}, func() Store { return &RedisStore{} }},
	// Valkey speaks the Redis protocol: fall back to redis-server when
	// valkey-server is not installed (the log says which one ran).
	{"valkey", []string{"valkey-server", "redis-server"}, func() Store { return &ValkeyStore{} }},
}

func forEachBackend(t *testing.T, fn func(t *testing.T, st Store, client *redis.Client)) {
	for _, b := range kvBackends {
		t.Run(b.name, func(t *testing.T) {
			client := startServer(t, b.servers...)
			st := b.newStore()
			if err := st.InitStore(client, "sess:"); err != nil {
				t.Fatal(err)
			}
			fn(t, st, client)
		})
	}
}

func TestKVStoreInitStoreRejectsWrongClient(t *testing.T) {
	for _, b := range kvBackends {
		if err := b.newStore().InitStore("nope", "x"); err == nil {
			t.Errorf("%s: wrong client type accepted", b.name)
		}
		if err := b.newStore().InitStore((*redis.Client)(nil), "x"); err == nil {
			t.Errorf("%s: nil client accepted", b.name)
		}
	}
}

func TestKVStoreBasics(t *testing.T) {
	forEachBackend(t, func(t *testing.T, st Store, client *redis.Client) {
		ctx := context.Background()

		if _, err := st.Init(time.Now().Add(-time.Second)); err == nil {
			t.Error("expiration in the past accepted")
		}

		sid, err := st.Init(time.Now().Add(time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if !st.Check(sid) {
			t.Fatal("new session not found")
		}
		if err := st.Set(sid, "user", "andrea"); err != nil {
			t.Fatal(err)
		}
		if v, err := st.Get(sid, "user"); err != nil || v != "andrea" {
			t.Errorf("Get = %v, %v", v, err)
		}

		// Set must keep the TTL.
		ttl := client.TTL(ctx, "sess:"+sid).Val()
		if ttl <= 0 || ttl > time.Hour {
			t.Errorf("TTL after Set = %v", ttl)
		}

		if err := st.RemoveKey(sid, "user"); err != nil {
			t.Fatal(err)
		}
		if _, err := st.Get(sid, "user"); err == nil || err.Error() != ErrKeyNotFound {
			t.Errorf("Get removed key: err = %v", err)
		}
		if err := st.RemoveKey(sid, "user"); err == nil || err.Error() != ErrKeyNotFound {
			t.Errorf("RemoveKey missing key: err = %v", err)
		}

		// Refresh keeps data and TTL, and drops the old sid.
		_ = st.Set(sid, "k", "v")
		newSid, err := st.Refresh(sid)
		if err != nil {
			t.Fatal(err)
		}
		if st.Check(sid) {
			t.Error("old sid still valid after Refresh")
		}
		if v, _ := st.Get(newSid, "k"); v != "v" {
			t.Errorf("data lost after Refresh: %v", v)
		}
		if ttl := client.TTL(ctx, "sess:"+newSid).Val(); ttl <= 0 {
			t.Errorf("TTL lost after Refresh: %v", ttl)
		}
		if _, err := st.Refresh("missing"); err == nil || err.Error() != ErrSessionNotFound {
			t.Errorf("Refresh missing: err = %v", err)
		}

		if err := st.Delete(newSid); err != nil || st.Check(newSid) {
			t.Errorf("Delete failed: %v", err)
		}
		if err := st.Set(newSid, "a", 1); err == nil || err.Error() != ErrSessionNotFound {
			t.Errorf("Set on deleted session: err = %v", err)
		}
	})
}

func TestKVStoreNoExpiry(t *testing.T) {
	forEachBackend(t, func(t *testing.T, st Store, client *redis.Client) {
		sid, err := st.Init(time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		_ = st.Set(sid, "a", 1)
		if ttl := client.TTL(context.Background(), "sess:"+sid).Val(); ttl != -1 {
			t.Errorf("TTL = %v, want -1 (no expiry)", ttl)
		}
	})
}

// Concurrent Set calls on different keys of the same session must not lose writes.
func TestKVStoreConcurrentSet(t *testing.T) {
	forEachBackend(t, func(t *testing.T, st Store, client *redis.Client) {
		sid, err := st.Init(time.Now().Add(time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		const n = 20
		var wg sync.WaitGroup
		setErr := make([]error, n)
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				setErr[i] = st.Set(sid, fmt.Sprintf("k%d", i), i)
			}(i)
		}
		wg.Wait()

		for i := 0; i < n; i++ {
			if setErr[i] != nil {
				// Under heavy contention the only acceptable failure is a reported conflict.
				if setErr[i].Error() != ErrSessionConflict {
					t.Errorf("Set k%d: %v", i, setErr[i])
				}
				continue
			}
			// Every successful Set must be visible: no lost updates.
			if _, err := st.Get(sid, fmt.Sprintf("k%d", i)); err != nil {
				t.Errorf("key k%d lost after successful Set: %v", i, err)
			}
		}
	})
}
